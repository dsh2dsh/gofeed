package atom

import (
	"encoding/base64"
	"fmt"
	"io"
	"iter"
	"maps"
	"strings"
	"time"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
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
	p   *xml.Parser
	err error
}

var emptyAttrs = map[string]string{}

// NewParser creates a new Atom parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an xml feed into an atom.Feed
func (self *Parser) Parse(r io.Reader, opts ...options.Option) (*Feed, error) {
	self.p = xml.NewParser(
		xpp.NewXMLPullParser(r, false, shared.NewReaderLabel))

	if _, err := self.p.FindRoot(); err != nil {
		return nil, fmt.Errorf("gofeed/atom: %w", err)
	}

	feed := self.root()
	if err := self.Err(); err != nil {
		return nil, err
	}
	return feed, nil
}

func (self *Parser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/atom: xml parser errored: %w", self.p.Err())
	}
	return nil
}

func (self *Parser) root() *Feed {
	children := self.makeChildrenSeq(self.p.Name)
	if children == nil {
		return nil
	}

	atom := &Feed{
		Language: self.language(),
		Entries:  []*Entry{},
		Version:  self.version(),
	}

	for name := range children {
		self.feedBody(name, atom)
	}

	if self.err != nil {
		return nil
	}
	return atom
}

func (self *Parser) makeChildrenSeq(name string) iter.Seq[string] {
	children, err := self.p.MakeChildrenSeq(name)
	if err != nil {
		self.err = err
		return nil
	}

	return func(yield func(string) bool) {
		for name := range children {
			if err := self.Err(); err != nil {
				self.err = err
				return
			}

			self.resolveAttrs()
			if !yield(name) {
				break
			}
		}

		if err := self.Err(); err != nil {
			self.err = err
			return
		}
	}
}

// resolveAttrs resolves relative URI attributes according to xml:base.
func (self *Parser) resolveAttrs() {
	if self.p.BaseStack.Top() == nil {
		return
	}

	for i := range self.p.Attrs {
		attr := &self.p.Attrs[i]
		lowerName := strings.ToLower(attr.Name.Local)
		if _, ok := atomUriAttrs[lowerName]; ok {
			absURL, err := self.p.XmlBaseResolveUrl(attr.Value)
			if err == nil && absURL != nil {
				attr.Value = absURL.String()
			}
			// Continue processing even if URL resolution fails (e.g., for non-HTTP
			// URIs like at://)
		}
	}
}

func (self *Parser) feedBody(name string, atom *Feed) {
	if e, ok := self.extensions(atom.Extensions); ok {
		atom.Extensions = e
		return
	}

	switch name {
	case "title":
		atom.Title = self.text(name)
	case "id":
		atom.ID = self.text(name)
	case "updated", "modified":
		atom.Updated, atom.UpdatedParsed = self.parseDate(name)
	case "subtitle", "tagline":
		atom.Subtitle = self.text(name)
	case "link":
		atom.Links = self.appendLink(name, atom.Links)
	case "generator":
		atom.Generator = self.generator(name)
	case "icon":
		atom.Icon = self.text(name)
	case "logo":
		atom.Logo = self.text(name)
	case "rights", "copyright":
		atom.Rights = self.text(name)
	case "contributor":
		atom.Contributors = self.appendPerson(name, atom.Contributors)
	case "author":
		atom.Authors = self.appendPerson(name, atom.Authors)
	case "category":
		atom.Categories = self.appendCategory(name, atom.Categories)
	case "entry":
		atom.Entries = self.appendEntry(name, atom.Entries)
	default:
		// For non-standard Atom feed elements, add them to extensions
		// under a special "_custom" namespace prefix
		if e, ok := self.parseCustomExtInto(name, atom.Extensions); ok {
			atom.Extensions = e
		}
	}
}

func (self *Parser) extensions(e ext.Extensions) (ext.Extensions, bool) {
	if self.p.ExtensionPrefix() == "" {
		return e, false
	}

	e, err := shared.ParseExtension(e, self.p.XMLPullParser)
	if err != nil {
		self.err = err
	}
	return e, true
}

