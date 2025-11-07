package itunes

import (
	"fmt"
	"iter"
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

func (self *feedParser) image(name string) (href string) {
	err := self.p.WithSkip(name, func() error {
		href = self.p.Attribute("href")
		return nil
	})
	if err != nil {
		self.err = err
		return ""
	}
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

func (self *feedParser) category(name string) *ext.ITunesCategory {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	c := &ext.ITunesCategory{Text: self.p.Attribute("text")}
	for name := range children {
		switch name {
		case "category":
			c.Subcategory = self.category(name)
		default:
			self.p.Skip(name)
		}
	}

	if self.err != nil {
		return nil
	}
	return c
}

func (self *feedParser) makeChildrenSeq(name string) iter.Seq[string] {
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

func (self *feedParser) owner(name string) (owner *ext.ITunesOwner) {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return nil
	}

	owner = new(ext.ITunesOwner)
	for name := range children {
		switch name {
		case "name":
			owner.Name = self.p.Text()
		case "email":
			owner.Email = self.p.Text()
		default:
			self.p.Skip(name)
		}
	}

	if self.err != nil {
		return nil
	}
	return owner
}
