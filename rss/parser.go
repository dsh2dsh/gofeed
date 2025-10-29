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
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is a RSS Parser
type Parser struct {
	p   *xpp.XMLPullParser
	err error
}

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
			case "channel":
				rp.parseChannelTo(&channel)
			case "item":
				items = rp.parseItemTo(items)
			case "textinput":
				rp.parseTextInputTo(&textinput)
			case "image":
				rp.parseImageTo(&image)
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

	if n := len(items); n > 0 {
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

func (rp *Parser) parseChannelTo(ref **Feed) {
	channel, err := rp.parseChannel()
	if err != nil {
		rp.err = err
		return
	}
	*ref = channel
}

func (rp *Parser) parseChannel() (rss *Feed, err error) {
	if err = rp.expect(xpp.StartTag, "channel"); err != nil {
		return nil, err
	}

	rss = &Feed{
		Items: []*Item{},
	}
	extensions := ext.Extensions{}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if shared.IsExtension(rp.p) {
				ext, err := shared.ParseExtension(extensions, rp.p)
				if err != nil {
					return nil, err
				}
				extensions = ext
				continue
			}

			switch strings.ToLower(rp.p.Name) {
			case "title":
				rp.parseTextTo(&rss.Title)
			case "description":
				rp.parseTextTo(&rss.Description)
			case "link":
				rss.Links = rp.parseLinkTo(&rss.Link, rss.Links)
			case "language":
				rp.parseTextTo(&rss.Language)
			case "copyright":
				rp.parseTextTo(&rss.Copyright)
			case "managingeditor":
				rp.parseTextTo(&rss.ManagingEditor)
			case "webmaster":
				rp.parseTextTo(&rss.WebMaster)
			case "pubdate":
				rp.parseDateTo(&rss.PubDate, &rss.PubDateParsed)
			case "lastbuilddate":
				rp.parseDateTo(&rss.LastBuildDate, &rss.LastBuildDateParsed)
			case "generator":
				rp.parseTextTo(&rss.Generator)
			case "docs":
				rp.parseTextTo(&rss.Docs)
			case "ttl":
				rp.parseTextTo(&rss.TTL)
			case "rating":
				rp.parseTextTo(&rss.Rating)
			case "skiphours":
				rss.SkipHours = rp.parseSkipHoursTo(rss.SkipHours)
			case "skipdays":
				rss.SkipDays = rp.parseSkipDaysTo(rss.SkipDays)
			case "item":
				rss.Items = rp.parseItemTo(rss.Items)
			case "cloud":
				rp.parseCloudTo(&rss.Cloud)
			case "category":
				rss.Categories = rp.parseCategoryTo(rss.Categories)
			case "image":
				rp.parseImageTo(&rss.Image)
			case "textinput":
				rp.parseTextInputTo(&rss.TextInput)
			case "items":
				// Skip RDF items element - it's a structural element
				// that contains item references, not actual content
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			default:
				// For non-standard RSS channel elements, add them to extensions
				// under a special "_custom" namespace prefix
				extensitons2, ok := rp.parseCustomExtInto(extensions)
				if !ok {
					continue
				}
				extensions = extensitons2
			}
		}
	}

	if err = rp.expect(xpp.EndTag, "channel"); err != nil {
		return nil, err
	}

	if len(extensions) > 0 {
		rss.Extensions = extensions
		if itunes, ok := rss.Extensions["itunes"]; ok {
			rss.ITunesExt = ext.NewITunesFeedExtension(itunes)
		}
		if dc, ok := rss.Extensions["dc"]; ok {
			rss.DublinCoreExt = ext.NewDublinCoreExtension(dc)
		}
	}
	return rss, nil
}

func (rp *Parser) parseTextTo(ref *string) {
	result, err := shared.ParseText(rp.p)
	if err != nil {
		rp.err = err
		return
	}
	*ref = result
}

func (rp *Parser) parseItemTo(items []*Item) []*Item {
	item, err := rp.parseItem()
	if err != nil {
		rp.err = err
		return items
	}
	return append(items, item)
}

