package gofeed_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2"
	"github.com/dsh2dsh/gofeed/v2/options"
	"github.com/dsh2dsh/gofeed/v2/rss"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		file      string
		feedType  string
		feedTitle string
		hasError  bool
	}{
		{"atom03_feed.xml", "atom", "Feed Title", false},
		{"atom10_feed.xml", "atom", "Feed Title", false},
		{"rss_feed.xml", "rss", "Feed Title", false},
		{"rss_feed_bom.xml", "rss", "Feed Title", false},
		{"rss_feed_leading_spaces.xml", "rss", "Feed Title", false},
		{"rdf_feed.xml", "rss", "Feed Title", false},
		{"sample.json", "json", "title", false},
		{"json10_feed.json", "json", "title", false},
		{"json11_feed.json", "json", "title", false},
		{"unknown_feed.xml", "", "", true},
		{"empty_feed.xml", "", "", true},
		{"invalid.json", "", "", true},
		{"invalidutf8.xml", "rss", "Android Authority", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			// Get feed content
			b, err := os.ReadFile(path.Join("testdata/parser/", tt.file))
			require.NoError(t, err)

			// Get actual value
			fp := gofeed.NewParser()
			feed, err := fp.Parse(bytes.NewReader(b))

			if tt.hasError {
				require.Error(t, err)
				assert.Nil(t, feed)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, feed)
			assert.Equal(t, feed.FeedType, tt.feedType)
			assert.Equal(t, feed.Title, tt.feedTitle)
		})
	}
}

// to detect race conditions, run with go test -race
func TestParser_Concurrent(t *testing.T) {
	feedTests := []string{
		"atom03_feed.xml", "atom10_feed.xml", "rss_feed.xml", "rss_feed_bom.xml",
		"rss_feed_leading_spaces.xml", "rdf_feed.xml", "json10_feed.json",
		"json11_feed.json",
	}

	fp := gofeed.NewParser()
	var wg sync.WaitGroup
	for _, test := range feedTests {
		t.Logf("\nTesting concurrently %s... ", test)

		// Get feed content
		path := "testdata/parser/" + test
		f, _ := os.ReadFile(path)
		wg.Go(func() { fp.Parse(bytes.NewReader(f)) })
	}
	wg.Wait()
}

// Examples

func ExampleParser_Parse() {
	feedData := `<rss version="2.0">
<channel>
<title>Sample Feed</title>
</channel>
</rss>`
	fp := gofeed.NewParser()
	feed, err := fp.Parse(strings.NewReader(feedData))
	if err != nil {
		panic(err)
	}
	fmt.Println(feed.Title)
}

// TestParserConcurrentParseString shares one Parser across goroutines. Before
// the lazy-init fix this races on the
// AtomTranslator/RSSTranslator/JSONTranslator fields under -race.
func TestParserConcurrentParseString(t *testing.T) {
	p := gofeed.NewParser()
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			const concurrencyFeed = `<rss version="2.0"><channel><title>t</title><item><title>i</title></item></channel></rss>`
			_, err := p.Parse(strings.NewReader(concurrencyFeed))
			require.NoError(t, err)
		})
	}
	wg.Wait()
}

func TestParserKeepOriginalFeed(t *testing.T) {
	const feed = `<rss version="2.0"><channel><title>t</title><item><title>i</title></item></channel></rss>`

	// Off by default: OriginalFeed() is nil.
	p := gofeed.NewParser()
	f, err := p.Parse(strings.NewReader(feed))
	require.NoError(t, err)
	require.NotNil(t, feed)
	assert.Nil(t, f.OriginalFeed, "want nil when KeepOriginalFeed is off")

	// On: OriginalFeed() returns the source *rss.Feed.
	f, err = p.Parse(strings.NewReader(feed), options.WithKeepOriginalFeed(true))
	require.NoError(t, err)
	require.NotNil(t, feed)

	orig, ok := f.OriginalFeed.(*rss.Feed)
	require.True(t, ok, "want *rss.Feed")
	assert.Equal(t, "t", orig.Title, "original feed title")
}

// An I/O error from the reader must surface as itself, not be masked as a
// failed type detection (issue #311).
func TestParser_Parse_ReaderError(t *testing.T) {
	boom := errors.New("boom")
	r := io.MultiReader(strings.NewReader(`<rss version="2.0"><channel>`),
		iotest.ErrReader(boom))

	_, err := gofeed.NewParser().Parse(r)
	assert.ErrorIs(t, err, boom)
}

// A feed much larger than the detection window must parse completely: the
// format parser reads from the start of the stream, not from after the peek.
func TestParser_Parse_LargeFeed(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`<rss version="2.0"><channel><title>big</title>`)
	for i := range 2000 {
		fmt.Fprintf(&sb,
			`<item><title>item %d</title><guid>g%d</guid></item>`, i, i)
	}
	sb.WriteString(`</channel></rss>`)

	feed, err := gofeed.NewParser().Parse(strings.NewReader(sb.String()))
	require.NoError(t, err)
	require.NotNil(t, feed)
	assert.Len(t, feed.Items, 2000)
	assert.Equal(t, "item 1999", feed.Items[1999].Title)
}

func TestParser_Parse_RootBeyondDetectionWindow(t *testing.T) {
	pad := "<!-- " + strings.Repeat("x", 8192) + " -->"
	feed, err := gofeed.NewParser().Parse(
		strings.NewReader(pad + `<rss version="2.0"><channel></channel></rss>`))
	require.NoError(t, err)
	require.NotNil(t, feed)
	assert.Equal(t, "rss", feed.FeedType)
}
