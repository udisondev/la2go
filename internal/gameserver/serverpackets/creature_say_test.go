package serverpackets

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestCreatureSay_Write(t *testing.T) {
	pkt := NewCreatureSay(12345, 0, "TestPlayer", "Hello World")
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if len(data) < 9 {
		t.Fatalf("packet too short: %d bytes", len(data))
	}

	// Verify opcode
	if data[0] != OpcodeCreatureSay {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeCreatureSay)
	}

	// Verify objectID
	objectID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objectID != 12345 {
		t.Errorf("objectID = %d, want 12345", objectID)
	}

	// Verify chatType
	chatType := int32(binary.LittleEndian.Uint32(data[5:9]))
	if chatType != 0 {
		t.Errorf("chatType = %d, want 0", chatType)
	}

	// Verify senderName (UTF-16LE null-terminated starting at offset 9)
	name, offset := readUTF16String(t, data, 9)
	if name != "TestPlayer" {
		t.Errorf("senderName = %q, want %q", name, "TestPlayer")
	}

	// Verify text
	text, _ := readUTF16String(t, data, offset)
	if text != "Hello World" {
		t.Errorf("text = %q, want %q", text, "Hello World")
	}
}

func TestCreatureSay_Write_Shout(t *testing.T) {
	pkt := NewCreatureSay(0, 1, "System", "Server restarting")
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	objectID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objectID != 0 {
		t.Errorf("objectID = %d, want 0", objectID)
	}

	chatType := int32(binary.LittleEndian.Uint32(data[5:9]))
	if chatType != 1 {
		t.Errorf("chatType = %d, want 1 (SHOUT)", chatType)
	}
}

// readUTF16String reads a UTF-16LE null-terminated string from data at given offset.
// Returns the decoded string and the offset after the null terminator.
func readUTF16String(t *testing.T, data []byte, offset int) (string, int) {
	t.Helper()

	var runes []uint16
	pos := offset
	for {
		if pos+2 > len(data) {
			t.Fatalf("unexpected end of data reading string at offset %d", offset)
		}
		r := binary.LittleEndian.Uint16(data[pos:])
		pos += 2
		if r == 0 {
			break
		}
		runes = append(runes, r)
	}
	return string(utf16.Decode(runes)), pos
}
