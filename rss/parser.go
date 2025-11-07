package rss

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"strings"
	"time"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/dublincore"
	"github.com/dsh2dsh/gofeed/v2/internal/itunes"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
	"github.com/dsh2dsh/gofeed/v2/options"
)

const (
	dcNS     = "dc"
	itunesNS = "itunes"
)

var emptyAttrs = map[string]string{}

// Parser is a RSS Parser
type Parser struct {
	p   *xml.Parser
	err error
}

// NewParser creates a new RSS parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an xml feed into an rss.Feed
func (self *Parser) Parse(r io.Reader, opts ...options.Option) (*Feed, error) {
	self.p = xml.NewParser(
		xpp.NewXMLPullParser(r, false, shared.NewReaderLabel))

	if _, err := self.p.FindRoot(); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	feed := self.root(self.p.Name)
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
		return fmt.Errorf("gofeed/rss: xml parser errored: %w", self.p.Err())
	}
	return nil
}

func (self *Parser) root(name string) (channel *Feed) {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	// Items found in feed root
	var ti *TextInput
	var image *Image
	items := []*Item{}
	ver := self.parseVersion(name)

	for name := range children {
		// Skip any extensions found in the feed root.
		if shared.IsExtension(self.p.XMLPullParser) {
			self.p.Skip(name)
			continue
		}

		switch name {
		case "channel":
			channel = self.channel(name)
		case "item":
			items = self.appendItem(name, items)
		case "textinput":
			ti = self.textInput(name)
		case "image":
			image = self.image(name)
		default:
			self.p.Skip(name)
		}
	}
	if self.err != nil {
		return nil
	}

	if channel == nil {
		channel = &Feed{Items: items}
	} else if n := len(items); n != 0 {
		channel.Items = append(channel.Items, items...)
	}

	if ti != nil {
		channel.TextInput = ti
	}

	if image != nil {
		channel.Image = image
	}

	channel.Version = ver
	return channel
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

func (self *Parser) channel(name string) *Feed {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	rss := &Feed{Items: []*Item{}}
	for name := range children {
		self.channelBody(name, rss)
	}

	if self.err != nil {
		return nil
	}
	return rss
}

func (self *Parser) channelBody(name string, rss *Feed) {
	if self.parseChannelExt(rss) {
		return
	}

	switch name {
	case "title":
		rss.Title = self.p.Text()
	case "description":
		rss.Description = self.p.Text()
	case "link":
		rss.Links = self.appendLink(name, rss.Links)
	case "language":
		rss.Language = self.p.Text()
	case "copyright":
		rss.Copyright = self.p.Text()
	case "managingeditor":
		rss.ManagingEditor = self.p.Text()
	case "webmaster":
		rss.WebMaster = self.p.Text()
	case "pubdate":
		rss.PubDate, rss.PubDateParsed = self.parseDate(name)
	case "lastbuilddate":
		rss.LastBuildDate, rss.LastBuildDateParsed = self.parseDate(name)
	case "generator":
		rss.Generator = self.p.Text()
	case "docs":
		rss.Docs = self.p.Text()
	case "ttl":
		rss.TTL = self.p.Text()
	case "rating":
		rss.Rating = self.p.Text()
	case "skiphours":
		rss.SkipHours = self.appendSkip(name, "hour", rss.SkipHours)
	case "skipdays":
		rss.SkipDays = self.appendSkip(name, "day", rss.SkipDays)
	case "item":
		rss.Items = self.appendItem(name, rss.Items)
	case "cloud":
		rss.Cloud = self.cloud(name)
	case "category":
		rss.Categories = self.appendCategory(name, rss.Categories)
	case "image":
		rss.Image = self.image(name)
	case "textinput":
		rss.TextInput = self.textInput(name)
	case "items":
		// Skip RDF items element - it's a structural element
		// that contains item references, not actual content
		self.p.Skip(name)
	default:
		// For non-standard RSS channel elements, add them to extensions
		// under a special "_custom" namespace prefix
		if e, ok := self.parseCustomExtInto(name, rss.Extensions); ok {
			rss.Extensions = e
		}
	}
}

func (self *Parser) appendItem(name string, items []*Item) []*Item {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return items
	}

	item := new(Item)
	for name := range children {
		self.itemBody(name, item)
	}

	if self.err != nil {
		return items
	}
	return append(items, item)
}

