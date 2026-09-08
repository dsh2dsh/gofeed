package json_test

import (
	"bytes"
	jsonEnc "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2/json"
)

func TestParser_Parse(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.json")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		if strings.HasSuffix(name, "_expected") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s... ", name)

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", name))
			require.NoError(t, err)

			// Parse actual feed
			fp := json.NewParser()
			actual, err := fp.Parse(bytes.NewReader(f), nil)
			require.NoError(t, err)

			// Get json encoded expected feed result
			e, err := os.ReadFile(fmt.Sprintf("testdata/%s_expected.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected json.Feed
			require.NoError(t, jsonEnc.Unmarshal(e, &expected))
			assert.Equal(t, &expected, actual)
		})
	}
}

// TODO: Remove redundant tests
func TestParser_ParseInvalidAndStruct(t *testing.T) {
	name := "invalid"
	fmt.Printf("Testing %s... ", name)

	// Get actual source feed
	ff := fmt.Sprintf("testdata/invalid/%s.json", name)
	fmt.Println(ff)
	f, _ := os.ReadFile(ff)

	// Parse actual feed
	fp := json.NewParser()
	_, err := fp.Parse(bytes.NewReader(f), nil)
	require.Error(t, err)

	name = "version_json_10"
	fmt.Printf("Testing %s... ", name)

	// Get actual source feed
	ff = fmt.Sprintf("testdata/%s.json", name)
	fmt.Println(ff)
	f, _ = os.ReadFile(ff)

	// Parse actual feed
	actual, _ := fp.Parse(bytes.NewReader(f), nil)

	assert.Equal(t, "1.0", actual.Version)
	assert.Equal(t, "title", actual.Title)
	assert.Equal(t, "https://sample-json-feed.com", actual.HomePageURL)
	assert.Equal(t, "https://sample-json-feed.com/feed.json", actual.FeedURL)
	assert.Equal(t, "description", actual.Description)
	assert.Equal(t, "user_comment", actual.UserComment)
	assert.Equal(t, "https://sample-json-feed.com/feed.json?next=500", actual.NextURL)
	assert.Equal(t, "https://sample-json-feed.com/icon.png", actual.Icon)
	assert.Equal(t, "https://sample-json-feed.com/favicon.png", actual.Favicon)
	assert.Equal(t, "author_name", actual.Author.Name)
	assert.Equal(t, "https://sample-feed-author.com", actual.Author.URL)
	assert.Equal(t, "https://sample-feed-author.com/me.png", actual.Author.Avatar)
	assert.False(t, actual.Expired)
	assert.Equal(t, "id", actual.Items[0].ID)
	assert.Equal(t, "https://sample-json-feed.com/id", actual.Items[0].URL)
	assert.Equal(t, "https://sample-json-feed.com/external", actual.Items[0].ExternalURL)
	assert.Equal(t, "title", actual.Items[0].Title)
	assert.Contains(t, actual.Items[0].ContentHTML, "content_html")
	assert.Equal(t, "content_text", actual.Items[0].ContentText)
	assert.Equal(t, "summary", actual.Items[0].Summary)
	assert.Equal(t, "https://sample-json-feed.com/image.png", actual.Items[0].Image)
	assert.Equal(t, "https://sample-json-feed.com/banner_image.png", actual.Items[0].BannerImage)
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Items[0].DatePublished)
	assert.Equal(t, "2019-10-12T07:20:50.52Z", actual.Items[0].DateModified)
	assert.Equal(t, "author_name", actual.Items[0].Author.Name)
	assert.Equal(t, "https://sample-feed-author.com", actual.Items[0].Author.URL)
	assert.Equal(t, "https://sample-feed-author.com/me.png", actual.Items[0].Author.Avatar)
	assert.Equal(t, "tag1", actual.Items[0].Tags[0])
	assert.Equal(t, "tag2", actual.Items[0].Tags[1])
	assert.Equal(t, "https://sample-json-feed.com/attachment", (*actual.Items[0].Attachments)[0].URL)
	assert.Equal(t, "audio/mpeg", (*actual.Items[0].Attachments)[0].MimeType)
	assert.Equal(t, "title", (*actual.Items[0].Attachments)[0].Title)
	assert.Equal(t, int64(100), (*actual.Items[0].Attachments)[0].SizeInBytes)
	assert.Equal(t, int64(100), (*actual.Items[0].Attachments)[0].DurationInSeconds)

	assert.Contains(t, actual.String(), "https://sample-json-feed.com/attachment")
}

// An I/O error from the reader must surface as itself, not as a misleading JSON
// syntax error from a truncated buffer (issue #311).
func TestParser_Parse_ReaderError(t *testing.T) {
	boom := errors.New("boom")
	r := io.MultiReader(
		strings.NewReader(`{"version":"https://jsonfeed.org/version/1"`),
		iotest.ErrReader(boom))

	_, err := json.NewParser().Parse(r)
	require.ErrorIs(t, err, boom)
}

func TestParser_NullOptionalJSONFields(t *testing.T) {
	for i, items := range [...]string{
		"null",
		`[{"id":"a","content_text":"text","author":null,"authors":null}]`,
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			data := `{"version":"https://jsonfeed.org/version/1.1","title":"Null fields","author":null,"authors":null,"items":` + items + `}`
			feed, err := json.NewParser().Parse(strings.NewReader(data))
			require.NoError(t, err)
			require.NotNil(t, feed)
			require.Empty(t, feed.Authors)
			if items == "null" {
				require.Empty(t, feed.Items)
				return
			}

			require.Len(t, feed.Items, 1)
			require.Equal(t, "a", feed.Items[0].ID)
			require.Equal(t, "text", feed.Items[0].Content())
			require.Empty(t, feed.Items[0].Authors)
		})
	}
}
