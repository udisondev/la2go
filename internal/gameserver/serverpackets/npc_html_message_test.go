package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestNpcHtmlMessage_Write(t *testing.T) {
	html := "<html><body>Hello World</body></html>"
	pkt := NewNpcHtmlMessage(12345, html)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if len(data) < 5 {
		t.Fatalf("packet too short: %d bytes", len(data))
	}

	// Verify opcode
	if data[0] != OpcodeNpcHtmlMessage {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeNpcHtmlMessage)
	}

	// Verify npcObjectID
	npcObjID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if npcObjID != 12345 {
		t.Errorf("npcObjectID = %d, want 12345", npcObjID)
	}

	// Verify HTML string is present (UTF-16LE null-terminated)
	htmlStr, offset := readUTF16String(t, data, 5)
	if htmlStr != html {
		t.Errorf("html = %q, want %q", htmlStr, html)
	}

	// Verify itemID at end
	if offset+4 > len(data) {
		t.Fatalf("not enough data for itemID at offset %d", offset)
	}
	itemID := int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
	if itemID != 0 {
		t.Errorf("itemID = %d, want 0", itemID)
	}
}

func TestNpcHtmlMessage_Write_WithItemID(t *testing.T) {
	pkt := NpcHtmlMessage{
		NpcObjectID: 0,
		Html:        "test",
		ItemID:      999,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// NpcObjectID should be 0
	npcObjID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if npcObjID != 0 {
		t.Errorf("npcObjectID = %d, want 0", npcObjID)
	}

	// Read past HTML string
	_, offset := readUTF16String(t, data, 5)

	// ItemID
	itemID := int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
	if itemID != 999 {
		t.Errorf("itemID = %d, want 999", itemID)
	}
}

func TestNpcHtmlMessage_Write_EmptyHtml(t *testing.T) {
	pkt := NewNpcHtmlMessage(1, "")
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Opcode + npcObjID(4) + empty string (null terminator = 2 bytes) + itemID(4)
	if data[0] != OpcodeNpcHtmlMessage {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeNpcHtmlMessage)
	}
}
