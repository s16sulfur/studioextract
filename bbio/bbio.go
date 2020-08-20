package bbio

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

// intDataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func intDataSize(data interface{}) int {
	switch data := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return 1
	case []bool:
		return len(data)
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, uint16, *int16, *uint16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, uint32, *int32, *uint32:
		return 4
	case []int32:
		return 4 * len(data)
	case []uint32:
		return 4 * len(data)
	case int64, uint64, *int64, *uint64:
		return 8
	case []int64:
		return 8 * len(data)
	case []uint64:
		return 8 * len(data)
	case float32, *float32:
		return 4
	case float64, *float64:
		return 8
	case []float32:
		return 4 * len(data)
	case []float64:
		return 8 * len(data)
	}
	return 0
}

// read7BitEncodedInt is read out an Int32 7 bits at a time.  The high bit
// of the byte when on means to continue reading more bytes.
func read7BitEncodedInt(r *bytes.Reader) (int, int, error) {
	var count int
	var shift int
	var b int
	var n int

	for {
		if shift == 5*7 {
			return 0, n, errors.New("Bad format 7Bit Int32")
		}

		rb, err := r.ReadByte()
		n++
		if err != nil {
			return 0, n, err
		}

		b = int(rb)
		count |= (b & 0x7F) << shift
		shift += 7

		if (b & 0x80) == 0 {
			break
		}
	}
	return count, n, nil
}

// Reader implements of the Reader
type Reader struct {
	s   []byte
	pos int64
	r   *bytes.Reader
}

// NewReader implements for create Reader
func NewReader(sr io.Reader) *Reader {
	b, err := ioutil.ReadAll(sr)
	if err != nil {
		panic(err)
	}
	br := bytes.NewReader(b)
	return &Reader{s: b, r: br}
}

// NewReaderBytes implements for create Reader
func NewReaderBytes(b []byte) *Reader {
	br := bytes.NewReader(b)
	return &Reader{s: b, r: br}
}

// NewReaderFile implements for create Reader
func NewReaderFile(filename string) *Reader {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	br := bytes.NewReader(b)
	return &Reader{s: b, r: br}
}

// Seek implements the io.Seeker interface.
func (br *Reader) Seek(offset int64, whence int) (int64, error) {
	abs, err := br.r.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	br.pos = abs
	return abs, nil
}

// Position implements of the Reader
func (br *Reader) Position() int64 {
	return br.pos
}

