package serverpackets

import "testing"

func TestPledgeInfo_Write(t *testing.T) {
	p := NewPledgeInfo(1, "TestClan", "TestAlly")
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeInfo {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeInfo)
	}
	if len(data) < 10 {
		t.Errorf("data too short: %d bytes", len(data))
	}
}

func TestPledgeShowInfoUpdate_Write(t *testing.T) {
	p := NewPledgeShowInfoUpdate(1, 42, 5, 0, 0, 10, "Alliance", 11, 1)
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeShowInfoUpdate {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeShowInfoUpdate)
	}
}

func TestPledgeShowMemberListAll_Write(t *testing.T) {
	p := &PledgeShowMemberListAll{
		PledgeType: 0,
		ClanName:   "TestClan",
		LeaderName: "Leader",
		CrestID:    1,
		ClanLevel:  5,
		Reputation: 1000,
		Members: []PledgeMemberEntry{
			{Name: "Player1", Level: 40, ClassID: 5, Online: 1},
			{Name: "Player2", Level: 60, ClassID: 10, Online: 0},
		},
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeShowMemberListAll {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeShowMemberListAll)
	}
}

func TestPledgeShowMemberListUpdate_Write(t *testing.T) {
	p := &PledgeShowMemberListUpdate{
		Name: "Player1", Level: 40, ClassID: 5, Online: 1,
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeShowMemberListUpdate {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeShowMemberListUpdate)
	}
}

func TestPledgeShowMemberListDelete_Write(t *testing.T) {
	p := &PledgeShowMemberListDelete{Name: "Player1"}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeShowMemberListDelete {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeShowMemberListDelete)
	}
}

func TestAskJoinPledge_Write(t *testing.T) {
	p := NewAskJoinPledge(100, "TestClan")
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeAskJoinPledge {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeAskJoinPledge)
	}
}

func TestJoinPledge_Write(t *testing.T) {
	p := NewJoinPledge(42)
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeJoinPledge {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeJoinPledge)
	}
	if len(data) != 5 { // opcode(1) + int32(4)
		t.Errorf("len = %d, want 5", len(data))
	}
}

func TestPledgeReceiveWarList_Write(t *testing.T) {
	p := &PledgeReceiveWarList{
		Tab: 0,
		Entries: []PledgeWarEntry{
			{ClanName: "Enemy1"},
			{ClanName: "Enemy2"},
		},
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeReceiveWarList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeReceiveWarList)
	}
}

func TestPledgeReceiveWarList_Empty(t *testing.T) {
	p := &PledgeReceiveWarList{Tab: 1}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeReceiveWarList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeReceiveWarList)
	}
}

func TestPledgePowerGradeList_Write(t *testing.T) {
	p := &PledgePowerGradeList{
		Entries: []PledgePowerGradeEntry{
			{PowerGrade: 1, Privileges: 0xFFFFFF, MemberCount: 1},
			{PowerGrade: 2, Privileges: 0xFF, MemberCount: 5},
		},
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgePowerGradeList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgePowerGradeList)
	}
}

func TestPledgeSkillList_Write(t *testing.T) {
	p := &PledgeSkillList{
		Skills: []PledgeSkillEntry{
			{SkillID: 370, SkillLevel: 3},
		},
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeSkillList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeSkillList)
	}
}

func TestPledgeStatusChanged_Write(t *testing.T) {
	p := &PledgeStatusChanged{
		LeaderID: 100, ClanID: 1, CrestID: 5, AllyID: 0, AllyCrestID: 0,
	}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodePledgeStatusChanged {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePledgeStatusChanged)
	}
	// Regular packet (0xCD), not extended — no sub-opcode
	if data[0] != 0xCD {
		t.Errorf("opcode = 0x%02X, want 0xCD", data[0])
	}
}
