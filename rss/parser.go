package rss

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/dublincore"
	"github.com/dsh2dsh/gofeed/v2/internal/itunes"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is a RSS Parser
type Parser struct {
	p   *xpp.XMLPullParser
	err error
}

const (
	categoryTag  = "category"
	channelTag   = "channel"
	cloudTag     = "cloud"
	customKey    = "_custom"
	dcKey        = "dc"
	enclosureTag = "enclosure"
	guidTag      = "guid"
	imageTag     = "image"
	itemTag      = "item"
	itunesKey    = "itunes"
	skipDaysTag  = "skipdays"
	skipHoursTag = "skiphours"
	sourceTag    = "source"
	textInputTag = "textinput"
)

var emptyAttrs = map[string]string{}

// NewParser creates a new RSS parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an xml feed into an rss.Feed
func (rp *Parser) Parse(feed io.Reader, opts ...options.Option) (*Feed, error) {
	rp.p = xpp.NewXMLPullParser(feed, false, shared.NewReaderLabel)
	_, err := shared.FindRoot(rp.p)
	if err != nil {
		return nil, err
	}
	return rp.parseRoot()
}

func (rp *Parser) parseRoot() (*Feed, error) {
	if err := rp.expectRssRdf(xpp.StartTag); err != nil {
		return nil, err
	}

	// Items found in feed root
	var channel *Feed
	var textinput *TextInput
	var image *Image
	items := []*Item{}
	ver := rp.parseVersion()

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			// Skip any extensions found in the feed root.
			if shared.IsExtension(rp.p) {
				rp.p.Skip() //nolint:errcheck // upstream ignores err
				continue
			}

			switch strings.ToLower(rp.p.Name) {
			case channelTag:
				channel = rp.channel()
			case itemTag:
				items = rp.appendItem(items)
			case textInputTag:
				textinput = rp.textInput()
			case imageTag:
				image = rp.image()
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := rp.expectRssRdf(xpp.EndTag); err != nil {
		return nil, err
	}

	if channel == nil {
		channel = &Feed{Items: make([]*Item, 0, len(items))}
	}

	if n := len(items); n != 0 {
		channel.Items = append(slices.Grow(channel.Items, n), items...)
	}

	if textinput != nil {
		channel.TextInput = textinput
	}

	if image != nil {
		channel.Image = image
	}

	channel.Version = ver
	return channel, nil
}

func (rp *Parser) expect(event xpp.XMLEventType, name string) error {
	if err := rp.p.Expect(event, name); err != nil {
		return fmt.Errorf("gofeed/rss: %w", err)
	}
	return nil
}

func (rp *Parser) expectRssRdf(event xpp.XMLEventType) error {
	rssErr := rp.expect(event, "rss")
	rdfErr := rp.expect(event, "rdf")
	if rssErr != nil && rdfErr != nil {
		return fmt.Errorf("%w or %w", rssErr, rdfErr)
	}
	return nil
}

// nextTag iterates through the tokens until it reaches a StartTag or EndTag.
//
// nextTag is similar to goxpp's NextTag method except it wont throw an error if
// the next immediate token isnt a Start/EndTag. Instead, it will continue to
// consume tokens until it hits a Start/EndTag or EndDocument.
func (rp *Parser) nextTag() (xpp.XMLEventType, error) {
	if rp.err != nil {
		return 0, rp.err
	}

	for {
		event, err := rp.p.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/atom: %w", err)
		}

		switch event {
		case xpp.EndTag:
			return event, nil
		case xpp.StartTag:
			return event, nil
		case xpp.EndDocument:
			return event, errors.New(
				"gofeed/rss: failed to find NextTag before reaching the end of the document")
		}
	}
}

func (rp *Parser) channel() *Feed {
	channel, err := rp.parseChannel()
	if err != nil {
		rp.err = err
		return nil
	}
	return channel
}

