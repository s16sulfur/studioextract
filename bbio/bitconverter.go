package bbio

type bitConverter struct {
}

// BitConverter instance
var BitConverter bitConverter

// GetInt16Bytes implements for BitConverter
func (*bitConverter) GetInt16Bytes(v int16) (b []byte) {
	b = make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return
}

// GetUInt16Bytes implements for BitConverter
func (*bitConverter) GetUInt16Bytes(v uint16) (b []byte) {
	b = make([]byte, 2)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return
}

// GetInt32Bytes implements for BitConverter
func (*bitConverter) GetInt32Bytes(v int32) (b []byte) {
	b = make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	return
}

// GetUInt32Bytes implements for BitConverter
func (*bitConverter) GetUInt32Bytes(v uint32) (b []byte) {
	b = make([]byte, 4)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	return
}

// GetInt64Bytes implements for BitConverter
func (*bitConverter) GetInt64Bytes(v int64) (b []byte) {
	b = make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return
}

// GetUInt64Bytes implements for BitConverter
func (*bitConverter) GetUInt64Bytes(v uint64) (b []byte) {
	b = make([]byte, 8)
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return
}

// GetFloat32Bytes implements for BitConverter
func (*bitConverter) GetFloat32Bytes(v float32) (b []byte) {
	b = make([]byte, 4)
	tmp := uint32(v)
	b[0] = byte(tmp)
	b[1] = byte(tmp >> 8)
	b[2] = byte(tmp >> 16)
	b[3] = byte(tmp >> 24)
	return
}

// GetFloat64Bytes implements for BitConverter
func (*bitConverter) GetFloat64Bytes(v float64) (b []byte) {
	b = make([]byte, 8)
	tmp := uint64(v)
	b[0] = byte(tmp)
	b[1] = byte(tmp >> 8)
	b[2] = byte(tmp >> 16)
	b[3] = byte(tmp >> 24)
	b[4] = byte(tmp >> 32)
	b[5] = byte(tmp >> 40)
	b[6] = byte(tmp >> 48)
	b[7] = byte(tmp >> 56)
	return
}

// ToInt16 implements for BitConverter
func (*bitConverter) ToInt16(b []byte) int16 {
	if len(b) < 2 {
		return int16(0)
	}
	return int16(b[0]) | int16(b[1])<<8
}

// ToUInt16 implements for BitConverter
func (*bitConverter) ToUInt16(b []byte) uint16 {
	if len(b) < 2 {
		return uint16(0)
	}
	return uint16(b[0]) | uint16(b[1])<<8
}

// ToInt32 implements for BitConverter
func (*bitConverter) ToInt32(b []byte) int32 {
	if len(b) < 4 {
		return int32(0)
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

// ToUInt32 implements for BitConverter
func (*bitConverter) ToUInt32(b []byte) uint32 {
	if len(b) < 4 {
		return uint32(0)
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// ToInt64 implements for BitConverter
func (*bitConverter) ToInt64(b []byte) int64 {
	if len(b) < 8 {
		return int64(0)
	}

	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	return int64(hi)<<32 | int64(lo)
}

// ToUInt64 implements for BitConverter
func (*bitConverter) ToUInt64(b []byte) uint64 {
	if len(b) < 8 {
		return uint64(0)
	}

	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	return uint64(hi)<<32 | uint64(lo)
}

// ToFloat32 implements for BitConverter
func (*bitConverter) ToFloat32(b []byte) float32 {
	if len(b) < 4 {
		return float32(0)
	}
	tmp := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return float32(tmp)
}

// ToFloat64 implements for BitConverter
func (*bitConverter) ToFloat64(b []byte) float64 {
	if len(b) < 4 {
		return float64(0)
	}
	lo := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	hi := uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24
	tmp := uint64(hi)<<32 | uint64(lo)
	return float64(tmp)
}
