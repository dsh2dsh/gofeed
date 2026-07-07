package ext_test

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

	"github.com/dsh2dsh/gofeed/v2"
)

func TestParse(t *testing.T) {
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
				gofeed.Feed

				ErrorContains string `json:"errorContains"`
			}
			require.NoError(t, json.Unmarshal(e, &expected))

			// Get actual source feed
			f, err := os.ReadFile(fmt.Sprintf("testdata/%s.xml", name))
			require.NoError(t, err)

			// Parse actual feed
			feed, err := gofeed.NewParser().Parse(bytes.NewReader(f))

			if expected.ErrorContains != "" {
				t.Log(err)
				require.ErrorContains(t, err, expected.ErrorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, &expected.Feed, feed,
				"Feed file %s.xml did not match expected output %s.json", name, name)
		})
	}
}
