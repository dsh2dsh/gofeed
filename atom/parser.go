package atom

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

const (
	categoryTag  = "category"
	customKey    = "_custom"
	entryTag     = "entry"
	feedTag      = "feed"
	generatorTag = "generator"
	linkTag      = "link"
	sourceTag    = "source"
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

var emptyAttrs = map[string]string{}

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
	if err := ap.expect(xpp.StartTag, feedTag); err != nil {
		return nil, err
	}

	atom := &Feed{
		Language: ap.language(),
		Entries:  []*Entry{},
		Version:  ap.version(),
	}

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
				e, err := shared.ParseExtension(atom.Extensions, ap.p)
				if err != nil {
					return nil, err
				}
				atom.Extensions = e
				continue
			}

			switch tag := strings.ToLower(ap.p.Name); tag {
			case "title":
				atom.Title = ap.text()
			case "id":
				atom.ID = ap.text()
			case "updated", "modified":
				atom.Updated, atom.UpdatedParsed = ap.parseDate()
			case "subtitle", "tagline":
				atom.Subtitle = ap.text()
			case linkTag:
				atom.Links = ap.appendLink(atom.Links)
			case generatorTag:
				atom.Generator = ap.generator()
			case "icon":
				atom.Icon = ap.text()
			case "logo":
				atom.Logo = ap.text()
			case "rights", "copyright":
				atom.Rights = ap.text()
			case "contributor":
				atom.Contributors = ap.appendPerson(tag, atom.Contributors)
			case "author":
				atom.Authors = ap.appendPerson(tag, atom.Authors)
			case categoryTag:
				atom.Categories = ap.appendCategory(atom.Categories)
			case entryTag:
				atom.Entries = ap.appendEntry(atom.Entries)
			default:
				// For non-standard Atom feed elements, add them to extensions
				// under a special "_custom" namespace prefix
				if e, ok := ap.parseCustomExtInto(atom.Extensions); ok {
					atom.Extensions = e
				}
			}
		}
	}

	if err := ap.expect(xpp.EndTag, feedTag); err != nil {
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
	if ap.p.BaseStack.Top() == nil {
		return
	}

	for i := range ap.p.Attrs {
		attr := &ap.p.Attrs[i]
		lowerName := strings.ToLower(attr.Name.Local)
		if _, ok := atomUriAttrs[lowerName]; ok {
			absURL, err := ap.p.XmlBaseResolveUrl(attr.Value)
			if err == nil && absURL != nil {
				attr.Value = absURL.String()
			}
			// Continue processing even if URL resolution fails (e.g., for non-HTTP
			// URIs like at://)
		}
	}
}

func (ap *Parser) appendEntry(entries []*Entry) []*Entry {
	entry, err := ap.parseEntry()
	if err != nil {
		ap.err = err
		return entries
	}
	return append(entries, entry)
}

func (ap *Parser) parseEntry() (*Entry, error) {
	if err := ap.expect(xpp.StartTag, entryTag); err != nil {
		return nil, err
	}

	entry := new(Entry)

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
				e, err := shared.ParseExtension(entry.Extensions, ap.p)
				if err != nil {
					return nil, err
				}
				entry.Extensions = e
				continue
			}

			switch tag := strings.ToLower(ap.p.Name); tag {
			case "title":
				entry.Title = ap.text()
			case "id":
				entry.ID = ap.text()
			case "rights", "copyright":
				entry.Rights = ap.text()
			case "summary":
				entry.Summary = ap.text()
			case sourceTag:
				entry.Source = ap.source()
			case "updated", "modified":
				entry.Updated, entry.UpdatedParsed = ap.parseDate()
			case "contributor":
				entry.Contributors = ap.appendPerson(tag, entry.Contributors)
			case "author":
				entry.Authors = ap.appendPerson(tag, entry.Authors)
			case categoryTag:
				entry.Categories = ap.appendCategory(entry.Categories)
			case linkTag:
				entry.Links = ap.appendLink(entry.Links)
			case "published", "issued":
				entry.Published, entry.PublishedParsed = ap.parseDate()
			case "content":
				entry.Content = ap.content()
			default:
				// For non-standard Atom entry elements, add them to extensions
				// under a special "_custom" namespace prefix
				if e, ok := ap.parseCustomExtInto(entry.Extensions); ok {
					entry.Extensions = e
				}
			}
		}
	}

	if err := ap.expect(xpp.EndTag, entryTag); err != nil {
		return nil, err
	}
	return entry, nil
}

