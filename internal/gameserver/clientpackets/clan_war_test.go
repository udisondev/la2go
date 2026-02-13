package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestStartPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("EnemyClan")

	pkt, err := ParseRequestStartPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestStartPledgeWar: %v", err)
	}
	if pkt.ClanName != "EnemyClan" {
		t.Errorf("ClanName = %q, want %q", pkt.ClanName, "EnemyClan")
	}
}

func TestParseRequestStartPledgeWar_ShortData(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestStartPledgeWar(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestParseRequestReplyStartPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("Requester")
	w.WriteInt(1) // accept

	pkt, err := ParseRequestReplyStartPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestReplyStartPledgeWar: %v", err)
	}
	if pkt.Name != "Requester" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Requester")
	}
	if pkt.Answer != 1 {
		t.Errorf("Answer = %d, want 1", pkt.Answer)
	}
}

func TestParseRequestReplyStartPledgeWar_Deny(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("Requester")
	w.WriteInt(0) // deny

	pkt, err := ParseRequestReplyStartPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestReplyStartPledgeWar: %v", err)
	}
	if pkt.Answer != 0 {
		t.Errorf("Answer = %d, want 0", pkt.Answer)
	}
}

func TestParseRequestStopPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("TargetClan")

	pkt, err := ParseRequestStopPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestStopPledgeWar: %v", err)
	}
	if pkt.ClanName != "TargetClan" {
		t.Errorf("ClanName = %q, want %q", pkt.ClanName, "TargetClan")
	}
}

func TestParseRequestReplyStopPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("Requester")
	w.WriteInt(1) // accept

	pkt, err := ParseRequestReplyStopPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestReplyStopPledgeWar: %v", err)
	}
	if pkt.Name != "Requester" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Requester")
	}
	if pkt.Answer != 1 {
		t.Errorf("Answer = %d, want 1", pkt.Answer)
	}
}

func TestParseRequestSurrenderPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("StrongerClan")

	pkt, err := ParseRequestSurrenderPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestSurrenderPledgeWar: %v", err)
	}
	if pkt.ClanName != "StrongerClan" {
		t.Errorf("ClanName = %q, want %q", pkt.ClanName, "StrongerClan")
	}
}

func TestParseRequestReplySurrenderPledgeWar(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(64)
	w.WriteString("Surrenderer")
	w.WriteInt(1) // accept surrender

	pkt, err := ParseRequestReplySurrenderPledgeWar(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestReplySurrenderPledgeWar: %v", err)
	}
	if pkt.Name != "Surrenderer" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Surrenderer")
	}
	if pkt.Answer != 1 {
		t.Errorf("Answer = %d, want 1", pkt.Answer)
	}
}

func TestParseRequestReplySurrenderPledgeWar_ShortData(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestReplySurrenderPledgeWar(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}
