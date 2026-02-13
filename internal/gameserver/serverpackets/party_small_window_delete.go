package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePartySmallWindowDelete is the server packet opcode for removing a party member (S2C 0x51).
// Sent to remaining party members when someone leaves or is kicked.
//
// Java reference: PartySmallWindowDelete.java (opcode 0x51).
const OpcodePartySmallWindowDelete = 0x51

// PartySmallWindowDelete represents removing a single member from the party window.
//
// Packet structure (S2C 0x51):
//   - opcode   byte    0x51
//   - objectID int32   removed member's objectID
//   - name     string  removed member's name (UTF-16LE null-terminated)
type PartySmallWindowDelete struct {
	ObjectID int32
	Name     string
}

// NewPartySmallWindowDelete creates a PartySmallWindowDelete packet.
func NewPartySmallWindowDelete(objectID uint32, name string) PartySmallWindowDelete {
	return PartySmallWindowDelete{
		ObjectID: int32(objectID),
		Name:     name,
	}
}

// Write serializes the PartySmallWindowDelete packet to bytes.
func (p *PartySmallWindowDelete) Write() ([]byte, error) {
	// opcode(1) + objectID(4) + name(~32)
	w := packet.NewWriter(48)

	w.WriteByte(OpcodePartySmallWindowDelete)
	w.WriteInt(p.ObjectID)
	w.WriteString(p.Name)

	return w.Bytes(), nil
}