func (rp *Parser) parseChannel() (rss *Feed, err error) {
	if err = rp.expect(xpp.StartTag, channelTag); err != nil {
		return nil, err
	}

	rss = &Feed{Items: []*Item{}}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if rp.parseChannelExt(rss) {
				continue
			}

			switch strings.ToLower(rp.p.Name) {
			case "title":
				rss.Title = rp.text()
			case "description":
				rss.Description = rp.text()
			case "link":
				rss.Links = rp.appendLink(rss.Links)
			case "language":
				rss.Language = rp.text()
			case "copyright":
				rss.Copyright = rp.text()
			case "managingeditor":
				rss.ManagingEditor = rp.text()
			case "webmaster":
				rss.WebMaster = rp.text()
			case "pubdate":
				rss.PubDate, rss.PubDateParsed = rp.parseDate()
			case "lastbuilddate":
				rss.LastBuildDate, rss.LastBuildDateParsed = rp.parseDate()
			case "generator":
				rss.Generator = rp.text()
			case "docs":
				rss.Docs = rp.text()
			case "ttl":
				rss.TTL = rp.text()
			case "rating":
				rss.Rating = rp.text()
			case skipHoursTag:
				rss.SkipHours = rp.appendSkipHours(rss.SkipHours)
			case skipDaysTag:
				rss.SkipDays = rp.appendSkipDays(rss.SkipDays)
			case itemTag:
				rss.Items = rp.appendItem(rss.Items)
			case cloudTag:
				rss.Cloud = rp.cloud()
			case categoryTag:
				rss.Categories = rp.appendCategory(rss.Categories)
			case imageTag:
				rss.Image = rp.image()
			case textInputTag:
				rss.TextInput = rp.textInput()
			case "items":
				// Skip RDF items element - it's a structural element
				// that contains item references, not actual content
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			default:
				// For non-standard RSS channel elements, add them to extensions
				// under a special "_custom" namespace prefix
				if e, ok := rp.parseCustomExtInto(rss.Extensions); ok {
					rss.Extensions = e
				}
			}
		}
	}

	if err = rp.expect(xpp.EndTag, channelTag); err != nil {
		return nil, err
	}
	return rss, nil
}

func (rp *Parser) text() string {
	result, err := shared.ParseText(rp.p)
	if err != nil {
		rp.err = err
		return ""
	}
	return result
}

func (rp *Parser) appendItem(items []*Item) []*Item {
	item, err := rp.parseItem()
	if err != nil {
		rp.err = err
		return items
	}
	return append(items, item)
}

func (rp *Parser) parseItem() (item *Item, err error) {
	if err = rp.expect(xpp.StartTag, itemTag); err != nil {
		return nil, err
	}

	item = new(Item)

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if rp.parseItemExt(item) {
				continue
			}

			switch strings.ToLower(rp.p.Name) {
			case "title":
				item.Title = rp.text()
			case "description":
				item.Description = rp.text()
			case "encoded":
				prefix := shared.PrefixForNamespace(rp.p.Space, rp.p)
				if prefix == "content" {
					item.Content = rp.text()
				}
			case "link":
				item.Links = rp.appendLink(item.Links)
			case "author":
				item.Author = rp.text()
			case "comments":
				item.Comments = rp.text()
			case "pubdate":
				item.PubDate, item.PubDateParsed = rp.parseDate()
			case sourceTag:
				item.Source = rp.source()
			case enclosureTag:
				item.Enclosure = rp.enclosure()
			case guidTag:
				item.GUID = rp.guid()
			case categoryTag:
				item.Categories = rp.appendCategory(item.Categories)
			default:
				// For non-standard RSS elements, add them to extensions
				// under a special "_custom" namespace prefix
				if e, ok := rp.parseCustomExtInto(item.Extensions); ok {
					item.Extensions = e
				}
			}
		}
	}

	if err = rp.expect(xpp.EndTag, itemTag); err != nil {
		return nil, err
	}
	return item, nil
}

func (rp *Parser) appendLink(links []string) []string {
	result := rp.parseLink()
	if rp.err != nil {
		return links
	}
	return append(links, result)
}

func (rp *Parser) parseLink() string {
	href := rp.p.Attribute("href")
	url := rp.text()
	if rp.err != nil {
		return ""
	}

	if url == "" && href != "" {
		url = href
	}
	return url
}

func (rp *Parser) parseDate() (string, *time.Time) {
	result := rp.text()
	if rp.err != nil {
		return "", nil
	}

	date, err := shared.ParseDate(result)
	if err != nil {
		return result, nil
	}

	utcDate := date.UTC()
	return result, &utcDate
}

func (rp *Parser) source() *Source {
	result, err := rp.parseSource()
	if err != nil {
		rp.err = err
		return nil
	}
	return result
}

func (rp *Parser) parseSource() (source *Source, err error) {
	if err = rp.expect(xpp.StartTag, sourceTag); err != nil {
		return nil, err
	}

	source = &Source{URL: rp.p.Attribute("url")}

	result := rp.text()
	if rp.err != nil {
		return nil, rp.err
	}
	source.Title = result

	if err = rp.expect(xpp.EndTag, sourceTag); err != nil {
		return nil, err
	}
	return source, nil
}

func (rp *Parser) enclosure() *Enclosure {
	result, err := rp.parseEnclosure()
	if err != nil {
		rp.err = err
		return nil
	}
	return result
}

