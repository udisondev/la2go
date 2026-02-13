package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeChangeMoveType2 is the opcode for ChangeMoveType2 (C2S 0x1C).
// Java reference: ClientPackets.CHANGE_MOVE_TYPE(0x1C).
const OpcodeChangeMoveType2 = 0x1C

// ChangeMoveType2 represents a client request to toggle walk/run mode.
//
// Packet structure (body after opcode):
//   - typeRun (int32) — 1=run, 0=walk
type ChangeMoveType2 struct {
	TypeRun int32
}

// ParseChangeMoveType2 parses ChangeMoveType2 packet.
func ParseChangeMoveType2(data []byte) (*ChangeMoveType2, error) {
	r := packet.NewReader(data)

	typeRun, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading typeRun: %w", err)
	}

	return &ChangeMoveType2{TypeRun: typeRun}, nil
}
