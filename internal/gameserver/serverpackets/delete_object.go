package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeDeleteObject is the opcode for DeleteObject packet (S2C 0x12)
	OpcodeDeleteObject = 0x12
)

// DeleteObject packet (S2C 0x12) removes an object from client's view.
// Sent when object leaves visibility range or despawns.
type DeleteObject struct {
	ObjectID int32 // Object ID to delete
}

// NewDeleteObject creates DeleteObject packet for given object ID.
func NewDeleteObject(objectID int32) DeleteObject {
	return DeleteObject{
		ObjectID: objectID,
	}
}

// Write serializes DeleteObject packet to binary format.
func (p *DeleteObject) Write() ([]byte, error) {
	w := packet.NewWriter(8)

	w.WriteByte(OpcodeDeleteObject)
	w.WriteInt(p.ObjectID)

	return w.Bytes(), nil
}
