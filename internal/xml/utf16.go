package xml

import (
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	utf16le = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	utf16be = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
)

type maybeUTF16 struct {
	transform.NopResetter

	current transform.Transformer
}

var _ transform.Transformer = (*maybeUTF16)(nil)

//nolint:wrapcheck // return errors as is, because it just forwards calls.
func (self *maybeUTF16) Transform(dst, src []byte, atEOF bool,
) (nDst, nSrc int, err error) {
	if self.current != nil {
		return self.current.Transform(dst, src, atEOF)
	}

	switch {
	case len(src) < 2:
		if !atEOF {
			return 0, 0, transform.ErrShortSrc
		}
		self.current = transform.Nop
	case src[0] == '<' && src[1] == 0:
		self.current = utf16le.NewDecoder()
	case src[0] == 0 && src[1] == '<':
		self.current = utf16be.NewDecoder()
	default:
		self.current = transform.Nop
	}
	return self.current.Transform(dst, src, atEOF)
}
