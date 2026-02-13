package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExFishingHpRegen is the sub-opcode for ExFishingHpRegen (S2C 0xFE:0x16).
const SubOpcodeExFishingHpRegen int16 = 0x16

// ExFishingHpRegen updates the client about the current combat state each tick.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x16
//   - objectID (int32) — player object ID
//   - time (int32) — remaining time (seconds)
//   - fishHP (int32) — current fish HP
//   - hpMode (byte) — 0 = stop, 1 = raising
//   - goodUse (byte) — 0 = none, 1 = success, 2 = failed
//   - anim (byte) — 0 = none, 1 = pumping, 2 = reeling
//   - penalty (int32) — penalty damage applied
//   - hpBarColor (byte) — 0 = normal, 1 = deceptive
//
// Phase 29: Fishing System.
type ExFishingHpRegen struct {
	ObjectID   int32
	Time       int32
	FishHP     int32
	HPMode     int32
	GoodUse    int32
	Anim       int32
	Penalty    int32
	HPBarColor int32
}

// Write serializes ExFishingHpRegen to binary.
func (p ExFishingHpRegen) Write() ([]byte, error) {
	// 1 + 2 + 3*4 + 3*1 + 4 + 1 = 23
	w := packet.NewWriter(23)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExFishingHpRegen)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.Time)
	w.WriteInt(p.FishHP)
	w.WriteByte(byte(p.HPMode))
	w.WriteByte(byte(p.GoodUse))
	w.WriteByte(byte(p.Anim))
	w.WriteInt(p.Penalty)
	w.WriteByte(byte(p.HPBarColor))

	return w.Bytes(), nil
}
