package atom_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2/atom"
	"github.com/dsh2dsh/gofeed/v2/options"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("testdata/bench/large_atom.xml")
	require.NoError(t, err)

	feed, err := atom.NewParser().Parse(bytes.NewReader(data))
	require.NoError(t, err)
	assert.NotNil(t, feed)
}

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("testdata/bench/large_atom.xml")
	require.NoError(b, err)

	var bytesReader bytes.Reader

	b.ReportAllocs()
	for b.Loop() {
		var parser atom.Parser
		bytesReader.Reset(data)
		parser.Parse(&bytesReader, options.WithStrictChars(true))
	}
}

func TestParser_Parse(t *testing.T) {
	processTestFiles(t, "testdata", nil)
}

func processTestFiles(t *testing.T, dirPath string,
	parserFunc func(r io.Reader) (*atom.Feed, error),
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
				atom.Feed

				ErrorContains string `json:"errorContains"`
			}
			require.NoError(t, json.Unmarshal(e, &expected))

			// Get actual source feed
			f, err := os.ReadFile(path.Join(dirPath, name) + ".xml")
			require.NoError(t, err)

			// Parse actual feed
			if parserFunc == nil {
				parserFunc = func(r io.Reader) (*atom.Feed, error) {
					return atom.NewParser().Parse(r)
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
		func(r io.Reader) (*atom.Feed, error) {
			return atom.NewParser().Parse(r, options.WithSkipUnknownElements(true))
		})
}
