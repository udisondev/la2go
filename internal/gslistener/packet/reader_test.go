package packet

import (
	"encoding/binary"
	"testing"
)

func TestReader_ReadByte(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    byte
		wantErr bool
	}{
		{
			name:    "успешное чтение",
			data:    []byte{0x42, 0x00},
			want:    0x42,
			wantErr: false,
		},
		{
			name:    "пустые данные",
			data:    []byte{},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			got, err := r.ReadByte()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadByte() = 0x%02x, want 0x%02x", got, tt.want)
			}
		})
	}
}

func TestReader_ReadShort(t *testing.T) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, 0x1234)

	r := NewReader(data)
	got, err := r.ReadShort()
	if err != nil {
		t.Fatalf("ReadShort() error = %v", err)
	}

	want := int16(0x1234)
	if got != want {
		t.Errorf("ReadShort() = 0x%04x, want 0x%04x", got, want)
	}
}

func TestReader_ReadInt(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 0x12345678)

	r := NewReader(data)
	got, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt() error = %v", err)
	}

	want := int32(0x12345678)
	if got != want {
		t.Errorf("ReadInt() = 0x%08x, want 0x%08x", got, want)
	}
}

func TestReader_ReadString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "простая ASCII строка",
			input:   "Hello",
			wantErr: false,
		},
		{
			name:    "пустая строка",
			input:   "",
			wantErr: false,
		},
		{
			name:    "строка с Unicode",
			input:   "Привет",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Кодируем строку в UTF-16LE
			data := encodeUTF16LE(tt.input)

			r := NewReader(data)
			got, err := r.ReadString()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.input {
				t.Errorf("ReadString() = %q, want %q", got, tt.input)
			}
		})
	}
}

func TestReader_ReadBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	r := NewReader(data)

	got, err := r.ReadBytes(3)
	if err != nil {
		t.Fatalf("ReadBytes(3) error = %v", err)
	}

	want := []byte{0x01, 0x02, 0x03}
	if len(got) != len(want) {
		t.Fatalf("ReadBytes() len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadBytes()[%d] = 0x%02x, want 0x%02x", i, got[i], want[i])
		}
	}

	// Проверяем remaining
	if r.Remaining() != 2 {
		t.Errorf("Remaining() = %d, want 2", r.Remaining())
	}
}

func TestReader_Remaining(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	r := NewReader(data)

	if r.Remaining() != 4 {
		t.Errorf("initial Remaining() = %d, want 4", r.Remaining())
	}

	_, _ = r.ReadByte()
	if r.Remaining() != 3 {
		t.Errorf("after ReadByte() Remaining() = %d, want 3", r.Remaining())
	}

	_, _ = r.ReadShort()
	if r.Remaining() != 1 {
		t.Errorf("after ReadShort() Remaining() = %d, want 1", r.Remaining())
	}
}

// encodeUTF16LE кодирует строку в UTF-16LE с null terminator
func encodeUTF16LE(s string) []byte {
	runes := []rune(s)
	encoded := make([]byte, 0, (len(runes)+1)*2)

	for _, r := range runes {
		encoded = append(encoded, byte(r), byte(r>>8))
	}

	// Null terminator
	encoded = append(encoded, 0x00, 0x00)
	return encoded
}