// Len returns the number of bytes of the unread portion of the slice.
func (br *Reader) Len() int {
	return br.r.Len()
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (br *Reader) Size() int64 {
	return br.r.Size()
}

// Buffer implements of the Reader
func (br *Reader) Buffer() []byte {
	return br.s
}

// ReadToEnd implements of the Reader
func (br *Reader) ReadToEnd() (b []byte, err error) {
	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	capacity := int64(bytes.MinRead)
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	n, err := buf.ReadFrom(br.r)
	if err != nil {
		return nil, err
	}
	br.pos += int64(n)

	return buf.Bytes(), err
}

// Peek returns the next n bytes without advancing the reader
func (br *Reader) Peek(n int) (b []byte, err error) {
	p := br.pos + 1
	if (p + int64(n)) >= int64(len(br.s)) {
		err = io.EOF
		return
	}
	b = br.s[p : p+int64(n)]
	return
}

// BinaryRead implements the binary.Read
func (br *Reader) BinaryRead(data interface{}) error {
	if n := intDataSize(data); n != 0 {
		err := binary.Read(br.r, binary.LittleEndian, &data)
		if err != nil {
			return err
		}
		br.pos += int64(n)
	}
	return nil
}

// ReadAt implements the io.ReaderAt interface.
func (br *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("bbio.Reader.ReadAt: negative offset")
	}
	if off >= int64(len(br.s)) {
		return 0, io.EOF
	}
	n = copy(b, br.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

// Read implements the io.Reader interface.
func (br *Reader) Read(b []byte) (n int, err error) {
	n, err = br.r.Read(b)
	if err != nil {
		return
	}
	br.pos += int64(n)
	return
}

// ReadBoolean implements of the Reader
func (br *Reader) ReadBoolean() (bo bool, err error) {
	b, err := br.r.ReadByte()
	if err != nil {
		return false, err
	}

	br.pos++
	return b != 0, nil
}

// ReadByte implements the io.ByteReader interface.
func (br *Reader) ReadByte() (byte, error) {
	b, err := br.r.ReadByte()
	if err != nil {
		return 0, err
	}

	br.pos++
	return b, nil
}

// ReadBytes implements the Reader
func (br *Reader) ReadBytes(c int) ([]byte, error) {
	b := make([]byte, c)
	n, err := br.r.Read(b)
	if err != nil {
		return nil, err
	}

	br.pos += int64(n)
	return b, nil
}

// ReadChar implements of the Reader
func (br *Reader) ReadChar() (ch rune, err error) {
	var sz int
	ch, sz, err = br.r.ReadRune()
	if err != nil {
		return
	}

	br.pos += int64(sz)
	return
}

// ReadInt16 implements of the Reader
func (br *Reader) ReadInt16() (int16, error) {
	b := make([]byte, 2)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	return int16(b[0]) | int16(b[1])<<8, nil
}

// ReadUInt16 implements of the Reader
func (br *Reader) ReadUInt16() (uint16, error) {
	b := make([]byte, 2)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	return uint16(b[0]) | uint16(b[1])<<8, nil
}

// ReadInt32 implements of the Reader
func (br *Reader) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24, nil
}

// ReadUInt32 implements of the Reader
func (br *Reader) ReadUInt32() (uint32, error) {
	b := make([]byte, 4)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24, nil
}

// ReadInt64 implements of the Reader
func (br *Reader) ReadInt64() (int64, error) {
	b := make([]byte, 8)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	return int64(hi)<<32 | int64(lo), nil
}

// ReadUInt64 implements of the Reader
func (br *Reader) ReadUInt64() (uint64, error) {
	b := make([]byte, 8)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	return uint64(hi)<<32 | uint64(lo), nil
}

// ReadSingle implements of the Reader
func (br *Reader) ReadSingle() (float32, error) {
	b := make([]byte, 4)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	tmp := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return float32(tmp), nil
}

// ReadDouble implements of the Reader
func (br *Reader) ReadDouble() (float64, error) {
	b := make([]byte, 8)
	n, err := br.r.Read(b)
	if err != nil {
		return 0, err
	}

	br.pos += int64(n)
	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	tmp := uint64(hi)<<32 | uint64(lo)
	return float64(tmp), nil
}

// ReadString implements of the Reader
func (br *Reader) ReadString() (s string, err error) {
	var slen, sn int
	s = ""
	slen, sn, err = read7BitEncodedInt(br.r)
	if err != nil {
		return
	}
	if slen < 0 {
		err = errors.New("Invalid string len")
		return
	}
	if slen == 0 {
		return
	}

	var currPos, readLen int
	currPos = 0
	readLen = 0

	maxCharBytesSize := 128

	charBuffer := bytes.NewBufferString("")

	for {
		if (slen - currPos) > maxCharBytesSize {
			readLen = maxCharBytesSize
		} else {
			readLen = slen - currPos
		}

		charBytes := make([]byte, readLen)
		n, e := br.r.Read(charBytes)
		if e != nil {
			err = e
			return
		}
		if n == 0 {
			err = io.EOF
			break
		}

		charBuffer.Grow(readLen)
		_, ce := charBuffer.Write(charBytes)
		if ce != nil {
			err = ce
			break
		}

		if currPos == 0 && n == slen {
			break
		}

		currPos += n
		if currPos >= slen {
			break
		}
	}

	br.pos += int64(slen + sn)

	s = charBuffer.String()
	return
}

// ReadSlice implements of the Reader
func (br *Reader) ReadSlice(delim []byte) (b []byte, err error) {
	pos := int(br.pos)
	i := bytes.Index(br.s[pos:], delim)
	if i <= 0 {
		return
	}
	sz := (i + len(delim)) - pos
	b = make([]byte, sz)
	n, err := br.r.Read(b)
	if err != nil {
		return
	}

	br.pos += int64(n)
	return
}

// ReadStringSlice implements of the Reader
func (br *Reader) ReadStringSlice(delim byte) (s string, err error) {
	s = ""
	pos := int(br.pos)

	i := bytes.IndexByte(br.s[pos:], delim)
	if i <= 0 {
		return
	}
	sz := i - pos

	b := make([]byte, sz)
	n, err := br.r.Read(b)
	if err != nil {
		return
	}

	br.pos += int64(n)
	s = string(bytes.Trim(b, "\x00"))
	return
}

// ReadStringFixed implements of the Reader
func (br *Reader) ReadStringFixed(count int, trimZero bool) (s string, err error) {
	b := make([]byte, count)
	n, err := br.r.Read(b)
	if err != nil {
		return
	}

	br.pos += int64(n)
	if trimZero {
		s = string(bytes.Trim(b, "\x00"))
	} else {
		s = string(b)
	}
	return
}

// WriteTo implements the io.WriterTo interface.
func (br *Reader) WriteTo(w io.Writer) (n int64, err error) {
	n, err = br.r.WriteTo(w)
	return
}

// Index returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func (br *Reader) Index(sep []byte) int {
	return bytes.Index(br.s, sep)
}

// LastIndex returns the index of the last instance of sep in s, or -1 if sep is not present in s.
func (br *Reader) LastIndex(sep []byte) int {
	return bytes.LastIndex(br.s, sep)
}

