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

		fmt.Printf("Testing %s... ", name)

		// Get actual source feed
		ff := fmt.Sprintf("testdata/%s.xml", name)
		f, _ := os.ReadFile(ff)

		// Parse actual feed
		fp := rss.NewParser()
		actual, _ := fp.Parse(bytes.NewReader(f), nil)

		// Get json encoded expected feed result
		ef := fmt.Sprintf("testdata/%s.json", name)
		e, _ := os.ReadFile(ef)

		// Unmarshal expected feed
		expected := &rss.Feed{}
		json.Unmarshal(e, &expected)

		if assert.Equal(t, expected, actual, "Feed file %s.xml did not match expected output %s.json", name, name) {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("Failed\n")
		}
	}
}

// TODO: Examples
