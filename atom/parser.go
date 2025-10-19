package atom

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"strings"

	xpp "github.com/mmcdole/goxpp"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

// Atom elements which contain URIs
// https://tools.ietf.org/html/rfc4287
var atomUriElements = map[string]struct{}{
	"icon": {},
	"id":   {},
	"logo": {},
	"uri":  {},
	"url":  {}, // atom 0.3
}

// List of xml attributes that contain URIs to be resolved relative to xml:base.
// From the Atom spec https://tools.ietf.org/html/rfc4287
var atomUriAttrs = map[string]struct{}{
	"href":   {},
	"scheme": {},
	"src":    {},
	"uri":    {},
}

// Parser is an Atom Parser
type Parser struct {
	p *xpp.XMLPullParser
}

// NewParser creates a new Atom parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an xml feed into an atom.Feed
func (ap *Parser) Parse(feed io.Reader, opts ...options.Option) (*Feed, error) {
	ap.p = xpp.NewXMLPullParser(feed, false, shared.NewReaderLabel)
	_, err := shared.FindRoot(ap.p)
	if err != nil {
		return nil, err
	}
	return ap.parseRoot()
}

func (ap *Parser) parseRoot() (*Feed, error) {
	if err := ap.expect(xpp.StartTag, "feed"); err != nil {
		return nil, err
	}

	atom := &Feed{}
	atom.Entries = []*Entry{}
	atom.Version = ap.parseVersion()
	atom.Language = ap.parseLanguage()

	contributors := []*Person{}
	authors := []*Person{}
	categories := []*Category{}
	links := []*Link{}
	extensions := ext.Extensions{}

	for {
		tok, err := ap.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(ap.p.Name)

			switch {
			case shared.IsExtension(ap.p):
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
			case name == "title":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Title = result
			case name == "id":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.ID = result
			case name == "updated" || name == "modified":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Updated = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					atom.UpdatedParsed = &utcDate
				}
			case name == "subtitle" || name == "tagline":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Subtitle = result
			case name == "link":
				result, err := ap.parseLink()
				if err != nil {
					return nil, err
				}
				links = append(links, result)
			case name == "generator":
				result, err := ap.parseGenerator()
				if err != nil {
					return nil, err
				}
				atom.Generator = result
			case name == "icon":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Icon = result
			case name == "logo":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Logo = result
			case name == "rights" || name == "copyright":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				atom.Rights = result
			case name == "contributor":
				result, err := ap.parsePerson("contributor")
				if err != nil {
					return nil, err
				}
				contributors = append(contributors, result)
			case name == "author":
				result, err := ap.parsePerson("author")
				if err != nil {
					return nil, err
				}
				authors = append(authors, result)
			case name == "category":
				result, err := ap.parseCategory()
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			case name == "entry":
				result, err := ap.parseEntry()
				if err != nil {
					return nil, err
				}
				atom.Entries = append(atom.Entries, result)
			default:
				// For non-standard Atom feed elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  ap.p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range ap.p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(ap.p)
				if err != nil {
					ap.p.Skip() //nolint:errcheck // upstream ignores err
					continue
				}
				customExt.Value = result

				// Initialize extensions map if needed
				if extensions == nil {
					extensions = make(ext.Extensions)
				}
				if extensions["_custom"] == nil {
					extensions["_custom"] = make(map[string][]ext.Extension)
				}

				// Add to extensions
				extensions["_custom"][ap.p.Name] = append(
					extensions["_custom"][ap.p.Name], customExt)
			}
		}
	}

	if len(categories) > 0 {
		atom.Categories = categories
	}

	if len(authors) > 0 {
		atom.Authors = authors
	}

	if len(contributors) > 0 {
		atom.Contributors = contributors
	}

	if len(links) > 0 {
		atom.Links = links
	}

	if len(extensions) > 0 {
		atom.Extensions = extensions
	}

	if err := ap.expect(xpp.EndTag, "feed"); err != nil {
		return nil, err
	}

	return atom, nil
}

