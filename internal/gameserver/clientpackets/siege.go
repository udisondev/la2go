package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Siege-related client packet opcodes.
const (
	OpcodeRequestSiegeInfo              = 0x47 // Request siege info for a castle
	OpcodeRequestSiegeAttackerList      = 0xA2 // Request list of siege attackers
	OpcodeRequestSiegeDefenderList      = 0xA3 // Request list of siege defenders
	OpcodeRequestJoinSiege              = 0xA4 // Register/unregister for siege
	OpcodeRequestConfirmSiegeWaitingList = 0xA5 // Approve/reject pending defenders
)

// RequestSiegeInfo requests siege information for a castle.
//
// Packet structure: castleID (int32).
type RequestSiegeInfo struct {
	CastleID int32
}

// ParseRequestSiegeInfo parses the RequestSiegeInfo packet.
func ParseRequestSiegeInfo(data []byte) (*RequestSiegeInfo, error) {
	r := packet.NewReader(data)
	castleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading castleID: %w", err)
	}
	return &RequestSiegeInfo{CastleID: castleID}, nil
}

// RequestSiegeAttackerList requests the attacker list for a castle.
//
// Packet structure: castleID (int32).
type RequestSiegeAttackerList struct {
	CastleID int32
}

// ParseRequestSiegeAttackerList parses the packet.
func ParseRequestSiegeAttackerList(data []byte) (*RequestSiegeAttackerList, error) {
	r := packet.NewReader(data)
	castleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading castleID: %w", err)
	}
	return &RequestSiegeAttackerList{CastleID: castleID}, nil
}

// RequestSiegeDefenderList requests the defender list for a castle.
//
// Packet structure: castleID (int32).
type RequestSiegeDefenderList struct {
	CastleID int32
}

// ParseRequestSiegeDefenderList parses the packet.
func ParseRequestSiegeDefenderList(data []byte) (*RequestSiegeDefenderList, error) {
	r := packet.NewReader(data)
	castleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading castleID: %w", err)
	}
	return &RequestSiegeDefenderList{CastleID: castleID}, nil
}

// RequestJoinSiege registers or unregisters a clan from a siege.
//
// Packet structure:
//   - castleID    int32
//   - isAttacker  int32  (1=attacker, 0=defender)
//   - isJoining   int32  (1=join, 0=leave)
type RequestJoinSiege struct {
	CastleID   int32
	IsAttacker bool
	IsJoining  bool
}

// ParseRequestJoinSiege parses the packet.
func ParseRequestJoinSiege(data []byte) (*RequestJoinSiege, error) {
	r := packet.NewReader(data)
	castleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading castleID: %w", err)
	}
	attacker, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading isAttacker: %w", err)
	}
	joining, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading isJoining: %w", err)
	}
	return &RequestJoinSiege{
		CastleID:   castleID,
		IsAttacker: attacker == 1,
		IsJoining:  joining == 1,
	}, nil
}

// RequestConfirmSiegeWaitingList approves/rejects pending defenders.
//
// Packet structure:
//   - isApproval int32  (1=approve, 0=reject)
//   - castleID   int32
//   - clanID     int32
type RequestConfirmSiegeWaitingList struct {
	IsApproval bool
	CastleID   int32
	ClanID     int32
}

// ParseRequestConfirmSiegeWaitingList parses the packet.
func ParseRequestConfirmSiegeWaitingList(data []byte) (*RequestConfirmSiegeWaitingList, error) {
	r := packet.NewReader(data)
	approval, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading isApproval: %w", err)
	}
	castleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading castleID: %w", err)
	}
	clanID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading clanID: %w", err)
	}
	return &RequestConfirmSiegeWaitingList{
		IsApproval: approval == 1,
		CastleID:   castleID,
		ClanID:     clanID,
	}, nil
}