func (self *Parser) itemBody(name string, item *Item) {
	if self.parseItemExt(item) {
		return
	}

	switch name {
	case "title":
		item.Title = self.p.Text()
	case "description":
		item.Description = self.p.Text()
	case "encoded":
		prefix := shared.PrefixForNamespace(self.p.Space, self.p.XMLPullParser)
		if prefix == "content" {
			item.Content = self.p.Text()
		}
	case "link":
		item.Links = self.appendLink(name, item.Links)
	case "author":
		item.Author = self.p.Text()
	case "comments":
		item.Comments = self.p.Text()
	case "pubdate":
		item.PubDate, item.PubDateParsed = self.parseDate(name)
	case "source":
		item.Source = self.source(name)
	case "enclosure":
		item.Enclosure = self.enclosure(name)
	case "guid":
		item.GUID = self.guid(name)
	case "category":
		item.Categories = self.appendCategory(name, item.Categories)
	default:
		// For non-standard RSS elements, add them to extensions
		// under a special "_custom" namespace prefix
		if e, ok := self.parseCustomExtInto(name, item.Extensions); ok {
			item.Extensions = e
		}
	}
}

func (self *Parser) appendLink(name string, links []string) []string {
	var url string
	err := self.p.WithText(name,
		func() error {
			url = self.p.Attribute("href")
			return nil
		},
		func(s string) error {
			if s != "" {
				url = s
			}
			return nil
		})
	if err != nil {
		self.err = err
		return links
	}
	return append(links, url)
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

func (self *Parser) source(name string) (source *Source) {
	err := self.p.WithText(name,
		func() error {
			source = &Source{URL: self.p.Attribute("url")}
			return nil
		},
		func(s string) error {
			source.Title = s
			return nil
		})
	if err != nil {
		self.err = err
		return nil
	}
	return source
}

func (self *Parser) enclosure(name string) *Enclosure {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	enclosure := self.makeEnclosure()
	for name := range children {
		// Ignore any enclosure tag
		self.p.Skip(name)
	}

	if self.err != nil {
		return nil
	}
	return enclosure
}

func (self *Parser) makeEnclosure() *Enclosure {
	enclosure := new(Enclosure)
	for name, value := range self.p.AttributeSeq() {
		switch name {
		case "url":
			enclosure.URL = value
		case "length":
			enclosure.Length = value
		case "type":
			enclosure.Type = value
		}
	}
	return enclosure
}

func (self *Parser) image(name string) *Image {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	image := new(Image)
	for name := range children {
		self.imageBody(name, image)
	}

	if self.err != nil {
		return nil
	}
	return image
}

func (self *Parser) imageBody(name string, image *Image) {
	switch name {
	case "url":
		image.URL = self.p.Text()
	case "title":
		image.Title = self.p.Text()
	case "link":
		image.Link = self.p.Text()
	case "width":
		image.Width = self.p.Text()
	case "height":
		image.Height = self.p.Text()
	case "description":
		image.Description = self.p.Text()
	default:
		self.p.Skip(name)
	}
}

func (self *Parser) guid(name string) (guid *GUID) {
	err := self.p.WithText(name,
		func() error {
			guid = &GUID{IsPermalink: self.p.Attribute("isPermalink")}
			return nil
		},
		func(s string) error {
			guid.Value = s
			return nil
		})
	if err != nil {
		self.err = err
		return nil
	}
	return guid
}

func (self *Parser) appendCategory(name string, categories []*Category,
) []*Category {
	var c *Category
	err := self.p.WithText(name,
		func() error {
			c = &Category{Domain: self.p.Attribute("domain")}
			return nil
		},
		func(s string) error {
			c.Value = s
			return nil
		})
	if err != nil {
		self.err = err
		return categories
	}
	return append(categories, c)
}

func (self *Parser) textInput(name string) (ti *TextInput) {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	ti = new(TextInput)
	for name := range children {
		self.textInputBody(name, ti)
	}

	if self.err != nil {
		return nil
	}
	return ti
}

func (self *Parser) textInputBody(name string, ti *TextInput) {
	switch name {
	case "title":
		ti.Title = self.p.Text()
	case "description":
		ti.Description = self.p.Text()
	case "name":
		ti.Name = self.p.Text()
	case "link":
		ti.Link = self.p.Text()
	default:
		self.p.Skip(name)
	}
}

func (self *Parser) appendSkip(name, unit string, values []string) []string {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return values
	}

	for name := range children {
		switch name {
		case unit:
			s := self.p.Text()
			if self.p.Err() == nil {
				values = append(values, s)
			}
		default:
			self.p.Skip(name)
		}
	}
	return values
}

