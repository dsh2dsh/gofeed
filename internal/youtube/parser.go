package youtube

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	"github.com/dsh2dsh/gofeed/v2/ext"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type parser struct {
	p  *xml.Parser
	yt *ext.Youtube

	err error
}

func Parse(p *xml.Parser, yt *ext.Youtube) (*ext.Youtube, error) {
	if yt == nil {
		yt = &ext.Youtube{}
	}

	self := parser{p: p, yt: yt}
	return self.Parse()
}

func (self *parser) Parse() (*ext.Youtube, error) {
	name := strings.ToLower(self.p.Name)
	self.body(name)
	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.p.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/youtube: unexpected state at the end: %w", err)
	}
	return self.yt, nil
}

func (self *parser) body(name string) {
	switch name {
	case "channelid":
		self.yt.ChannelId = self.p.Text()
	case "videoid":
		self.yt.VideoId = self.p.Text()
	default:
		self.p.Skip(name)
	}
}

func (self *parser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/youtube: xml parser errored: %w",
			self.p.Err())
	}
	return nil
}
