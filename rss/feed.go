package rss

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
)

// Feed is an RSS Feed
type Feed struct {
	Title               string                   `json:"title,omitempty"`
	Link                string                   `json:"link,omitempty"`
	Links               []string                 `json:"links,omitempty"`
	Description         string                   `json:"description,omitempty"`
	Language            string                   `json:"language,omitempty"`
	Copyright           string                   `json:"copyright,omitempty"`
	ManagingEditor      string                   `json:"managingEditor,omitempty"`
	WebMaster           string                   `json:"webMaster,omitempty"`
	PubDate             string                   `json:"pubDate,omitempty"`
	PubDateParsed       *time.Time               `json:"pubDateParsed,omitempty"`
	LastBuildDate       string                   `json:"lastBuildDate,omitempty"`
	LastBuildDateParsed *time.Time               `json:"lastBuildDateParsed,omitempty"`
	Categories          []*Category              `json:"categories,omitempty"`
	Generator           string                   `json:"generator,omitempty"`
	Docs                string                   `json:"docs,omitempty"`
	TTL                 string                   `json:"ttl,omitempty"`
	Image               *Image                   `json:"image,omitempty"`
	Rating              string                   `json:"rating,omitempty"`
	SkipHours           []string                 `json:"skipHours,omitempty"`
	SkipDays            []string                 `json:"skipDays,omitempty"`
	Cloud               *Cloud                   `json:"cloud,omitempty"`
	TextInput           *TextInput               `json:"textInput,omitempty"`
	DublinCoreExt       *ext.DublinCoreExtension `json:"dcExt,omitempty"`
	ITunesExt           *ext.ITunesFeedExtension `json:"itunesExt,omitempty"`
	Extensions          ext.Extensions           `json:"extensions,omitempty"`
	Items               []*Item                  `json:"items"`
	Version             string                   `json:"version"`
}

// Image is an image that represents the feed
type Image struct {
	URL         string `json:"url,omitempty"`
	Link        string `json:"link,omitempty"`
	Title       string `json:"title,omitempty"`
	Width       string `json:"width,omitempty"`
	Height      string `json:"height,omitempty"`
	Description string `json:"description,omitempty"`
}

// Category is category metadata for Feeds and Entries
type Category struct {
	Domain string `json:"domain,omitempty"`
	Value  string `json:"value,omitempty"`
}

// TextInput specifies a text input box that
// can be displayed with the channel
type TextInput struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Link        string `json:"link,omitempty"`
}

