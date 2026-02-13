package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeRevive is the S2C opcode 0x07.
// Sent when a character is revived/respawned.
const OpcodeRevive byte = 0x07

// Revive notifies the client that a character has been revived.
// Java reference: serverpackets/Revive.java
type Revive struct {
	ObjectID int32 // revived character's objectID
}

// Write serializes the Revive packet.
func (p *Revive) Write() ([]byte, error) {
	w := packet.NewWriter(5) // 1 + 4
	w.WriteByte(OpcodeRevive)
	w.WriteInt(p.ObjectID)
	return w.Bytes(), nil
}
