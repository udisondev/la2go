package packet

import (
	"encoding/binary"
	"testing"
)

func TestWriter_WriteByte(t *testing.T) {
	w := NewWriter(16)

	if err := w.WriteByte(0x42); err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}

	data := w.Bytes()
	if len(data) != 1 {
		t.Fatalf("expected length 1, got %d", len(data))
	}
	if data[0] != 0x42 {
		t.Errorf("expected byte 0x42, got 0x%02X", data[0])
	}
}

func TestWriter_WriteShort(t *testing.T) {
	w := NewWriter(16)

	w.WriteShort(0x1234)

	data := w.Bytes()
	if len(data) != 2 {
		t.Fatalf("expected length 2, got %d", len(data))
	}

	val := int16(binary.LittleEndian.Uint16(data))
	if val != 0x1234 {
		t.Errorf("expected 0x1234, got 0x%04X", val)
	}
}

func TestWriter_WriteInt(t *testing.T) {
	w := NewWriter(16)

	w.WriteInt(0x12345678)

	data := w.Bytes()
	if len(data) != 4 {
		t.Fatalf("expected length 4, got %d", len(data))
	}

	val := int32(binary.LittleEndian.Uint32(data))
	if val != 0x12345678 {
		t.Errorf("expected 0x12345678, got 0x%08X", val)
	}
}

func TestWriter_WriteLong(t *testing.T) {
	w := NewWriter(16)

	w.WriteLong(0x123456789ABCDEF0)

	data := w.Bytes()
	if len(data) != 8 {
		t.Fatalf("expected length 8, got %d", len(data))
	}

	val := int64(binary.LittleEndian.Uint64(data))
	if val != 0x123456789ABCDEF0 {
		t.Errorf("expected 0x123456789ABCDEF0, got 0x%016X", val)
	}
}

func TestWriter_WriteString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []uint16 // UTF-16LE encoding + null terminator
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []uint16{0x0000},
		},
		{
			name:     "ASCII string",
			input:    "hello",
			expected: []uint16{0x0068, 0x0065, 0x006C, 0x006C, 0x006F, 0x0000},
		},
		{
			name:     "Russian string",
			input:    "привет",
			expected: []uint16{0x043F, 0x0440, 0x0438, 0x0432, 0x0435, 0x0442, 0x0000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWriter(64)
			w.WriteString(tt.input)

			data := w.Bytes()
			expectedLen := len(tt.expected) * 2 // uint16 = 2 bytes
			if len(data) != expectedLen {
				t.Fatalf("expected length %d, got %d", expectedLen, len(data))
			}

			for i, expected := range tt.expected {
				offset := i * 2
				val := binary.LittleEndian.Uint16(data[offset:])
				if val != expected {
					t.Errorf("at index %d: expected 0x%04X, got 0x%04X", i, expected, val)
				}
			}
		})
	}
}

func TestWriter_WriteBytes(t *testing.T) {
	w := NewWriter(16)

	input := []byte{0x11, 0x22, 0x33, 0x44}
	w.WriteBytes(input)

	data := w.Bytes()
	if len(data) != 4 {
		t.Fatalf("expected length 4, got %d", len(data))
	}

	for i, expected := range input {
		if data[i] != expected {
			t.Errorf("at index %d: expected 0x%02X, got 0x%02X", i, expected, data[i])
		}
	}
}

func TestWriter_Multiple(t *testing.T) {
	w := NewWriter(32)

	if err := w.WriteByte(0x2E); err != nil { // opcode KeyPacket
		t.Fatalf("WriteByte failed: %v", err)
	}
	if err := w.WriteByte(0x01); err != nil { // protocol version
		t.Fatalf("WriteByte failed: %v", err)
	}
	w.WriteInt(0x12345678)         // some data
	w.WriteString("test")          // string

	data := w.Bytes()

	// Verify opcode
	if data[0] != 0x2E {
		t.Errorf("expected opcode 0x2E, got 0x%02X", data[0])
	}

	// Verify protocol version
	if data[1] != 0x01 {
		t.Errorf("expected protocol version 0x01, got 0x%02X", data[1])
	}

	// Verify int32
	val := int32(binary.LittleEndian.Uint32(data[2:]))
	if val != 0x12345678 {
		t.Errorf("expected int32 0x12345678, got 0x%08X", val)
	}

	// Verify string (UTF-16LE: 't' 'e' 's' 't' '\0')
	expectedString := []uint16{0x0074, 0x0065, 0x0073, 0x0074, 0x0000}
	for i, expected := range expectedString {
		offset := 6 + i*2 // 1+1+4 = 6 bytes before string
		val := binary.LittleEndian.Uint16(data[offset:])
		if val != expected {
			t.Errorf("at string index %d: expected 0x%04X, got 0x%04X", i, expected, val)
		}
	}
}

func TestWriter_Reset(t *testing.T) {
	w := NewWriter(16)

	w.WriteInt(0x12345678)
	if w.Len() != 4 {
		t.Fatalf("expected length 4 before reset, got %d", w.Len())
	}

	w.Reset()

	if w.Len() != 0 {
		t.Errorf("expected length 0 after reset, got %d", w.Len())
	}

	w.WriteShort(0x1234)
	data := w.Bytes()
	if len(data) != 2 {
		t.Errorf("expected length 2 after reset+write, got %d", len(data))
	}
}