func (ap *Parser) expect(event xpp.XMLEventType, name string) error {
	if err := ap.p.Expect(event, name); err != nil {
		return fmt.Errorf("gofeed/atom: %w", err)
	}
	return nil
}

// nextTag iterates through the tokens until it reaches a StartTag or EndTag. It
// resolves urls in tag attributes relative to the current xml:base.
//
// nextTag is similar to goxpp's NextTag method except it wont throw an error if
// the next immediate token isnt a Start/EndTag. Instead, it will continue to
// consume tokens until it hits a Start/EndTag or EndDocument.
func (ap *Parser) nextTag() (xpp.XMLEventType, error) {
	for {
		event, err := ap.p.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/atom: %w", err)
		}

		switch event {
		case xpp.EndTag:
			return event, nil
		case xpp.StartTag:
			ap.resolveAttrs()
			return event, nil
		case xpp.EndDocument:
			return event, errors.New(
				"gofeed/atom: failed to find NextTag before reaching the end of the document")
		}
	}
}

// resolveAttrs resolves relative URI attributes according to xml:base.
func (ap *Parser) resolveAttrs() {
	for i := range ap.p.Attrs {
		attr := &ap.p.Attrs[i]
		lowerName := strings.ToLower(attr.Name.Local)
		if _, ok := atomUriAttrs[lowerName]; ok {
			absURL, err := xmlBaseResolveUrl(ap.p.BaseStack.Top(), attr.Value)
			if err == nil {
				attr.Value = absURL.String()
			}
			// Continue processing even if URL resolution fails (e.g., for non-HTTP
			// URIs like at://)
		}
	}
}

func (ap *Parser) parseEntry() (*Entry, error) {
	if err := ap.expect(xpp.StartTag, "entry"); err != nil {
		return nil, err
	}
	entry := &Entry{}

	contributors := []*Person{}
	authors := []*Person{}
	categories := []*Category{}
	links := []*Link{}
	extensions := ext.Extensions{}

	for {
		tok, err := ap.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(ap.p.Name)

			switch {
			case shared.IsExtension(ap.p):
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
			case name == "title":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.Title = result
			case name == "id":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.ID = result
			case name == "rights" || name == "copyright":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.Rights = result
			case name == "summary":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.Summary = result
			case name == "source":
				result, err := ap.parseSource()
				if err != nil {
					return nil, err
				}
				entry.Source = result
			case name == "updated" || name == "modified":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.Updated = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					entry.UpdatedParsed = &utcDate
				}
			case name == "contributor":
				result, err := ap.parsePerson("contributor")
				if err != nil {
					return nil, err
				}
				contributors = append(contributors, result)
			case name == "author":
				result, err := ap.parsePerson("author")
				if err != nil {
					return nil, err
				}
				authors = append(authors, result)
			case name == "category":
				result, err := ap.parseCategory()
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			case name == "link":
				result, err := ap.parseLink()
				if err != nil {
					return nil, err
				}
				links = append(links, result)
			case name == "published" || name == "issued":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				entry.Published = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					entry.PublishedParsed = &utcDate
				}
			case name == "content":
				result, err := ap.parseContent()
				if err != nil {
					return nil, err
				}
				entry.Content = result
			default:
				// For non-standard Atom entry elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  ap.p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range ap.p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(ap.p)
				if err != nil {
					ap.p.Skip() //nolint:errcheck // upstream ignores err
					continue
				}
				customExt.Value = result

				// Initialize extensions map if needed
				if extensions == nil {
					extensions = make(ext.Extensions)
				}
				if extensions["_custom"] == nil {
					extensions["_custom"] = make(map[string][]ext.Extension)
				}

				// Add to extensions
				extensions["_custom"][ap.p.Name] = append(
					extensions["_custom"][ap.p.Name], customExt)
			}
		}
	}

	if len(categories) > 0 {
		entry.Categories = categories
	}

	if len(authors) > 0 {
		entry.Authors = authors
	}

	if len(links) > 0 {
		entry.Links = links
	}

	if len(contributors) > 0 {
		entry.Contributors = contributors
	}

	if len(extensions) > 0 {
		entry.Extensions = extensions
	}

	if err := ap.expect(xpp.EndTag, "entry"); err != nil {
		return nil, err
	}

	return entry, nil
}

