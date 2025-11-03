package rss_test

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

	"github.com/dsh2dsh/gofeed/v2/rss"
)

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("testdata/bench/large_rss.xml")
	require.NoError(b, err)

	b.ReportAllocs()
	for b.Loop() {
		rss.NewParser().Parse(bytes.NewReader(data), nil)
	}
}

func TestParser_Parse(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.xml")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Run(name, func(t *testing.T) {
			t.Logf("Testing %s... ", name)

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/%s.xml", name))
			require.NoError(t, err)

			// Parse actual feed
			actual, err := rss.NewParser().Parse(bytes.NewReader(f))
			require.NoError(t, err)

			// Get json encoded expected feed result
			e, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", name))
			require.NoError(t, err)

			// Unmarshal expected feed
			var expected rss.Feed
			require.NoError(t, json.Unmarshal(e, &expected))

			assert.Equal(t, &expected, actual,
				"Feed file %s.xml did not match expected output %s.json", name, name)
		})
	}
}
