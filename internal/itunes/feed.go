package itunes

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type feedParser struct {
	xpp    *xml.Parser
	itunes *ext.ITunesFeedExtension

	err error
}

func ParseFeed(p *xml.Parser, itunes *ext.ITunesFeedExtension,
) (*ext.ITunesFeedExtension, error) {
	if itunes == nil {
		itunes = &ext.ITunesFeedExtension{}
	}

	self := feedParser{xpp: p, itunes: itunes}
	return self.Parse()
}

func (self *feedParser) Parse() (*ext.ITunesFeedExtension, error) {
	name := strings.ToLower(self.xpp.Name)
	switch name {
	case "author":
		self.itunes.Author = self.xpp.Text()
	case "block":
		self.itunes.Block = self.xpp.Text()
	case "explicit":
		self.itunes.Explicit = self.xpp.Text()
	case "keywords":
		self.itunes.Keywords = self.xpp.Text()
	case "subtitle":
		self.itunes.Subtitle = self.xpp.Text()
	case "summary":
		self.itunes.Summary = self.xpp.Text()
	case "complete":
		self.itunes.Complete = self.xpp.Text()
	case "new-feed-url":
		self.itunes.NewFeedURL = self.xpp.Text()
	case "type":
		self.itunes.Type = self.xpp.Text()
	case "image":
		self.itunes.Image = self.image()
	case "category":
		self.itunes.Categories = self.appendCategory(name, self.itunes.Categories)
	case "owner":
		self.itunes.Owner = self.owner(name)
	default:
		self.xpp.Skip(name)
	}

	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.xpp.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/itunes: unexpected state at the end of feed: %w", err)
	}
	return self.itunes, nil
}

func (self *feedParser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.xpp.Err() != nil:
		return fmt.Errorf("gofeed/itunes: xml parser errored: %w", self.xpp.Err())
	}
	return nil
}

func (self *feedParser) image() string {
	return self.xpp.Attribute("href")
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
	err := self.xpp.ParsingElement(name,
		func() error {
			c = &ext.ITunesCategory{Text: self.xpp.Attribute("text")}
			return nil
		},
		func() error {
			switch tag := strings.ToLower(self.xpp.Name); tag {
			case name:
				c.Subcategory = self.category(tag)
			default:
				self.xpp.Skip(tag)
			}
			return nil
		})
	if err != nil {
		self.err = err
		return nil
	}
	return c
}

func (self *feedParser) owner(name string) (owner *ext.ITunesOwner) {
	err := self.xpp.ParsingElement(name,
		func() error {
			owner = new(ext.ITunesOwner)
			return nil
		},
		func() error {
			switch tag := strings.ToLower(self.xpp.Name); tag {
			case "name":
				owner.Name = self.xpp.Text()
			case "email":
				owner.Email = self.xpp.Text()
			default:
				self.xpp.Skip(tag)
			}
			return nil
		})
	if err != nil {
		self.err = err
		return nil
	}
	return owner
}