// FindAll return the all occurrences of sep
func (br *Reader) FindAll(sep []byte) (f []int) {
	index := len(br.s)
	tmp := br.s
	for true {
		match := bytes.LastIndex(tmp[0:index], sep)
		if match == -1 {
			break
		} else {
			index = match
			f = append(f, match)
		}
	}

	sort.Ints(f)
	return
}

// write7BitEncodedInt is write out an int 7 bits at a time.  The high bit of the byte,
// when on, tells reader to continue reading more bytes.
func write7BitEncodedInt(w io.Writer, val int) (n int, err error) {
	var v uint
	var n1, n2 int
	v = uint(val)
	n = 0

	for v >= 0x80 {
		b := make([]byte, 1)
		b[0] = byte(v | 0x80)
		n1, err = w.Write(b)
		if err != nil {
			return
		}
		n += n1
		v >>= 7
	}

	b := make([]byte, 1)
	b[0] = byte(v)
	n2, err = w.Write(b)
	if err != nil {
		return
	}
	n += n2
	return
}

// Writer implements of the Writer
type Writer struct {
	pos int64
	w   *bufio.Writer
}

// NewWriter implements for create Writer
func NewWriter(w io.Writer) *Writer {
	bw := bufio.NewWriter(w)
	return &Writer{w: bw}
}

// NewWriterSize returns a new Writer whose buffer has at least the specified
// size. If the argument io.Writer is already a Writer with large enough
// size, it returns the underlying Writer.
func NewWriterSize(w io.Writer, size int) *Writer {
	bw := bufio.NewWriterSize(w, size)
	return &Writer{w: bw}
}

// NewWriterFile implements for create Writer
func NewWriterFile(filename string) *Writer {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	return &Writer{w: w}
}

// Position implements of the Writer.BaseStream
func (bw *Writer) Position() int64 {
	return bw.pos
}

// Size returns the size of the underlying buffer in bytes.
func (bw *Writer) Size() int {
	return bw.Size()
}

// Flush writes any buffered data to the underlying io.Writer.
func (bw *Writer) Flush() error {
	return bw.w.Flush()
}

// ReadFrom implements io.ReaderFrom interface. If the underlying writer
// supports the ReadFrom method, and b has no buffered data yet,
// this calls the underlying ReadFrom without buffering.
func (bw *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = bw.w.ReadFrom(r)
	if err != nil {
		return
	}
	bw.pos += n
	return
}

// BinaryWrite implements from binary.Write
func (bw *Writer) BinaryWrite(data interface{}) error {
	if n := intDataSize(data); n != 0 {
		err := binary.Write(bw.w, binary.LittleEndian, &data)
		if err != nil {
			return err
		}
		bw.pos += int64(n)
	}
	return nil
}

// WriteByte implements the io.ByteWriter interface.
func (bw *Writer) WriteByte(v byte) error {
	err := bw.w.WriteByte(v)
	if err != nil {
		return err
	}
	bw.pos++
	return nil
}

// Write implements the io.Writer interface.
func (bw *Writer) Write(v []byte) (n int, err error) {
	n, err = bw.w.Write(v)
	if err != nil {
		return
	}
	bw.pos += int64(n)
	return
}

// WriteFloat implements of the Writer
func (bw *Writer) WriteFloat(v float32) error {
	b := make([]byte, 4)
	tmp := uint32(v)
	b[0] = byte(tmp)
	b[1] = byte(tmp >> 8)
	b[2] = byte(tmp >> 16)
	b[3] = byte(tmp >> 24)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteDouble implements of the Writer
func (bw *Writer) WriteDouble(v float64) error {
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
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteShort implements of the Writer
func (bw *Writer) WriteShort(v int16) error {
	b := make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteUShort implements of the Writer
func (bw *Writer) WriteUShort(v uint16) error {
	b := make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteInt implements of the Writer
func (bw *Writer) WriteInt(v int32) error {
	b := make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteUInt implements of the Writer
func (bw *Writer) WriteUInt(v uint32) error {
	b := make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteLong implements of the Writer
func (bw *Writer) WriteLong(v int64) error {
	b := make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteULong implements of the Writer
func (bw *Writer) WriteULong(v uint64) error {
	b := make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	n, err := bw.w.Write(b)
	if err != nil {
		return err
	}
	bw.pos += int64(n)
	return nil
}

// WriteString implements this io.StringWriter interface.
func (bw *Writer) WriteString(v string) (n int, err error) {
	bStr := []byte(v)
	slen := len(bStr)

	sn, err := write7BitEncodedInt(bw.w, slen)
	if err != nil {
		return
	}
	n += sn
	bw.pos += int64(sn)

	wn, err := bw.w.Write(bStr)
	if err != nil {
		return
	}
	n += wn
	bw.pos += int64(n)
	return
}
