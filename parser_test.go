package gofeed_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2"
)

func TestParser_Parse(t *testing.T) {
	feedTests := []struct {
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
	}

	for _, test := range feedTests {
		fmt.Printf("Testing %s... ", test.file)

		// Get feed content
		path := "testdata/parser/universal/" + test.file
		f, _ := os.ReadFile(path)

		// Get actual value
		fp := gofeed.NewParser()
		feed, err := fp.Parse(bytes.NewReader(f), nil)

		if test.hasError {
			require.Error(t, err)
			assert.Nil(t, feed)
		} else {
			assert.NotNil(t, feed)
			require.NoError(t, err)
			assert.Equal(t, feed.FeedType, test.feedType)
			assert.Equal(t, feed.Title, test.feedTitle)
		}
	}
}

func TestParser_ParseString(t *testing.T) {
	feedTests := []struct {
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
		{"unknown_feed.xml", "", "", true},
		{"empty_feed.xml", "", "", true},
		{"invalid.json", "", "", true},
	}

	for _, test := range feedTests {
		fmt.Printf("Testing %s... ", test.file)

		// Get feed content
		path := "testdata/parser/universal/" + test.file
		f, _ := os.ReadFile(path)

		// Get actual value
		fp := gofeed.NewParser()
		feed, err := fp.ParseString(string(f), nil)

		if test.hasError {
			require.Error(t, err)
			assert.Nil(t, feed)
		} else {
			assert.NotNil(t, feed)
			require.NoError(t, err)
			assert.Equal(t, feed.FeedType, test.feedType)
			assert.Equal(t, feed.Title, test.feedTitle)
		}
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
		path := "testdata/parser/universal/" + test
		f, _ := os.ReadFile(path)

		wg.Go(func() { fp.ParseString(string(f), nil) })
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
	feed, err := fp.Parse(strings.NewReader(feedData), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(feed.Title)
}

func ExampleParser_ParseString() {
	feedData := `<rss version="2.0">
<channel>
<title>Sample Feed</title>
</channel>
</rss>`
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(feedData, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(feed.Title)
}