func (ap *Parser) parseSource() (*Source, error) {
	if err := ap.expect(xpp.StartTag, "source"); err != nil {
		return nil, err
	}

	source := &Source{}

	contributors := []*Person{}
	authors := []*Person{}
	categories := []*Category{}
	links := []*Link{}
	extensions := ext.Extensions{}

	for {
		tok, err := ap.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(ap.p.Name)

			switch {
			case shared.IsExtension(ap.p):
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
			case name == "title":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Title = result
			case name == "id":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.ID = result
			case name == "updated" || name == "modified":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Updated = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					source.UpdatedParsed = &utcDate
				}
			case name == "subtitle" || name == "tagline":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Subtitle = result
			case name == "link":
				result, err := ap.parseLink()
				if err != nil {
					return nil, err
				}
				links = append(links, result)
			case name == "generator":
				result, err := ap.parseGenerator()
				if err != nil {
					return nil, err
				}
				source.Generator = result
			case name == "icon":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Icon = result
			case name == "logo":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Logo = result
			case name == "rights" || name == "copyright":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				source.Rights = result
			case name == "contributor":
				result, err := ap.parsePerson("contributor")
				if err != nil {
					return nil, err
				}
				contributors = append(contributors, result)
			case name == "author":
				result, err := ap.parsePerson("author")
				if err != nil {
					return nil, err
				}
				authors = append(authors, result)
			case name == "category":
				result, err := ap.parseCategory()
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			default:
				err := ap.p.Skip()
				if err != nil {
					return nil, fmt.Errorf("gofeed/atom: %w", err)
				}
			}
		}
	}

	if len(categories) > 0 {
		source.Categories = categories
	}

	if len(authors) > 0 {
		source.Authors = authors
	}

	if len(contributors) > 0 {
		source.Contributors = contributors
	}

	if len(links) > 0 {
		source.Links = links
	}

	if len(extensions) > 0 {
		source.Extensions = extensions
	}

	if err := ap.expect(xpp.EndTag, "source"); err != nil {
		return nil, err
	}

	return source, nil
}

func (ap *Parser) parseContent() (*Content, error) {
	c := &Content{}
	c.Type = ap.p.Attribute("type")
	c.Src = ap.p.Attribute("src")

	text, err := ap.parseAtomText()
	if err != nil {
		return nil, err
	}
	c.Value = text

	return c, nil
}

func (ap *Parser) parsePerson(name string) (*Person, error) {
	if err := ap.expect(xpp.StartTag, name); err != nil {
		return nil, err
	}

	person := &Person{}

	for {
		tok, err := ap.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(ap.p.Name)

			switch name {
			case "name":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				person.Name = result
			case "email":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				person.Email = result
			case "uri", "url", "homepage":
				result, err := ap.parseAtomText()
				if err != nil {
					return nil, err
				}
				person.URI = result
			default:
				err := ap.p.Skip()
				if err != nil {
					return nil, fmt.Errorf("gofeed/atom: %w", err)
				}
			}
		}
	}

	if err := ap.expect(xpp.EndTag, name); err != nil {
		return nil, err
	}

	return person, nil
}

