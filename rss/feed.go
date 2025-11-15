package rss

import (
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/dsh2dsh/gofeed/v2/atom"
	"github.com/dsh2dsh/gofeed/v2/ext"
	"github.com/dsh2dsh/gofeed/v2/internal/json"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
)

// Feed is an RSS Feed
type Feed struct {
	Title               string                   `json:"title,omitempty"`
	Links               []string                 `json:"links,omitempty"`
	AtomLinks           []*atom.Link             `json:"atomLinks,omitempty"`
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
	Media               *ext.Media               `json:"media,omitempty"`
	Extensions          ext.Extensions           `json:"extensions,omitempty"`
	Items               []*Item                  `json:"items,omitempty"`
	Version             string                   `json:"version,omitempty"`
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
	s, _ := json.MarshalString(self)
	return s
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

func (self *Feed) Link() string {
	if len(self.Links) != 0 {
		return self.Links[0]
	}
	return ""
}

func (self *Feed) FeedLink() (link string) {
	for _, l := range self.AtomLinks {
		if l.Rel == "self" {
			return l.Href
		}
	}
	return ""
}

func (self *Feed) LinkSeq() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, link := range self.Links {
			if !yield(link) {
				return
			}
		}

		for _, link := range self.AtomLinks {
			switch link.Rel {
			case "", "alternate", "self":
				if !yield(link.Href) {
					return
				}
			}
		}
	}
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
	}

	if self.DublinCoreExt != nil {
		switch {
		case self.DublinCoreExt.Author != "":
			name, address = shared.ParseNameAddress(self.DublinCoreExt.Author)
			return name, address, true
		case self.DublinCoreExt.Creator != "":
			name, address = shared.ParseNameAddress(self.DublinCoreExt.Creator)
			return name, address, true
		}
	}

	if self.ITunesExt != nil {
		switch {
		case self.ITunesExt.Author != "":
			name, address = shared.ParseNameAddress(self.ITunesExt.Author)
			return name, address, true
		case self.ITunesExt.Owner != nil:
			owner := self.ITunesExt.Owner
			return owner.Name, owner.Email, true
		}
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

	if self.Media == nil {
		return nil
	}

	for _, c := range self.Media.Contents {
		hasImage := strings.HasPrefix(c.Type, "image/") || c.Medium == "image"
		if hasImage {
			return &Image{URL: c.URL}
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

func (self *Feed) AllCategories() iter.Seq[string] {
	return self.categoriesIter
}

func (self *Feed) categoriesIter(yield func(string) bool) {
	for _, c := range self.Categories {
		if !yield(c.Value) {
			return
		}
	}

	if itunes := self.ITunesExt; itunes != nil {
		if itunes.Keywords != "" {
			for s := range strings.SplitSeq(itunes.Keywords, ",") {
				if !yield(s) {
					return
				}
			}
		}

		for _, c := range itunes.Categories {
			if !yield(c.Text) {
				return
			}
			if s := c.Subcategory; s != nil {
				if !yield(s.Text) {
					return
				}
			}
		}
	}

	if dc := self.DublinCoreExt; dc != nil && dc.Subject != "" {
		if !yield(dc.Subject) {
			return
		}
	}

	if media := self.Media; media != nil {
		for s := range media.AllCategories() {
			if !yield(s) {
				return
			}
		}
	}
}

func (self *Feed) GetTTL() int {
	if self.TTL == "" {
		return 0
	}

	ttl, err := strconv.Atoi(self.TTL)
	if err != nil {
		return 0
	}
	return ttl
}

// Item is an RSS Item
type Item struct {
	Title         string                   `json:"title,omitempty"`
	Links         []string                 `json:"links,omitempty"`
	AtomLinks     []*atom.Link             `json:"atomLinks,omitempty"`
	Description   string                   `json:"description,omitempty"`
	Content       string                   `json:"content,omitempty"`
	Author        string                   `json:"author,omitempty"`
	Categories    []*Category              `json:"categories,omitempty"`
	Comments      string                   `json:"comments,omitempty"`
	Enclosure     *Enclosure               `json:"enclosure,omitempty"`
	GUID          *GUID                    `json:"guid,omitempty"`
	PubDate       string                   `json:"pubDate,omitempty"`
	PubDateParsed *time.Time               `json:"pubDateParsed,omitempty"`
	Source        *Source                  `json:"source,omitempty"`
	DublinCoreExt *ext.DublinCoreExtension `json:"dcExt,omitempty"`
	ITunesExt     *ext.ITunesItemExtension `json:"itunesExt,omitempty"`
	Media         *ext.Media               `json:"media,omitempty"`
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

func (self *Item) GetContent() string {
	if self.Content != "" {
		return self.Content
	}
	return self.GetDescription()
}

func (self *Item) GetDescription() string {
	switch {
	case self.Description != "":
		return self.Description
	case self.DublinCoreExt != nil && self.DublinCoreExt.Description != "":
		return self.DublinCoreExt.Description
	}

	if self.ITunesExt != nil {
		switch {
		case self.ITunesExt.Summary != "":
			return self.ITunesExt.Summary
		case self.ITunesExt.Subtitle != "":
			return self.ITunesExt.Subtitle
		}
	}

	if self.Media != nil {
		return self.Media.Description()
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
	if self.Author != "" {
		name, address = shared.ParseNameAddress(self.Author)
		return name, address, true
	}

	if self.DublinCoreExt != nil {
		switch {
		case self.DublinCoreExt.Author != "":
			name, address = shared.ParseNameAddress(self.DublinCoreExt.Author)
			return name, address, true
		case self.DublinCoreExt.Creator != "":
			name, address = shared.ParseNameAddress(self.DublinCoreExt.Creator)
			return name, address, true
		}
	}

	if self.ITunesExt != nil && self.ITunesExt.Author != "" {
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

	if self.Media != nil {
		for _, c := range self.Media.Contents {
			hasImage := strings.Contains(c.Type, "image") ||
				strings.Contains(c.Medium, "image")
			if hasImage {
				return c.URL
			}
		}
	}

	enc := self.Enclosure
	if enc != nil && strings.HasPrefix(enc.Type, "image/") {
		return enc.URL
	}
	return ""
}

func (self *Item) AllCategories() iter.Seq[string] {
	return self.categoriesIter
}

func (self *Item) categoriesIter(yield func(string) bool) {
	for _, c := range self.Categories {
		if !yield(c.Value) {
			return
		}
	}

	if itunes := self.ITunesExt; itunes != nil {
		if itunes.Keywords != "" {
			for s := range strings.SplitSeq(itunes.Keywords, ",") {
				if !yield(s) {
					return
				}
			}
		}
	}

	if dc := self.DublinCoreExt; dc != nil {
		if dc.Subject != "" {
			if !yield(dc.Subject) {
				return
			}
		}
	}

	if media := self.Media; media != nil {
		for s := range media.AllCategories() {
			if !yield(s) {
				return
			}
		}
	}
}

func (self *Item) Link() string {
	for _, s := range self.Links {
		if s != "" {
			return s
		}
	}

	for _, link := range self.AtomLinks {
		if link.Href != "" {
			switch link.Rel {
			case "", "alternate":
				return link.Href
			}
		}
	}

	if guid := self.GUID; guid != nil {
		if s := guid.IsPermalink; s == "true" || s == "" {
			return guid.Value
		}
	}
	return ""
}

func (self *Item) AllEnclosures() iter.Seq[Enclosure] {
	return func(yield func(Enclosure) bool) {
		if self.Enclosure != nil && self.Enclosure.URL != "" {
			if !yield(*self.Enclosure) {
				return
			}
		}

		if self.Media == nil {
			return
		}

		for enc := range self.mediaThumbnails() {
			if !yield(enc) {
				return
			}
		}

		for enc := range self.mediaContents() {
			if !yield(enc) {
				return
			}
		}

		for enc := range self.mediaPeerLinks() {
			if !yield(enc) {
				return
			}
		}
	}
}

func (self *Item) mediaThumbnails() iter.Seq[Enclosure] {
	return func(yield func(Enclosure) bool) {
		for thumbnail := range self.Media.AllThumbnails() {
			enc := Enclosure{URL: thumbnail, Type: "image/*"}
			if enc.URL != "" && !yield(enc) {
				return
			}
		}
	}
}

func (self *Item) mediaContents() iter.Seq[Enclosure] {
	return func(yield func(Enclosure) bool) {
		for content := range self.Media.AllContents() {
			enc := Enclosure{
				URL:    content.URL,
				Length: content.FileSize,
				Type:   content.Type,
			}

			if enc.Type == "" {
				switch content.Medium {
				case "image":
					enc.Type = "image/*"
				case "video":
					enc.Type = "video/*"
				case "audio":
					enc.Type = "audio/*"
				default:
					enc.Type = "application/octet-stream"
				}
			}

			if enc.URL != "" && !yield(enc) {
				return
			}
		}
	}
}

func (self *Item) mediaPeerLinks() iter.Seq[Enclosure] {
	return func(yield func(Enclosure) bool) {
		for pl := range self.Media.AllPeerLinks() {
			enc := Enclosure{URL: pl.URL, Type: pl.Type}
			if enc.Type == "" {
				enc.Type = "application/octet-stream"
			}
			if enc.URL != "" && !yield(enc) {
				return
			}
		}
	}
}
