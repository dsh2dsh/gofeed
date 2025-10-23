package gofeed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/dsh2dsh/gofeed/v2/atom"
	"github.com/dsh2dsh/gofeed/v2/json"
	"github.com/dsh2dsh/gofeed/v2/options"
	"github.com/dsh2dsh/gofeed/v2/rss"
)

// ErrFeedTypeNotDetected is returned when the detection system can not figure
// out the Feed format
var ErrFeedTypeNotDetected = errors.New("failed to detect feed type")

// Parser is a universal feed parser that detects
// a given feed type, parsers it, and translates it
// to the universal feed type.
type Parser struct {
	AtomTranslator Translator
	RSSTranslator  Translator
	JSONTranslator Translator

	opts *options.Parse
}

// NewParser creates a universal feed parser.
func NewParser() *Parser { return &Parser{} }

// Parse parses a RSS or Atom or JSON feed into the universal gofeed.Feed. It
// takes an io.Reader which should return the xml/json content.
func (f *Parser) Parse(feed io.Reader, opts ...options.Option) (*Feed, error) {
	f.opts = options.Default().Apply(opts...)

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(feed); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFeedTypeNotDetected, err)
	}
	feedType := DetectFeedBytes(buf.Bytes())

	switch feedType {
	case FeedTypeAtom:
		return f.parseAtomFeed(&buf)
	case FeedTypeRSS:
		return f.parseRSSFeed(&buf)
	case FeedTypeJSON:
		return f.parseJSONFeed(&buf)
	}
	return nil, ErrFeedTypeNotDetected
}

func (f *Parser) parseAtomFeed(feed io.Reader) (*Feed, error) {
	af, err := atom.NewParser().Parse(feed, options.From(f.opts))
	if err != nil {
		return nil, err
	}

	tr := f.AtomTranslator
	if tr == nil {
		tr = &DefaultAtomTranslator{}
	}

	result, err := tr.Translate(af, f.opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: translate atom: %w", err)
	}

	if f.keepOriginalFeed() {
		result.OriginalFeed = af
	}
	return result, nil
}

func (f *Parser) keepOriginalFeed() bool { return f.opts.KeepOriginalFeed }

func (f *Parser) parseRSSFeed(feed io.Reader) (*Feed, error) {
	rf, err := rss.NewParser().Parse(feed, options.From(f.opts))
	if err != nil {
		return nil, err
	}

	tr := f.RSSTranslator
	if tr == nil {
		tr = &DefaultRSSTranslator{}
	}

	result, err := tr.Translate(rf, f.opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: translate rss: %w", err)
	}

	if f.keepOriginalFeed() {
		result.OriginalFeed = rf
	}
	return result, nil
}

func (f *Parser) parseJSONFeed(feed io.Reader) (*Feed, error) {
	jf, err := json.NewParser().Parse(feed, options.From(f.opts))
	if err != nil {
		return nil, err
	}

	tr := f.JSONTranslator
	if tr == nil {
		tr = &DefaultJSONTranslator{}
	}

	result, err := tr.Translate(jf, f.opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: translate json: %w", err)
	}

	if f.keepOriginalFeed() {
		result.OriginalFeed = jf
	}
	return result, nil
}
