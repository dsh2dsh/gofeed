package shared

import (
	"regexp"
)

var (
	emailNameRgx = regexp.MustCompile(`^([^@]+@[^\s]+)\s+\(([^@]+)\)$`)
	nameEmailRgx = regexp.MustCompile(`^([^@]+)\s+\(([^@]+@[^)]+)\)$`)
	nameOnlyRgx  = regexp.MustCompile(`^([^@()]+)$`)
	emailOnlyRgx = regexp.MustCompile(`^([^@()]+@[^@()]+)$`)
)

// ParseNameAddress parses name/email strings commonly found in RSS feeds of the
// format "Example Name (example@site.com)" and other variations of this format.
func ParseNameAddress(s string) (name, address string) {
	if s == "" {
		return "", ""
	}

	if m := emailNameRgx.FindStringSubmatch(s); m != nil {
		return m[2], m[1]
	}

	if m := nameEmailRgx.FindStringSubmatch(s); m != nil {
		return m[1], m[2]
	}

	if m := nameOnlyRgx.FindStringSubmatch(s); m != nil {
		return m[1], ""
	}

	if m := emailOnlyRgx.FindStringSubmatch(s); m != nil {
		return "", m[1]
	}
	return s, ""
}
