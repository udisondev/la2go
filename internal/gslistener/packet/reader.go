package packet

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// Reader предоставляет методы для чтения данных из пакета GS→LS.
// Использует Little-Endian byte order для всех многобайтовых значений.
type Reader struct {
	data []byte
	pos  int
}

// NewReader создаёт новый Reader для чтения пакета.
func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
		pos:  0,
	}
}

// ReadByte читает 1 байт.
func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("ReadByte: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// ReadShort читает int16 (2 байта, LE).
func (r *Reader) ReadShort() (int16, error) {
	if r.pos+2 > len(r.data) {
		return 0, fmt.Errorf("ReadShort: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	val := int16(binary.LittleEndian.Uint16(r.data[r.pos:]))
	r.pos += 2
	return val, nil
}

// ReadInt читает int32 (4 байта, LE).
func (r *Reader) ReadInt() (int32, error) {
	if r.pos+4 > len(r.data) {
		return 0, fmt.Errorf("ReadInt: not enough data (pos=%d, len=%d)", r.pos, len(r.data))
	}
	val := int32(binary.LittleEndian.Uint32(r.data[r.pos:]))
	r.pos += 4
	return val, nil
}

// ReadString читает UTF-16LE null-terminated строку.
// Читает пары байт (uint16 LE) до тех пор, пока не встретит 0x0000.
func (r *Reader) ReadString() (string, error) {
	var utf16Runes []uint16

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

	// Декодируем UTF-16 в UTF-8
	decoded := utf16.Decode(utf16Runes)
	return string(decoded), nil
}

// ReadBytes читает n байт.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("ReadBytes: negative count %d", n)
	}
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("ReadBytes: not enough data (pos=%d, need=%d, len=%d)", r.pos, n, len(r.data))
	}

	bytes := make([]byte, n)
	copy(bytes, r.data[r.pos:r.pos+n])
	r.pos += n
	return bytes, nil
}

// Remaining возвращает количество непрочитанных байт.
func (r *Reader) Remaining() int {
	return len(r.data) - r.pos
}

// Position возвращает текущую позицию чтения.
func (r *Reader) Position() int {
	return r.pos
}
