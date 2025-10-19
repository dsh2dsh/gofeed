package rss

import (
	"errors"
	"fmt"
	"io"
	"strings"

	xpp "github.com/mmcdole/goxpp"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is a RSS Parser
type Parser struct {
	p *xpp.XMLPullParser
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
	rssErr := rp.expect(xpp.StartTag, "rss")
	rdfErr := rp.expect(xpp.StartTag, "rdf")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%w or %w", rssErr, rdfErr)
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

			name := strings.ToLower(rp.p.Name)

			switch name {
			case "channel":
				channel, err = rp.parseChannel()
				if err != nil {
					return nil, err
				}
			case "item":
				item, err := rp.parseItem()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			case "textinput":
				textinput, err = rp.parseTextInput()
				if err != nil {
					return nil, err
				}
			case "image":
				image, err = rp.parseImage()
				if err != nil {
					return nil, err
				}
			default:
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	rssErr = rp.expect(xpp.EndTag, "rss")
	rdfErr = rp.expect(xpp.EndTag, "rdf")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%w or %w", rssErr, rdfErr)
	}

	if channel == nil {
		channel = &Feed{}
		channel.Items = []*Item{}
	}

	if len(items) > 0 {
		channel.Items = append(channel.Items, items...)
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

// nextTag iterates through the tokens until it reaches a StartTag or EndTag.
//
// nextTag is similar to goxpp's NextTag method except it wont throw an error if
// the next immediate token isnt a Start/EndTag. Instead, it will continue to
// consume tokens until it hits a Start/EndTag or EndDocument.
func (rp *Parser) nextTag() (xpp.XMLEventType, error) {
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

func (rp *Parser) parseChannel() (rss *Feed, err error) {
	if err = rp.expect(xpp.StartTag, "channel"); err != nil {
		return nil, err
	}

	rss = &Feed{}
	rss.Items = []*Item{}

	extensions := ext.Extensions{}
	categories := []*Category{}
	links := []string{}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(rp.p.Name)

			switch {
			case shared.IsExtension(rp.p):
				ext, err := shared.ParseExtension(extensions, rp.p)
				if err != nil {
					return nil, err
				}
				extensions = ext
			case name == "title":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Title = result
			case name == "description":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Description = result
			case name == "link":
				result, err := rp.parseLink()
				if err != nil {
					return nil, err
				}
				rss.Link = result
				links = append(links, result)
			case name == "language":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Language = result
			case name == "copyright":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Copyright = result
			case name == "managingeditor":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.ManagingEditor = result
			case name == "webmaster":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.WebMaster = result
			case name == "pubdate":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.PubDate = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					rss.PubDateParsed = &utcDate
				}
			case name == "lastbuilddate":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.LastBuildDate = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					rss.LastBuildDateParsed = &utcDate
				}
			case name == "generator":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Generator = result
			case name == "docs":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Docs = result
			case name == "ttl":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.TTL = result
			case name == "rating":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				rss.Rating = result
			case name == "skiphours":
				result, err := rp.parseSkipHours()
				if err != nil {
					return nil, err
				}
				rss.SkipHours = result
			case name == "skipdays":
				result, err := rp.parseSkipDays()
				if err != nil {
					return nil, err
				}
				rss.SkipDays = result
			case name == "item":
				result, err := rp.parseItem()
				if err != nil {
					return nil, err
				}
				rss.Items = append(rss.Items, result)
			case name == "cloud":
				result, err := rp.parseCloud()
				if err != nil {
					return nil, err
				}
				rss.Cloud = result
			case name == "category":
				result, err := rp.parseCategory()
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			case name == "image":
				result, err := rp.parseImage()
				if err != nil {
					return nil, err
				}
				rss.Image = result
			case name == "textinput":
				result, err := rp.parseTextInput()
				if err != nil {
					return nil, err
				}
				rss.TextInput = result
			case name == "items":
				// Skip RDF items element - it's a structural element
				// that contains item references, not actual content
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			default:
				// For non-standard RSS channel elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  rp.p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range rp.p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(rp.p)
				if err != nil {
					rp.p.Skip() //nolint:errcheck // upstream ignores err
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
				extensions["_custom"][rp.p.Name] = append(
					extensions["_custom"][rp.p.Name], customExt)
			}
		}
	}

	if err = rp.expect(xpp.EndTag, "channel"); err != nil {
		return nil, err
	}

	if len(categories) > 0 {
		rss.Categories = categories
	}

	if len(links) > 0 {
		rss.Links = links
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

func (rp *Parser) parseItem() (item *Item, err error) {
	if err = rp.expect(xpp.StartTag, "item"); err != nil {
		return nil, err
	}

	item = &Item{}
	extensions := ext.Extensions{}
	categories := []*Category{}
	enclosures := []*Enclosure{}
	links := []string{}

	for {
		tok, err := rp.nextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(rp.p.Name)

			switch {
			case shared.IsExtension(rp.p):
				ext, err := shared.ParseExtension(extensions, rp.p)
				if err != nil {
					return nil, err
				}
				item.Extensions = ext
			case name == "title":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				item.Title = result
			case name == "description":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				item.Description = result
			case name == "encoded":
				space := strings.TrimSpace(rp.p.Space)
				prefix := shared.PrefixForNamespace(space, rp.p)
				if prefix == "content" {
					result, err := shared.ParseText(rp.p)
					if err != nil {
						return nil, err
					}
					item.Content = result
				}
			case name == "link":
				result, err := rp.parseLink()
				if err != nil {
					return nil, err
				}
				item.Link = result
				links = append(links, result)
			case name == "author":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				item.Author = result
			case name == "comments":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				item.Comments = result
			case name == "pubdate":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				item.PubDate = result
				date, err := shared.ParseDate(result)
				if err == nil {
					utcDate := date.UTC()
					item.PubDateParsed = &utcDate
				}
			case name == "source":
				result, err := rp.parseSource()
				if err != nil {
					return nil, err
				}
				item.Source = result
			case name == "enclosure":
				result, err := rp.parseEnclosure()
				if err != nil {
					return nil, err
				}
				item.Enclosure = result
				enclosures = append(enclosures, result)
			case name == "guid":
				result, err := rp.parseGUID()
				if err != nil {
					return nil, err
				}
				item.GUID = result
			case name == "category":
				result, err := rp.parseCategory()
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			default:
				// For non-standard RSS elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  rp.p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range rp.p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(rp.p)
				if err != nil {
					continue
				}
				customExt.Value = result

				if extensions["_custom"] == nil {
					extensions["_custom"] = make(map[string][]ext.Extension)
				}

				// Add to extensions
				extensions["_custom"][rp.p.Name] = append(
					extensions["_custom"][rp.p.Name], customExt)
			}
		}
	}

	if len(enclosures) > 0 {
		item.Enclosures = enclosures
	}

	if len(categories) > 0 {
		item.Categories = categories
	}

	if len(links) > 0 {
		item.Links = links
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

	if err = rp.expect(xpp.EndTag, "item"); err != nil {
		return nil, err
	}
	return item, nil
}

func (rp *Parser) parseLink() (url string, err error) {
	href := rp.p.Attribute("href")
	url, err = shared.ParseText(rp.p)
	if err != nil {
		return "", err
	}
	if url == "" && href != "" {
		url = href
	}
	return url, err
}

func (rp *Parser) parseSource() (source *Source, err error) {
	if err = rp.expect(xpp.StartTag, "source"); err != nil {
		return nil, err
	}

	source = &Source{}
	source.URL = rp.p.Attribute("url")

	result, err := shared.ParseText(rp.p)
	if err != nil {
		return source, err
	}
	source.Title = result

	if err = rp.expect(xpp.EndTag, "source"); err != nil {
		return nil, err
	}
	return source, nil
}

func (rp *Parser) parseEnclosure() (enclosure *Enclosure, err error) {
	if err = rp.expect(xpp.StartTag, "enclosure"); err != nil {
		return nil, err
	}

	enclosure = &Enclosure{}
	enclosure.URL = rp.p.Attribute("url")
	enclosure.Length = rp.p.Attribute("length")
	enclosure.Type = rp.p.Attribute("type")

	// Ignore any enclosure tag
	for {
		_, err := rp.p.Next()
		if err != nil {
			return enclosure, fmt.Errorf("gofeed/rss: %w", err)
		}

		if rp.p.Event == xpp.EndTag && rp.p.Name == "enclosure" {
			break
		}
	}
	return enclosure, nil
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
			name := strings.ToLower(rp.p.Name)

			switch name {
			case "url":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.URL = result
			case "title":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.Title = result
			case "link":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.Link = result
			case "width":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.Width = result
			case "height":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.Height = result
			case "description":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				image.Description = result
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

func (rp *Parser) parseGUID() (guid *GUID, err error) {
	if err = rp.expect(xpp.StartTag, "guid"); err != nil {
		return nil, err
	}

	guid = &GUID{}
	guid.IsPermalink = rp.p.Attribute("isPermalink")

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

func (rp *Parser) parseCategory() (cat *Category, err error) {
	if err = rp.expect(xpp.StartTag, "category"); err != nil {
		return nil, err
	}

	cat = &Category{}
	cat.Domain = rp.p.Attribute("domain")

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
			name := strings.ToLower(rp.p.Name)

			switch name {
			case "title":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				ti.Title = result
			case "description":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				ti.Description = result
			case "name":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				ti.Name = result
			case "link":
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				ti.Link = result
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
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(rp.p.Name)
			if name == "hour" {
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				hours = append(hours, result)
			} else {
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := rp.expect(xpp.EndTag, "skiphours"); err != nil {
		return nil, err
	}
	return hours, nil
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
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(rp.p.Name)
			if name == "day" {
				result, err := shared.ParseText(rp.p)
				if err != nil {
					return nil, err
				}
				days = append(days, result)
			} else {
				rp.p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := rp.expect(xpp.EndTag, "skipdays"); err != nil {
		return nil, err
	}
	return days, nil
}

func (rp *Parser) parseCloud() (*Cloud, error) {
	if err := rp.expect(xpp.StartTag, "cloud"); err != nil {
		return nil, err
	}

	cloud := &Cloud{}
	cloud.Domain = rp.p.Attribute("domain")
	cloud.Port = rp.p.Attribute("port")
	cloud.Path = rp.p.Attribute("path")
	cloud.RegisterProcedure = rp.p.Attribute("registerProcedure")
	cloud.Protocol = rp.p.Attribute("protocol")

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
