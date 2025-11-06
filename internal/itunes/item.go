package itunes

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type itemParser struct {
	xpp    *xml.Parser
	itunes *ext.ITunesItemExtension

	err error
}

func ParseItem(p *xml.Parser, itunes *ext.ITunesItemExtension,
) (*ext.ITunesItemExtension, error) {
	if itunes == nil {
		itunes = &ext.ITunesItemExtension{}
	}

	self := itemParser{xpp: p, itunes: itunes}
	return self.Parse()
}

func (self *itemParser) Parse() (*ext.ITunesItemExtension, error) {
	name := strings.ToLower(self.xpp.Name)
	switch name {
	case "author":
		self.itunes.Author = self.xpp.Text()
	case "block":
		self.itunes.Block = self.xpp.Text()
	case "duration":
		self.itunes.Duration = self.xpp.Text()
	case "explicit":
		self.itunes.Explicit = self.xpp.Text()
	case "subtitle":
		self.itunes.Subtitle = self.xpp.Text()
	case "summary":
		self.itunes.Summary = self.xpp.Text()
	case "keywords":
		self.itunes.Keywords = self.xpp.Text()
	case "isClosedCaptioned":
		self.itunes.IsClosedCaptioned = self.xpp.Text()
	case "episode":
		self.itunes.Episode = self.xpp.Text()
	case "season":
		self.itunes.Season = self.xpp.Text()
	case "order":
		self.itunes.Order = self.xpp.Text()
	case "episodeType":
		self.itunes.EpisodeType = self.xpp.Text()
	case "image":
		self.itunes.Image = self.image()
	default:
		self.xpp.Skip(name)
	}

	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.xpp.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/itunes: unexpected state at the end of item: %w", err)
	}
	return self.itunes, nil
}

func (self *itemParser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.xpp.Err() != nil:
		return fmt.Errorf("gofeed/itunes: xml parser errored: %w", self.xpp.Err())
	}
	return nil
}

func (self *itemParser) image() string {
	return self.xpp.Attribute("href")
}
