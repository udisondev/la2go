package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeChangeWaitType2 is the opcode for ChangeWaitType2 (C2S 0x1D).
// Java reference: ClientPackets.CHANGE_WAIT_TYPE(0x1D).
const OpcodeChangeWaitType2 = 0x1D

// ChangeWaitType2 represents a client request to toggle sit/stand mode.
//
// Packet structure (body after opcode):
//   - typeStand (int32) — 1=stand, 0=sit
type ChangeWaitType2 struct {
	TypeStand int32
}

// ParseChangeWaitType2 parses ChangeWaitType2 packet.
func ParseChangeWaitType2(data []byte) (*ChangeWaitType2, error) {
	r := packet.NewReader(data)

	typeStand, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading typeStand: %w", err)
	}

	return &ChangeWaitType2{TypeStand: typeStand}, nil
}
