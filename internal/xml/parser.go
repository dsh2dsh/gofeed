package xml

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"
)

type Parser struct {
	*xpp.XMLPullParser

	err error
}

func NewParser(p *xpp.XMLPullParser) *Parser {
	return &Parser{XMLPullParser: p}
}

func (self *Parser) Err() error { return self.err }

// FindRoot iterates through the tokens of an xml document until it encounters
// its first StartTag event. It returns an error if it reaches EndDocument
// before finding a tag.
func (self *Parser) FindRoot() (event xpp.XMLEventType, err error) {
	for {
		event, err = self.XMLPullParser.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/internal/xml: looking for root: %w", err)
		}

		if event == xpp.StartTag {
			break
		} else if event == xpp.EndDocument {
			return event, errors.New(
				"gofeed/internal/xml: failed to find root node before document end")
		}
	}
	return event, nil
}

// Text is a helper function for parsing the text from the current element of
// the XMLPullParser.
func (self *Parser) Text() string {
	s, err := self.NextText()
	if err != nil {
		self.err = fmt.Errorf("gofeed/internal/xml: parse text: %w", err)
		return ""
	}
	return strings.TrimSpace(s)
}

func (self *Parser) Skip(tag string) {
	if err := self.XMLPullParser.Skip(); err != nil {
		self.err = fmt.Errorf(
			"gofeed/internal/xml: skip unknown element %q: %w", tag, err)
	}
}

func (self *Parser) Expect(event xpp.XMLEventType, name string) error {
	if err := self.XMLPullParser.Expect(event, name); err != nil {
		return fmt.Errorf("gofeed/internal/xml: expect %q tag, got %q: %w",
			name, self.Name, err)
	}
	return nil
}

// Next iterates through the tokens until it reaches a StartTag or EndTag.
//
// Next is similar to goxpp's NextTag method except it wont throw an error if
// the next immediate token isnt a Start/EndTag. Instead, it will continue to
// consume tokens until it hits a Start/EndTag or EndDocument.
func (self *Parser) Next() (xpp.XMLEventType, error) {
	if self.err != nil {
		return 0, self.err
	}

	for {
		event, err := self.XMLPullParser.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/internal/xml: looking for next tag: %w",
				err)
		}

		switch event {
		case xpp.EndTag:
			return event, nil
		case xpp.StartTag:
			return event, nil
		case xpp.EndDocument:
			return event, errors.New(
				"gofeed/internal/xml: looking for next tag, got unexpected end of the document")
		}
	}
}

func (self *Parser) ParsingElement(name string, init func() error,
	yield func() error,
) error {
	if err := self.Expect(xpp.StartTag, name); err != nil {
		return err
	}

	if init != nil {
		if err := init(); err != nil {
			return err
		}
	}

	for {
		event, err := self.Next()
		if err != nil {
			return fmt.Errorf("gofeed/internal/xml: next %q element: %w", name, err)
		} else if event == xpp.EndTag {
			break
		}

		if err := yield(); err != nil {
			return err
		}
	}

	if err := self.Expect(xpp.EndTag, name); err != nil {
		return err
	}
	return nil
}

func (self *Parser) WithText(name string, init func() error,
	yield func(string) error,
) error {
	if err := self.Expect(xpp.StartTag, name); err != nil {
		return err
	}

	if init != nil {
		if err := init(); err != nil {
			return err
		}
	}

	if s := self.Text(); self.err == nil {
		if err := yield(s); err != nil {
			return err
		}
	}

	if self.err != nil {
		return self.err
	}
	return self.Expect(xpp.EndTag, name)
}

func (self *Parser) MakeChildrenSeq(name string) (iter.Seq[string], error) {
	if err := self.Expect(xpp.StartTag, name); err != nil {
		return nil, err
	}

	return func(yield func(string) bool) {
		for {
			event, err := self.Next()
			switch {
			case err != nil:
				self.err = fmt.Errorf("next child of %q element: %w", name, err)
				return
			case event == xpp.EndTag:
				if self.err == nil {
					self.err = self.Expect(xpp.EndTag, name)
				}
				return
			case !yield(strings.ToLower(self.Name)):
				return
			}
		}
	}, nil
}

func (self *Parser) AttributeSeq() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i := range self.Attrs {
			attr := &self.Attrs[i]
			lowerName := strings.ToLower(attr.Name.Local)
			if !yield(lowerName, attr.Value) {
				return
			}
		}
	}
}

func (self *Parser) WithSkip(name string, yield func() error) error {
	if err := self.Expect(xpp.StartTag, name); err != nil {
		return err
	}

	if err := yield(); err != nil {
		return err
	}
	self.Skip(name)

	if self.err != nil {
		return self.err
	}
	return self.Expect(xpp.EndTag, name)
}
