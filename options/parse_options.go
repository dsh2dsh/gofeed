package options

// Parse configures how feeds are parsed
type Parse struct {
	// Keep reference to the original format-specific feed
	KeepOriginalFeed bool
}

// Default returns sensible defaults
func Default() *Parse { return &Parse{} }

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
func From(v *Parse) Option {
	return func(opts *Parse) { *opts = *v }
}
