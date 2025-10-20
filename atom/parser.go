package atom

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"strings"
	"time"

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
	p   *xpp.XMLPullParser
	err error
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

	atom := &Feed{
		Language: ap.parseLanguage(),
		Entries:  []*Entry{},
		Version:  ap.parseVersion(),
	}
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
			if shared.IsExtension(ap.p) {
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
				continue
			}

			switch strings.ToLower(ap.p.Name) {
			case "title":
				ap.parseAtomTextTo(&atom.Title)
			case "id":
				ap.parseAtomTextTo(&atom.ID)
			case "updated", "modified":
				ap.parseAtomDateTo(&atom.Updated, &atom.UpdatedParsed)
			case "subtitle", "tagline":
				ap.parseAtomTextTo(&atom.Subtitle)
			case "link":
				atom.Links = ap.parseLinkTo(atom.Links)
			case "generator":
				ap.parseGeneratorTo(&atom.Generator)
			case "icon":
				ap.parseAtomTextTo(&atom.Icon)
			case "logo":
				ap.parseAtomTextTo(&atom.Logo)
			case "rights", "copyright":
				ap.parseAtomTextTo(&atom.Rights)
			case "contributor":
				atom.Contributors = ap.parsePersonTo("contributor", atom.Contributors)
			case "author":
				atom.Authors = ap.parsePersonTo("author", atom.Authors)
			case "category":
				atom.Categories = ap.parseCategoryTo(atom.Categories)
			case "entry":
				atom.Entries = ap.parseEntryTo(atom.Entries)
			default:
				// For non-standard Atom feed elements, add them to extensions
				// under a special "_custom" namespace prefix
				extensitons2, ok := ap.parseCustomExtInto(extensions)
				if !ok {
					continue
				}
				extensions = extensitons2
			}
		}
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
	if ap.err != nil {
		return 0, ap.err
	}

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
	base := ap.p.BaseStack.Top()
	if base == nil {
		return
	}

	for i := range ap.p.Attrs {
		attr := &ap.p.Attrs[i]
		lowerName := strings.ToLower(attr.Name.Local)
		if _, ok := atomUriAttrs[lowerName]; ok {
			absURL, err := xmlBaseResolveUrl(base, attr.Value)
			if err == nil {
				attr.Value = absURL.String()
			}
			// Continue processing even if URL resolution fails (e.g., for non-HTTP
			// URIs like at://)
		}
	}
}

func (ap *Parser) parseEntryTo(entries []*Entry) []*Entry {
	entry, err := ap.parseEntry()
	if err != nil {
		ap.err = err
		return entries
	}
	return append(entries, entry)
}

func (ap *Parser) parseEntry() (*Entry, error) {
	if err := ap.expect(xpp.StartTag, "entry"); err != nil {
		return nil, err
	}

	entry := &Entry{}
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
			if shared.IsExtension(ap.p) {
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
				continue
			}

			switch strings.ToLower(ap.p.Name) {
			case "title":
				ap.parseAtomTextTo(&entry.Title)
			case "id":
				ap.parseAtomTextTo(&entry.ID)
			case "rights", "copyright":
				ap.parseAtomTextTo(&entry.Rights)
			case "summary":
				ap.parseAtomTextTo(&entry.Summary)
			case "source":
				ap.parseSourceTo(&entry.Source)
			case "updated", "modified":
				ap.parseAtomDateTo(&entry.Updated, &entry.UpdatedParsed)
			case "contributor":
				entry.Contributors = ap.parsePersonTo("contributor", entry.Contributors)
			case "author":
				entry.Authors = ap.parsePersonTo("author", entry.Authors)
			case "category":
				entry.Categories = ap.parseCategoryTo(entry.Categories)
			case "link":
				entry.Links = ap.parseLinkTo(entry.Links)
			case "published", "issued":
				ap.parseAtomDateTo(&entry.Published, &entry.PublishedParsed)
			case "content":
				ap.parseContentTo(&entry.Content)
			default:
				// For non-standard Atom entry elements, add them to extensions
				// under a special "_custom" namespace prefix
				extensions2, ok := ap.parseCustomExtInto(extensions)
				if !ok {
					continue
				}
				extensions = extensions2
			}
		}
	}

	if len(extensions) > 0 {
		entry.Extensions = extensions
	}

	if err := ap.expect(xpp.EndTag, "entry"); err != nil {
		return nil, err
	}
	return entry, nil
}

