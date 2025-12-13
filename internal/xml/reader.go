package xml

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

type ValidReader struct {
	io.Reader

	runeReader    io.RuneReader
	charsetReader CharsetReaderFunc

	buf  [utf8.UTFMax]byte
	i, n int
}

type CharsetReaderFunc func(charset string, input io.Reader) (io.Reader, error)

var (
	_ io.ByteReader = (*ValidReader)(nil)
	_ io.Reader     = (*ValidReader)(nil)
)

func (self *ValidReader) WithCharsetReader(charsetReader CharsetReaderFunc,
) *ValidReader {
	self.charsetReader = charsetReader
	return self
}

func (self *ValidReader) WithReader(r io.Reader) *ValidReader {
	if rr, ok := r.(io.RuneReader); ok {
		self.Reader = r
		self.runeReader = rr
		return self
	}

	buf := bufio.NewReader(r)
	self.Reader = buf
	self.runeReader = buf
	return self
}

func (self *ValidReader) ReadByte() (byte, error) {
	if self.n > 0 && self.i < self.n {
		b := self.buf[self.i]
		self.i++
		return b, nil
	}

	//nolint:wrapcheck // fwd as is
	for {
		r, size, err := self.runeReader.ReadRune()
		switch {
		case err != nil:
			return 0, err
		case (r == unicode.ReplacementChar && size == 1) || !inXMLCharacterRange(r):
			continue
		case size == 1:
			self.i, self.n = 0, 0
			return byte(r), nil
		}

		self.i = 1
		self.n = utf8.EncodeRune(self.buf[:], r)
		return self.buf[0], nil
	}
}

// Decide whether the given rune is in the XML Character Range, per
// the Char production of https://www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
//
// Extracted from encoding/xml/xml.go.
func inXMLCharacterRange(r rune) bool {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

func (self *ValidReader) CharsetReader(enc string, _ io.Reader,
) (io.Reader, error) {
	charsetReader := self.charsetReader
	if charsetReader == nil {
		charsetReader = charset.NewReaderLabel
	}

	r, err := charsetReader(enc, self.Reader)
	if err != nil {
		return nil, fmt.Errorf(
			"gofeed: unable create charset converter charset=%q: %w", enc, err)
	}
	return self.WithReader(r), nil
}
