package gofeed_test

import (
	"bytes"
	jsonEncoding "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2"
	"github.com/dsh2dsh/gofeed/v2/atom"
	"github.com/dsh2dsh/gofeed/v2/json"
	"github.com/dsh2dsh/gofeed/v2/rss"
)

func TestDefaultRSSTranslator_Translate(t *testing.T) {
	files, _ := filepath.Glob("testdata/translator/rss/*.xml")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s ... ", base)

			// Get json encoded expected feed result
			e, err := os.ReadFile(fmt.Sprintf("testdata/translator/rss/%s.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected struct {
				gofeed.Feed

				ErrorContains string `json:"errorContains"`
			}
			require.NoError(t, jsonEncoding.Unmarshal(e, &expected))

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/translator/rss/%s.xml", name))
			require.NoError(t, err)

			// Parse actual feed
			rssFeed, err := rss.NewParser().Parse(bytes.NewReader(f))
			require.NoError(t, err)

			var translator gofeed.DefaultRSSTranslator
			actual, err := translator.Translate(rssFeed, nil)

			if expected.ErrorContains != "" {
				t.Log(err)
				require.ErrorContains(t, err, expected.ErrorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, &expected.Feed, actual,
				"Feed file %s.xml did not match expected output %s.json", name, name)
		})
	}
}

func TestDefaultRSSTranslator_Translate_WrongType(t *testing.T) {
	var translator gofeed.DefaultRSSTranslator
	af, err := translator.Translate("wrong type", nil)
	require.Nil(t, af)
	require.Error(t, err)
}

func TestDefaultAtomTranslator_Translate(t *testing.T) {
	files, _ := filepath.Glob("testdata/translator/atom/*.xml")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s ... ", base)

			// Get json encoded expected feed result
			e, err := os.ReadFile(fmt.Sprintf("testdata/translator/atom/%s.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected struct {
				gofeed.Feed

				ErrorContains string `json:"errorContains"`
			}
			require.NoError(t, jsonEncoding.Unmarshal(e, &expected))

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/translator/atom/%s.xml", name))
			require.NoError(t, err)

			// Parse actual feed
			atomFeed, err := atom.NewParser().Parse(bytes.NewReader(f))
			require.NoError(t, err)

			var translator gofeed.DefaultAtomTranslator
			actual, err := translator.Translate(atomFeed, nil)

			if expected.ErrorContains != "" {
				t.Log(err)
				require.ErrorContains(t, err, expected.ErrorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, &expected.Feed, actual,
				"Feed file %s.xml did not match expected output %s.json", name, name)
		})
	}
}

func TestDefaultAtomTranslator_Translate_WrongType(t *testing.T) {
	translator := &gofeed.DefaultAtomTranslator{}
	af, err := translator.Translate("wrong type", nil)
	assert.Nil(t, af)
	assert.Error(t, err)
}

func TestDefaultJSONTranslator_Translate(t *testing.T) {
	files, _ := filepath.Glob("testdata/translator/json/*.json")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		if strings.HasSuffix(name, "expected") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s... ", name)

			// Get actual source feed
			f, err := os.ReadFile(
				fmt.Sprintf("testdata/translator/json/%s.json", name))
			require.NoError(t, err)

			// Parse actual feed
			fp := json.NewParser()
			jsonFeed, err := fp.Parse(bytes.NewReader(f))
			require.NoError(t, err)

			var translator gofeed.DefaultJSONTranslator
			actual, err := translator.Translate(jsonFeed, nil)
			require.NoError(t, err)

			// Get json encoded expected feed result
			e, err := os.ReadFile(
				fmt.Sprintf("testdata/translator/json/%s_expected.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected gofeed.Feed
			require.NoError(t, jsonEncoding.Unmarshal(e, &expected),
				"unmarshal expected json")
			assert.Equal(t, &expected, actual)
		})
	}
}

