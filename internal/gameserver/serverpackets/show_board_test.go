package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/game/bbs"
)

func TestShowBoard_Write_Show(t *testing.T) {
	pkt := NewShowBoard("101", "<html>Hello</html>")

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeShowBoard {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeShowBoard)
	}
	if data[1] != 0x01 {
		t.Errorf("show byte = 0x%02X; want 0x01", data[1])
	}
}

func TestShowBoard_Write_Hide(t *testing.T) {
	pkt := NewShowBoardHide()

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeShowBoard {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeShowBoard)
	}
	if data[1] != 0x00 {
		t.Errorf("show byte = 0x%02X; want 0x00", data[1])
	}
}

func TestShowBoard_Content(t *testing.T) {
	pkt := NewShowBoard("101", "<html>test</html>")

	expected := bbs.FormatContent("101", "<html>test</html>")
	if pkt.Content != expected {
		t.Errorf("Content = %q; want %q", pkt.Content, expected)
	}
}

func TestShowBoard_Opcode(t *testing.T) {
	if OpcodeShowBoard != 0x6E {
		t.Errorf("OpcodeShowBoard = 0x%02X; want 0x6E", OpcodeShowBoard)
	}
}
