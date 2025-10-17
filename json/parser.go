package json

import (
	"bytes"
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
func (ap *Parser) Parse(feed io.Reader, opts ...options.Option) (*Feed, error) {
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(feed); err != nil {
		return nil, fmt.Errorf("gofeed/json: %w", err)
	}

	jsonFeed := &Feed{}
	err := json.Unmarshal(buffer.Bytes(), jsonFeed)
	if err != nil {
		return nil, fmt.Errorf("gofeed/json: unable unmarshal feed: %w", err)
	}
	return jsonFeed, nil
}
