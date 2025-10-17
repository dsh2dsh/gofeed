package shared

import (
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

func NewReaderLabel(label string, input io.Reader) (io.Reader, error) {
	conv, err := charset.NewReaderLabel(label, input)
	if err != nil {
		return nil, fmt.Errorf(
			"gofeed: unable create charset converter for %q: %w", label, err)
	}
	return conv, nil
}