// Cloud allows processes to register with a
// cloud to be notified of updates to the channel,
// implementing a lightweight publish-subscribe protocol
// for RSS feeds
type Cloud struct {
	Domain            string `json:"domain,omitempty"`
	Port              string `json:"port,omitempty"`
	Path              string `json:"path,omitempty"`
	RegisterProcedure string `json:"registerProcedure,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
}

func (self *Feed) String() string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	_ = enc.Encode(self)
	return b.String()
}

func (self *Feed) GetTitle() string {
	switch {
	case self.Title != "":
		return self.Title
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Title
	}
	return ""
}

func (self *Feed) GetDescription() string {
	switch {
	case self.Description != "":
		return self.Description
	case self.ITunesExt != nil && self.ITunesExt.Summary != "":
		return self.ITunesExt.Summary
	}
	return ""
}

func (self *Feed) GetLink() string {
	switch {
	case self.Link != "":
		return self.Link
	case self.ITunesExt != nil && self.ITunesExt.Subtitle != "":
		return self.ITunesExt.Subtitle
	}
	return ""
}

func (self *Feed) GetFeedLink() (link string) {
	atomExtensions := ext.ElementsSeq(self.Extensions, "atom", "atom10", "atom03")
	for ex := range atomExtensions {
		if links, ok := ex["link"]; ok {
			for i := range links {
				l := &links[i]
				if l.Attrs["rel"] == "self" {
					return l.Attrs["href"]
				}
			}
		}
	}
	return ""
}

func (self *Feed) GetLinks() (links []string) {
	if n := len(self.Links); n > 0 {
		links = make([]string, 0, n)
		links = append(links, self.Links...)
	}

	atomExtensions := ext.ElementsSeq(self.Extensions, "atom", "atom10", "atom03")
	for ex := range atomExtensions {
		if lks, ok := ex["link"]; ok {
			for i := range lks {
				l := &lks[i]
				ok := l.Attrs["rel"] == "" || l.Attrs["rel"] == "alternate" ||
					l.Attrs["rel"] == "self"
				if ok {
					links = append(links, l.Attrs["href"])
				}
			}
		}
	}
	return links
}

func (self *Feed) GetUpdated() string {
	switch {
	case self.LastBuildDate != "":
		return self.LastBuildDate
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Date
	}
	return ""
}

func (self *Feed) GetUpdatedParsed() *time.Time {
	if self.LastBuildDateParsed != nil {
		return self.LastBuildDateParsed
	}

	if self.DublinCoreExt == nil || self.DublinCoreExt.Date == "" {
		return nil
	}

	if date, err := shared.ParseDate(self.DublinCoreExt.Date); err == nil {
		return &date
	}
	return nil
}

func (self *Feed) GetAuthor() (name, address string, ok bool) {
	switch {
	case self.ManagingEditor != "":
		name, address = shared.ParseNameAddress(self.ManagingEditor)
		return name, address, true
	case self.WebMaster != "":
		name, address = shared.ParseNameAddress(self.WebMaster)
		return name, address, true
	case self.DublinCoreExt != nil && self.DublinCoreExt.Author != "":
		name, address = shared.ParseNameAddress(self.DublinCoreExt.Author)
		return name, address, true
	case self.DublinCoreExt != nil && self.DublinCoreExt.Creator != "":
		name, address = shared.ParseNameAddress(self.DublinCoreExt.Creator)
		return name, address, true
	case self.ITunesExt != nil && self.ITunesExt.Author != "":
		name, address = shared.ParseNameAddress(self.ITunesExt.Author)
		return name, address, true
	}
	return name, address, false
}

func (self *Feed) GetLanguage() string {
	switch {
	case self.Language != "":
		return self.Language
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Language
	}
	return ""
}

func (self *Feed) GetImage() *Image {
	if self.Image != nil {
		return self.Image
	}

	if self.ITunesExt != nil && self.ITunesExt.Image != "" {
		return &Image{URL: self.ITunesExt.Image}
	}

	media, ok := self.Extensions["media"]
	if !ok {
		return nil
	}

	content, ok := media["content"]
	if !ok {
		return nil
	}

	for i := range content {
		c := &content[i]
		hasImage := strings.HasPrefix(c.Attrs["type"], "image/") ||
			c.Attrs["medium"] == "image"
		if hasImage {
			return &Image{URL: c.Attrs["url"]}
		}
	}
	return nil
}

func (self *Feed) GetCopyright() string {
	switch {
	case self.Copyright != "":
		return self.Copyright
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Rights
	}
	return ""
}

func (self *Feed) GetCategories() []string {
	var cats []string
	if len(self.Categories) != 0 {
		cats = make([]string, 0, len(self.Categories))
		for _, c := range self.Categories {
			cats = append(cats, c.Value)
		}
	}

	if self.ITunesExt != nil && self.ITunesExt.Keywords != "" {
		cats = slices.Grow(cats, strings.Count(self.ITunesExt.Keywords, ",")+1)
		cats = slices.AppendSeq(cats, strings.SplitSeq(
			self.ITunesExt.Keywords, ","))
	}

	if self.ITunesExt != nil && len(self.ITunesExt.Categories) != 0 {
		cats = slices.Grow(cats, len(self.ITunesExt.Categories))
		for _, c := range self.ITunesExt.Categories {
			cats = append(cats, c.Text)
			if s := c.Subcategory; s != nil {
				cats = append(cats, s.Text)
			}
		}
	}

	if self.DublinCoreExt != nil && self.DublinCoreExt.Subject != "" {
		cats = append(cats, self.DublinCoreExt.Subject)
	}
	return cats
}

// Item is an RSS Item
type Item struct {
	Title         string                   `json:"title,omitempty"`
	Link          string                   `json:"link,omitempty"`
	Links         []string                 `json:"links,omitempty"`
	Description   string                   `json:"description,omitempty"`
	Content       string                   `json:"content,omitempty"`
	Author        string                   `json:"author,omitempty"`
	Categories    []*Category              `json:"categories,omitempty"`
	Comments      string                   `json:"comments,omitempty"`
	Enclosure     *Enclosure               `json:"enclosure,omitempty"`
	Enclosures    []*Enclosure             `json:"enclosures,omitempty"`
	GUID          *GUID                    `json:"guid,omitempty"`
	PubDate       string                   `json:"pubDate,omitempty"`
	PubDateParsed *time.Time               `json:"pubDateParsed,omitempty"`
	Source        *Source                  `json:"source,omitempty"`
	DublinCoreExt *ext.DublinCoreExtension `json:"dcExt,omitempty"`
	ITunesExt     *ext.ITunesItemExtension `json:"itunesExt,omitempty"`
	Extensions    ext.Extensions           `json:"extensions,omitempty"`
}

// Enclosure is a media object that is attached to
// the item
type Enclosure struct {
	URL    string `json:"url,omitempty"`
	Length string `json:"length,omitempty"`
	Type   string `json:"type,omitempty"`
}

// GUID is a unique identifier for an item
type GUID struct {
	Value       string `json:"value,omitempty"`
	IsPermalink string `json:"isPermalink,omitempty"`
}

// Source contains feed information for another
// feed if a given item came from that feed
type Source struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

func (self *Item) GetTitle() string {
	switch {
	case self.Title != "":
		return self.Title
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Title
	}
	return ""
}

func (self *Item) GetDescription() string {
	switch {
	case self.Description != "":
		return self.Description
	case self.DublinCoreExt != nil && self.DublinCoreExt.Description != "":
		return self.DublinCoreExt.Description
	case self.ITunesExt != nil && self.ITunesExt.Summary != "":
		return self.ITunesExt.Summary
	}
	return ""
}

func (self *Item) GetPublished() string {
	switch {
	case self.PubDate != "":
		return self.PubDate
	case self.DublinCoreExt != nil:
		return self.DublinCoreExt.Date
	}
	return ""
}

func (self *Item) GetPublishedParsed() *time.Time {
	if self.PubDateParsed != nil {
		return self.PubDateParsed
	}

	if self.DublinCoreExt == nil || self.DublinCoreExt.Date == "" {
		return nil
	}

	pubDateParsed, err := shared.ParseDate(self.DublinCoreExt.Date)
	if err == nil {
		return &pubDateParsed
	}
	return nil
}

func (self *Item) GetAuthor() (name, address string, ok bool) {
	switch {
	case self.Author != "":
		name, address = shared.ParseNameAddress(self.Author)
		return name, address, true
	case self.DublinCoreExt != nil && self.DublinCoreExt.Author != "":
		name, address = shared.ParseNameAddress(self.DublinCoreExt.Author)
		return name, address, true
	case self.DublinCoreExt != nil && self.DublinCoreExt.Creator != "":
		name, address = shared.ParseNameAddress(self.DublinCoreExt.Creator)
		return name, address, true
	case self.ITunesExt != nil && self.ITunesExt.Author != "":
		name, address = shared.ParseNameAddress(self.ITunesExt.Author)
		return name, address, true
	}
	return name, address, false
}

func (self *Item) GetGUID() string {
	if self.GUID != nil {
		return self.GUID.Value
	}
	return ""
}

func (self *Item) ImageURL() string {
	if self.ITunesExt != nil && self.ITunesExt.Image != "" {
		return self.ITunesExt.Image
	}

	if media, ok := self.Extensions["media"]; ok {
		if content, ok := media["content"]; ok {
			for _, c := range content {
				hasImage := strings.Contains(c.Attrs["type"], "image") ||
					strings.Contains(c.Attrs["medium"], "image")
				if hasImage {
					return c.Attrs["url"]
				}
			}
		}
	}

	for _, enc := range self.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") {
			return enc.URL
		}
	}
	return ""
}

func (self *Item) GetCategories() []string {
	var cats []string
	if self.Categories != nil {
		cats = make([]string, 0, len(self.Categories))
		for _, c := range self.Categories {
			cats = append(cats, c.Value)
		}
	}

	if self.ITunesExt != nil && self.ITunesExt.Keywords != "" {
		cats = slices.Grow(cats, strings.Count(self.ITunesExt.Keywords, ",")+1)
		cats = slices.AppendSeq(cats, strings.SplitSeq(
			self.ITunesExt.Keywords, ","))
	}

	if self.DublinCoreExt != nil && self.DublinCoreExt.Subject != "" {
		cats = append(cats, self.DublinCoreExt.Subject)
	}
	return cats
}
