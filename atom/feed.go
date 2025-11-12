package atom

import (
	"strings"
	"time"

	"github.com/dsh2dsh/gofeed/v2/ext"
	"github.com/dsh2dsh/gofeed/v2/internal/json"
)

// Feed is an Atom Feed
type Feed struct {
	Title         string         `json:"title,omitempty"`
	ID            string         `json:"id,omitempty"`
	Updated       string         `json:"updated,omitempty"`
	UpdatedParsed *time.Time     `json:"updatedParsed,omitempty"`
	Subtitle      string         `json:"subtitle,omitempty"`
	Links         []*Link        `json:"links,omitempty"`
	Language      string         `json:"language,omitempty"`
	Generator     *Generator     `json:"generator,omitempty"`
	Icon          string         `json:"icon,omitempty"`
	Logo          string         `json:"logo,omitempty"`
	Rights        string         `json:"rights,omitempty"`
	Contributors  []*Person      `json:"contributors,omitempty"`
	Authors       []*Person      `json:"authors,omitempty"`
	Categories    []*Category    `json:"categories,omitempty"`
	Entries       []*Entry       `json:"entries,omitempty"`
	Extensions    ext.Extensions `json:"extensions,omitempty"`
	Version       string         `json:"version,omitempty"`
}

// Link is an Atom link that defines a reference
// from an entry or feed to a Web resource
type Link struct {
	Href     string `json:"href,omitempty"`
	Hreflang string `json:"hreflang,omitempty"`
	Rel      string `json:"rel,omitempty"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Length   string `json:"length,omitempty"`
}

// Generator identifies the agent used to generate a
// feed, for debugging and other purposes.
type Generator struct {
	Value   string `json:"value,omitempty"`
	URI     string `json:"uri,omitempty"`
	Version string `json:"version,omitempty"`
}

// Person represents a person in an Atom feed
// for things like Authors, Contributors, etc
type Person struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URI   string `json:"uri,omitempty"`
}

// Category is category metadata for Feeds and Entries
type Category struct {
	Term   string `json:"term,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Label  string `json:"label,omitempty"`
}

func (self *Feed) String() string {
	s, _ := json.MarshalString(self)
	return s
}

func (self *Feed) GetLink() string {
	if l := firstLinkWithType("alternate", self.Links); l != nil {
		return l.Href
	}
	return ""
}

func (self *Feed) GetFeedLink() string {
	if feedLink := firstLinkWithType("self", self.Links); feedLink != nil {
		return feedLink.Href
	}
	return ""
}

func (self *Feed) GetLinks() (links []string) {
	for _, l := range self.Links {
		if l.Rel == "" || l.Rel == "alternate" || l.Rel == "self" {
			links = append(links, l.Href)
		}
	}
	return links
}

func (self *Feed) GetAuthor() *Person { return firstPerson(self.Authors) }

func (self *Feed) ImageURL() string {
	if self.Logo != "" {
		return self.Logo
	}
	return self.Icon
}

func (self *Feed) GetGenerator() string {
	if self.Generator == nil {
		return ""
	}

	generator := self.Generator.Value
	if s := self.Generator.Version; s != "" {
		generator += " v" + s
	}
	if s := self.Generator.URI; s != "" {
		generator += " " + s
	}
	return strings.TrimSpace(generator)
}

func (self *Feed) GetCategories() []string {
	if len(self.Categories) == 0 {
		return nil
	}

	categories := make([]string, len(self.Categories))
	for i, c := range self.Categories {
		if c.Label != "" {
			categories[i] = c.Label
		} else {
			categories[i] = c.Term
		}
	}
	return categories
}

func firstLinkWithType(linkType string, links []*Link) *Link {
	for _, link := range links {
		if link.Rel == linkType {
			return link
		}
	}
	return nil
}

func firstPerson(persons []*Person) *Person {
	if len(persons) == 0 {
		return nil
	}
	return persons[0]
}

// Entry is an Atom Entry
type Entry struct {
	Title           string         `json:"title,omitempty"`
	ID              string         `json:"id,omitempty"`
	Updated         string         `json:"updated,omitempty"`
	UpdatedParsed   *time.Time     `json:"updatedParsed,omitempty"`
	Summary         string         `json:"summary,omitempty"`
	Authors         []*Person      `json:"authors,omitempty"`
	Contributors    []*Person      `json:"contributors,omitempty"`
	Categories      []*Category    `json:"categories,omitempty"`
	Links           []*Link        `json:"links,omitempty"`
	Rights          string         `json:"rights,omitempty"`
	Published       string         `json:"published,omitempty"`
	PublishedParsed *time.Time     `json:"publishedParsed,omitempty"`
	Source          *Source        `json:"source,omitempty"`
	Content         *Content       `json:"content,omitempty"`
	Media           *ext.Media     `json:"media,omitempty"`
	Extensions      ext.Extensions `json:"extensions,omitempty"`
}

// Content either contains or links to the content of
// the entry
type Content struct {
	Src   string `json:"src,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Source contains the feed information for another
// feed if a given entry came from that feed.
type Source struct {
	Title         string         `json:"title,omitempty"`
	ID            string         `json:"id,omitempty"`
	Updated       string         `json:"updated,omitempty"`
	UpdatedParsed *time.Time     `json:"updatedParsed,omitempty"`
	Subtitle      string         `json:"subtitle,omitempty"`
	Links         []*Link        `json:"links,omitempty"`
	Generator     *Generator     `json:"generator,omitempty"`
	Icon          string         `json:"icon,omitempty"`
	Logo          string         `json:"logo,omitempty"`
	Rights        string         `json:"rights,omitempty"`
	Contributors  []*Person      `json:"contributors,omitempty"`
	Authors       []*Person      `json:"authors,omitempty"`
	Categories    []*Category    `json:"categories,omitempty"`
	Extensions    ext.Extensions `json:"extensions,omitempty"`
}

func (self *Entry) GetContent() string {
	if self.Content != nil {
		return self.Content.Value
	}

	if self.Summary != "" {
		return self.Summary
	}

	if self.Media != nil {
		return self.Media.Description()
	}
	return ""
}

func (self *Entry) GetLink() string {
	if l := firstLinkWithType("alternate", self.Links); l != nil {
		return l.Href
	}
	return ""
}

func (self *Entry) GetLinks() []string {
	if len(self.Links) == 0 {
		return nil
	}

	var links []string
	for _, l := range self.Links {
		if l.Rel == "" || l.Rel == "alternate" || l.Rel == "self" {
			links = append(links, l.Href)
		}
	}
	return links
}

func (self *Entry) GetPublished() string {
	if self.Published != "" {
		return self.Published
	}
	return self.Updated
}

func (self *Entry) GetPublishedParsed() *time.Time {
	if self.PublishedParsed != nil {
		return self.PublishedParsed
	}
	return self.UpdatedParsed
}

func (self *Entry) GetAuthor() *Person { return firstPerson(self.Authors) }

func (self *Entry) GetCategories() []string {
	if len(self.Categories) == 0 {
		return nil
	}

	categories := make([]string, len(self.Categories))
	for i, c := range self.Categories {
		if c.Label != "" {
			categories[i] = c.Label
		} else {
			categories[i] = c.Term
		}
	}
	return categories
}