func (ap *Parser) source() *Source {
	src, err := ap.parseSource()
	if err != nil {
		ap.err = err
		return nil
	}
	return src
}

func (ap *Parser) parseSource() (*Source, error) {
	if err := ap.expect(xpp.StartTag, sourceTag); err != nil {
		return nil, err
	}

	source := new(Source)

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
				e, err := shared.ParseExtension(source.Extensions, ap.p)
				if err != nil {
					return nil, err
				}
				source.Extensions = e
				continue
			}

			switch tag := strings.ToLower(ap.p.Name); tag {
			case "title":
				source.Title = ap.text()
			case "id":
				source.ID = ap.text()
			case "updated", "modified":
				source.Updated, source.UpdatedParsed = ap.parseDate()
			case "subtitle", "tagline":
				source.Subtitle = ap.text()
			case linkTag:
				source.Links = ap.appendLink(source.Links)
			case generatorTag:
				source.Generator = ap.generator()
			case "icon":
				source.Icon = ap.text()
			case "logo":
				source.Logo = ap.text()
			case "rights", "copyright":
				source.Rights = ap.text()
			case "contributor":
				source.Contributors = ap.appendPerson(tag, source.Contributors)
			case "author":
				source.Authors = ap.appendPerson(tag, source.Authors)
			case categoryTag:
				source.Categories = ap.appendCategory(source.Categories)
			default:
				if err := ap.p.Skip(); err != nil {
					return nil, fmt.Errorf("gofeed/atom: %w", err)
				}
			}
		}
	}

	if err := ap.expect(xpp.EndTag, sourceTag); err != nil {
		return nil, err
	}
	return source, nil
}

func (ap *Parser) content() *Content {
	c := new(Content)
	for _, attr := range ap.p.Attrs {
		switch attr.Name.Local {
		case "type":
			c.Type = attr.Value
		case "src":
			c.Src = attr.Value
		}
	}

	if c.Value = ap.text(); ap.err != nil {
		return nil
	}
	return c
}

func (ap *Parser) appendPerson(name string, persons []*Person) []*Person {
	p, err := ap.parsePerson(name)
	if err != nil {
		ap.err = err
		return persons
	}
	return append(persons, p)
}

func (ap *Parser) parsePerson(tag string) (*Person, error) {
	if err := ap.expect(xpp.StartTag, tag); err != nil {
		return nil, err
	}

	person := new(Person)

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
				person.Name = ap.text()
			case "email":
				person.Email = ap.text()
			case "uri", "url", "homepage":
				person.URI = ap.text()
			default:
				if err := ap.p.Skip(); err != nil {
					return nil, fmt.Errorf("gofeed/atom: %w", err)
				}
			}
		}
	}

	if err := ap.expect(xpp.EndTag, tag); err != nil {
		return nil, err
	}
	return person, nil
}

func (ap *Parser) appendLink(links []*Link) []*Link {
	link, err := ap.parseLink()
	if err != nil {
		ap.err = err
		return links
	}
	return append(links, link)
}

