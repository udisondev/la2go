package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestOpcodeRequestShowBoard(t *testing.T) {
	if OpcodeRequestShowBoard != 0x57 {
		t.Errorf("OpcodeRequestShowBoard = 0x%02X; want 0x57", OpcodeRequestShowBoard)
	}
}

func TestParseRequestShowBoard(t *testing.T) {
	// 4 bytes — unused int32 (little-endian)
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 42)

	pkt, err := ParseRequestShowBoard(data)
	if err != nil {
		t.Fatalf("ParseRequestShowBoard: %v", err)
	}
	if pkt.Unknown != 42 {
		t.Errorf("Unknown = %d; want 42", pkt.Unknown)
	}
}

func TestParseRequestShowBoard_Empty(t *testing.T) {
	// Пустые данные — Java тоже игнорирует
	pkt, err := ParseRequestShowBoard(nil)
	if err != nil {
		t.Fatalf("ParseRequestShowBoard: %v", err)
	}
	if pkt.Unknown != 0 {
		t.Errorf("Unknown = %d; want 0", pkt.Unknown)
	}
}
