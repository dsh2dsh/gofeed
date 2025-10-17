package options

// ParseOptions configures how feeds are parsed
type ParseOptions struct {
	// Keep reference to the original format-specific feed
	KeepOriginalFeed bool

	// Whether to parse dates (can be disabled for performance)
	ParseDates bool

	// Parsing behavior options
	StrictnessOptions StrictnessOptions
}

// StrictnessOptions controls parsing strictness
type StrictnessOptions struct {
	AllowInvalidDates    bool
	AllowMissingRequired bool
	AllowUnescapedMarkup bool
}

// DefaultParseOptions returns sensible defaults
func DefaultParseOptions() *ParseOptions {
	return &ParseOptions{
		KeepOriginalFeed: false,
		ParseDates:       true,
		StrictnessOptions: StrictnessOptions{
			AllowInvalidDates:    true,
			AllowMissingRequired: true,
			AllowUnescapedMarkup: true,
		},
	}
}
