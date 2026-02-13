package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeDie is the S2C opcode 0x06.
// Sent when a character dies — shows death dialog with respawn options.
const OpcodeDie byte = 0x06

// Die notifies the client that a character has died.
// Java reference: serverpackets/Die.java
type Die struct {
	ObjectID    int32 // dying character's objectID
	CanTeleport bool  // whether player can teleport (revive available)
	ToHideaway  bool  // clan hall respawn available
	ToCastle    bool  // castle respawn available
	ToSiegeHQ   bool  // siege HQ respawn available
	Sweepable   bool  // blue glow (spoil active)
	FixedRes    bool  // fixed resurrection (scroll of resurrection)
}

// Write serializes the Die packet.
func (p *Die) Write() ([]byte, error) {
	w := packet.NewWriter(25) // 1 + 6*4
	w.WriteByte(OpcodeDie)
	w.WriteInt(p.ObjectID)
	w.WriteInt(boolToInt32(p.CanTeleport))
	w.WriteInt(boolToInt32(p.ToHideaway))
	w.WriteInt(boolToInt32(p.ToCastle))
	w.WriteInt(boolToInt32(p.ToSiegeHQ))
	w.WriteInt(boolToInt32(p.Sweepable))
	w.WriteInt(boolToInt32(p.FixedRes))
	return w.Bytes(), nil
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
