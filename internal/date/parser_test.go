package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateRFC822Zones(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		offset int // seconds east of UTC
		hours  int // expected hour after converting to UTC
	}{
		{
			name:   "EDT",
			input:  "Mon, 21 Apr 2025 06:00:00 EDT",
			offset: -4 * 3600,
			hours:  10,
		},
		{
			name:   "CDT",
			input:  "Mon, 21 Apr 2025 06:00:00 CDT",
			offset: -5 * 3600,
			hours:  11,
		},
		{
			name:   "MDT",
			input:  "Mon, 21 Apr 2025 06:00:00 MDT",
			offset: -6 * 3600,
			hours:  12,
		},
		{
			name:   "PDT",
			input:  "Mon, 21 Apr 2025 06:00:00 PDT",
			offset: -7 * 3600,
			hours:  13,
		},
		{
			name:   "EST",
			input:  "Mon, 21 Apr 2025 06:00:00 EST",
			offset: -5 * 3600,
			hours:  11,
		},
		{
			name:   "CST",
			input:  "Mon, 21 Apr 2025 06:00:00 CST",
			offset: -6 * 3600,
			hours:  12,
		},
		{
			name:   "MST",
			input:  "Mon, 21 Apr 2025 06:00:00 MST",
			offset: -7 * 3600,
			hours:  13,
		},
		{
			name:   "PST",
			input:  "Mon, 21 Apr 2025 06:00:00 PST",
			offset: -8 * 3600,
			hours:  14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.input)
			require.NoError(t, err)

			_, offset := parsed.Zone()
			assert.Equal(t, tt.offset, offset)
			assert.Equal(t, 6, parsed.Hour())
			assert.Equal(t, tt.hours, parsed.UTC().Hour())
		})
	}
}

func TestParseDateNumericOffsetUnchanged(t *testing.T) {
	// Numeric offsets must not be affected by the RFC 822 zone fix.
	input := "Mon, 21 Apr 2025 06:00:00 -0400"
	parsed, err := Parse(input)
	require.NoError(t, err)

	_, offset := parsed.Zone()
	assert.Equal(t, -4*3600, offset)
}

func TestParseDateGMTUnchanged(t *testing.T) {
	// GMT should still parse as offset 0.
	input := "Mon, 21 Apr 2025 06:00:00 GMT"
	parsed, err := Parse(input)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 21, 6, 0, 0, 0, time.UTC), parsed.UTC())
}
