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

// ParseNameAddress parses name/email strings commonly
// found in RSS feeds of the format "Example Name (example@site.com)"
// and other variations of this format.
func ParseNameAddress(nameAddressText string) (name, address string) {
	if nameAddressText == "" {
		return "", ""
	}

	if m := emailNameRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[2], m[1]
	}

	if m := nameEmailRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[1], m[2]
	}

	if m := nameOnlyRgx.FindStringSubmatch(nameAddressText); m != nil {
		return m[1], ""
	}

	if m := emailOnlyRgx.FindStringSubmatch(nameAddressText); m != nil {
		return "", m[1]
	}
	return "", ""
}
