package dublincore

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/shared"
)

type parser struct {
	xpp *xpp.XMLPullParser
	dc  *ext.DublinCoreExtension

	err error
}

func Parse(p *xpp.XMLPullParser, dc *ext.DublinCoreExtension,
) (*ext.DublinCoreExtension, error) {
	if dc == nil {
		dc = &ext.DublinCoreExtension{}
	}

	self := parser{xpp: p, dc: dc}
	return self.Parse()
}

func (self *parser) Parse() (*ext.DublinCoreExtension, error) {
	tag := strings.ToLower(self.xpp.Name)
	switch tag {
	case "title":
		self.dc.Title = self.text()
	case "creator":
		self.dc.Creator = self.text()
	case "author":
		self.dc.Author = self.text()
	case "subject":
		self.dc.Subject = self.text()
	case "description":
		self.dc.Description = self.text()
	case "publisher":
		self.dc.Publisher = self.text()
	case "contributor":
		self.dc.Contributor = self.text()
	case "date":
		self.dc.Date = self.text()
	case "type":
		self.dc.Type = self.text()
	case "format":
		self.dc.Format = self.text()
	case "identifier":
		self.dc.Identifier = self.text()
	case "source":
		self.dc.Source = self.text()
	case "language":
		self.dc.Language = self.text()
	case "relation":
		self.dc.Relation = self.text()
	case "coverage":
		self.dc.Coverage = self.text()
	case "rights":
		self.dc.Rights = self.text()
	default:
		self.skip(tag)
	}

	if self.err != nil {
		return nil, self.err
	}

	if err := self.xpp.Expect(xpp.EndTag, tag); err != nil {
		return nil, fmt.Errorf("gofeed/dublincore: expect end tag %q: %w", tag, err)
	}
	return self.dc, nil
}

func (self *parser) text() string {
	s, err := shared.ParseText(self.xpp)
	if err != nil {
		self.err = err
		return ""
	}
	return s
}

func (self *parser) skip(tag string) {
	if err := self.xpp.Skip(); err != nil {
		self.err = fmt.Errorf(
			"gofeed/dublincore: skip unknown element %q: %w", tag, err)
	}
}
