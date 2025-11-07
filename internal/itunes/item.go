package itunes

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type itemParser struct {
	p      *xml.Parser
	itunes *ext.ITunesItemExtension

	err error
}

func ParseItem(p *xml.Parser, itunes *ext.ITunesItemExtension,
) (*ext.ITunesItemExtension, error) {
	if itunes == nil {
		itunes = &ext.ITunesItemExtension{}
	}

	self := itemParser{p: p, itunes: itunes}
	return self.Parse()
}

func (self *itemParser) Parse() (*ext.ITunesItemExtension, error) {
	name := strings.ToLower(self.p.Name)
	self.body(name)
	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.p.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/itunes: unexpected state at the end of item: %w", err)
	}
	return self.itunes, nil
}

func (self *itemParser) body(name string) {
	switch name {
	case "author":
		self.itunes.Author = self.p.Text()
	case "block":
		self.itunes.Block = self.p.Text()
	case "duration":
		self.itunes.Duration = self.p.Text()
	case "explicit":
		self.itunes.Explicit = self.p.Text()
	case "subtitle":
		self.itunes.Subtitle = self.p.Text()
	case "summary":
		self.itunes.Summary = self.p.Text()
	case "keywords":
		self.itunes.Keywords = self.p.Text()
	case "isClosedCaptioned":
		self.itunes.IsClosedCaptioned = self.p.Text()
	case "episode":
		self.itunes.Episode = self.p.Text()
	case "season":
		self.itunes.Season = self.p.Text()
	case "order":
		self.itunes.Order = self.p.Text()
	case "episodeType":
		self.itunes.EpisodeType = self.p.Text()
	case "image":
		self.itunes.Image = self.image(name)
	default:
		self.p.Skip(name)
	}
}

func (self *itemParser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/itunes: xml parser errored: %w", self.p.Err())
	}
	return nil
}

func (self *itemParser) image(name string) (href string) {
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
