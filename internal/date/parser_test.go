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

func FuzzParse(f *testing.F) {
	f.Add("2017-12-22T22:09:49+00:00")
	f.Add("Fri, 31 Mar 2023 20:19:00 America/Los_Angeles")
	f.Fuzz(func(t *testing.T, date string) {
		Parse(date)
	})
}

func TestParseEmptyDate(t *testing.T) {
	_, err := Parse("  ")
	require.Error(t, err)
}

func TestParseInvalidDate(t *testing.T) {
	_, err := Parse("invalid")
	require.Error(t, err)
}

func TestParseAtomDate(t *testing.T) {
	date, err := Parse("2017-12-22T22:09:49+00:00")
	require.NoError(t, err)
	assert.Equal(t, int64(1513980589), date.Unix())

	_, offset := date.Zone()
	assert.Zero(t, offset)
}

func TestParseRSSDateGMT(t *testing.T) {
	date, err := Parse("Tue, 03 Jun 2003 09:39:21 GMT")
	require.NoError(t, err)
	assert.Equal(t, time.Unix(1054633161, 0), date)
}

func TestParseRSSDatePST(t *testing.T) {
	date, err := Parse("Wed, 26 Dec 2018 10:00:54 PST")
	require.NoError(t, err)
	assert.Equal(t, int64(1545847254), date.Unix())
	assert.Equal(t, "PST", date.Location().String())

	name, offset := date.Zone()
	assert.Equal(t, "PST", name)
	assert.Equal(t, -28800, offset)
}

func TestParseRSSDateEST(t *testing.T) {
	date, err := Parse("Wed, 10 Feb 2021 22:46:00 EST")
	require.NoError(t, err)
	assert.Equal(t, int64(1613015160), date.Unix())
	assert.Equal(t, "EST", date.Location().String())

	name, offset := date.Zone()
	assert.Equal(t, "EST", name)
	assert.Equal(t, -18000, offset)
}

func TestParseRSSDateOffset(t *testing.T) {
	date, err := Parse("Sun, 28 Oct 2018 13:48:00 +0100")
	require.NoError(t, err)
	assert.Equal(t, int64(1540730880), date.Unix())

	_, offset := date.Zone()
	assert.Equal(t, 3600, offset)
}
