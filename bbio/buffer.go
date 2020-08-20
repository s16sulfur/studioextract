package bbio

import (
	"bytes"
	"io"
)

// bufWrite7BitEncodedInt is write out an int 7 bits at a time.  The high bit of the byte,
// when on, tells reader to continue reading more bytes.
func bufWrite7BitEncodedInt(buf *bytes.Buffer, val int) (n int, err error) {
	var v uint
	v = uint(val)
	n = 0

	for v >= 0x80 {
		b := byte(v | 0x80)
		err = buf.WriteByte(b)
		if err != nil {
			return
		}
		n++
		v >>= 7
	}

	b := byte(v)
	err = buf.WriteByte(b)
	if err != nil {
		return
	}
	n++
	return
}

// Buffer structure
type Buffer struct {
	buf *bytes.Buffer
}

// NewBuffer create Buffer structure
func NewBuffer() *Buffer {
	b := make([]byte, 0)
	nBuf := bytes.NewBuffer(b)
	return &Buffer{buf: nBuf}
}

// Len returns the number of bytes of the unread portion of the buffer
func (bl *Buffer) Len() int {
	return bl.buf.Len()
}

// Truncate discards all but the first n unread bytes from the buffer
func (bl *Buffer) Truncate(n int) {
	bl.buf.Truncate(n)
}

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
func (bl *Buffer) Bytes() []byte {
	return bl.buf.Bytes()
}

// Reset resets the buffer to be empty
func (bl *Buffer) Reset() {
	bl.buf.Reset()
}

// Grow grows the buffer's capacity
func (bl *Buffer) Grow(n int) {
	bl.buf.Grow(n)
}

// Write appends the contents of p to the buffer
func (bl *Buffer) Write(p []byte) (n int, err error) {
	return bl.buf.Write(p)
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
func (bl *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	return bl.buf.WriteTo(w)
}

// WriteByte appends the byte c to the buffer
func (bl *Buffer) WriteByte(c byte) error {
	return bl.buf.WriteByte(c)
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer
func (bl *Buffer) WriteRune(r rune) (n int, err error) {
	return bl.buf.WriteRune(r)
}

// WriteString appends the string to the buffer
func (bl *Buffer) WriteString(v string) (n int, err error) {
	bStr := []byte(v)
	slen := len(bStr)
	sn, err := bufWrite7BitEncodedInt(bl.buf, slen)
	if err != nil {
		return
	}
	n += sn
	wn, err := bl.buf.Write(bStr)
	if err != nil {
		return
	}
	n += wn
	return
}

// PutFloat implements for Buffer
func (bl *Buffer) PutFloat(v float32) (n int, err error) {
	b := make([]byte, 4)
	tmp := uint32(v)
	b[0] = byte(tmp)
	b[1] = byte(tmp >> 8)
	b[2] = byte(tmp >> 16)
	b[3] = byte(tmp >> 24)
	return bl.buf.Write(b)
}

// PutDouble implements for Buffer
func (bl *Buffer) PutDouble(v float64) (n int, err error) {
	b := make([]byte, 8)
	tmp := uint64(v)
	b[0] = byte(tmp)
	b[1] = byte(tmp >> 8)
	b[2] = byte(tmp >> 16)
	b[3] = byte(tmp >> 24)
	b[4] = byte(tmp >> 32)
	b[5] = byte(tmp >> 40)
	b[6] = byte(tmp >> 48)
	b[7] = byte(tmp >> 56)
	return bl.buf.Write(b)
}

// PutShort implements for Buffer
func (bl *Buffer) PutShort(v int16) (n int, err error) {
	b := make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return bl.buf.Write(b)
}

// PutUShort implements for Buffer
func (bl *Buffer) PutUShort(v uint16) (n int, err error) {
	b := make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return bl.buf.Write(b)
}

// PutInt implements for Buffer
func (bl *Buffer) PutInt(v int32) (n int, err error) {
	b := make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	return bl.buf.Write(b)
}

// PutUInt implements for Buffer
func (bl *Buffer) PutUInt(v uint32) (n int, err error) {
	b := make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	return bl.buf.Write(b)
}

// PutLong implements for Buffer
func (bl *Buffer) PutLong(v int64) (n int, err error) {
	b := make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return bl.buf.Write(b)
}

// PutULong implements for Buffer
func (bl *Buffer) PutULong(v uint64) (n int, err error) {
	b := make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return bl.buf.Write(b)
}

// PutFloatAll implements for Buffer
func (bl *Buffer) PutFloatAll(v ...float32) (n int, err error) {
	l := len(v)
	for i := 0; i < l; i++ {
		b := make([]byte, 4)
		tmp := uint32(v[i])
		b[0] = byte(tmp)
		b[1] = byte(tmp >> 8)
		b[2] = byte(tmp >> 16)
		b[3] = byte(tmp >> 24)
		p, werr := bl.buf.Write(b)
		if werr != nil {
			err = werr
			return
		}
		n += p
	}
	return
}

// PutIntAll implements for Buffer
func (bl *Buffer) PutIntAll(v ...int) (n int, err error) {
	l := len(v)
	for i := 0; i < l; i++ {
		t := v[i]
		b := make([]byte, 4)
		b[0] = byte(t)
		b[1] = byte(t >> 8)
		b[2] = byte(t >> 16)
		b[3] = byte(t >> 24)
		p, werr := bl.buf.Write(b)
		if werr != nil {
			err = werr
			return
		}
		n += p
	}
	return
}
