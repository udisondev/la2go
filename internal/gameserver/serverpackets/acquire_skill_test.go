package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestAcquireSkillInfo_Write(t *testing.T) {
	t.Parallel()

	pkt := &AcquireSkillInfo{
		SkillID:   56,
		Level:     3,
		SpCost:    5000,
		SkillType: 0, // CLASS
		Reqs: []AcquireSkillReq{
			{Type: 99, ItemID: 57, Count: 10, Unk: 50},
		},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data[0] != OpcodeAcquireSkillInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAcquireSkillInfo)
	}

	skillID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if skillID != 56 {
		t.Errorf("SkillID = %d; want 56", skillID)
	}

	level := int32(binary.LittleEndian.Uint32(data[5:9]))
	if level != 3 {
		t.Errorf("Level = %d; want 3", level)
	}

	spCost := int32(binary.LittleEndian.Uint32(data[9:13]))
	if spCost != 5000 {
		t.Errorf("SpCost = %d; want 5000", spCost)
	}

	skillType := int32(binary.LittleEndian.Uint32(data[13:17]))
	if skillType != 0 {
		t.Errorf("SkillType = %d; want 0", skillType)
	}

	reqCount := int32(binary.LittleEndian.Uint32(data[17:21]))
	if reqCount != 1 {
		t.Errorf("ReqCount = %d; want 1", reqCount)
	}

	// First requirement starts at offset 21
	reqType := int32(binary.LittleEndian.Uint32(data[21:25]))
	if reqType != 99 {
		t.Errorf("Req.Type = %d; want 99", reqType)
	}

	itemID := int32(binary.LittleEndian.Uint32(data[25:29]))
	if itemID != 57 {
		t.Errorf("Req.ItemID = %d; want 57", itemID)
	}

	count := int64(binary.LittleEndian.Uint64(data[29:37]))
	if count != 10 {
		t.Errorf("Req.Count = %d; want 10", count)
	}
}

func TestAcquireSkillInfo_WriteNoReqs(t *testing.T) {
	t.Parallel()

	pkt := &AcquireSkillInfo{
		SkillID:   100,
		Level:     1,
		SpCost:    200,
		SkillType: 0,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reqCount := int32(binary.LittleEndian.Uint32(data[17:21]))
	if reqCount != 0 {
		t.Errorf("ReqCount = %d; want 0", reqCount)
	}

	// total: 1 (opcode) + 5*4 = 21 bytes
	if len(data) != 21 {
		t.Errorf("len(data) = %d; want 21", len(data))
	}
}

func TestAcquireSkillList_Write(t *testing.T) {
	t.Parallel()

	pkt := &AcquireSkillList{
		SkillType: 0,
		Skills: []AcquireSkillEntry{
			{SkillID: 1, NextLevel: 2, MaxLevel: 5, SpCost: 1000, Requirements: 0},
			{SkillID: 3, NextLevel: 1, MaxLevel: 3, SpCost: 500, Requirements: 0},
		},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data[0] != OpcodeAcquireSkillList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAcquireSkillList)
	}

	skillType := int32(binary.LittleEndian.Uint32(data[1:5]))
	if skillType != 0 {
		t.Errorf("SkillType = %d; want 0", skillType)
	}

	count := int32(binary.LittleEndian.Uint32(data[5:9]))
	if count != 2 {
		t.Errorf("count = %d; want 2", count)
	}

	// First skill at offset 9
	sid := int32(binary.LittleEndian.Uint32(data[9:13]))
	if sid != 1 {
		t.Errorf("Skills[0].SkillID = %d; want 1", sid)
	}
}

func TestAcquireSkillList_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &AcquireSkillList{
		SkillType: 1,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := int32(binary.LittleEndian.Uint32(data[5:9]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}

	// 1 + 4 + 4 = 9 bytes
	if len(data) != 9 {
		t.Errorf("len(data) = %d; want 9", len(data))
	}
}
