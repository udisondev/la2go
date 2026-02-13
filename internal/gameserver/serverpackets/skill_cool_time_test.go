package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestSkillCoolTime_Write(t *testing.T) {
	t.Parallel()

	pkt := &SkillCoolTime{
		CoolDowns: []SkillCoolDown{
			{SkillID: 100, SkillLevel: 3, ReuseTime: 30, RemainingTime: 15},
			{SkillID: 200, SkillLevel: 1, ReuseTime: 60, RemainingTime: 45},
		},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data[0] != OpcodeSkillCoolTime {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeSkillCoolTime)
	}

	count := int32(binary.LittleEndian.Uint32(data[1:5]))
	if count != 2 {
		t.Errorf("count = %d; want 2", count)
	}

	// First cooldown at offset 5
	sid := int32(binary.LittleEndian.Uint32(data[5:9]))
	if sid != 100 {
		t.Errorf("CoolDowns[0].SkillID = %d; want 100", sid)
	}

	remaining := int32(binary.LittleEndian.Uint32(data[17:21]))
	if remaining != 15 {
		t.Errorf("CoolDowns[0].RemainingTime = %d; want 15", remaining)
	}

	// Second cooldown at offset 21
	sid2 := int32(binary.LittleEndian.Uint32(data[21:25]))
	if sid2 != 200 {
		t.Errorf("CoolDowns[1].SkillID = %d; want 200", sid2)
	}
}

func TestSkillCoolTime_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &SkillCoolTime{}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := int32(binary.LittleEndian.Uint32(data[1:5]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}

	if len(data) != 5 {
		t.Errorf("len(data) = %d; want 5", len(data))
	}
}