func (self *Parser) appendEntry(name string, entries []*Entry) []*Entry {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return entries
	}

	entry := new(Entry)
	for name := range children {
		self.entryBody(name, entry)
	}

	if self.err != nil {
		return entries
	}
	return append(entries, entry)
}

func (self *Parser) entryBody(name string, entry *Entry) {
	if e, ok := self.extensions(entry.Extensions); ok {
		entry.Extensions = e
		return
	}

	switch name {
	case "title":
		entry.Title = self.text(name)
	case "id":
		entry.ID = self.text(name)
	case "rights", "copyright":
		entry.Rights = self.text(name)
	case "summary":
		entry.Summary = self.text(name)
	case "source":
		entry.Source = self.source(name)
	case "updated", "modified":
		entry.Updated, entry.UpdatedParsed = self.parseDate(name)
	case "contributor":
		entry.Contributors = self.appendPerson(name, entry.Contributors)
	case "author":
		entry.Authors = self.appendPerson(name, entry.Authors)
	case "category":
		entry.Categories = self.appendCategory(name, entry.Categories)
	case "link":
		entry.Links = self.appendLink(name, entry.Links)
	case "published", "issued":
		entry.Published, entry.PublishedParsed = self.parseDate(name)
	case "content":
		entry.Content = self.content(name)
	default:
		// For non-standard Atom entry elements, add them to extensions
		// under a special "_custom" namespace prefix
		if e, ok := self.parseCustomExtInto(name, entry.Extensions); ok {
			entry.Extensions = e
		}
	}
}

func (self *Parser) source(name string) *Source {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	source := new(Source)
	for name := range children {
		self.sourceBody(name, source)
	}

	if self.err != nil {
		return nil
	}
	return source
}

func (self *Parser) sourceBody(name string, source *Source) {
	if e, ok := self.extensions(source.Extensions); ok {
		source.Extensions = e
		return
	}

	switch name {
	case "title":
		source.Title = self.text(name)
	case "id":
		source.ID = self.text(name)
	case "updated", "modified":
		source.Updated, source.UpdatedParsed = self.parseDate(name)
	case "subtitle", "tagline":
		source.Subtitle = self.text(name)
	case "link":
		source.Links = self.appendLink(name, source.Links)
	case "generator":
		source.Generator = self.generator(name)
	case "icon":
		source.Icon = self.text(name)
	case "logo":
		source.Logo = self.text(name)
	case "rights", "copyright":
		source.Rights = self.text(name)
	case "contributor":
		source.Contributors = self.appendPerson(name, source.Contributors)
	case "author":
		source.Authors = self.appendPerson(name, source.Authors)
	case "category":
		source.Categories = self.appendCategory(name, source.Categories)
	default:
		self.p.Skip(name)
	}
}

func (self *Parser) content(name string) (c *Content) {
	c = new(Content)
	for name, value := range self.p.AttributeSeq() {
		switch name {
		case "type":
			c.Type = value
		case "src":
			c.Src = value
		}
	}
	c.Value = self.text(name)
	return c
}

func (self *Parser) appendPerson(name string, persons []*Person) []*Person {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return persons
	}

	person := new(Person)
	for name := range children {
		self.personBody(name, person)
	}

	if self.err != nil {
		return persons
	}
	return append(persons, person)
}

func (self *Parser) personBody(name string, person *Person) {
	switch name {
	case "name":
		person.Name = self.text(name)
	case "email":
		person.Email = self.text(name)
	case "uri", "url", "homepage":
		person.URI = self.text(name)
	default:
		self.p.Skip(name)
	}
}

func (self *Parser) appendLink(name string, links []*Link) []*Link {
	l := &Link{Rel: "alternate"}
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			switch name {
			case "href":
				l.Href = value
			case "hreflang":
				l.Hreflang = value
			case "type":
				l.Type = value
			case "length":
				l.Length = value
			case "title":
				l.Title = value
			case "rel":
				l.Rel = value
			}
		}
		return nil
	})
	if err != nil {
		self.err = err
		return links
	}
	return append(links, l)
}

func (self *Parser) appendCategory(name string, categories []*Category,
) []*Category {
	c := new(Category)
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			switch name {
			case "term":
				c.Term = value
			case "scheme":
				c.Scheme = value
			case "label":
				c.Label = value
			}
		}
		return nil
	})
	if err != nil {
		self.err = err
		return categories
	}
	return append(categories, c)
}