func (rp *Parser) parseItem() (item *Item, err error) {
	if err = rp.expect(xpp.StartTag, "item"); err != nil {
		return nil, err
	}

	item = &Item{}
	extensions := ext.Extensions{}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if shared.IsExtension(rp.p) {
				ext, err := shared.ParseExtension(extensions, rp.p)
				if err != nil {
					return nil, err
				}
				item.Extensions = ext
				continue
			}

			switch strings.ToLower(rp.p.Name) {
			case "title":
				rp.parseTextTo(&item.Title)
			case "description":
				rp.parseTextTo(&item.Description)
			case "encoded":
				space := strings.TrimSpace(rp.p.Space)
				prefix := shared.PrefixForNamespace(space, rp.p)
				if prefix == "content" {
					rp.parseTextTo(&item.Content)
				}
			case "link":
				item.Links = rp.parseLinkTo(&item.Link, item.Links)
			case "author":
				rp.parseTextTo(&item.Author)
			case "comments":
				rp.parseTextTo(&item.Comments)
			case "pubdate":
				rp.parseDateTo(&item.PubDate, &item.PubDateParsed)
			case "source":
				rp.parseSourceTo(&item.Source)
			case "enclosure":
				item.Enclosures = rp.parseEnclosureTo(&item.Enclosure, item.Enclosures)
			case "guid":
				rp.parseGUIDTo(&item.GUID)
			case "category":
				item.Categories = rp.parseCategoryTo(item.Categories)
			default:
				// For non-standard RSS elements, add them to extensions
				// under a special "_custom" namespace prefix
				extensitons2, ok := rp.parseCustomExtInto(extensions)
				if !ok {
					continue
				}
				extensions = extensitons2
			}
		}
	}

	if err = rp.expect(xpp.EndTag, "item"); err != nil {
		return nil, err
	}

	if len(extensions) > 0 {
		item.Extensions = extensions
		if itunes, ok := item.Extensions["itunes"]; ok {
			item.ITunesExt = ext.NewITunesItemExtension(itunes)
		}
		if dc, ok := item.Extensions["dc"]; ok {
			item.DublinCoreExt = ext.NewDublinCoreExtension(dc)
		}
	}
	return item, nil
}

func (rp *Parser) parseLinkTo(ref *string, links []string) []string {
	result, err := rp.parseLink()
	if err != nil {
		rp.err = err
		return links
	}
	*ref = result
	return append(links, result)
}

func (rp *Parser) parseLink() (string, error) {
	href := rp.p.Attribute("href")
	url, err := shared.ParseText(rp.p)
	if err != nil {
		return "", err
	}

	if url == "" && href != "" {
		return href, nil
	}
	return url, nil
}

func (rp *Parser) parseDateTo(strRef *string, dtRef **time.Time) {
	result, err := shared.ParseText(rp.p)
	if err != nil {
		rp.err = err
		return
	}
	*strRef = result

	date, err := shared.ParseDate(result)
	if err != nil {
		rp.err = err
		return
	}

	utcDate := date.UTC()
	*dtRef = &utcDate
}

func (rp *Parser) parseSourceTo(ref **Source) {
	result, err := rp.parseSource()
	if err != nil {
		rp.err = err
		return
	}
	*ref = result
}

func (rp *Parser) parseSource() (source *Source, err error) {
	if err = rp.expect(xpp.StartTag, "source"); err != nil {
		return nil, err
	}

	source = &Source{
		URL: rp.p.Attribute("url"),
	}

	result, err := shared.ParseText(rp.p)
	if err != nil {
		return nil, err
	}
	source.Title = result

	if err = rp.expect(xpp.EndTag, "source"); err != nil {
		return nil, err
	}
	return source, nil
}

func (rp *Parser) parseEnclosureTo(ref **Enclosure, enclosures []*Enclosure,
) []*Enclosure {
	result, err := rp.parseEnclosure()
	if err != nil {
		rp.err = err
		return enclosures
	}
	*ref = result
	return append(enclosures, result)
}

func (rp *Parser) parseEnclosure() (enclosure *Enclosure, err error) {
	if err = rp.expect(xpp.StartTag, "enclosure"); err != nil {
		return nil, err
	}

	enclosure = &Enclosure{
		URL:    rp.p.Attribute("url"),
		Length: rp.p.Attribute("length"),
		Type:   rp.p.Attribute("type"),
	}

	// Ignore any enclosure tag
	for {
		_, err := rp.p.Next()
		if err != nil {
			return nil, fmt.Errorf("gofeed/rss: %w", err)
		} else if rp.p.Event == xpp.EndTag && rp.p.Name == "enclosure" {
			break
		}
	}
	return enclosure, nil
}

func (rp *Parser) parseImageTo(ref **Image) {
	img, err := rp.parseImage()
	if err != nil {
		rp.err = err
		return
	}
	*ref = img
}

func (rp *Parser) parseImage() (image *Image, err error) {
	if err = rp.expect(xpp.StartTag, "image"); err != nil {
		return nil, err
	}

	image = &Image{}

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
				rp.parseTextTo(&image.URL)
			case "title":
				rp.parseTextTo(&image.Title)
			case "link":
				rp.parseTextTo(&image.Link)
			case "width":
				rp.parseTextTo(&image.Width)
			case "height":
				rp.parseTextTo(&image.Height)
			case "description":
				rp.parseTextTo(&image.Description)
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err = rp.expect(xpp.EndTag, "image"); err != nil {
		return nil, err
	}
	return image, nil
}

func (rp *Parser) parseGUIDTo(ref **GUID) {
	result, err := rp.parseGUID()
	if err != nil {
		rp.err = err
		return
	}
	*ref = result
}

