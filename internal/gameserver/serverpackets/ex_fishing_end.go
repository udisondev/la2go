package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExFishingEnd is the sub-opcode for ExFishingEnd (S2C 0xFE:0x14).
const SubOpcodeExFishingEnd int16 = 0x14

// ExFishingEnd notifies the client that fishing combat has ended.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x14
//   - objectID (int32) — player object ID
//   - win (byte) — 0 = lose, 1 = win
//
// Phase 29: Fishing System.
type ExFishingEnd struct {
	ObjectID int32
	Win      bool
}

// Write serializes ExFishingEnd to binary.
func (p ExFishingEnd) Write() ([]byte, error) {
	// 1 + 2 + 4 + 1 = 8
	w := packet.NewWriter(8)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExFishingEnd)
	w.WriteInt(p.ObjectID)

	if p.Win {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}

	return w.Bytes(), nil
}
