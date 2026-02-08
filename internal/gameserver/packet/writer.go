package packet

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
)

// Writer provides methods for writing packet data.
// Uses Little-Endian byte order for all multi-byte values.
type Writer struct {
	buf *bytes.Buffer
}

// writerPool reduces allocations by reusing Writers.
// Get() returns a Writer with Reset() called, Put() returns it to pool.
var writerPool = sync.Pool{
	New: func() any {
		return &Writer{
			buf: bytes.NewBuffer(make([]byte, 0, 512)),
		}
	},
}

// Get returns a Writer from the pool (already Reset).
func Get() *Writer {
	w := writerPool.Get().(*Writer)
	w.Reset()
	return w
}

// Put returns a Writer to the pool for reuse.
// IMPORTANT: Do not use the Writer after calling Put.
func (w *Writer) Put() {
	writerPool.Put(w)
}

// NewWriter creates a new packet writer with the given initial capacity.
func NewWriter(capacity int) *Writer {
	return &Writer{
		buf: bytes.NewBuffer(make([]byte, 0, capacity)),
	}
}

// WriteByte writes a single byte.
func (w *Writer) WriteByte(b byte) error {
	return w.buf.WriteByte(b)
}

// WriteShort writes an int16 (2 bytes, LE).
// Optimized: manual encoding instead of binary.Write.
func (w *Writer) WriteShort(val int16) {
	w.buf.WriteByte(byte(val))
	w.buf.WriteByte(byte(val >> 8))
}

// WriteInt writes an int32 (4 bytes, LE).
// Optimized: manual encoding instead of binary.Write.
func (w *Writer) WriteInt(val int32) {
	w.buf.WriteByte(byte(val))
	w.buf.WriteByte(byte(val >> 8))
	w.buf.WriteByte(byte(val >> 16))
	w.buf.WriteByte(byte(val >> 24))
}

// WriteLong writes an int64 (8 bytes, LE).
// Optimized: manual encoding instead of binary.Write.
func (w *Writer) WriteLong(val int64) {
	w.buf.WriteByte(byte(val))
	w.buf.WriteByte(byte(val >> 8))
	w.buf.WriteByte(byte(val >> 16))
	w.buf.WriteByte(byte(val >> 24))
	w.buf.WriteByte(byte(val >> 32))
	w.buf.WriteByte(byte(val >> 40))
	w.buf.WriteByte(byte(val >> 48))
	w.buf.WriteByte(byte(val >> 56))
}

// WriteDouble writes a float64 (8 bytes, LE).
// Uses binary.LittleEndian.PutUint64 for correct IEEE 754 encoding.
func (w *Writer) WriteDouble(val float64) {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], math.Float64bits(val))
	w.buf.Write(tmp[:])
}

// WriteString writes a UTF-16LE null-terminated string.
// Each Go rune is encoded as uint16 (may need surrogates for runes > 0xFFFF).
// Optimized: manual encoding instead of binary.Write to reduce allocations.
func (w *Writer) WriteString(s string) {
	// Pre-allocate buffer space (estimate: len(s) runes → len(s)*2 bytes + 2 null terminator)
	// Actual size may be larger if surrogates are needed (runes > 0xFFFF)
	estimatedSize := len(s)*2 + 2
	if w.buf.Cap()-w.buf.Len() < estimatedSize {
		w.buf.Grow(estimatedSize)
	}

	// Manual UTF-16LE encoding (zero allocations for BMP runes)
	for _, r := range s {
		if r <= 0xFFFF {
			// Basic Multilingual Plane (BMP) — single uint16
			w.buf.WriteByte(byte(r))
			w.buf.WriteByte(byte(r >> 8))
		} else {
			// Supplementary planes — surrogate pair
			r := r - 0x10000
			high := uint16((r >> 10) + 0xD800)
			low := uint16((r & 0x3FF) + 0xDC00)
			w.buf.WriteByte(byte(high))
			w.buf.WriteByte(byte(high >> 8))
			w.buf.WriteByte(byte(low))
			w.buf.WriteByte(byte(low >> 8))
		}
	}

	// Null terminator (UTF-16LE: 0x0000)
	w.buf.WriteByte(0x00)
	w.buf.WriteByte(0x00)
}

// WriteBytes writes raw bytes.
func (w *Writer) WriteBytes(data []byte) {
	_, _ = w.buf.Write(data)
}

// Bytes returns the accumulated packet data.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// Len returns the current length of the packet.
func (w *Writer) Len() int {
	return w.buf.Len()
}

// Reset clears the buffer for reuse.
func (w *Writer) Reset() {
	w.buf.Reset()
}
