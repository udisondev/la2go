package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestPartySpelled_Write(t *testing.T) {
	t.Parallel()

	pkt := &PartySpelled{
		Type:     0, // player
		ObjectID: 12345,
		Effects: []PartyEffect{
			{SkillID: 1204, SkillLevel: 3, Duration: 1800},
			{SkillID: 1068, SkillLevel: 1, Duration: 900},
		},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data[0] != OpcodePartySpelled {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodePartySpelled)
	}

	creatureType := int32(binary.LittleEndian.Uint32(data[1:5]))
	if creatureType != 0 {
		t.Errorf("Type = %d; want 0", creatureType)
	}

	objectID := int32(binary.LittleEndian.Uint32(data[5:9]))
	if objectID != 12345 {
		t.Errorf("ObjectID = %d; want 12345", objectID)
	}

	count := int32(binary.LittleEndian.Uint32(data[9:13]))
	if count != 2 {
		t.Errorf("EffectCount = %d; want 2", count)
	}

	// First effect at offset 13: skillID(4) + skillLevel(2) + duration(4) = 10 bytes
	sid := int32(binary.LittleEndian.Uint32(data[13:17]))
	if sid != 1204 {
		t.Errorf("Effects[0].SkillID = %d; want 1204", sid)
	}

	level := int16(binary.LittleEndian.Uint16(data[17:19]))
	if level != 3 {
		t.Errorf("Effects[0].SkillLevel = %d; want 3", level)
	}

	duration := int32(binary.LittleEndian.Uint32(data[19:23]))
	if duration != 1800 {
		t.Errorf("Effects[0].Duration = %d; want 1800", duration)
	}
}

func TestPartySpelled_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &PartySpelled{
		Type:     1, // pet
		ObjectID: 999,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := int32(binary.LittleEndian.Uint32(data[9:13]))
	if count != 0 {
		t.Errorf("EffectCount = %d; want 0", count)
	}

	// 1 + 4 + 4 + 4 = 13
	if len(data) != 13 {
		t.Errorf("len(data) = %d; want 13", len(data))
	}
}
