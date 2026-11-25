package rss_test

import (
	"bytes"
	"encoding/json/v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"

	"github.com/dsh2dsh/gofeed/v2/options"
	"github.com/dsh2dsh/gofeed/v2/rss"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("testdata/bench/large_rss.xml")
	require.NoError(t, err)

	feed, err := rss.NewParser().Parse(bytes.NewReader(data))
	require.NoError(t, err)
	assert.NotNil(t, feed)
}

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("testdata/bench/large_rss.xml")
	require.NoError(b, err)

	var bytesReader bytes.Reader

	b.ReportAllocs()
	for b.Loop() {
		var parser rss.Parser
		bytesReader.Reset(data)
		parser.Parse(&bytesReader, options.WithStrictChars(true))
	}
}

func TestParser_Parse(t *testing.T) {
	processTestFiles(t, "testdata", nil)
}

func processTestFiles(t *testing.T, dirPath string,
	parserFunc func(r io.Reader) (*rss.Feed, error),
) {
	files, _ := filepath.Glob(path.Join(dirPath, "*.xml"))
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s ... ", base)

			// Get json encoded expected feed result
			e, err := os.ReadFile(path.Join(dirPath, name) + ".json")
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected struct {
				rss.Feed

				ErrorContains string `json:"errorContains"`
			}
			require.NoError(t, json.Unmarshal(e, &expected))

			// Get actual source feed
			f, err := os.ReadFile(path.Join(dirPath, name) + ".xml")
			require.NoError(t, err)

			// Parse actual feed
			if parserFunc == nil {
				parserFunc = func(r io.Reader) (*rss.Feed, error) {
					return rss.NewParser().Parse(r)
				}
			}
			actual, err := parserFunc(bytes.NewReader(f))

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

func TestParser_Parse_withSkipUnknownElements(t *testing.T) {
	processTestFiles(t, "testdata/skip_unknown_elements",
		func(r io.Reader) (*rss.Feed, error) {
			return rss.NewParser().Parse(r, options.WithSkipUnknownElements(true))
		})
}

func TestParser_Parse_utf16Encoding(t *testing.T) {
	document := `<?xml version="1.0" encoding="utf-16"?><rss version="2.0"><channel><title>Título</title><item><title>i</title></item></channel></rss>`

	tests := []struct {
		name string
		e    unicode.Endianness
		b    unicode.BOMPolicy
	}{
		{"utf16le BOM", unicode.LittleEndian, unicode.UseBOM},
		{"utf16be BOM", unicode.BigEndian, unicode.UseBOM},
		{"utf16le no BOM", unicode.LittleEndian, unicode.IgnoreBOM},
		{"utf16be no BOM", unicode.BigEndian, unicode.IgnoreBOM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := unicode.UTF16(tt.e, tt.b).
				NewEncoder().
				Bytes([]byte(document))
			require.NoError(t, err, "unable create UTF16 encoder")

			feed, err := rss.NewParser().Parse(bytes.NewReader(data))
			require.NoError(t, err)
			assert.Equal(t, "Título", feed.Title, "Incorrect feed title")
			assert.Len(t, feed.Items, 1, "Incorrect number of items")
		})
	}
}

func TestParser_Parse_utf8BOM(t *testing.T) {
	document := `<?xml version="1.0" encoding="utf-8"?><rss version="2.0"><channel><title>Título</title><item><title>i</title></item></channel></rss>`

	data, err := unicode.UTF8BOM.NewEncoder().Bytes([]byte(document))
	require.NoError(t, err, "unable create UTF-8 encoder")

	feed, err := rss.NewParser().Parse(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, "Título", feed.Title, "Incorrect feed title")
	assert.Len(t, feed.Items, 1, "Incorrect number of items")
}
