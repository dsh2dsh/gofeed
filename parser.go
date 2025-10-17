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
	rp             *rss.Parser
	ap             *atom.Parser
	jp             *json.Parser
}

// NewParser creates a universal feed parser.
func NewParser() *Parser {
	fp := Parser{
		rp: rss.NewParser(),
		ap: atom.NewParser(),
		jp: json.NewParser(),
	}
	return &fp
}

// Parse parses a RSS or Atom or JSON feed into
// the universal gofeed.Feed.  It takes an
// io.Reader which should return the xml/json content.
func (f *Parser) Parse(feed io.Reader, opts *options.ParseOptions) (*Feed, error) {
	if opts == nil {
		opts = options.DefaultParseOptions()
	}

	// Wrap the feed io.Reader in a io.TeeReader
	// so we can capture all the bytes read by the
	// DetectFeedType function and construct a new
	// reader with those bytes intact for when we
	// attempt to parse the feeds.
	var buf bytes.Buffer
	tee := io.TeeReader(feed, &buf)
	feedType := DetectFeedType(tee)

	// Glue the read bytes from the detect function
	// back into a new reader
	r := io.MultiReader(&buf, feed)

	switch feedType {
	case FeedTypeAtom:
		return f.parseAtomFeed(r, opts)
	case FeedTypeRSS:
		return f.parseRSSFeed(r, opts)
	case FeedTypeJSON:
		return f.parseJSONFeed(r, opts)
	}

	return nil, ErrFeedTypeNotDetected
}

func (f *Parser) parseAtomFeed(feed io.Reader, opts *options.ParseOptions) (*Feed, error) {
	af, err := f.ap.Parse(feed, opts)
	if err != nil {
		return nil, err
	}

	result, err := f.atomTrans().Translate(af, opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: %w", err)
	}

	if opts != nil && opts.KeepOriginalFeed {
		result.OriginalFeed = af
	}

	return result, nil
}

func (f *Parser) parseRSSFeed(feed io.Reader, opts *options.ParseOptions) (*Feed, error) {
	rf, err := f.rp.Parse(feed, opts)
	if err != nil {
		return nil, err
	}

	result, err := f.rssTrans().Translate(rf, opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: %w", err)
	}

	if opts != nil && opts.KeepOriginalFeed {
		result.OriginalFeed = rf
	}

	return result, nil
}

func (f *Parser) parseJSONFeed(feed io.Reader, opts *options.ParseOptions) (*Feed, error) {
	jf, err := f.jp.Parse(feed, opts)
	if err != nil {
		return nil, err
	}

	result, err := f.jsonTrans().Translate(jf, opts)
	if err != nil {
		return nil, fmt.Errorf("gofeed: %w", err)
	}

	if opts != nil && opts.KeepOriginalFeed {
		result.OriginalFeed = jf
	}

	return result, nil
}

func (f *Parser) atomTrans() Translator {
	if f.AtomTranslator != nil {
		return f.AtomTranslator
	}
	f.AtomTranslator = &DefaultAtomTranslator{}
	return f.AtomTranslator
}

func (f *Parser) rssTrans() Translator {
	if f.RSSTranslator != nil {
		return f.RSSTranslator
	}
	f.RSSTranslator = &DefaultRSSTranslator{}
	return f.RSSTranslator
}

func (f *Parser) jsonTrans() Translator {
	if f.JSONTranslator != nil {
		return f.JSONTranslator
	}
	f.JSONTranslator = &DefaultJSONTranslator{}
	return f.JSONTranslator
}
