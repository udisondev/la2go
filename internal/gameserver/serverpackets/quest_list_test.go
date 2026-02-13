package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestQuestList_Write_Empty(t *testing.T) {
	pkt := NewQuestList()
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// 1 opcode + 2 count = 3
	if len(data) != 3 {
		t.Fatalf("packet length = %d, want 3", len(data))
	}

	if data[0] != OpcodeQuestList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeQuestList)
	}

	count := int16(binary.LittleEndian.Uint16(data[1:3]))
	if count != 0 {
		t.Errorf("quest count = %d, want 0", count)
	}
}

func TestQuestList_Write_WithQuests(t *testing.T) {
	entries := []QuestEntry{
		{QuestID: 303, State: 1},
		{QuestID: 255, State: 2},
	}
	pkt := NewQuestListWithEntries(entries)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// 1 opcode + 2 count + 2 * (4 + 4) = 19
	if len(data) != 19 {
		t.Fatalf("packet length = %d, want 19", len(data))
	}

	count := int16(binary.LittleEndian.Uint16(data[1:3]))
	if count != 2 {
		t.Errorf("quest count = %d, want 2", count)
	}

	// Quest 1
	questID1 := int32(binary.LittleEndian.Uint32(data[3:7]))
	if questID1 != 303 {
		t.Errorf("quest1 ID = %d, want 303", questID1)
	}
	state1 := int32(binary.LittleEndian.Uint32(data[7:11]))
	if state1 != 1 {
		t.Errorf("quest1 state = %d, want 1", state1)
	}

	// Quest 2
	questID2 := int32(binary.LittleEndian.Uint32(data[11:15]))
	if questID2 != 255 {
		t.Errorf("quest2 ID = %d, want 255", questID2)
	}
	state2 := int32(binary.LittleEndian.Uint32(data[15:19]))
	if state2 != 2 {
		t.Errorf("quest2 state = %d, want 2", state2)
	}
}
