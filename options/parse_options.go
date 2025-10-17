package options

// Parse configures how feeds are parsed
type Parse struct {
	// Keep reference to the original format-specific feed
	KeepOriginalFeed bool

	// Whether to parse dates (can be disabled for performance)
	ParseDates bool

	// Parsing behavior options
	StrictnessOptions Strictness
}

// Strictness controls parsing strictness
type Strictness struct {
	AllowInvalidDates    bool
	AllowMissingRequired bool
	AllowUnescapedMarkup bool
}

// Default returns sensible defaults
func Default() *Parse {
	return &Parse{
		KeepOriginalFeed: false,
		ParseDates:       true,
		StrictnessOptions: Strictness{
			AllowInvalidDates:    true,
			AllowMissingRequired: true,
			AllowUnescapedMarkup: true,
		},
	}
}
