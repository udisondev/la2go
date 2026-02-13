package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExFishingStart is the sub-opcode for ExFishingStart (S2C 0xFE:0x13).
const SubOpcodeExFishingStart int16 = 0x13

// ExFishingStart notifies the client that a fishing cast has started.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x13
//   - objectID (int32) — player object ID
//   - fishType (int32) — fish group/type
//   - x, y, z (int32) — lure landing coordinates
//   - nightLure (byte) — 0/1
//   - reserved (byte) — 0
//
// Phase 29: Fishing System.
type ExFishingStart struct {
	ObjectID  int32
	FishType  int32
	X, Y, Z   int32
	NightLure bool
}

// Write serializes ExFishingStart to binary.
func (p ExFishingStart) Write() ([]byte, error) {
	// 1 + 2 + 5*4 + 2 = 25
	w := packet.NewWriter(25)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExFishingStart)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.FishType)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)

	if p.NightLure {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}
	w.WriteByte(0) // reserved

	return w.Bytes(), nil
}
