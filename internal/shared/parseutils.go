package shared

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"

	xpp "github.com/dsh2dsh/goxpp/v2"
)

var (
	emailNameRgx = regexp.MustCompile(`^([^@]+@[^\s]+)\s+\(([^@]+)\)$`)
	nameEmailRgx = regexp.MustCompile(`^([^@]+)\s+\(([^@]+@[^)]+)\)$`)
	nameOnlyRgx  = regexp.MustCompile(`^([^@()]+)$`)
	emailOnlyRgx = regexp.MustCompile(`^([^@()]+@[^@()]+)$`)
)

const (
	CDATA_START = "<![CDATA["
	CDATA_END   = "]]>"
)

// FindRoot iterates through the tokens of an xml document until
// it encounters its first StartTag event.  It returns an error
// if it reaches EndDocument before finding a tag.
func FindRoot(p *xpp.XMLPullParser) (event xpp.XMLEventType, err error) {
	for {
		event, err = p.Next()
		if err != nil {
			return event, fmt.Errorf("gofeed/internal/shared: %w", err)
		}
		if event == xpp.StartTag {
			break
		}

		if event == xpp.EndDocument {
			return event, errors.New("failed to find root node before document end")
		}
	}
	return event, nil
}

// StripCDATA removes CDATA tags from the string
// content outside of CDATA tags is passed via DecodeEntities
func StripCDATA(str string) string {
	var buf strings.Builder
	buf.Grow(len(str))

	curr := 0

	for curr < len(str) {

		start := indexAt(str, CDATA_START, curr)

		if start == -1 {
			buf.WriteString(html.UnescapeString(str[curr:]))
			return buf.String()
		}

		end := indexAt(str, CDATA_END, start)

		if end == -1 {
			buf.WriteString(html.UnescapeString(str[curr:]))
			return buf.String()
		}

		buf.WriteString(str[start+len(CDATA_START) : end])

		curr = curr + end + len(CDATA_END)
	}

	return buf.String()
}

func indexAt(str, substr string, start int) int {
	idx := strings.Index(str[start:], substr)
	if idx > -1 {
		idx += start
	}
	return idx
}

// ParseNameAddress parses name/email strings commonly
// found in RSS feeds of the format "Example Name (example@site.com)"
// and other variations of this format.
func ParseNameAddress(nameAddressText string) (name, address string) {
	if nameAddressText == "" {
		return "", ""
	}

	if m := emailNameRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[2], m[1]
	}

	if m := nameEmailRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[1], m[2]
	}

	if m := nameOnlyRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[1], ""
	}

	if m := emailOnlyRgx.FindStringSubmatch(nameAddressText); m != nil {
		return "", m[1]
	}
	return "", ""
}