/*

func TestDefaultJSONTranslator_Translate(t *testing.T) {
	name := "sample"
	fmt.Printf("Testing %s... ", name)

	// Get actual source feed
	ff := fmt.Sprintf("testdata/translator/json/%s.json", name)
	fmt.Println(ff)
	f, _ := ioutil.ReadFile(ff)

	// Parse actual feed
	translator := &gofeed.DefaultJSONTranslator{}
	fp := json.NewParser()
	feed, _ := fp.Parse(bytes.NewReader(f), nil)
	actual, _ := translator.Translate(feed, nil)

	assert.Equal(t, "title", actual.Title)
	assert.Equal(t, "description", actual.Description)
	assert.Equal(t, "https://sample-json-feed.com", actual.Link)
	assert.Equal(t, "https://sample-json-feed.com/feed.json", actual.FeedLink)
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Updated)
	assert.Equal(t, "2019-10-12T07:20:50Z", actual.UpdatedParsed.Format(time.RFC3339))
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Published)
	assert.Equal(t, "2019-10-12T07:20:50Z", actual.PublishedParsed.Format(time.RFC3339))
	assert.Equal(t, "author_name", actual.Author.Name)
	assert.Equal(t, "", actual.Author.Email)
	assert.Equal(t, "", actual.Language)
	assert.Equal(t, "https://sample-json-feed.com/icon.png", actual.Image.URL)
	assert.Equal(t, "", actual.Image.Title)
	assert.Equal(t, "", actual.Copyright)
	assert.Equal(t, "", actual.Generator)
	assert.Equal(t, 0, len(actual.Categories))
	assert.Equal(t, (*ext.DublinCoreExtension)(nil), actual.DublinCoreExt)
	assert.Equal(t, (*ext.ITunesFeedExtension)(nil), actual.ITunesExt)
	assert.Equal(t, ext.Extensions(nil), actual.Extensions)
	assert.Equal(t, "json", actual.FeedType)
	assert.Equal(t, "1.0", actual.FeedVersion)
	assert.Equal(t, "title", actual.Items[0].Title)
	assert.Equal(t, "summary", actual.Items[0].Description)
	assert.Equal(t, "<p>content_html</p>", actual.Items[0].Content)
	assert.Equal(t, "https://sample-json-feed.com/id", actual.Items[0].Link)
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Items[0].Updated)
	assert.Equal(t, "2019-10-12T07:20:50Z", actual.Items[0].UpdatedParsed.Format(time.RFC3339))
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Items[0].Published)
	assert.Equal(t, "2019-10-12T07:20:50Z", actual.Items[0].PublishedParsed.Format(time.RFC3339))
	assert.Equal(t, "author_name", actual.Items[0].Author.Name)
	assert.Equal(t, "", actual.Items[0].Author.Email)
	assert.Equal(t, "id", actual.Items[0].GUID)
	assert.Equal(t, "https://sample-json-feed.com/image.png", actual.Items[0].Image.URL)
	assert.Equal(t, "", actual.Items[0].Image.Title)
	assert.Equal(t, "tag1", actual.Items[0].Categories[0])
	assert.Equal(t, "tag2", actual.Items[0].Categories[1])
	assert.Equal(t, "https://sample-json-feed.com/attachment", (actual.Items[0].Enclosures)[0].URL)
	assert.Equal(t, "100", (actual.Items[0].Enclosures)[0].Length)
	assert.Equal(t, "audio/mpeg", (actual.Items[0].Enclosures)[0].Type)
	assert.Equal(t, (*ext.DublinCoreExtension)(nil), actual.Items[0].DublinCoreExt)
	assert.Equal(t, (*ext.ITunesItemExtension)(nil), actual.Items[0].ITunesExt)
	assert.Equal(t, ext.Extensions(nil), actual.Items[0].Extensions)

	name = "sample2"
	fmt.Printf("Testing %s... ", name)

	// Get actual source feed
	ff = fmt.Sprintf("testdata/translator/json/%s.json", name)
	fmt.Println(ff)
	f, _ = ioutil.ReadFile(ff)

	// Parse actual feed
	feed, _ = fp.Parse(bytes.NewReader(f), nil)
	actual, _ = translator.Translate(feed, nil)

	assert.Equal(t, "content_text", actual.Items[0].Content)
	assert.Equal(t, "https://sample-json-feed.com/banner_image.png", actual.Items[0].Image.URL)

}
*/

func TestDefaultJSONTranslator_Translate_WrongType(t *testing.T) {
	translator := &gofeed.DefaultJSONTranslator{}
	af, err := translator.Translate("wrong type", nil)
	assert.Nil(t, af)
	assert.Error(t, err)
}

// DisableContentImageScan turns off the HTML-parsing fallback that finds a
// first <img> in feed and item content; explicit images are unaffected.
func TestDisableContentImageScan(t *testing.T) {
	feed := `<rss version="2.0"><channel>
		<description><![CDATA[<p><img src="http://example.org/feed.png"/></p>]]></description>
		<item><description><![CDATA[<img src="http://example.org/item.png">]]></description></item>
	</channel></rss>`

	rssFeed, err := rss.NewParser().Parse(strings.NewReader(feed))
	require.NoError(t, err)
	require.NotNil(t, rssFeed)

	var def gofeed.DefaultRSSTranslator
	out, err := def.Translate(rssFeed, nil)
	require.NoError(t, err)
	assert.Nil(t, out.Image)
	assert.Nil(t, out.Items[0].Image)
}
