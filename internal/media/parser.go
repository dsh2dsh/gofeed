package media

import (
	"fmt"
	"iter"
	"strconv"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	"github.com/dsh2dsh/gofeed/v2/ext"
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
	m := self.media
	switch name {
	case "group":
		m.Groups = self.appendGroup(name, m.Groups)
	case "content":
		m.Contents = self.appendContent(name, m.Contents)
	case "category":
		m.Categories = self.appendCategory(name, m.Categories)
	case "thumbnail":
		m.ThumbnailsEx = self.appendThumbnail(name, m.ThumbnailsEx,
			func(t *ext.MediaThumbnail) {
				m.Thumbnails = append(m.Thumbnails, t.URL)
			})
	case "title":
		m.Titles = self.appendDescription(name, m.Titles)
	case "description":
		m.Descriptions = self.appendDescription(name, m.Descriptions)
	case "peerlink":
		m.PeerLinks = self.appendPeerLink(name, m.PeerLinks)
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

	if s := strings.TrimSpace(label); s == "" {
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
		var err error
		switch name {
		case "url":
			c.URL = value
		case "type":
			c.Type = value
		case "filesize":
			c.FileSize = value
		case "medium":
			c.Medium = value
		case "height":
			err = parseIntTo(name, value, &c.Height)
		case "width":
			err = parseIntTo(name, value, &c.Width)
		}
		if err != nil {
			self.err = err
			return contents
		}
	}

	for name := range children {
		switch name {
		case "category":
			c.Categories = self.appendCategory(name, c.Categories)
		case "thumbnail":
			c.ThumbnailsEx = self.appendThumbnail(name, c.ThumbnailsEx,
				func(t *ext.MediaThumbnail) {
					c.Thumbnails = append(c.Thumbnails, t.URL)
				})
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

	if self.err != nil || c.URL == "" {
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

func (self *parser) appendThumbnail(name string,
	thumbnails []ext.MediaThumbnail, okFunc func(*ext.MediaThumbnail),
) []ext.MediaThumbnail {
	var t ext.MediaThumbnail
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			var err error
			switch name {
			case "url":
				t.URL = value
			case "height":
				err = parseIntTo(name, value, &t.Height)
			case "width":
				err = parseIntTo(name, value, &t.Width)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	switch {
	case err != nil:
		self.err = err
		fallthrough
	case t.URL == "":
		return thumbnails
	}

	if okFunc != nil {
		okFunc(&t)
	}
	return append(thumbnails, t)
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

	if link.URL == "" {
		return links
	}
	return append(links, link)
}

func (self *parser) appendGroup(name string, groups []ext.MediaGroup,
) []ext.MediaGroup {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return groups
	}

	var g ext.MediaGroup
	for name := range children {
		switch name {
		case "category":
			g.Categories = self.appendCategory(name, g.Categories)
		case "content":
			g.Contents = self.appendContent(name, g.Contents)
		case "thumbnail":
			g.ThumbnailsEx = self.appendThumbnail(name, g.ThumbnailsEx,
				func(t *ext.MediaThumbnail) {
					g.Thumbnails = append(g.Thumbnails, t.URL)
				})
		case "title":
			g.Titles = self.appendDescription(name, g.Titles)
		case "description":
			g.Descriptions = self.appendDescription(name, g.Descriptions)
		case "peerlink":
			g.PeerLinks = self.appendPeerLink(name, g.PeerLinks)
		case "community":
			g.Community = self.community(name)
		default:
			self.p.Skip(name)
		}
	}

	if self.err != nil {
		return groups
	}
	return append(groups, g)
}

func (self *parser) community(name string) (community ext.MediaCommunity) {
	children := self.makeChildrenSeq(name)
	if children == nil {
		return community
	}

	for name := range children {
		switch name {
		case "starrating":
			community.StarRating = self.starRating(name)
		case "statistics":
			community.Statistics = self.statistics(name)
		default:
			self.p.Skip(name)
		}
	}
	return community
}

func (self *parser) starRating(name string) (rating ext.MediaStarRating) {
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			var err error
			switch name {
			case "average":
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("gofeed/media: parse %v=%q as float: %w",
						name, value, err)
				}
				rating.Average = v
			case "count":
				err = parseIntTo(name, value, &rating.Count)
			case "min":
				err = parseIntTo(name, value, &rating.Min)
			case "max":
				err = parseIntTo(name, value, &rating.Max)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		self.err = err
	}
	return rating
}

func parseIntTo(name, value string, to *int) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("gofeed/media: parse %v=%q as int: %w", name, value, err)
	}
	*to = n
	return nil
}

func (self *parser) statistics(name string) (stat ext.MediaStatistics) {
	err := self.p.WithSkip(name, func() error {
		for name, value := range self.p.AttributeSeq() {
			var err error
			switch name {
			case "views":
				err = parseIntTo(name, value, &stat.Views)
			case "favorites":
				err = parseIntTo(name, value, &stat.Favorites)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		self.err = err
	}
	return stat
}
