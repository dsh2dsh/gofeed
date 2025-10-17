package options

import (
	"net/http"
	"time"
)

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

// RequestOptions configures HTTP requests for ParseURL
type RequestOptions struct {
	UserAgent       string
	Timeout         time.Duration
	IfNoneMatch     string    // ETag for conditional requests
	IfModifiedSince time.Time // For conditional requests
	Client          *http.Client
	AuthConfig      *Auth
}

// Auth is a structure allowing to use the BasicAuth during the HTTP request. It
// must be instantiated with your new Parser.
type Auth struct {
	Username string
	Password string
}

// Empty return true if this Auth has not Username and Password.
func (self *Auth) Empty() bool {
	return self.Username == "" && self.Password == ""
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