func (ap *Parser) parseSourceTo(ref **Source) {
	src, err := ap.parseSource()
	if err != nil {
		ap.err = err
		return
	}
	*ref = src
}

func (ap *Parser) parseSource() (*Source, error) {
	if err := ap.expect(xpp.StartTag, "source"); err != nil {
		return nil, err
	}

	source := &Source{}
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
			if shared.IsExtension(ap.p) {
				e, err := shared.ParseExtension(extensions, ap.p)
				if err != nil {
					return nil, err
				}
				extensions = e
				continue
			}

			switch strings.ToLower(ap.p.Name) {
			case "title":
				ap.parseAtomTextTo(&source.Title)
			case "id":
				ap.parseAtomTextTo(&source.ID)
			case "updated", "modified":
				ap.parseAtomDateTo(&source.Updated, &source.UpdatedParsed)
			case "subtitle", "tagline":
				ap.parseAtomTextTo(&source.Subtitle)
			case "link":
				source.Links = ap.parseLinkTo(source.Links)
			case "generator":
				ap.parseGeneratorTo(&source.Generator)
			case "icon":
				ap.parseAtomTextTo(&source.Icon)
			case "logo":
				ap.parseAtomTextTo(&source.Logo)
			case "rights", "copyright":
				ap.parseAtomTextTo(&source.Rights)
			case "contributor":
				source.Contributors = ap.parsePersonTo("contributor",
					source.Contributors)
			case "author":
				source.Authors = ap.parsePersonTo("author", source.Authors)
			case "category":
				source.Categories = ap.parseCategoryTo(source.Categories)
			default:
				if err := ap.p.Skip(); err != nil {
					return nil, fmt.Errorf("gofeed/atom: %w", err)
				}
			}
		}
	}

	if len(extensions) > 0 {
		source.Extensions = extensions
	}

	if err := ap.expect(xpp.EndTag, "source"); err != nil {
		return nil, err
	}
	return source, nil
}

func (ap *Parser) parseContentTo(ref **Content) {
	c := &Content{
		Type: ap.p.Attribute("type"),
		Src:  ap.p.Attribute("src"),
	}

	if ap.parseAtomTextTo(&c.Value); ap.err != nil {
		return
	}
	*ref = c
}

