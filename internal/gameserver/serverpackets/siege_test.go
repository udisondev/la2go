package serverpackets

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestSiegeInfo_Write(t *testing.T) {
	t.Parallel()

	siegeDate := time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC)
	p := &SiegeInfo{
		CastleID:    1,
		CanManage:   true,
		OwnerClanID: 42,
		OwnerName:   "TestClan",
		LeaderName:  "Leader",
		AllyID:      10,
		AllyName:    "Alliance",
		SiegeDate:   siegeDate,
		TimeRegOver: true,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeInfo {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeInfo)
	}
	if len(data) < 20 {
		t.Fatalf("data too short: %d bytes", len(data))
	}

	castleID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if castleID != 1 {
		t.Errorf("castleID = %d, want 1", castleID)
	}

	canManage := int32(binary.LittleEndian.Uint32(data[5:9]))
	if canManage != 1 {
		t.Errorf("canManage = %d, want 1", canManage)
	}

	ownerClanID := int32(binary.LittleEndian.Uint32(data[9:13]))
	if ownerClanID != 42 {
		t.Errorf("ownerClanID = %d, want 42", ownerClanID)
	}
}

func TestSiegeInfo_Write_NoManage(t *testing.T) {
	t.Parallel()

	p := &SiegeInfo{
		CastleID:    2,
		CanManage:   false,
		OwnerClanID: 0,
		TimeRegOver: false,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeInfo {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeInfo)
	}

	canManage := int32(binary.LittleEndian.Uint32(data[5:9]))
	if canManage != 0 {
		t.Errorf("canManage = %d, want 0", canManage)
	}
}

func TestSiegeAttackerList_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &SiegeAttackerList{
		CastleID:  1,
		Attackers: nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeAttackerList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeAttackerList)
	}

	castleID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if castleID != 1 {
		t.Errorf("castleID = %d, want 1", castleID)
	}

	// offset: 1 (opcode) + 4 (castleID) + 4 (unk) + 4 (unk) + 4 (unk) = 17
	count := int32(binary.LittleEndian.Uint32(data[17:21]))
	if count != 0 {
		t.Errorf("attacker count = %d, want 0", count)
	}
}

func TestSiegeAttackerList_Write_WithEntries(t *testing.T) {
	t.Parallel()

	p := &SiegeAttackerList{
		CastleID: 1,
		Attackers: []SiegeAttackerEntry{
			{ClanID: 100, ClanName: "Clan1", LeaderName: "Leader1", CrestID: 5, AllyID: 10, AllyName: "Ally1", AllyCrestID: 15},
			{ClanID: 200, ClanName: "Clan2", LeaderName: "Leader2"},
		},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeAttackerList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeAttackerList)
	}

	count := int32(binary.LittleEndian.Uint32(data[17:21]))
	if count != 2 {
		t.Errorf("attacker count = %d, want 2", count)
	}
}

func TestSiegeDefenderList_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &SiegeDefenderList{
		CastleID:  3,
		Defenders: nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeDefenderList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeDefenderList)
	}

	castleID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if castleID != 3 {
		t.Errorf("castleID = %d, want 3", castleID)
	}

	count := int32(binary.LittleEndian.Uint32(data[17:21]))
	if count != 0 {
		t.Errorf("defender count = %d, want 0", count)
	}
}

func TestSiegeDefenderList_Write_WithEntries(t *testing.T) {
	t.Parallel()

	p := &SiegeDefenderList{
		CastleID: 1,
		Defenders: []SiegeDefenderEntry{
			{ClanID: 500, ClanName: "Owner", LeaderName: "Boss", CrestID: 1, Type: DefenderTypeOwner},
			{ClanID: 200, ClanName: "Def", LeaderName: "D1", Type: DefenderTypeApproved},
			{ClanID: 300, ClanName: "Pend", LeaderName: "P1", Type: DefenderTypePending},
		},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeSiegeDefenderList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSiegeDefenderList)
	}

	count := int32(binary.LittleEndian.Uint32(data[17:21]))
	if count != 3 {
		t.Errorf("defender count = %d, want 3", count)
	}
}
