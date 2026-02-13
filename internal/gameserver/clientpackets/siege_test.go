package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestSiegeInfo(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(8)
	w.WriteInt(3) // castleID

	pkt, err := ParseRequestSiegeInfo(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.CastleID != 3 {
		t.Errorf("CastleID = %d, want 3", pkt.CastleID)
	}
}

func TestParseRequestSiegeInfo_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestSiegeInfo(nil)
	if err == nil {
		t.Error("Parse(nil) = nil; want error")
	}
}

func TestParseRequestSiegeAttackerList(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(8)
	w.WriteInt(5) // castleID

	pkt, err := ParseRequestSiegeAttackerList(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.CastleID != 5 {
		t.Errorf("CastleID = %d, want 5", pkt.CastleID)
	}
}

func TestParseRequestSiegeAttackerList_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestSiegeAttackerList(nil)
	if err == nil {
		t.Error("Parse(nil) = nil; want error")
	}
}

func TestParseRequestSiegeDefenderList(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(8)
	w.WriteInt(7) // castleID

	pkt, err := ParseRequestSiegeDefenderList(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.CastleID != 7 {
		t.Errorf("CastleID = %d, want 7", pkt.CastleID)
	}
}

func TestParseRequestSiegeDefenderList_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestSiegeDefenderList(nil)
	if err == nil {
		t.Error("Parse(nil) = nil; want error")
	}
}

func TestParseRequestJoinSiege_Attacker(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(16)
	w.WriteInt(1) // castleID
	w.WriteInt(1) // isAttacker
	w.WriteInt(1) // isJoining

	pkt, err := ParseRequestJoinSiege(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.CastleID != 1 {
		t.Errorf("CastleID = %d, want 1", pkt.CastleID)
	}
	if !pkt.IsAttacker {
		t.Error("IsAttacker = false, want true")
	}
	if !pkt.IsJoining {
		t.Error("IsJoining = false, want true")
	}
}

func TestParseRequestJoinSiege_Defender(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(16)
	w.WriteInt(2) // castleID
	w.WriteInt(0) // isAttacker = false (defender)
	w.WriteInt(1) // isJoining

	pkt, err := ParseRequestJoinSiege(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.IsAttacker {
		t.Error("IsAttacker = true, want false")
	}
}

func TestParseRequestJoinSiege_Leave(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(16)
	w.WriteInt(1) // castleID
	w.WriteInt(1) // isAttacker
	w.WriteInt(0) // isJoining = false (leave)

	pkt, err := ParseRequestJoinSiege(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.IsJoining {
		t.Error("IsJoining = true, want false")
	}
}

func TestParseRequestJoinSiege_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestJoinSiege(nil)
	if err == nil {
		t.Error("Parse(nil) = nil; want error")
	}
}

func TestParseRequestConfirmSiegeWaitingList_Approve(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(16)
	w.WriteInt(1)  // isApproval = true
	w.WriteInt(3)  // castleID
	w.WriteInt(42) // clanID

	pkt, err := ParseRequestConfirmSiegeWaitingList(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !pkt.IsApproval {
		t.Error("IsApproval = false, want true")
	}
	if pkt.CastleID != 3 {
		t.Errorf("CastleID = %d, want 3", pkt.CastleID)
	}
	if pkt.ClanID != 42 {
		t.Errorf("ClanID = %d, want 42", pkt.ClanID)
	}
}

func TestParseRequestConfirmSiegeWaitingList_Reject(t *testing.T) {
	t.Parallel()
	w := packet.NewWriter(16)
	w.WriteInt(0)  // isApproval = false
	w.WriteInt(1)  // castleID
	w.WriteInt(99) // clanID

	pkt, err := ParseRequestConfirmSiegeWaitingList(w.Bytes())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkt.IsApproval {
		t.Error("IsApproval = true, want false")
	}
}

func TestParseRequestConfirmSiegeWaitingList_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestConfirmSiegeWaitingList(nil)
	if err == nil {
		t.Error("Parse(nil) = nil; want error")
	}
}
