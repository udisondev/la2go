package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf16"
)

// DefaultStringCapacity — типичная длина L2 account name (characters).
// Большинство account names ≤16 символов, pre-allocation снижает allocations.
const DefaultStringCapacity = 16

// Reader provides methods for reading packet data.
// Uses Little-Endian byte order for all multi-byte values.
type Reader struct {
	data []byte
	pos  int
}

// NewReader creates a new packet reader.
func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
		pos:  0,
	}
}

// ReadByte reads a single byte.
func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("ReadByte: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// ReadShort reads an int16 (2 bytes, LE).
func (r *Reader) ReadShort() (int16, error) {
	if r.pos+2 > len(r.data) {
		return 0, fmt.Errorf("ReadShort: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	val := int16(binary.LittleEndian.Uint16(r.data[r.pos:]))
	r.pos += 2
	return val, nil
}

// ReadInt reads an int32 (4 bytes, LE).
func (r *Reader) ReadInt() (int32, error) {
	if r.pos+4 > len(r.data) {
		return 0, fmt.Errorf("ReadInt: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	val := int32(binary.LittleEndian.Uint32(r.data[r.pos:]))
	r.pos += 4
	return val, nil
}

// ReadLong reads an int64 (8 bytes, LE).
func (r *Reader) ReadLong() (int64, error) {
	if r.pos+8 > len(r.data) {
		return 0, fmt.Errorf("ReadLong: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	val := int64(binary.LittleEndian.Uint64(r.data[r.pos:]))
	r.pos += 8
	return val, nil
}

// ReadDouble reads a float64 (8 bytes, LE).
func (r *Reader) ReadDouble() (float64, error) {
	if r.pos+8 > len(r.data) {
		return 0, fmt.Errorf("ReadDouble: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	bits := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return math.Float64frombits(bits), nil
}

// ReadString reads a UTF-16LE null-terminated string.
// Pre-allocates buffer для типичных L2 account names (≤16 chars).
func (r *Reader) ReadString() (string, error) {
	// Pre-allocate с реалистичной capacity для снижения allocations
	utf16Runes := make([]uint16, 0, DefaultStringCapacity)

	for {
		if r.pos+2 > len(r.data) {
			return "", fmt.Errorf("ReadString: unexpected end of data (pos=%d, len=%d)", r.pos, len(r.data))
		}

		rune := binary.LittleEndian.Uint16(r.data[r.pos:])
		r.pos += 2

		if rune == 0 {
			// Null terminator
			break
		}

		utf16Runes = append(utf16Runes, rune)
	}

	// Decode UTF-16 → UTF-8
	decoded := utf16.Decode(utf16Runes)
	return string(decoded), nil
}

// ReadBytes reads n bytes (ZERO-COPY — returns subslice of internal data).
// IMPORTANT: Returned slice shares underlying array with Reader.data.
// Caller MUST NOT modify returned bytes. Use ReadBytesCopy() if mutation needed.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("ReadBytes: negative count %d", n)
	}
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("ReadBytes: not enough data (pos=%d, need=%d, len=%d)", r.pos, n, len(r.data))
	}

	// Zero-copy: return subslice
	bytes := r.data[r.pos : r.pos+n]
	r.pos += n
	return bytes, nil
}

// ReadBytesCopy reads n bytes and returns a MUTABLE COPY.
// Use this when you need to modify returned bytes.
// For read-only access, prefer ReadBytes() (zero-copy).
func (r *Reader) ReadBytesCopy(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("ReadBytesCopy: negative count %d", n)
	}
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("ReadBytesCopy: not enough data (pos=%d, need=%d, len=%d)", r.pos, n, len(r.data))
	}

	// Allocate new slice and copy
	bytes := make([]byte, n)
	copy(bytes, r.data[r.pos:r.pos+n])
	r.pos += n
	return bytes, nil
}

// Remaining returns the number of unread bytes.
func (r *Reader) Remaining() int {
	return len(r.data) - r.pos
}

// Position returns the current read position.
func (r *Reader) Position() int {
	return r.pos
}
