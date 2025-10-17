package rss

import (
	"fmt"
	"io"
	"strings"

	xpp "github.com/mmcdole/goxpp"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is a RSS Parser
type Parser struct{}

// NewParser creates a new RSS parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an xml feed into an rss.Feed
func (rp *Parser) Parse(feed io.Reader, opts ...options.Option) (*Feed, error) {
	p := xpp.NewXMLPullParser(feed, false, shared.NewReaderLabel)
	_, err := shared.FindRoot(p)
	if err != nil {
		return nil, err
	}
	return rp.parseRoot(p)
}

func (rp *Parser) parseRoot(p *xpp.XMLPullParser) (*Feed, error) {
	rssErr := p.Expect(xpp.StartTag, "rss")
	rdfErr := p.Expect(xpp.StartTag, "rdf")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%s or %s", rssErr.Error(), rdfErr.Error())
	}

	// Items found in feed root
	var channel *Feed
	var textinput *TextInput
	var image *Image
	items := []*Item{}

	ver := rp.parseVersion(p)

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			// Skip any extensions found in the feed root.
			if shared.IsExtension(p) {
				p.Skip() //nolint:errcheck // upstream ignores err
				continue
			}

			name := strings.ToLower(p.Name)

			switch name {
			case "channel":
				channel, err = rp.parseChannel(p)
				if err != nil {
					return nil, err
				}
			case "item":
				item, err := rp.parseItem(p)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			case "textinput":
				textinput, err = rp.parseTextInput(p)
				if err != nil {
					return nil, err
				}
			case "image":
				image, err = rp.parseImage(p)
				if err != nil {
					return nil, err
				}
			default:
				p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	rssErr = p.Expect(xpp.EndTag, "rss")
	rdfErr = p.Expect(xpp.EndTag, "rdf")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%s or %s", rssErr.Error(), rdfErr.Error())
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

func (rp *Parser) parseChannel(p *xpp.XMLPullParser) (rss *Feed, err error) {
	if err = p.Expect(xpp.StartTag, "channel"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	rss = &Feed{}
	rss.Items = []*Item{}

	extensions := ext.Extensions{}
	categories := []*Category{}
	links := []string{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(p.Name)

			switch {
			case shared.IsExtension(p):
				ext, err := shared.ParseExtension(extensions, p)
				if err != nil {
					return nil, err
				}
				extensions = ext
			case name == "title":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Title = result
			case name == "description":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Description = result
			case name == "link":
				result, err := rp.parseLink(p)
				if err != nil {
					return nil, err
				}
				rss.Link = result
				links = append(links, result)
			case name == "language":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Language = result
			case name == "copyright":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Copyright = result
			case name == "managingeditor":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.ManagingEditor = result
			case name == "webmaster":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.WebMaster = result
			case name == "pubdate":
				result, err := shared.ParseText(p)
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
				result, err := shared.ParseText(p)
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
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Generator = result
			case name == "docs":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Docs = result
			case name == "ttl":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.TTL = result
			case name == "rating":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				rss.Rating = result
			case name == "skiphours":
				result, err := rp.parseSkipHours(p)
				if err != nil {
					return nil, err
				}
				rss.SkipHours = result
			case name == "skipdays":
				result, err := rp.parseSkipDays(p)
				if err != nil {
					return nil, err
				}
				rss.SkipDays = result
			case name == "item":
				result, err := rp.parseItem(p)
				if err != nil {
					return nil, err
				}
				rss.Items = append(rss.Items, result)
			case name == "cloud":
				result, err := rp.parseCloud(p)
				if err != nil {
					return nil, err
				}
				rss.Cloud = result
			case name == "category":
				result, err := rp.parseCategory(p)
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			case name == "image":
				result, err := rp.parseImage(p)
				if err != nil {
					return nil, err
				}
				rss.Image = result
			case name == "textinput":
				result, err := rp.parseTextInput(p)
				if err != nil {
					return nil, err
				}
				rss.TextInput = result
			case name == "items":
				// Skip RDF items element - it's a structural element
				// that contains item references, not actual content
				p.Skip() //nolint:errcheck // upstream ignores err
			default:
				// For non-standard RSS channel elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(p)
				if err != nil {
					p.Skip() //nolint:errcheck // upstream ignores err
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
				extensions["_custom"][p.Name] = append(extensions["_custom"][p.Name], customExt)
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "channel"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
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

func (rp *Parser) parseItem(p *xpp.XMLPullParser) (item *Item, err error) {
	if err = p.Expect(xpp.StartTag, "item"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	item = &Item{}
	extensions := ext.Extensions{}
	categories := []*Category{}
	enclosures := []*Enclosure{}
	links := []string{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(p.Name)

			switch {
			case shared.IsExtension(p):
				ext, err := shared.ParseExtension(extensions, p)
				if err != nil {
					return nil, err
				}
				item.Extensions = ext
			case name == "title":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				item.Title = result
			case name == "description":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				item.Description = result
			case name == "encoded":
				space := strings.TrimSpace(p.Space)
				prefix := shared.PrefixForNamespace(space, p)
				if prefix == "content" {
					result, err := shared.ParseText(p)
					if err != nil {
						return nil, err
					}
					item.Content = result
				}
			case name == "link":
				result, err := rp.parseLink(p)
				if err != nil {
					return nil, err
				}
				item.Link = result
				links = append(links, result)
			case name == "author":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				item.Author = result
			case name == "comments":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				item.Comments = result
			case name == "pubdate":
				result, err := shared.ParseText(p)
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
				result, err := rp.parseSource(p)
				if err != nil {
					return nil, err
				}
				item.Source = result
			case name == "enclosure":
				result, err := rp.parseEnclosure(p)
				if err != nil {
					return nil, err
				}
				item.Enclosure = result
				enclosures = append(enclosures, result)
			case name == "guid":
				result, err := rp.parseGUID(p)
				if err != nil {
					return nil, err
				}
				item.GUID = result
			case name == "category":
				result, err := rp.parseCategory(p)
				if err != nil {
					return nil, err
				}
				categories = append(categories, result)
			default:
				// For non-standard RSS elements, add them to extensions
				// under a special "_custom" namespace prefix
				customExt := ext.Extension{
					Name:  p.Name,
					Attrs: make(map[string]string),
				}

				// Copy attributes
				for _, attr := range p.Attrs {
					customExt.Attrs[attr.Name.Local] = attr.Value
				}

				// Parse the text content
				result, err := shared.ParseText(p)
				if err != nil {
					continue
				}
				customExt.Value = result

				if extensions["_custom"] == nil {
					extensions["_custom"] = make(map[string][]ext.Extension)
				}

				// Add to extensions
				extensions["_custom"][p.Name] = append(extensions["_custom"][p.Name], customExt)
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

	if err = p.Expect(xpp.EndTag, "item"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return item, nil
}

func (rp *Parser) parseLink(p *xpp.XMLPullParser) (url string, err error) {
	href := p.Attribute("href")
	url, err = shared.ParseText(p)
	if err != nil {
		return "", err
	}
	if url == "" && href != "" {
		url = href
	}
	return url, err
}

func (rp *Parser) parseSource(p *xpp.XMLPullParser) (source *Source, err error) {
	if err = p.Expect(xpp.StartTag, "source"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	source = &Source{}
	source.URL = p.Attribute("url")

	result, err := shared.ParseText(p)
	if err != nil {
		return source, err
	}
	source.Title = result

	if err = p.Expect(xpp.EndTag, "source"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}
	return source, nil
}

func (rp *Parser) parseEnclosure(p *xpp.XMLPullParser) (enclosure *Enclosure, err error) {
	if err = p.Expect(xpp.StartTag, "enclosure"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	enclosure = &Enclosure{}
	enclosure.URL = p.Attribute("url")
	enclosure.Length = p.Attribute("length")
	enclosure.Type = p.Attribute("type")

	// Ignore any enclosure tag
	for {
		_, err := p.Next()
		if err != nil {
			return enclosure, fmt.Errorf("gofeed/rss: %w", err)
		}

		if p.Event == xpp.EndTag && p.Name == "enclosure" {
			break
		}
	}

	return enclosure, nil
}

func (rp *Parser) parseImage(p *xpp.XMLPullParser) (image *Image, err error) {
	if err = p.Expect(xpp.StartTag, "image"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	image = &Image{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return image, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)

			switch name {
			case "url":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.URL = result
			case "title":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Title = result
			case "link":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Link = result
			case "width":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Width = result
			case "height":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Height = result
			case "description":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Description = result
			default:
				p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "image"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return image, nil
}

func (rp *Parser) parseGUID(p *xpp.XMLPullParser) (guid *GUID, err error) {
	if err = p.Expect(xpp.StartTag, "guid"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	guid = &GUID{}
	guid.IsPermalink = p.Attribute("isPermalink")

	result, err := shared.ParseText(p)
	if err != nil {
		return nil, err
	}
	guid.Value = result

	if err = p.Expect(xpp.EndTag, "guid"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return guid, nil
}

func (rp *Parser) parseCategory(p *xpp.XMLPullParser) (cat *Category, err error) {
	if err = p.Expect(xpp.StartTag, "category"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	cat = &Category{}
	cat.Domain = p.Attribute("domain")

	result, err := shared.ParseText(p)
	if err != nil {
		return nil, err
	}

	cat.Value = result

	if err = p.Expect(xpp.EndTag, "category"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}
	return cat, nil
}

func (rp *Parser) parseTextInput(p *xpp.XMLPullParser) (*TextInput, error) {
	if err := p.Expect(xpp.StartTag, "textinput"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	ti := &TextInput{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)

			switch name {
			case "title":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				ti.Title = result
			case "description":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				ti.Description = result
			case "name":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				ti.Name = result
			case "link":
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				ti.Link = result
			default:
				p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := p.Expect(xpp.EndTag, "textinput"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return ti, nil
}

func (rp *Parser) parseSkipHours(p *xpp.XMLPullParser) ([]string, error) {
	if err := p.Expect(xpp.StartTag, "skiphours"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	hours := []string{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)
			if name == "hour" {
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				hours = append(hours, result)
			} else {
				p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := p.Expect(xpp.EndTag, "skiphours"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return hours, nil
}

func (rp *Parser) parseSkipDays(p *xpp.XMLPullParser) ([]string, error) {
	if err := p.Expect(xpp.StartTag, "skipdays"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	days := []string{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)
			if name == "day" {
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				days = append(days, result)
			} else {
				p.Skip() //nolint:errcheck // upstream ignores err
			}
		}
	}

	if err := p.Expect(xpp.EndTag, "skipdays"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return days, nil
}

func (rp *Parser) parseCloud(p *xpp.XMLPullParser) (*Cloud, error) {
	if err := p.Expect(xpp.StartTag, "cloud"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	cloud := &Cloud{}
	cloud.Domain = p.Attribute("domain")
	cloud.Port = p.Attribute("port")
	cloud.Path = p.Attribute("path")
	cloud.RegisterProcedure = p.Attribute("registerProcedure")
	cloud.Protocol = p.Attribute("protocol")

	shared.NextTag(p) //nolint:errcheck // upstream ignores err

	if err := p.Expect(xpp.EndTag, "cloud"); err != nil {
		return nil, fmt.Errorf("gofeed/rss: %w", err)
	}

	return cloud, nil
}

func (rp *Parser) parseVersion(p *xpp.XMLPullParser) string {
	name := strings.ToLower(p.Name)
	switch name {
	case "rss":
		return p.Attribute("version")
	case "rdf":
		switch p.Attribute("xmlns") {
		case "http://channel.netscape.com/rdf/simple/0.9/",
			"http://my.netscape.com/rdf/simple/0.9/":
			return "0.9"
		case "http://purl.org/rss/1.0/":
			return "1.0"
		}
	}
	return ""
}