func (self *Parser) cloud(name string) (cloud *Cloud) {
	err := self.p.WithSkip(name, func() error {
		cloud = self.makeCloud()
		return nil
	})
	if err != nil {
		self.err = err
		return nil
	}
	return cloud
}

func (self *Parser) makeCloud() *Cloud {
	cloud := new(Cloud)
	for name, value := range self.p.AttributeSeq() {
		switch name {
		case "domain":
			cloud.Domain = value
		case "port":
			cloud.Port = value
		case "path":
			cloud.Path = value
		case "registerprocedure":
			cloud.RegisterProcedure = value
		case "protocol":
			cloud.Protocol = value
		}
	}
	return cloud
}

func (self *Parser) parseVersion(name string) string {
	switch strings.ToLower(name) {
	case "rss":
		return self.p.Attribute("version")
	case "rdf":
		switch self.p.Attribute("xmlns") {
		case "http://channel.netscape.com/rdf/simple/0.9/",
			"http://my.netscape.com/rdf/simple/0.9/":
			return "0.9"
		case "http://purl.org/rss/1.0/":
			return "1.0"
		}
	}
	return ""
}

func (self *Parser) parseCustomExtInto(name string, extensions ext.Extensions,
) (ext.Extensions, bool) {
	custom := ext.Extension{Name: self.p.Name, Attrs: emptyAttrs}
	// Copy attributes
	if n := len(self.p.Attrs); n != 0 {
		custom.Attrs = make(map[string]string, n)
		maps.Insert(custom.Attrs, self.p.AttributeSeq())
	}

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

func (self *Parser) parseChannelExt(rss *Feed) bool {
	switch shared.PrefixForNamespace(self.p.Space, self.p.XMLPullParser) {
	case "", "rss", "rdf", "content":
		return false
	case dcNS:
		rss.DublinCoreExt = self.dublinCore(rss.DublinCoreExt)
	case itunesNS:
		rss.ITunesExt = self.itunesFeed(rss.ITunesExt)
	default:
		rss.Extensions = self.extensions(rss.Extensions)
	}
	return true
}

func (self *Parser) dublinCore(dc *ext.DublinCoreExtension,
) *ext.DublinCoreExtension {
	dc, err := dublincore.Parse(self.p, dc)
	if err != nil {
		self.err = err
	}
	return dc
}

func (self *Parser) itunesFeed(feed *ext.ITunesFeedExtension,
) *ext.ITunesFeedExtension {
	feed, err := itunes.ParseFeed(self.p, feed)
	if err != nil {
		self.err = err
	}
	return feed
}

func (self *Parser) extensions(e ext.Extensions) ext.Extensions {
	e, err := shared.ParseExtension(e, self.p.XMLPullParser)
	if err != nil {
		self.err = err
	}
	return e
}

func (self *Parser) parseItemExt(item *Item) bool {
	switch shared.PrefixForNamespace(self.p.Space, self.p.XMLPullParser) {
	case "", "rss", "rdf", "content":
		return false
	case dcNS:
		item.DublinCoreExt = self.dublinCore(item.DublinCoreExt)
	case itunesNS:
		item.ITunesExt = self.itunesItem(item.ITunesExt)
	default:
		item.Extensions = self.extensions(item.Extensions)
	}
	return true
}

func (self *Parser) itunesItem(item *ext.ITunesItemExtension,
) *ext.ITunesItemExtension {
	item, err := itunes.ParseItem(self.p, item)
	if err != nil {
		self.err = err
	}
	return item
}
