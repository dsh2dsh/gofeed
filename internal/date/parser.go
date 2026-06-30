package date

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/itlightning/dateparse"
)

// ParseDate parses a given date string using a large list of commonly found
// feed date formats.
func Parse(ds string) (time.Time, error) {
	ds = strings.TrimSpace(ds)
	if ds == "" {
		return time.Time{}, errors.New("date string is empty")
	}

	t, err := dateparse.ParseAny(ds, dateparse.SimpleErrorMessages(true))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date %q: %w", ds, err)
	}

	tz, offset := t.Zone()
	if offset != 0 || tz == "" || tz == "UTC" {
		return t, nil
	}

	// This is a format match! Now try to load the timezone name
	loc := locationFromCache(tz)
	// We couldn't load the TZ name. Just use UTC instead...
	if loc == time.UTC {
		return t, nil
	}

	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(),
		t.Second(), t.Nanosecond(), loc)
	return t2, nil
}
