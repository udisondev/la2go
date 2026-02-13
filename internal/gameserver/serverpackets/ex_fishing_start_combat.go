package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExFishingStartCombat is the sub-opcode for ExFishingStartCombat (S2C 0xFE:0x15).
const SubOpcodeExFishingStartCombat int16 = 0x15

// ExFishingStartCombat notifies the client that the fishing combat phase has begun.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x15
//   - objectID (int32) — player object ID
//   - time (int32) — combat duration (seconds)
//   - hp (int32) — fish max HP
//   - mode (byte) — 0 = resting, 1 = fighting
//   - lureType (byte) — 0 = newbie, 1 = normal, 2 = night
//   - deceptiveMode (byte) — 0 = normal, 1 = deceptive
//
// Phase 29: Fishing System.
type ExFishingStartCombat struct {
	ObjectID      int32
	Time          int32
	HP            int32
	Mode          int32
	LureType      int32
	DeceptiveMode int32
}

// Write serializes ExFishingStartCombat to binary.
func (p ExFishingStartCombat) Write() ([]byte, error) {
	// 1 + 2 + 3*4 + 3 = 18
	w := packet.NewWriter(18)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExFishingStartCombat)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.Time)
	w.WriteInt(p.HP)
	w.WriteByte(byte(p.Mode))
	w.WriteByte(byte(p.LureType))
	w.WriteByte(byte(p.DeceptiveMode))

	return w.Bytes(), nil
}
