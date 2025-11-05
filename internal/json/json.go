package json

import (
	"encoding/json"
	"fmt"
	"strings"
)

func MarshalString(v any) (string, error) {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return "", fmt.Errorf("gofeed/internal/json: marshal to string: %w", err)
	}
	return b.String(), nil
}
