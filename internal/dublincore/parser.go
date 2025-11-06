package dublincore

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	ext "github.com/dsh2dsh/gofeed/v2/extensions"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type parser struct {
	xpp *xml.Parser
	dc  *ext.DublinCoreExtension

	err error
}

func Parse(p *xml.Parser, dc *ext.DublinCoreExtension,
) (*ext.DublinCoreExtension, error) {
	if dc == nil {
		dc = &ext.DublinCoreExtension{}
	}

	self := parser{xpp: p, dc: dc}
	return self.Parse()
}

func (self *parser) Parse() (*ext.DublinCoreExtension, error) {
	name := strings.ToLower(self.xpp.Name)
	switch name {
	case "title":
		self.dc.Title = self.xpp.Text()
	case "creator":
		self.dc.Creator = self.xpp.Text()
	case "author":
		self.dc.Author = self.xpp.Text()
	case "subject":
		self.dc.Subject = self.xpp.Text()
	case "description":
		self.dc.Description = self.xpp.Text()
	case "publisher":
		self.dc.Publisher = self.xpp.Text()
	case "contributor":
		self.dc.Contributor = self.xpp.Text()
	case "date":
		self.dc.Date = self.xpp.Text()
	case "type":
		self.dc.Type = self.xpp.Text()
	case "format":
		self.dc.Format = self.xpp.Text()
	case "identifier":
		self.dc.Identifier = self.xpp.Text()
	case "source":
		self.dc.Source = self.xpp.Text()
	case "language":
		self.dc.Language = self.xpp.Text()
	case "relation":
		self.dc.Relation = self.xpp.Text()
	case "coverage":
		self.dc.Coverage = self.xpp.Text()
	case "rights":
		self.dc.Rights = self.xpp.Text()
	default:
		self.xpp.Skip(name)
	}

	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.xpp.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/dublincore: unexpected state at the end: %w", err)
	}
	return self.dc, nil
}

func (self *parser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.xpp.Err() != nil:
		return fmt.Errorf("gofeed/dublincore: xml parser errored: %w",
			self.xpp.Err())
	}
	return nil
}
