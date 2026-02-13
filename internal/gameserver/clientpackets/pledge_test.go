package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestJoinPledge(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteInt(42)  // objectID
	w.WriteInt(-1)  // pledgeType (academy)

	pkt, err := ParseRequestJoinPledge(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestJoinPledge: %v", err)
	}
	if pkt.ObjectID != 42 {
		t.Errorf("ObjectID = %d, want 42", pkt.ObjectID)
	}
	if pkt.PledgeType != -1 {
		t.Errorf("PledgeType = %d, want -1", pkt.PledgeType)
	}
}

func TestParseRequestAnswerJoinPledge(t *testing.T) {
	w := packet.NewWriter(8)
	w.WriteInt(1) // accept

	pkt, err := ParseRequestAnswerJoinPledge(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.Answer != 1 {
		t.Errorf("Answer = %d, want 1", pkt.Answer)
	}
}

func TestParseRequestOustPledgeMember(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("TestPlayer")

	pkt, err := ParseRequestOustPledgeMember(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.Name != "TestPlayer" {
		t.Errorf("Name = %q, want %q", pkt.Name, "TestPlayer")
	}
}

func TestParseRequestPledgeInfo(t *testing.T) {
	w := packet.NewWriter(8)
	w.WriteInt(5)

	pkt, err := ParseRequestPledgeInfo(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.ClanID != 5 {
		t.Errorf("ClanID = %d, want 5", pkt.ClanID)
	}
}

func TestParseRequestPledgeCrest(t *testing.T) {
	w := packet.NewWriter(8)
	w.WriteInt(42)

	pkt, err := ParseRequestPledgeCrest(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.CrestID != 42 {
		t.Errorf("CrestID = %d, want 42", pkt.CrestID)
	}
}

func TestParseRequestPledgeSetMemberPowerGrade(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("TestPlayer")
	w.WriteInt(3) // grade

	pkt, err := ParseRequestPledgeSetMemberPowerGrade(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.MemberName != "TestPlayer" {
		t.Errorf("MemberName = %q, want %q", pkt.MemberName, "TestPlayer")
	}
	if pkt.PowerGrade != 3 {
		t.Errorf("PowerGrade = %d, want 3", pkt.PowerGrade)
	}
}

func TestParseRequestPledgeReorganizeMember(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("TestPlayer")
	w.WriteInt(100) // Royal Guard 1

	pkt, err := ParseRequestPledgeReorganizeMember(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.MemberName != "TestPlayer" {
		t.Errorf("MemberName = %q, want %q", pkt.MemberName, "TestPlayer")
	}
	if pkt.NewPledgeType != 100 {
		t.Errorf("NewPledgeType = %d, want 100", pkt.NewPledgeType)
	}
}

func TestParseRequestPledgePower(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteInt(3)     // grade
	w.WriteInt(0xFF)  // privs

	pkt, err := ParseRequestPledgePower(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.PowerGrade != 3 {
		t.Errorf("PowerGrade = %d, want 3", pkt.PowerGrade)
	}
	if pkt.Privileges != 0xFF {
		t.Errorf("Privileges = %d, want 255", pkt.Privileges)
	}
}

func TestParseRequestPledgeWarList(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteInt(1) // page
	w.WriteInt(0) // tab

	pkt, err := ParseRequestPledgeWarList(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.Page != 1 {
		t.Errorf("Page = %d, want 1", pkt.Page)
	}
	if pkt.Tab != 0 {
		t.Errorf("Tab = %d, want 0", pkt.Tab)
	}
}

func TestParseRequestJoinPledge_ShortData(t *testing.T) {
	_, err := ParseRequestJoinPledge([]byte{0x01}) // too short
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseRequestAnswerJoinPledge_ShortData(t *testing.T) {
	_, err := ParseRequestAnswerJoinPledge(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}
