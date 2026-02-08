package packet

import (
	"encoding/binary"
	"testing"
)

func TestReader_ReadByte(t *testing.T) {
	data := []byte{0x42}
	r := NewReader(data)

	val, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}

	if val != 0x42 {
		t.Errorf("expected 0x42, got 0x%02X", val)
	}

	if r.Remaining() != 0 {
		t.Errorf("expected 0 remaining bytes, got %d", r.Remaining())
	}
}

func TestReader_ReadShort(t *testing.T) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, 0x1234)

	r := NewReader(data)

	val, err := r.ReadShort()
	if err != nil {
		t.Fatalf("ReadShort failed: %v", err)
	}

	if val != 0x1234 {
		t.Errorf("expected 0x1234, got 0x%04X", val)
	}
}

func TestReader_ReadInt(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 0x12345678)

	r := NewReader(data)

	val, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt failed: %v", err)
	}

	if val != 0x12345678 {
		t.Errorf("expected 0x12345678, got 0x%08X", val)
	}
}

func TestReader_ReadLong(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0x123456789ABCDEF0)

	r := NewReader(data)

	val, err := r.ReadLong()
	if err != nil {
		t.Fatalf("ReadLong failed: %v", err)
	}

	if val != 0x123456789ABCDEF0 {
		t.Errorf("expected 0x123456789ABCDEF0, got 0x%016X", val)
	}
}

func TestReader_ReadString(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16 // UTF-16LE encoding + null terminator
		expected string
	}{
		{
			name:     "empty string",
			input:    []uint16{0x0000},
			expected: "",
		},
		{
			name:     "ASCII string",
			input:    []uint16{0x0068, 0x0065, 0x006C, 0x006C, 0x006F, 0x0000},
			expected: "hello",
		},
		{
			name:     "Russian string",
			input:    []uint16{0x043F, 0x0440, 0x0438, 0x0432, 0x0435, 0x0442, 0x0000},
			expected: "привет",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode UTF-16LE
			data := make([]byte, len(tt.input)*2)
			for i, r := range tt.input {
				binary.LittleEndian.PutUint16(data[i*2:], r)
			}

			r := NewReader(data)

			val, err := r.ReadString()
			if err != nil {
				t.Fatalf("ReadString failed: %v", err)
			}

			if val != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, val)
			}
		})
	}
}

func TestReader_ReadBytes(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}
	r := NewReader(data)

	val, err := r.ReadBytes(4)
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	if len(val) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(val))
	}

	for i, expected := range data {
		if val[i] != expected {
			t.Errorf("at index %d: expected 0x%02X, got 0x%02X", i, expected, val[i])
		}
	}
}

func TestReader_ReadByte_NotEnoughData(t *testing.T) {
	data := []byte{}
	r := NewReader(data)

	_, err := r.ReadByte()
	if err == nil {
		t.Error("expected error when reading byte from empty buffer")
	}
}

func TestReader_ReadInt_NotEnoughData(t *testing.T) {
	data := []byte{0x11, 0x22}
	r := NewReader(data)

	_, err := r.ReadInt()
	if err == nil {
		t.Error("expected error when reading int32 from 2-byte buffer")
	}
}

func TestReader_ReadString_NotEnoughData(t *testing.T) {
	data := []byte{0x68, 0x00, 0x65} // incomplete UTF-16LE
	r := NewReader(data)

	_, err := r.ReadString()
	if err == nil {
		t.Error("expected error when reading incomplete string")
	}
}

func TestReader_Remaining(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55}
	r := NewReader(data)

	if r.Remaining() != 5 {
		t.Errorf("expected 5 remaining bytes, got %d", r.Remaining())
	}

	_, _ = r.ReadByte()
	if r.Remaining() != 4 {
		t.Errorf("expected 4 remaining bytes after ReadByte, got %d", r.Remaining())
	}

	_, _ = r.ReadInt()
	if r.Remaining() != 0 {
		t.Errorf("expected 0 remaining bytes after ReadInt, got %d", r.Remaining())
	}
}

func TestReader_Position(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55}
	r := NewReader(data)

	if r.Position() != 0 {
		t.Errorf("expected position 0, got %d", r.Position())
	}

	_, _ = r.ReadByte()
	if r.Position() != 1 {
		t.Errorf("expected position 1 after ReadByte, got %d", r.Position())
	}

	_, _ = r.ReadInt()
	if r.Position() != 5 {
		t.Errorf("expected position 5 after ReadInt, got %d", r.Position())
	}
}