func (rp *Parser) parseEnclosure() (enclosure *Enclosure, err error) {
	if err = rp.expect(xpp.StartTag, enclosureTag); err != nil {
		return nil, err
	}

	enclosure = new(Enclosure)
	for _, attr := range rp.p.Attrs {
		switch v := attr.Name.Local; v {
		case "url":
			enclosure.URL = attr.Value
		case "length":
			enclosure.Length = attr.Value
		case "type":
			enclosure.Type = attr.Value
		}
	}

	// Ignore any enclosure tag
	for {
		_, err := rp.p.Next()
		if err != nil {
			return nil, fmt.Errorf("gofeed/rss: %w", err)
		} else if rp.p.Event == xpp.EndTag && rp.p.Name == enclosureTag {
			break
		}
	}
	return enclosure, nil
}

func (rp *Parser) image() *Image {
	img, err := rp.parseImage()
	if err != nil {
		rp.err = err
		return nil
	}
	return img
}

func (rp *Parser) parseImage() (image *Image, err error) {
	if err = rp.expect(xpp.StartTag, imageTag); err != nil {
		return nil, err
	}

	image = new(Image)

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return image, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			switch strings.ToLower(rp.p.Name) {
			case "url":
				image.URL = rp.text()
			case "title":
				image.Title = rp.text()
			case "link":
				image.Link = rp.text()
			case "width":
				image.Width = rp.text()
			case "height":
				image.Height = rp.text()
			case "description":
				image.Description = rp.text()
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err = rp.expect(xpp.EndTag, imageTag); err != nil {
		return nil, err
	}
	return image, nil
}

func (rp *Parser) guid() *GUID {
	result, err := rp.parseGUID()
	if err != nil {
		rp.err = err
		return nil
	}
	return result
}

func (rp *Parser) parseGUID() (guid *GUID, err error) {
	if err = rp.expect(xpp.StartTag, guidTag); err != nil {
		return nil, err
	}

	guid = &GUID{IsPermalink: rp.p.Attribute("isPermalink")}

	result := rp.text()
	if rp.err != nil {
		return nil, rp.err
	}
	guid.Value = result

	if err = rp.expect(xpp.EndTag, guidTag); err != nil {
		return nil, err
	}
	return guid, nil
}

func (rp *Parser) appendCategory(cats []*Category) []*Category {
	cat, err := rp.parseCategory()
	if err != nil {
		rp.err = err
		return cats
	}
	return append(cats, cat)
}

func (rp *Parser) parseCategory() (cat *Category, err error) {
	if err = rp.expect(xpp.StartTag, categoryTag); err != nil {
		return nil, err
	}

	cat = &Category{Domain: rp.p.Attribute("domain")}

	result := rp.text()
	if rp.err != nil {
		return nil, rp.err
	}
	cat.Value = result

	if err = rp.expect(xpp.EndTag, categoryTag); err != nil {
		return nil, err
	}
	return cat, nil
}

func (rp *Parser) textInput() *TextInput {
	ti, err := rp.parseTextInput()
	if err != nil {
		rp.err = err
		return nil
	}
	return ti
}

func (rp *Parser) parseTextInput() (*TextInput, error) {
	if err := rp.expect(xpp.StartTag, textInputTag); err != nil {
		return nil, err
	}

	ti := new(TextInput)

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			switch strings.ToLower(rp.p.Name) {
			case "title":
				ti.Title = rp.text()
			case "description":
				ti.Description = rp.text()
			case "name":
				ti.Name = rp.text()
			case "link":
				ti.Link = rp.text()
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := rp.expect(xpp.EndTag, textInputTag); err != nil {
		return nil, err
	}
	return ti, nil
}

func (rp *Parser) appendSkipHours(hours []string) []string {
	skipHours, err := rp.parseSkipSomething(skipHoursTag, "hour")
	if err != nil {
		rp.err = err
		return hours
	}
	return append(slices.Grow(hours, len(skipHours)), skipHours...)
}

func (rp *Parser) parseSkipSomething(tag, unit string) ([]string, error) {
	if err := rp.expect(xpp.StartTag, tag); err != nil {
		return nil, err
	}

	hours := []string{}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		} else if tok != xpp.StartTag {
			continue
		}

		name := strings.ToLower(rp.p.Name)
		if name != unit {
			rp.p.Skip() //nolint:errcheck // upstream ignores err
			continue
		}

		result := rp.text()
		if rp.err != nil {
			return nil, rp.err
		}
		hours = append(hours, result)
	}

	if err := rp.expect(xpp.EndTag, tag); err != nil {
		return nil, err
	}
	return hours, nil
}