func (ap *Parser) parseLink() (*Link, error) {
	if err := ap.expect(xpp.StartTag, "link"); err != nil {
		return nil, err
	}

	l := &Link{}
	l.Href = ap.p.Attribute("href")
	l.Hreflang = ap.p.Attribute("hreflang")
	l.Type = ap.p.Attribute("type")
	l.Length = ap.p.Attribute("length")
	l.Title = ap.p.Attribute("title")
	l.Rel = ap.p.Attribute("rel")
	if l.Rel == "" {
		l.Rel = "alternate"
	}

	if err := ap.p.Skip(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	if err := ap.expect(xpp.EndTag, "link"); err != nil {
		return nil, err
	}
	return l, nil
}

func (ap *Parser) parseCategory() (*Category, error) {
	if err := ap.expect(xpp.StartTag, "category"); err != nil {
		return nil, err
	}

	c := &Category{}
	c.Term = ap.p.Attribute("term")
	c.Scheme = ap.p.Attribute("scheme")
	c.Label = ap.p.Attribute("label")

	if err := ap.p.Skip(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	if err := ap.expect(xpp.EndTag, "category"); err != nil {
		return nil, err
	}
	return c, nil
}

func (ap *Parser) parseGenerator() (*Generator, error) {
	if err := ap.expect(xpp.StartTag, "generator"); err != nil {
		return nil, err
	}

	g := &Generator{}

	uri := ap.p.Attribute("uri") // Atom 1.0
	url := ap.p.Attribute("url") // Atom 0.3

	if uri != "" {
		g.URI = uri
	} else if url != "" {
		g.URI = url
	}

	g.Version = ap.p.Attribute("version")

	result, err := ap.parseAtomText()
	if err != nil {
		return nil, err
	}

	g.Value = result

	if err := ap.expect(xpp.EndTag, "generator"); err != nil {
		return nil, err
	}

	return g, nil
}

func (ap *Parser) parseAtomText() (string, error) {
	var text struct {
		Type     string `xml:"type,attr"`
		Mode     string `xml:"mode,attr"`
		InnerXML string `xml:",innerxml"`

		XHTML struct {
			XMLName  xml.Name `xml:"div"`
			InnerXML string   `xml:",innerxml"`
		} `xml:"http://www.w3.org/1999/xhtml div"`
	}

	// get current base URL before it is clobbered by DecodeElement
	base := ap.p.BaseStack.Top()
	err := ap.p.DecodeElement(&text)
	if err != nil {
		return "", fmt.Errorf("gofeed/atom: %w", err)
	}

	lowerType := strings.ToLower(text.Type)
	lowerMode := strings.ToLower(text.Mode)

	var result string
	xhtmlType := strings.Contains(lowerType, "xhtml") || lowerType == "html"
	if xhtmlType && text.XHTML.XMLName.Local == "div" {
		result = strings.TrimSpace(text.XHTML.InnerXML)
	} else {
		result = strings.TrimSpace(text.InnerXML)
	}

	if strings.Contains(result, "<![CDATA[") {
		result = shared.StripCDATA(result)
	} else {
		// decode non-CDATA contents depending on type

		switch {
		case lowerType == "text" ||
			strings.HasPrefix(lowerType, "text/") ||
			(lowerType == "" && lowerMode == ""):
			result = html.UnescapeString(result)
		case strings.Contains(lowerType, "xhtml"):
		// do nothing
		case lowerType == "html":
			result = html.UnescapeString(result)
		default:
			decodedStr, err := base64.StdEncoding.DecodeString(result)
			if err == nil {
				result = string(decodedStr)
			}
		}
	}

	// resolve relative URIs in URI-containing elements according to xml:base
	name := strings.ToLower(ap.p.Name)
	if _, ok := atomUriElements[name]; ok {
		resolved, err := xmlBaseResolveUrl(base, result)
		if resolved != nil && err == nil {
			result = resolved.String()
		}
	}
	return result, nil
}

func (ap *Parser) parseLanguage() string { return ap.p.Attribute("lang") }

func (ap *Parser) parseVersion() string {
	ver := ap.p.Attribute("version")
	if ver != "" {
		return ver
	}

	ns := ap.p.Attribute("xmlns")
	if ns == "http://purl.org/atom/ns#" {
		return "0.3"
	}

	if ns == "http://www.w3.org/2005/Atom" {
		return "1.0"
	}

	return ""
}
