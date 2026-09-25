package parsers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNameAddress(t *testing.T) {
	tests := []struct {
		in   string
		want [2]string
	}{
		{"", [2]string{"", ""}},
		{"John Doe", [2]string{"John Doe", ""}},
		{"john@example.com", [2]string{"", "john@example.com"}},
		{"john@example.com (John Doe)", [2]string{"John Doe", "john@example.com"}},
		{"John Doe (john@example.com)", [2]string{"John Doe", "john@example.com"}},
		{"john@example.com (John (JD) Doe)", [2]string{"John (JD) Doe", "john@example.com"}},
		{"john@example.com (John@Home)", [2]string{"John@Home", "john@example.com"}},
		{"John Doe <john@example.com>", [2]string{"John Doe", "john@example.com"}},
		{`"Doe, John" <john@example.com>`, [2]string{"Doe, John", "john@example.com"}},
		{"<john@example.com>", [2]string{"", "john@example.com"}},
		{"Smith & Sons (Ltd)", [2]string{"Smith & Sons (Ltd)", ""}},
		{"@jack", [2]string{"@jack", ""}},
		{"Jane Doe (@jane)", [2]string{"Jane Doe (@jane)", ""}},
		{"John (Johnny) Doe", [2]string{"John (Johnny) Doe", ""}},
		{"Jane <not an address>", [2]string{"Jane <not an address>", ""}},
		{"John Doe john@example.com", [2]string{"John Doe john@example.com", ""}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var got [2]string
			got[0], got[1] = ParseNameAddress(tt.in)
			assert.Equal(t, tt.want, got, tt.in)
		})
	}
}
