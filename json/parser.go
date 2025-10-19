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
	feed := &Feed{}
	if err := json.NewDecoder(r).Decode(feed); err != nil {
		return nil, fmt.Errorf("gofeed/json: unable unmarshal feed: %w", err)
	}
	return feed, nil
}
