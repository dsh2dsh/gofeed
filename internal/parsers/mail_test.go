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
		{"john@example.com (John Doe)", [2]string{"John Doe", "john@example.com"}},
		{"John Doe <john@example.com>", [2]string{"John Doe", "john@example.com"}},
		{"John Doe (john@example.com)", [2]string{"John Doe", "john@example.com"}},
		{"Smith & Sons (Ltd)", [2]string{"Smith & Sons (Ltd)", ""}},
		{"@jack", [2]string{"@jack", ""}},
		{"John Doe", [2]string{"John Doe", ""}},
		{"user@localhost", [2]string{"", "user@localhost"}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var got [2]string
			got[0], got[1] = ParseNameAddress(tt.in)
			assert.Equal(t, tt.want, got, tt.in)
		})
	}
}
