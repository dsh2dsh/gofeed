package parsers

import (
	"net/mail"
	"regexp"
)

var nameAddressRe = []struct {
	re         *regexp.Regexp
	nameIndex  int
	emailIndex int
}{
	// John Doe (user@localhost)
	{regexp.MustCompile(`^([^@]+)\s+\(([^@]+@[^)]+)\)$`), 1, 2},
}

// ParseNameAddress parses name/email strings commonly found in RSS feeds of the
// format "Example Name (example@site.com)" and other variations of this format.
func ParseNameAddress(s string) (name, address string) {
	if s == "" {
		return "", ""
	}

	if a, err := mail.ParseAddress(s); err == nil {
		return a.Name, a.Address
	}

	for _, item := range nameAddressRe {
		m := item.re.FindStringSubmatchIndex(s)
		if len(m) == 0 {
			continue
		}

		if i := item.nameIndex * 2; i > 0 {
			name = s[m[i]:m[i+1]]
		}
		if i := item.emailIndex * 2; i > 0 {
			address = s[m[i]:m[i+1]]
		}
		return name, address
	}
	return s, ""
}
