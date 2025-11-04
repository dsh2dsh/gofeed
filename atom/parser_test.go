package atom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsh2dsh/gofeed/v2/atom"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("testdata/bench/large_atom.xml")
	require.NoError(t, err)

	feed, err := atom.NewParser().Parse(bytes.NewReader(data), nil)
	require.NoError(t, err)
	assert.NotNil(t, feed)
}

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("testdata/bench/large_atom.xml")
	require.NoError(b, err)

	b.ReportAllocs()
	for b.Loop() {
		atom.NewParser().Parse(bytes.NewReader(data), nil)
	}
}

func TestParser_Parse(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.xml")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s ... ", base)

			// Get json encoded expected feed result
			e, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected struct {
				atom.Feed

				ErrorContains string `json:"errorContains"`
			}
			json.Unmarshal(e, &expected)

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/%s.xml", name))
			require.NoError(t, err)

			// Parse actual feed
			actual, err := atom.NewParser().Parse(bytes.NewReader(f), nil)

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