func (rp *Parser) parseGUID() (guid *GUID, err error) {
	if err = rp.expect(xpp.StartTag, "guid"); err != nil {
		return nil, err
	}

	guid = &GUID{IsPermalink: rp.p.Attribute("isPermalink")}

	result, err := shared.ParseText(rp.p)
	if err != nil {
		return nil, err
	}
	guid.Value = result

	if err = rp.expect(xpp.EndTag, "guid"); err != nil {
		return nil, err
	}
	return guid, nil
}

func (rp *Parser) parseCategoryTo(cats []*Category) []*Category {
	cat, err := rp.parseCategory()
	if err != nil {
		rp.err = err
		return cats
	}
	return append(cats, cat)
}

func (rp *Parser) parseCategory() (cat *Category, err error) {
	if err = rp.expect(xpp.StartTag, "category"); err != nil {
		return nil, err
	}

	cat = &Category{Domain: rp.p.Attribute("domain")}

	result, err := shared.ParseText(rp.p)
	if err != nil {
		return nil, err
	}
	cat.Value = result

	if err = rp.expect(xpp.EndTag, "category"); err != nil {
		return nil, err
	}
	return cat, nil
}

func (rp *Parser) parseTextInputTo(ref **TextInput) {
	ti, err := rp.parseTextInput()
	if err != nil {
		rp.err = err
		return
	}
	*ref = ti
}

func (rp *Parser) parseTextInput() (*TextInput, error) {
	if err := rp.expect(xpp.StartTag, "textinput"); err != nil {
		return nil, err
	}

	ti := &TextInput{}

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
				rp.parseTextTo(&ti.Title)
			case "description":
				rp.parseTextTo(&ti.Description)
			case "name":
				rp.parseTextTo(&ti.Name)
			case "link":
				rp.parseTextTo(&ti.Link)
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := rp.expect(xpp.EndTag, "textinput"); err != nil {
		return nil, err
	}
	return ti, nil
}

func (rp *Parser) parseSkipHoursTo(hours []string) []string {
	skipHours, err := rp.parseSkipHours()
	if err != nil {
		rp.err = err
		return hours
	}
	return append(slices.Grow(hours, len(skipHours)), skipHours...)
}

func (rp *Parser) parseSkipHours() ([]string, error) {
	if err := rp.expect(xpp.StartTag, "skiphours"); err != nil {
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
		if name != "hour" {
			rp.p.Skip() //nolint:errcheck // upstream ignores err
			continue
		}

		result, err := shared.ParseText(rp.p)
		if err != nil {
			return nil, err
		}
		hours = append(hours, result)
	}

	if err := rp.expect(xpp.EndTag, "skiphours"); err != nil {
		return nil, err
	}
	return hours, nil
}

func (rp *Parser) parseSkipDaysTo(days []string) []string {
	skipDays, err := rp.parseSkipDays()
	if err != nil {
		rp.err = err
		return days
	}
	return append(slices.Grow(days, len(skipDays)), skipDays...)
}

func (rp *Parser) parseSkipDays() ([]string, error) {
	if err := rp.expect(xpp.StartTag, "skipdays"); err != nil {
		return nil, err
	}

	days := []string{}

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
		if name != "day" {
			rp.p.Skip() //nolint:errcheck // upstream ignores err
			continue
		}

		result, err := shared.ParseText(rp.p)
		if err != nil {
			return nil, err
		}
		days = append(days, result)
	}

	if err := rp.expect(xpp.EndTag, "skipdays"); err != nil {
		return nil, err
	}
	return days, nil
}

func (rp *Parser) parseCloudTo(ref **Cloud) {
	result, err := rp.parseCloud()
	if err != nil {
		rp.err = err
		return
	}
	*ref = result
}

func (rp *Parser) parseCloud() (*Cloud, error) {
	if err := rp.expect(xpp.StartTag, "cloud"); err != nil {
		return nil, err
	}

	cloud := &Cloud{
		Domain:            rp.p.Attribute("domain"),
		Port:              rp.p.Attribute("port"),
		Path:              rp.p.Attribute("path"),
		RegisterProcedure: rp.p.Attribute("registerProcedure"),
		Protocol:          rp.p.Attribute("protocol"),
	}

	rp.nextTag() //nolint:errcheck // upstream ignores err

	if err := rp.expect(xpp.EndTag, "cloud"); err != nil {
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
		Attrs: make(map[string]string),
	}

	// Copy attributes
	for _, attr := range rp.p.Attrs {
		custom.Attrs[attr.Name.Local] = attr.Value
	}

	// Parse the text content
	result, err := shared.ParseText(rp.p)
	if err != nil {
		rp.p.Skip() //nolint:errcheck // upstream ignores err
		return nil, false
	}
	custom.Value = result

	// Initialize extensions map if needed
	if extensions == nil {
		extensions = make(ext.Extensions)
	}
	if extensions["_custom"] == nil {
		extensions["_custom"] = make(map[string][]ext.Extension)
	}

	// Add to extensions
	extensions["_custom"][rp.p.Name] = append(
		extensions["_custom"][rp.p.Name], custom)
	return extensions, true
}
