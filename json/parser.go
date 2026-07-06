package json

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is an JSON Feed Parser
type Parser struct{}

// NewParser creates a new JSON Feed parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an json feed into an json.Feed
func (ap *Parser) Parse(r io.Reader, opts ...options.Option) (*Feed, error) {
	feed := &Feed{}
	if err := json.NewDecoder(r).Decode(feed); err != nil {
		return nil, fmt.Errorf("gofeed/json: unable unmarshal feed: %w", err)
	}
	return feed, nil
}

var _ json.Unmarshaler = (*Feed)(nil)

func (self *Feed) UnmarshalJSON(b []byte) error {
	type alias Feed
	aux := struct {
		*alias

		Authors arrayOrSingle[Author] `json:"authors,omitempty"`
	}{alias: (*alias)(self)}

	if err := json.Unmarshal(b, &aux); err != nil {
		return fmt.Errorf("unmarshal json feed: %w", err)
	}

	self.Authors = aux.Authors
	return nil
}

var _ json.Unmarshaler = (*Item)(nil)

func (self *Item) UnmarshalJSON(b []byte) error {
	type alias Item
	aux := struct {
		*alias

		ID      asString              `json:"id,omitzero"`
		Authors arrayOrSingle[Author] `json:"authors,omitempty"`
	}{alias: (*alias)(self)}

	if err := json.Unmarshal(b, &aux); err != nil {
		return fmt.Errorf("unmarshal json feed item: %w", err)
	}

	self.ID = aux.ID.Value
	self.Authors = aux.Authors
	return nil
}

type arrayOrSingle[T any] []*T

var _ json.Unmarshaler = (*arrayOrSingle[any])(nil)

func (self *arrayOrSingle[T]) UnmarshalJSON(b []byte) error {
	var items []*T
	err := json.Unmarshal(b, &items)
	if err == nil {
		*self = items
		return nil
	}

	item := new(T)
	if err2 := json.Unmarshal(b, item); err2 != nil {
		return fmt.Errorf(
			"unmarshal array of objects or single object: %w: %w", err, err2)
	}
	*self = append(*self, item)
	return nil
}

type asString struct {
	Value string
}

var _ json.Unmarshaler = (*asString)(nil)

func (self *asString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &self.Value)
	if err == nil {
		return nil
	}

	var raw json.RawMessage
	if err2 := json.Unmarshal(b, &raw); err2 != nil {
		return fmt.Errorf("unmarshal as string value: %w: %w", err, err2)
	}

	self.Value = strings.TrimSpace(string(raw))
	return nil
}
