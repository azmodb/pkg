// Package wire provides low-level routines for parsing and producing 9P2000
// wire-format messages.
//
// The wire package is to be used for making higher-level 9P2000 libraries. The
// parsing routines within make very few assumptions or decisions, so that it
// may be used for a wide variety of higher-level packages.
package wire

import (
	"errors"
	"io"
)

// ParseError converts an error code into an error value. This returns nil if n
// is a non-negative number.
func ParseError(n int) error {
	if n >= 0 {
		return nil
	}
	switch n {
	case errUnexpectedEOF:
		return io.ErrUnexpectedEOF
	}
	return errors.New("parse error")
}

const (
	_ = -iota
	errUnexpectedEOF
)

// ConsumeBytes parses b as a length-prefixed bytes value, reporting its length.
// This returns a negative length upon an error.
func ConsumeBytes(b []byte, target []byte) ([]byte, int) {
	m, n := ConsumeUint32(b)
	if n < 0 {
		return target, n // forward error code
	}
	if m > uint32(len(b[n:])) {
		return target, errUnexpectedEOF
	}
	if uint32(cap(target)) < m {
		target = make([]byte, m)
	}
	target = target[:m]

	n = copy(target, b[n:][:m])
	return target, 4 + n
}

// ConsumeString parses b as a length-prefixed string value, reporting its
// length. This returns a negative length upon an error.
func ConsumeString(b []byte) (string, int) {
	m, n := ConsumeUint16(b)
	if n < 0 {
		return "", n // forward error code
	}
	if m > uint16(len(b[n:])) {
		return "", errUnexpectedEOF
	}
	return string(b[n:][:m]), n + int(m)
}

// ConsumeUint64 parses b as a little-endian uint64, reporting its length. This
// returns a negative length upon an error.
func ConsumeUint64(b []byte) (uint64, int) {
	if len(b) < 8 {
		return 0, errUnexpectedEOF
	}
	return uint64(b[0])<<0 |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56, 8
}

// ConsumeUint32 parses b as a little-endian uint32, reporting its length. This
// returns a negative length upon an error.
func ConsumeUint32(b []byte) (uint32, int) {
	if len(b) < 4 {
		return 0, errUnexpectedEOF
	}
	return uint32(b[0])<<0 |
		uint32(b[1])<<8 |
		uint32(b[2])<<16 |
		uint32(b[3])<<24, 4
}

// ConsumeUint16 parses b as a little-endian uint16, reporting its length. This
// returns a negative length upon an error.
func ConsumeUint16(b []byte) (uint16, int) {
	if len(b) < 2 {
		return 0, errUnexpectedEOF
	}
	return uint16(b[0])<<0 | uint16(b[1])<<8, 2
}

// ConsumeUint8 parses b as a little-endian uint8, reporting its length. This
// returns a negative length upon an error.
func ConsumeUint8(b []byte) (uint8, int) {
	if len(b) < 1 {
		return 0, errUnexpectedEOF
	}
	return b[0], 1
}

// PutBytes appends v to b as a length-prefixed bytes value.
func PutBytes(b []byte, v []byte) []byte {
	return append(PutUint32(b, uint32(len(v))), v...)
}

// PutString appends v to b as a length-prefixed string value.
func PutString(b []byte, v string) []byte {
	return append(PutUint16(b, uint16(len(v))), v...)
}

// PutUint64 appends v to b as a little-endian uint64.
func PutUint64(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

// PutUint32 appends v to b as a little-endian uint32.
func PutUint32(b []byte, v uint32) []byte {
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
	)
}

// PutUint16 appends v to b as a little-endian uint16.
func PutUint16(b []byte, v uint16) []byte {
	return append(b, byte(v>>0), byte(v>>8))
}

// PutUint8 appends v to b as a little-endian uint8.
func PutUint8(b []byte, v uint8) []byte {
	return append(b, v)
}

type Option func(*Buffer)

// Buffer is a buffer for encoding and decoding the wire format. It may be
// eused between invocations to reduce memory usage.
type Buffer struct {
	data []byte
	err  error
}

// NewBuffer allocates a new Buffer initialized with data, where the contents
// of data are considered the unread portion of the buffer.
func NewBuffer(data []byte, opts ...Option) *Buffer {
	if data == nil {
		data = make([]byte, 0, 128) // TOOD
	}
	b := &Buffer{}
	for _, opt := range opts {
		opt(b)
	}
	b.SetBuf(data)
	return b
}