func (self *Parser) generator(name string) (g *Generator) {
	err := self.p.WithText(name,
		func() error {
			g = new(Generator)
			for name, value := range self.p.AttributeSeq() {
				switch name {
				case "url", "uri":
					if value != "" {
						g.URI = value
					}
				case "version":
					g.Version = value
				}
			}
			return nil
		},
		func(s string) error {
			g.Value = s
			return nil
		})
	if err != nil {
		self.err = err
		return nil
	}
	return g
}

func (self *Parser) text(name string) string {
	s, err := self.parseAtomText(name)
	if err != nil {
		self.err = err
		return ""
	}
	return s
}

func (self *Parser) parseAtomText(name string) (string, error) {
	attrs := self.textAttributes()
	if attrs.XHTML() {
		return self.xhtmlContent()
	}

	xmlBaseResolver := self.p.XmlBaseResolver()

	var result string
	err := self.p.WithText(name, nil, func(s string) error {
		result = s
		return nil
	})
	if err != nil {
		return "", err
	}

	if attrs.Encoded() {
		if b, err := base64.StdEncoding.DecodeString(result); err == nil {
			return string(b), nil
		}
	}

	// resolve relative URIs in URI-containing elements according to xml:base
	if _, ok := atomUriElements[strings.ToLower(self.p.Name)]; !ok {
		return result, nil
	}

	if u, err := xmlBaseResolver(result); err == nil && u != nil {
		return u.String(), nil
	}
	return result, nil
}

func (self *Parser) textAttributes() textAttributes {
	var attrs textAttributes
	for name, value := range self.p.AttributeSeq() {
		switch name {
		case "mode":
			attrs.Mode = value
		case "type":
			attrs.Type = strings.ToLower(value)
		}
	}
	return attrs
}

func (self *Parser) xhtmlContent() (string, error) {
	var xhtmlContent struct {
		XHTML struct {
			InnerXML string `xml:",innerxml"`
		} `xml:"http://www.w3.org/1999/xhtml div"`
	}

	if err := self.p.DecodeElement(&xhtmlContent); err != nil {
		return "", fmt.Errorf("gofeed/atom: extract xhtml text from %q: %w",
			self.p.Name, err)
	}
	return strings.TrimSpace(xhtmlContent.XHTML.InnerXML), nil
}

func (self *Parser) language() string { return self.p.Attribute("lang") }

func (self *Parser) version() string {
	if ver := self.p.Attribute("version"); ver != "" {
		return ver
	}

	if ns := self.p.Attribute("xmlns"); ns == "http://purl.org/atom/ns#" {
		return "0.3"
	} else if ns == "http://www.w3.org/2005/Atom" {
		return "1.0"
	}
	return ""
}

func (self *Parser) parseDate(name string) (string, *time.Time) {
	var result string
	err := self.p.WithText(name, nil, func(s string) error {
		result = s
		return nil
	})
	if err != nil {
		self.err = err
		return "", nil
	}

	date, err := shared.ParseDate(result)
	if err != nil {
		return result, nil
	}

	utcDate := date.UTC()
	return result, &utcDate
}

func (self *Parser) parseCustomExtInto(name string, extensions ext.Extensions,
) (ext.Extensions, bool) {
	custom := ext.Extension{Name: self.p.Name, Attrs: emptyAttrs}
	// Copy attributes
	if n := len(self.p.Attrs); n != 0 {
		custom.Attrs = make(map[string]string, n)
		maps.Insert(custom.Attrs, self.p.AttributeSeq())
	}

	// Parse the text content
	err := self.p.WithText(name, nil, func(s string) error {
		custom.Value = s
		return nil
	})
	if err != nil {
		self.err = err
		return extensions, false
	}

	// Initialize extensions map if needed
	const customKey = "_custom"
	if extensions == nil {
		extensions = ext.Extensions{customKey: {self.p.Name: {custom}}}
	} else if m, ok := extensions[customKey]; !ok {
		extensions[customKey] = map[string][]ext.Extension{self.p.Name: {custom}}
	} else {
		m[self.p.Name] = append(m[self.p.Name], custom)
	}
	return extensions, true
}
