package gofeed

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/dsh2dsh/gofeed/v2/internal/xml"
)

// FeedType represents one of the possible feed
// types that we can detect.
type FeedType int

const (
	// FeedTypeUnknown represents a feed that could not have its
	// type determiend.
	FeedTypeUnknown FeedType = iota
	// FeedTypeAtom repesents an Atom feed
	FeedTypeAtom
	// FeedTypeRSS represents an RSS feed
	FeedTypeRSS
	// FeedTypeJSON represents a JSON feed
	FeedTypeJSON
)

// DetectFeedType attempts to determine the type of feed by looking for specific
// xml elements, unique to the various feed types. Internally just reads
// everything and calls [DetectFeedBytes].
func DetectFeedType(feed io.Reader) FeedType {
	var buffer bytes.Buffer
	buffer.ReadFrom(feed) //nolint:errcheck // upstream ignores err
	return DetectFeedBytes(buffer.Bytes())
}

// DetectFeedBytes attempts to determine the type of feed by looking for
// specific xml elements, unique to the various feed types.
func DetectFeedBytes(b []byte) FeedType {
	var firstChar byte
loop:
	for i, ch := range b {
		// ignore leading whitespace & byte order marks
		if unicode.IsSpace(rune(ch)) {
			continue
		}

		switch ch {
		case 0xFE, 0xFF, 0x00, 0xEF, 0xBB, 0xBF: // utf 8-16-32 bom
		default:
			firstChar = ch
			b = b[i:]
			break loop
		}
	}

	switch firstChar {
	case '<':
		// Check if it's an XML based feed
		p := xml.NewParser(bytes.NewReader(b))

		if _, err := p.FindRoot(); err != nil {
			return FeedTypeUnknown
		}

		switch strings.ToLower(p.Name) {
		case "rdf", "rss":
			return FeedTypeRSS
		case "feed":
			return FeedTypeAtom
		}
	case '{':
		// Check if document is valid JSON
		if json.Valid(b) {
			return FeedTypeJSON
		}
	}
	return FeedTypeUnknown
}
