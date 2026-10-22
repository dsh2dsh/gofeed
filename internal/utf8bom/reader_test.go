package utf8bom

import (
	"bufio"
	"bytes"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReader(t *testing.T) {
	const body = `{
		"version": "https://jsonfeed.org/version/1.1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"items": []
	}`
	data := slices.Concat(utf8BOM[:], []byte(body))

	tests := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bufio.Reader",
			reader: bufio.NewReader(bytes.NewReader(data)),
		},
		{
			name:   "bytes.Buffer",
			reader: bytes.NewBuffer(data),
		},
		{
			name:   "bytes.Reader",
			reader: bytes.NewReader(data),
		},
		{
			name:   "strings.Reader",
			reader: strings.NewReader(string(data)),
		},
		{
			name:   "default",
			reader: io.NopCloser(bytes.NewReader(data)),
		},
		{
			name:   "without utf8 BOM",
			reader: strings.NewReader(body),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReader(tt.reader)
			require.NoError(t, err)
			require.NotNil(t, r)

			got, err := io.ReadAll(r)
			require.NoError(t, err)
			assert.JSONEq(t, body, string(got))
		})
	}
}