func (ap *Parser) parseLink() (*Link, error) {
	if err := ap.expect(xpp.StartTag, linkTag); err != nil {
		return nil, err
	}

	l := &Link{Rel: "alternate"}

	for _, attr := range ap.p.Attrs {
		switch attr.Name.Local {
		case "href":
			l.Href = attr.Value
		case "hreflang":
			l.Hreflang = attr.Value
		case "type":
			l.Type = attr.Value
		case "length":
			l.Length = attr.Value
		case "title":
			l.Title = attr.Value
		case "rel":
			l.Rel = attr.Value
		}
	}

	if err := ap.p.Skip(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	if err := ap.expect(xpp.EndTag, linkTag); err != nil {
		return nil, err
	}
	return l, nil
}

func (ap *Parser) appendCategory(cats []*Category) []*Category {
	cat, err := ap.parseCategory()
	if err != nil {
		ap.err = err
		return cats
	}
	return append(cats, cat)
}

func (ap *Parser) parseCategory() (*Category, error) {
	if err := ap.expect(xpp.StartTag, categoryTag); err != nil {
		return nil, err
	}

	c := new(Category)
	for _, attr := range ap.p.Attrs {
		switch attr.Name.Local {
		case "term":
			c.Term = attr.Value
		case "scheme":
			c.Scheme = attr.Value
		case "label":
			c.Label = attr.Value
		}
	}

	if err := ap.p.Skip(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	if err := ap.expect(xpp.EndTag, categoryTag); err != nil {
		return nil, err
	}
	return c, nil
}

func (ap *Parser) generator() *Generator {
	g, err := ap.parseGenerator()
	if err != nil {
		ap.err = err
		return nil
	}
	return g
}

func (ap *Parser) parseGenerator() (*Generator, error) {
	if err := ap.expect(xpp.StartTag, generatorTag); err != nil {
		return nil, err
	}

	g := new(Generator)
	for _, attr := range ap.p.Attrs {
		switch attr.Name.Local {
		case "url", "uri":
			if attr.Value != "" {
				g.URI = attr.Value
			}
		case "version":
			g.Version = attr.Value
		}
	}

	if g.Value = ap.text(); ap.err != nil {
		return nil, ap.err
	}

	if err := ap.expect(xpp.EndTag, generatorTag); err != nil {
		return nil, err
	}
	return g, nil
}

func (ap *Parser) text() string {
	s, err := ap.parseAtomText()
	if err != nil {
		ap.err = err
		return ""
	}
	return s
}

func (ap *Parser) parseAtomText() (string, error) {
	attrs := ap.textAttributes()
	if attrs.XHTML() {
		return ap.xhtmlContent()
	}

	xmlBaseResolver := ap.p.XmlBaseResolver()

	result, err := shared.ParseText(ap.p)
	if err != nil {
		return "", fmt.Errorf("gofeed/atom: extract text from %q type=%s: %w",
			ap.p.Name, attrs.Type, err)
	}

	if attrs.Encoded() {
		if b, err := base64.StdEncoding.DecodeString(result); err == nil {
			return string(b), nil
		}
	}

	// resolve relative URIs in URI-containing elements according to xml:base
	if _, ok := atomUriElements[strings.ToLower(ap.p.Name)]; !ok {
		return result, nil
	}

	if u, err := xmlBaseResolver(result); err == nil && u != nil {
		return u.String(), nil
	}
	return result, nil
}

func (ap *Parser) textAttributes() textAttributes {
	var attrs textAttributes
	for _, attr := range ap.p.Attrs {
		switch attr.Name.Local {
		case "mode":
			attrs.Mode = attr.Value
		case "type":
			attrs.Type = strings.ToLower(attr.Value)
		}
	}
	return attrs
}

func (ap *Parser) xhtmlContent() (string, error) {
	var xhtmlContent struct {
		XHTML struct {
			XMLName  xml.Name `xml:"div"`
			InnerXML string   `xml:",innerxml"`
		} `xml:"http://www.w3.org/1999/xhtml div"`
	}

	if err := ap.p.DecodeElement(&xhtmlContent); err != nil {
		return "", fmt.Errorf("gofeed/atom: extract xhtml text from %q: %w",
			ap.p.Name, err)
	}
	return strings.TrimSpace(xhtmlContent.XHTML.InnerXML), nil
}

func (ap *Parser) language() string { return ap.p.Attribute("lang") }

func (ap *Parser) version() string {
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

func (ap *Parser) parseDate() (string, *time.Time) {
	s := ap.text()
	if ap.err != nil {
		return "", nil
	}

	date, err := shared.ParseDate(s)
	if err != nil {
		return s, nil
	}

	utcDate := date.UTC()
	return s, &utcDate
}

func (ap *Parser) parseCustomExtInto(extensions ext.Extensions) (ext.Extensions,
	bool,
) {
	custom := ext.Extension{
		Name:  ap.p.Name,
		Attrs: emptyAttrs,
	}

	// Copy attributes
	if n := len(ap.p.Attrs); n != 0 {
		custom.Attrs = make(map[string]string, n)
		for _, attr := range ap.p.Attrs {
			custom.Attrs[attr.Name.Local] = attr.Value
		}
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
		extensions = ext.Extensions{customKey: {ap.p.Name: {custom}}}
	} else if m, ok := extensions[customKey]; !ok {
		extensions[customKey] = map[string][]ext.Extension{ap.p.Name: {custom}}
	} else {
		m[ap.p.Name] = append(m[ap.p.Name], custom)
	}
	return extensions, true
}
