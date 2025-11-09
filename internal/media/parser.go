package media

import (
	"fmt"
	"iter"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type parser struct {
	p     *xml.Parser
	media *ext.Media

	err error
}

func Parse(p *xml.Parser, media *ext.Media) (*ext.Media, error) {
	if media == nil {
		media = new(ext.Media)
	}

	self := parser{p: p, media: media}
	return self.Parse()
}

func (self *parser) Parse() (*ext.Media, error) {
	name := strings.ToLower(self.p.Name)
	self.body(name)
	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.p.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/media: unexpected state at the end: %w", err)
	}
	return self.media, nil
}

func (self *parser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/media: xml parser errored: %w",
			self.p.Err())
	}
	return nil
}

func (self *parser) body(name string) {
	switch name {
	case "category":
		self.media.Categories = self.appendCategory(name, self.media.Categories)
	case "group":
		self.media.Groups = self.appendGroup(name, self.media.Groups)
	case "content":
		self.media.Contents = self.appendContent(name, self.media.Contents)
	case "thumbnail":
		self.media.Thumbnails = self.appendThumbnail(name, self.media.Thumbnails)
	case "title":
		self.media.Titles = self.appendDescription(name, self.media.Titles)
	case "description":
		self.media.Descriptions = self.appendDescription(name,
			self.media.Descriptions)
	case "peerlink":
		self.media.PeerLinks = self.appendPeerLink(name, self.media.PeerLinks)
	default:
		self.p.Skip(name)
	}
}

func (self *parser) appendCategory(name string, categories []string) []string {
	var label string
	err := self.p.WithSkip(name, func() error {
		label = self.p.Attribute("label")
		return nil
	})
	if err != nil {
		self.err = err
		return categories
	}
	return append(categories, label)
}

func (self *parser) appendContent(name string, contents []ext.MediaContent,
) []ext.MediaContent {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return contents
	}

	var c ext.MediaContent
	for name, value := range self.p.AttributeSeq() {
		switch name {
		case "url":
			c.URL = value
		case "type":
			c.Type = value
		case "filesize":
			c.FileSize = value
		case "medium":
			c.Medium = value
		}
	}

	for name := range children {
		switch name {
		case "category":
			c.Categories = self.appendCategory(name, c.Categories)
		case "thumbnail":
			c.Thumbnails = self.appendThumbnail(name, c.Thumbnails)
		case "title":
			c.Titles = self.appendDescription(name, c.Titles)
		case "description":
			c.Descriptions = self.appendDescription(name, c.Descriptions)
		case "peerlink":
			c.PeerLinks = self.appendPeerLink(name, c.PeerLinks)
		default:
			self.p.Skip(name)
		}
	}

	if self.err != nil {
		return contents
	}
	return append(contents, c)
}

func (self *parser) makeChildrenSeq(name string) iter.Seq[string] {
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

func (self *parser) appendThumbnail(name string, thumbnails []string) []string {
	var url string
	err := self.p.WithSkip(name, func() error {
		url = self.p.Attribute("url")
		return nil
	})
	if err != nil {
		self.err = err
		return thumbnails
	}
	return append(thumbnails, url)
}

func (self *parser) appendDescription(name string,
	descriptions []ext.MediaDescription,
) []ext.MediaDescription {
	var descr ext.MediaDescription
	err := self.p.WithText(name,
		func() error {
			descr.Type = self.p.Attribute("type")
			return nil
		},
		func(s string) error {
			descr.Text = s
			return nil
		})
	if err != nil {
		self.err = err
		return descriptions
	}
	return append(descriptions, descr)
}

func (self *parser) appendPeerLink(name string, links []ext.MediaPeerLink,
) []ext.MediaPeerLink {
	var link ext.MediaPeerLink
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			switch name {
			case "href":
				link.URL = value
			case "type":
				link.Type = value
			}
		}
		return nil
	})
	if err != nil {
		self.err = err
		return links
	}
	return append(links, link)
}

func (self *parser) appendGroup(name string, groups []ext.MediaGroup,
) []ext.MediaGroup {
	var g ext.MediaGroup
	switch name {
	case "category":
		g.Categories = self.appendCategory(name, g.Categories)
	case "content":
		g.Contents = self.appendContent(name, g.Contents)
	case "thumbnail":
		g.Thumbnails = self.appendThumbnail(name, g.Thumbnails)
	case "title":
		g.Titles = self.appendDescription(name, g.Titles)
	case "description":
		g.Descriptions = self.appendDescription(name, g.Descriptions)
	case "peerlink":
		g.PeerLinks = self.appendPeerLink(name, g.PeerLinks)
	default:
		self.p.Skip(name)
	}

	if err := self.Err(); err != nil {
		self.err = err
		return groups
	}
	return append(groups, g)
}
