package binary

import (
	"encoding/binary"
	"io"
)

// Buffer is a variable-sized buffer of bytes used to encode and decode
// 9P2000 messages.
type Buffer struct {
	data []byte
	err  error
}

// NewBuffer creates and initializes a new Buffer using data as its
// initial contents. The new Buffer takes ownership of data, and the
// caller should not use data after this call.
func NewBuffer(data []byte) *Buffer {
	return &Buffer{data: data}
}

// Len returns the number of bytes of the unread portion of the buffer.
func (b *Buffer) Len() int { return len(b.data) }

// Err returns the first error that was encountered by the Buffer.
func (b *Buffer) Err() error { return b.err }

// Bytes returns a slice of length b.Len() holding the unread portion of
// the buffer. The slice is valid for use only until the next buffer
// modification.
func (b *Buffer) Bytes() []byte { return b.data }

// Reset resets the buffer to be empty, but it retains the underlying
// storage for use by future writes.
func (b *Buffer) Reset() {
	b.err = nil
	b.data = b.data[:0]
}

const maxInt = int64(^uint(0) >> 1)

const bufBootstrapSize = 128

func (b *Buffer) grow(n int) (off int) {
	if b.data == nil {
		if n < bufBootstrapSize {
			b.data = make([]byte, n, bufBootstrapSize)
		} else {
			b.data = make([]byte, n)
		}
		return 0
	}

	off = len(b.data)
	if off+n > cap(b.data) {
		capacity := n + cap(b.data)*2
		if capacity > 4*1024*1024 {
			capacity = n + cap(b.data)
		}
		if int64(capacity) >= maxInt {
			panic("buffer too large")
		}

		data := make([]byte, off, capacity)
		copy(data[:off], b.data)
		b.data = data
	}

	b.data = b.data[:off+n]
	return off
}

// consume consumes n bytes from Buffer. If an error occurred it returns
// false.
func (b *Buffer) consume(n int) ([]byte, bool) {
	if b.err != nil {
		return nil, false
	}
	if len(b.data) < n {
		b.err = io.ErrUnexpectedEOF
		return nil, false
	}
	data := b.data[:n]
	b.data = b.data[n:]
	return data, true
}

var (
	puint64 = binary.LittleEndian.PutUint64
	puint32 = binary.LittleEndian.PutUint32
	puint16 = binary.LittleEndian.PutUint16

	guint64 = binary.LittleEndian.Uint64
	guint32 = binary.LittleEndian.Uint32
	guint16 = binary.LittleEndian.Uint16
)

// PutString16 writes a count-delimited string to the Buffer.
func (b *Buffer) PutString16(v string) {
	n := b.grow(2 + len(v))
	puint16(b.data[n:], uint16(len(v)))
	copy(b.data[n+2:], v)
}

// PutUint64 writes a unsigned 64-bit integer to to Buffer.
func (b *Buffer) PutUint64(v uint64) {
	n := b.grow(8)
	puint64(b.data[n:], v)
}

// PutUint32 writes a unsigned 32-bit integer to to Buffer.
func (b *Buffer) PutUint32(v uint32) {
	n := b.grow(4)
	puint32(b.data[n:], v)
}

// PutUint16 writes a unsigned 16-bit integer to to Buffer.
func (b *Buffer) PutUint16(v uint16) {
	n := b.grow(2)
	puint16(b.data[n:], v)
}

// PutUint8 writes a unsigned 8-bit integer to to Buffer.
func (b *Buffer) PutUint8(v uint8) {
	n := b.grow(1)
	b.data[n] = v
}

// String16 reads a count-delimited string from the Buffer.
func (b *Buffer) String16() string {
	data, ok := b.consume(2)
	if !ok {
		return ""
	}

	size := int(guint16(data))
	if len(b.data) < size {
		b.err = io.ErrUnexpectedEOF
		return ""
	}

	v := string(b.data[:size])
	b.data = b.data[size:]
	return v
}

// Uint64 reads a 64-bit integer from the Buffer.
func (b *Buffer) Uint64() uint64 {
	data, ok := b.consume(8)
	if !ok {
		return 0
	}
	return guint64(data)
}

// Uint32 reads a 32-bit integer from the Buffer.
func (b *Buffer) Uint32() uint32 {
	data, ok := b.consume(4)
	if !ok {
		return 0
	}
	return guint32(data)
}

// Uint16 reads a 16-bit integer from the Buffer.
func (b *Buffer) Uint16() uint16 {
	if b == nil {
		return 0
	}
	data, ok := b.consume(2)
	if !ok {
		return 0
	}
	return guint16(data)
}

// Uint8 reads a 8-bit integer from the Buffer.
func (b *Buffer) Uint8() uint8 {
	if b == nil {
		return 0
	}
	data, ok := b.consume(1)
	if !ok {
		return 0
	}
	return data[0]
}
