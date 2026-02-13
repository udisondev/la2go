package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestExShowQuestMark_Write(t *testing.T) {
	pkt := NewExShowQuestMark(303, 1)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// 1 opcode + 2 subop + 4 questID + 4 state = 11
	if len(data) != 11 {
		t.Fatalf("packet length = %d, want 11", len(data))
	}

	if data[0] != OpcodeExShowQuestMark {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeExShowQuestMark)
	}

	subOp := int16(binary.LittleEndian.Uint16(data[1:3]))
	if subOp != SubOpcodeExShowQuestMark {
		t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExShowQuestMark)
	}

	questID := int32(binary.LittleEndian.Uint32(data[3:7]))
	if questID != 303 {
		t.Errorf("questID = %d, want 303", questID)
	}

	state := int32(binary.LittleEndian.Uint32(data[7:11]))
	if state != 1 {
		t.Errorf("state = %d, want 1", state)
	}
}

func TestExShowQuestMark_Write_AllStates(t *testing.T) {
	tests := []struct {
		name  string
		state int32
	}{
		{"available", 0},
		{"in_progress", 1},
		{"completed", 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkt := NewExShowQuestMark(100, tc.state)
			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			got := int32(binary.LittleEndian.Uint32(data[7:11]))
			if got != tc.state {
				t.Errorf("state = %d, want %d", got, tc.state)
			}
		})
	}
}
