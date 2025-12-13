package gofeed_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2"
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
	fp.AtomTranslator = &gofeed.DefaultAtomTranslator{}
	fp.RSSTranslator = &gofeed.DefaultRSSTranslator{}
	fp.JSONTranslator = &gofeed.DefaultJSONTranslator{}
	wg := sync.WaitGroup{}
	for _, test := range feedTests {
		fmt.Printf("\nTesting concurrently %s... ", test)

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
