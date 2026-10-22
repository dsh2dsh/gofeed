package utf8bom

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var utf8BOM = [...]byte{0xEF, 0xBB, 0xBF}

func NewReader(r io.Reader) (io.Reader, error) {
	switch v := r.(type) {
	case *bufio.Reader:
		return fromBufioReader(v)
	case *bytes.Buffer:
		return fromBytesBuffer(v), nil
	case *bytes.Reader:
		return fromReadSeeker(v)
	case *strings.Reader:
		return fromReadSeeker(v)
	}
	return fromBufioReader(bufio.NewReader(r))
}

func fromBufioReader(r *bufio.Reader) (*bufio.Reader, error) {
	prefix, err := r.Peek(len(utf8BOM))
	if err != nil {
		return nil, fmt.Errorf("reading UTF8 BOM prefix: %w", err)
	}

	if !bytes.Equal(prefix, utf8BOM[:]) {
		return r, nil
	}

	if _, err := r.Discard(len(utf8BOM)); err != nil {
		return nil, fmt.Errorf("discarding UTF8 BOM prefix: %w", err)
	}
	return r, nil
}

func fromReadSeeker(r io.ReadSeeker) (io.ReadSeeker, error) {
	prefix := make([]byte, len(utf8BOM))
	_, err := r.Read(prefix)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return r, nil
		}
		return nil, fmt.Errorf("reading UTF8 BOM prefix: %w", err)
	}

	if bytes.Equal(prefix, utf8BOM[:]) {
		return r, nil
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("discarding UTF8 BOM prefix: %w", err)
	}
	return r, nil
}

func fromBytesBuffer(b *bytes.Buffer) *bytes.Buffer {
	if bytes.HasPrefix(b.Bytes(), utf8BOM[:]) {
		b.Next(len(utf8BOM))
	}
	return b
}
