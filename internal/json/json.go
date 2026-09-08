package json

import (
	"encoding/json/jsontext"
	jsonEnc "encoding/json/v2"
	"fmt"
	"strings"
)

func MarshalString(v any) (string, error) {
	var b strings.Builder
	enc := jsontext.NewEncoder(&b,
		jsontext.SpaceAfterColon(true),
		jsontext.SpaceAfterComma(true),
		jsontext.WithIndent("  "))

	if err := jsonEnc.MarshalEncode(enc, v); err != nil {
		return "", fmt.Errorf("gofeed/internal/json: marshal to string: %w", err)
	}
	return b.String(), nil
}
