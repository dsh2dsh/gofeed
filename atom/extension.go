package atom

import (
	"strings"

	"github.com/dsh2dsh/gofeed/v2/internal/xml"
	"github.com/dsh2dsh/gofeed/v2/options"
)

type ExtensionParser struct {
	atom Parser
	p    *xml.Parser
}

func NewExtension(p *xml.Parser, opts ...options.Option) *ExtensionParser {
	self := &ExtensionParser{
		atom: Parser{p: p},
		p:    p,
	}
	self.atom.opts.Apply(opts...)
	return self
}

func (self *ExtensionParser) ParseFeed(feed *Feed) (*Feed, error) {
	if feed == nil {
		feed = &Feed{}
	}
	self.atom.feed = feed
	self.atom.feedBody(strings.ToLower(self.p.Name))
	return feed, self.atom.Err()
}

func (self *ExtensionParser) ParseEntry(entry *Entry) (*Entry, error) {
	if entry == nil {
		entry = &Entry{}
	}

	self.atom.entryBody(strings.ToLower(self.p.Name), entry)
	return entry, self.atom.Err()
}
