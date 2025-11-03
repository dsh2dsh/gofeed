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

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("testdata/bench/large_atom.xml")
	require.NoError(b, err)

	b.ReportAllocs()
	for b.Loop() {
		atom.NewParser().Parse(bytes.NewReader(data), nil)
	}
}

// Tests

func TestParser_Parse(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.xml")
	for _, f := range files {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		t.Logf("Testing %s... ", name)

		// Get actual source feed
		f, err := os.ReadFile(fmt.Sprintf("testdata/%s.xml", name))
		require.NoError(t, err)

		// Parse actual feed
		actual, err := atom.NewParser().Parse(bytes.NewReader(f), nil)
		require.NoError(t, err)

		// Get json encoded expected feed result
		e, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", name))
		require.NoError(t, err)

		// Unmarshal expected feed
		var expected atom.Feed
		json.Unmarshal(e, &expected)

		ok := assert.Equal(t, &expected, actual,
			"Feed file %s.xml did not match expected output %s.json", name, name)
		if ok {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("Failed\n")
		}
	}
}

// TODO: Examples
