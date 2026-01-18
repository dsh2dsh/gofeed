package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/dsh2dsh/gofeed/v2/options"
)

// Parser is an JSON Feed Parser
type Parser struct{}

// NewParser creates a new JSON Feed parser
func NewParser() *Parser { return &Parser{} }

// Parse parses an json feed into an json.Feed
func (ap *Parser) Parse(r io.Reader, opts ...options.Option) (*Feed, error) {
	feed := parseFeed{Feed: new(Feed)}
	if err := json.NewDecoder(r).Decode(&feed); err != nil {
		return nil, fmt.Errorf("gofeed/json: unable unmarshal feed: %w", err)
	}
	return feed.Export(), nil
}

type parseFeed struct {
	*Feed

	Authors arrayOrSingle[Author] `json:"authors,omitempty"`
	Items   []*parseItem          `json:"items,omitempty"`
}

type parseItem struct {
	*Item

	Authors arrayOrSingle[Author] `json:"authors,omitempty"`
}

func (self *parseItem) Export() *Item {
	item := self.Item
	item.Authors = self.Authors
	return item
}

func (self *parseFeed) Export() *Feed {
	self.exportItems()
	feed := self.Feed
	feed.Authors = self.Authors
	return feed
}

func (self *parseFeed) exportItems() {
	if len(self.Items) == 0 {
		return
	}

	self.Feed.Items = make([]*Item, len(self.Items))
	for i, item := range self.Items {
		self.Feed.Items[i] = item.Export()
	}
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
