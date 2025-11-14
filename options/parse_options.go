package options

// Parse configures how feeds are parsed
type Parse struct {
	// Keep reference to the original format-specific feed
	KeepOriginalFeed bool

	// Skip any element or extension, which the parser doesn't know. So instead of
	// parse it into [ext.Extensions] map, the parser skips it.
	SkipUnknownElements bool
}

type Option func(opts *Parse)

// Apply applies every option from array of opts and returns self ref.
func (self *Parse) Apply(opts ...Option) *Parse {
	for _, fn := range opts {
		fn(self)
	}
	return self
}

// WithKeepOriginalFeed sets [Parse.KeepOriginalFeed] to given value. It defines
// keep or not reference to the original format-specific feed. By default
// doesn't keep.
func WithKeepOriginalFeed(v bool) Option {
	return func(opts *Parse) { opts.KeepOriginalFeed = v }
}

// From copies all given options.
func From(v Parse) Option {
	return func(opts *Parse) { *opts = v }
}

// WithSkipUnknownElements configures the parser to skip any element or
// extension, which the parser doesn't know. So instead of parse it into
// [ext.Extensions] map, the parser skips it.
func WithSkipUnknownElements(v bool) Option {
	return func(opts *Parse) { opts.SkipUnknownElements = v }
}