func (rp *Parser) appendSkipDays(days []string) []string {
	skipDays, err := rp.parseSkipSomething(skipDaysTag, "day")
	if err != nil {
		rp.err = err
		return days
	}
	return append(slices.Grow(days, len(skipDays)), skipDays...)
}

func (rp *Parser) cloud() *Cloud {
	result, err := rp.parseCloud()
	if err != nil {
		rp.err = err
		return nil
	}
	return result
}

func (rp *Parser) parseCloud() (*Cloud, error) {
	if err := rp.expect(xpp.StartTag, cloudTag); err != nil {
		return nil, err
	}

	cloud := new(Cloud)
	for _, attr := range rp.p.Attrs {
		switch v := attr.Name.Local; v {
		case "domain":
			cloud.Domain = attr.Value
		case "port":
			cloud.Port = attr.Value
		case "path":
			cloud.Path = attr.Value
		case "registerProcedure":
			cloud.RegisterProcedure = attr.Value
		case "protocol":
			cloud.Protocol = attr.Value
		}
	}
	rp.nextTag() //nolint:errcheck // upstream ignores err

	if err := rp.expect(xpp.EndTag, cloudTag); err != nil {
		return nil, err
	}
	return cloud, nil
}

func (rp *Parser) parseVersion() string {
	name := strings.ToLower(rp.p.Name)
	switch name {
	case "rss":
		return rp.p.Attribute("version")
	case "rdf":
		switch rp.p.Attribute("xmlns") {
		case "http://channel.netscape.com/rdf/simple/0.9/",
			"http://my.netscape.com/rdf/simple/0.9/":
			return "0.9"
		case "http://purl.org/rss/1.0/":
			return "1.0"
		}
	}
	return ""
}

func (rp *Parser) parseCustomExtInto(extensions ext.Extensions) (ext.Extensions,
	bool,
) {
	custom := ext.Extension{
		Name:  rp.p.Name,
		Attrs: emptyAttrs,
	}

	// Copy attributes
	if n := len(rp.p.Attrs); n != 0 {
		custom.Attrs = make(map[string]string, n)
		for _, attr := range rp.p.Attrs {
			custom.Attrs[attr.Name.Local] = attr.Value
		}
	}

	// Parse the text content
	result := rp.text()
	if rp.err != nil {
		return nil, false
	}
	custom.Value = result

	// Initialize extensions map if needed
	if extensions == nil {
		extensions = ext.Extensions{customKey: {rp.p.Name: {custom}}}
	} else if m, ok := extensions[customKey]; !ok {
		extensions[customKey] = map[string][]ext.Extension{rp.p.Name: {custom}}
	} else {
		m[rp.p.Name] = append(m[rp.p.Name], custom)
	}
	return extensions, true
}

func (rp *Parser) parseChannelExt(rss *Feed) bool {
	switch shared.PrefixForNamespace(rp.p.Space, rp.p) {
	case "", "rss", "rdf", "content":
		return false
	case dcKey:
		rss.DublinCoreExt = rp.dublinCore(rss.DublinCoreExt)
	case itunesKey:
		rss.ITunesExt = rp.itunesFeed(rss.ITunesExt)
	default:
		rss.Extensions = rp.extensions(rss.Extensions)
	}
	return true
}

func (rp *Parser) dublinCore(dc *ext.DublinCoreExtension,
) *ext.DublinCoreExtension {
	dc, err := dublincore.Parse(rp.p, dc)
	if err != nil {
		rp.err = err
	}
	return dc
}

func (rp *Parser) itunesFeed(feed *ext.ITunesFeedExtension,
) *ext.ITunesFeedExtension {
	feed, err := itunes.ParseFeed(rp.p, feed)
	if err != nil {
		rp.err = err
	}
	return feed
}

func (rp *Parser) extensions(e ext.Extensions) ext.Extensions {
	e, err := shared.ParseExtension(e, rp.p)
	if err != nil {
		rp.err = err
	}
	return e
}

func (rp *Parser) parseItemExt(item *Item) bool {
	switch shared.PrefixForNamespace(rp.p.Space, rp.p) {
	case "", "rss", "rdf", "content":
		return false
	case dcKey:
		item.DublinCoreExt = rp.dublinCore(item.DublinCoreExt)
	case itunesKey:
		item.ITunesExt = rp.itunesItem(item.ITunesExt)
	default:
		item.Extensions = rp.extensions(item.Extensions)
	}
	return true
}

func (rp *Parser) itunesItem(item *ext.ITunesItemExtension,
) *ext.ITunesItemExtension {
	item, err := itunes.ParseItem(rp.p, item)
	if err != nil {
		rp.err = err
	}
	return item
}