func (ap *Parser) parsePersonTo(name string, persons []*Person) []*Person {
	p, err := ap.parsePerson(name)
	if err != nil {
		ap.err = err
		return persons
	}
	return append(persons, p)
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
			switch strings.ToLower(ap.p.Name) {
			case "name":
				ap.parseAtomTextTo(&person.Name)
			case "email":
				ap.parseAtomTextTo(&person.Email)
			case "uri", "url", "homepage":
				ap.parseAtomTextTo(&person.URI)
			default:
				if err := ap.p.Skip(); err != nil {
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

func (ap *Parser) parseLinkTo(links []*Link) []*Link {
	link, err := ap.parseLink()
	if err != nil {
		ap.err = err
		return links
	}
	return append(links, link)
}

func (ap *Parser) parseLink() (*Link, error) {
	if err := ap.expect(xpp.StartTag, "link"); err != nil {
		return nil, err
	}

	l := &Link{
		Href:     ap.p.Attribute("href"),
		Hreflang: ap.p.Attribute("hreflang"),
		Type:     ap.p.Attribute("type"),
		Length:   ap.p.Attribute("length"),
		Title:    ap.p.Attribute("title"),
		Rel:      ap.p.Attribute("rel"),
	}
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

func (ap *Parser) parseCategoryTo(cats []*Category) []*Category {
	cat, err := ap.parseCategory()
	if err != nil {
		ap.err = err
		return cats
	}
	return append(cats, cat)
}

func (ap *Parser) parseCategory() (*Category, error) {
	if err := ap.expect(xpp.StartTag, "category"); err != nil {
		return nil, err
	}

	c := &Category{
		Term:   ap.p.Attribute("term"),
		Scheme: ap.p.Attribute("scheme"),
		Label:  ap.p.Attribute("label"),
	}

	if err := ap.p.Skip(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	if err := ap.expect(xpp.EndTag, "category"); err != nil {
		return nil, err
	}
	return c, nil
}

func (ap *Parser) parseGeneratorTo(ref **Generator) {
	g, err := ap.parseGenerator()
	if err != nil {
		ap.err = err
		return
	}
	*ref = g
}

func (ap *Parser) parseGenerator() (*Generator, error) {
	if err := ap.expect(xpp.StartTag, "generator"); err != nil {
		return nil, err
	}

	g := &Generator{
		Version: ap.p.Attribute("version"),
	}

	if uri := ap.p.Attribute("uri"); uri != "" {
		g.URI = uri // Atom 1.0
	} else if url := ap.p.Attribute("url"); url != "" {
		g.URI = url // Atom 0.3
	}

	if ap.parseAtomTextTo(&g.Value); ap.err != nil {
		return nil, ap.err
	}

	if err := ap.expect(xpp.EndTag, "generator"); err != nil {
		return nil, err
	}
	return g, nil
}

func (ap *Parser) parseAtomTextTo(ref *string) {
	s, err := ap.parseAtomText()
	if err != nil {
		ap.err = err
		return
	}
	*ref = s
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
	if _, ok := atomUriElements[name]; ok && base != nil {
		resolved, err := xmlBaseResolveUrl(base, result)
		if resolved != nil && err == nil {
			result = resolved.String()
		}
	}
	return result, nil
}

func (ap *Parser) parseLanguage() string { return ap.p.Attribute("lang") }

func (ap *Parser) parseVersion() string {
	if ver := ap.p.Attribute("version"); ver != "" {
		return ver
	}

	if ns := ap.p.Attribute("xmlns"); ns == "http://purl.org/atom/ns#" {
		return "0.3"
	} else if ns == "http://www.w3.org/2005/Atom" {
		return "1.0"
	}
	return ""
}

func (ap *Parser) parseAtomDateTo(strRef *string, dtRef **time.Time) {
	if ap.parseAtomTextTo(strRef); ap.err != nil {
		return
	}

	date, err := shared.ParseDate(*strRef)
	if err != nil {
		ap.err = err
		return
	}

	utcDate := date.UTC()
	*dtRef = &utcDate
}

func (ap *Parser) parseCustomExtInto(extensions ext.Extensions) (ext.Extensions,
	bool,
) {
	custom := ext.Extension{
		Name:  ap.p.Name,
		Attrs: make(map[string]string),
	}

	// Copy attributes
	for _, attr := range ap.p.Attrs {
		custom.Attrs[attr.Name.Local] = attr.Value
	}

	// Parse the text content
	result, err := shared.ParseText(ap.p)
	if err != nil {
		ap.p.Skip() //nolint:errcheck // upstream ignores err
		return nil, false
	}
	custom.Value = result

	// Initialize extensions map if needed
	if extensions == nil {
		extensions = make(ext.Extensions, 1)
	}
	if extensions["_custom"] == nil {
		extensions["_custom"] = make(map[string][]ext.Extension, 1)
	}

	// Add to extensions
	extensions["_custom"][ap.p.Name] = append(extensions["_custom"][ap.p.Name],
		custom)
	return extensions, true
}
