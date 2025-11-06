package itunes

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type feedParser struct {
	p      *xml.Parser
	itunes *ext.ITunesFeedExtension

	err error
}

func ParseFeed(p *xml.Parser, itunes *ext.ITunesFeedExtension,
) (*ext.ITunesFeedExtension, error) {
	if itunes == nil {
		itunes = &ext.ITunesFeedExtension{}
	}

	self := feedParser{p: p, itunes: itunes}
	return self.Parse()
}

func (self *feedParser) Parse() (*ext.ITunesFeedExtension, error) {
	name := strings.ToLower(self.p.Name)
	self.body(name)
	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.p.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/itunes: unexpected state at the end of feed: %w", err)
	}
	return self.itunes, nil
}

func (self *feedParser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/itunes: xml parser errored: %w", self.p.Err())
	}
	return nil
}

func (self *feedParser) body(name string) {
	switch name {
	case "author":
		self.itunes.Author = self.p.Text()
	case "block":
		self.itunes.Block = self.p.Text()
	case "explicit":
		self.itunes.Explicit = self.p.Text()
	case "keywords":
		self.itunes.Keywords = self.p.Text()
	case "subtitle":
		self.itunes.Subtitle = self.p.Text()
	case "summary":
		self.itunes.Summary = self.p.Text()
	case "complete":
		self.itunes.Complete = self.p.Text()
	case "new-feed-url":
		self.itunes.NewFeedURL = self.p.Text()
	case "type":
		self.itunes.Type = self.p.Text()
	case "image":
		self.itunes.Image = self.image(name)
	case "category":
		self.itunes.Categories = self.appendCategory(name, self.itunes.Categories)
	case "owner":
		self.itunes.Owner = self.owner(name)
	default:
		self.p.Skip(name)
	}
}

func (self *feedParser) image(name string) string {
	href := self.p.Attribute("href")
	self.p.Skip(name)
	return href
}

func (self *feedParser) appendCategory(name string,
	categories []*ext.ITunesCategory,
) []*ext.ITunesCategory {
	c := self.category(name)
	if self.err != nil {
		return categories
	}
	return append(categories, c)
}

func (self *feedParser) category(name string) (c *ext.ITunesCategory) {
	err := self.p.ParsingElement(name,
		func() error {
			c = &ext.ITunesCategory{Text: self.p.Attribute("text")}
			return nil
		},
		func() error { return self.categoryBody(name, c) })
	if err != nil {
		self.err = err
		return nil
	}
	return c
}

func (self *feedParser) categoryBody(name string, c *ext.ITunesCategory) error {
	switch tag := strings.ToLower(self.p.Name); tag {
	case name:
		c.Subcategory = self.category(tag)
	default:
		self.p.Skip(tag)
	}
	return nil
}

func (self *feedParser) owner(name string) (owner *ext.ITunesOwner) {
	err := self.p.ParsingElement(name,
		func() error {
			owner = new(ext.ITunesOwner)
			return nil
		},
		func() error { return self.ownerBody(owner) })
	if err != nil {
		self.err = err
		return nil
	}
	return owner
}

func (self *feedParser) ownerBody(owner *ext.ITunesOwner) error {
	switch name := strings.ToLower(self.p.Name); name {
	case "name":
		owner.Name = self.p.Text()
	case "email":
		owner.Email = self.p.Text()
	default:
		self.p.Skip(name)
	}
	return nil
}
