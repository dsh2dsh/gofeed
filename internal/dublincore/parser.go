package dublincore

import (
	"fmt"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"

	"github.com/dsh2dsh/gofeed/v2/ext"
	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

type parser struct {
	p  *xml.Parser
	dc *ext.DublinCoreExtension

	err error
}

func Parse(p *xml.Parser, dc *ext.DublinCoreExtension,
) (*ext.DublinCoreExtension, error) {
	if dc == nil {
		dc = &ext.DublinCoreExtension{}
	}

	self := parser{p: p, dc: dc}
	return self.Parse()
}

func (self *parser) Parse() (*ext.DublinCoreExtension, error) {
	name := strings.ToLower(self.p.Name)
	self.body(name)
	if err := self.Err(); err != nil {
		return nil, err
	}

	if err := self.p.Expect(xpp.EndTag, name); err != nil {
		return nil, fmt.Errorf(
			"gofeed/dublincore: unexpected state at the end: %w", err)
	}
	return self.dc, nil
}

func (self *parser) body(name string) {
	switch name {
	case "title":
		self.dc.Title = self.p.Text()
	case "creator":
		self.dc.Creator = self.p.Text()
	case "author":
		self.dc.Author = self.p.Text()
	case "subject":
		self.dc.Subject = self.p.Text()
	case "description":
		self.dc.Description = self.p.Text()
	case "publisher":
		self.dc.Publisher = self.p.Text()
	case "contributor":
		self.dc.Contributor = self.p.Text()
	case "date":
		self.dc.Date = self.p.Text()
	case "type":
		self.dc.Type = self.p.Text()
	case "format":
		self.dc.Format = self.p.Text()
	case "identifier":
		self.dc.Identifier = self.p.Text()
	case "source":
		self.dc.Source = self.p.Text()
	case "language":
		self.dc.Language = self.p.Text()
	case "relation":
		self.dc.Relation = self.p.Text()
	case "coverage":
		self.dc.Coverage = self.p.Text()
	case "rights":
		self.dc.Rights = self.p.Text()
	default:
		self.p.Skip(name)
	}
}

func (self *parser) Err() error {
	switch {
	case self.err != nil:
		return self.err
	case self.p.Err() != nil:
		return fmt.Errorf("gofeed/dublincore: xml parser errored: %w",
			self.p.Err())
	}
	return nil
}