// Reset clears the internal buffer of all written and unread data.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
	b.err = nil
}

// SetBuf sets data as the internal buffer, where the contents of data are
// considered the unread portion of b.
func (b *Buffer) SetBuf(data []byte) { b.data = data }

func (b *Buffer) setErr(err error) {
	if b.err == nil && err != nil {
		b.err = err
	}
}

// Err returns the first error that was encountered by b.
func (b *Buffer) Err() error { return b.err }

// Len returns the number of bytes of the unread portion of b.
func (b *Buffer) Len() int { return len(b.data) }

// PutBytes appends v to b as a length-prefixed bytes value.
func (b *Buffer) PutBytes(v []byte) { b.data = PutBytes(b.data, v) }

// PutString appends v to b as a length-prefixed string value.
func (b *Buffer) PutString(v string) { b.data = PutString(b.data, v) }

// PutUint64 appends v to b as a little-endian uint64.
func (b *Buffer) PutUint64(v uint64) { b.data = PutUint64(b.data, v) }

// PutUint32 appends v to b as a little-endian uint32.
func (b *Buffer) PutUint32(v uint32) { b.data = PutUint32(b.data, v) }

// PutUint16 appends v to b as a little-endian uint16.
func (b *Buffer) PutUint16(v uint16) { b.data = PutUint16(b.data, v) }

// PutUint8 appends v to b as a little-endian uint8.
func (b *Buffer) PutUint8(v uint8) { b.data = PutUint8(b.data, v) }

// Bytes decodes a 32-bit count-delimited bytes value from b.
func (b *Buffer) Bytes() []byte {
	if b.Err() != nil {
		return nil
	}

	v, n := ConsumeBytes(b.data, nil)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// String decodes a 16-bit count-delimited string value from b.
func (b *Buffer) String() string {
	if b.Err() != nil {
		return ""
	}

	v, n := ConsumeString(b.data)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// Uint64 decodes a 64-bit integer from b.
func (b *Buffer) Uint64() uint64 {
	if b.Err() != nil {
		return 0
	}

	v, n := ConsumeUint64(b.data)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// Uint32 decodes a 32-bit integer from b.
func (b *Buffer) Uint32() uint32 {
	if b.Err() != nil {
		return 0
	}

	v, n := ConsumeUint32(b.data)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// Uint16 decodes a 16-bit integer from b.
func (b *Buffer) Uint16() uint16 {
	if b.Err() != nil {
		return 0
	}

	v, n := ConsumeUint16(b.data)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// Uint8 decodes a 8-bit integer from b.
func (b *Buffer) Uint8() uint8 {
	if b.Err() != nil {
		return 0
	}

	v, n := ConsumeUint8(b.data)
	if n < 0 {
		b.setErr(ParseError(n))
	}
	b.data = b.data[n:]
	return v
}

// WriteString appends the contents of s to b, growing the buffer as needed. The
// return value n is the length of p; err is always nil.
func (b *Buffer) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	b.data = append(b.data, s...)
	return len(s), nil
}

// Write appends the contents of p to b, growing the buffer as needed. The
// return value n is the length of p; err is always nil.
func (b *Buffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

// WriteTo writes data to w until b is drained or an error occurs. The return
// value n is the number of bytes written; it always fits into an int, but it is
// int64 to match the io.WriterTo interface. Any error encountered during the
// write is also returned.
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	if len(b.data) == 0 {
		return 0, io.EOF
	}

	n, err := w.Write(b.data)
	if err != nil {
		return int64(n), err
	}
	if n != len(b.data) {
		return int64(n), io.ErrShortWrite
	}

	b.Reset() // Buffer is now empty; reset.
	return int64(n), nil
}

// Read reads the next len(p) bytes from b or until b is drained. The return
// value n is the number of bytes read. If the buffer has no data to return, err
// is io.EOF (unless len(p) is zero); otherwise it is nil.
func (b *Buffer) Read(p []byte) (int, error) {
	if len(b.data) == 0 {
		return 0, io.EOF
	}

	n := copy(p, b.data)
	b.data = b.data[n:]
	return n, nil
}
