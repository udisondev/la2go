package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeValidateLocation is the server packet opcode for ValidateLocation.
// Server sends this to correct client position when desync is detected.
const OpcodeValidateLocation = 0x61

// ValidateLocation represents the server's position correction packet.
// Sent when the server detects desynchronization between client and server positions.
//
// Packet structure:
//   - ObjectID (int32): Player's objectID (NOT characterID)
//   - X (int32): Server's authoritative X coordinate
//   - Y (int32): Server's authoritative Y coordinate
//   - Z (int32): Server's authoritative Z coordinate
//   - Heading (int32): Server's authoritative heading (uint16 stored as int32)
//
// Reference: ValidateLocation.java (L2J Mobius)
type ValidateLocation struct {
	ObjectID int32 // Player objectID (from Phase 4.15)
	X        int32 // Server's authoritative X
	Y        int32 // Server's authoritative Y
	Z        int32 // Server's authoritative Z
	Heading  int32 // Server's authoritative heading (uint16 → int32)
}

// NewValidateLocation creates a ValidateLocation packet from a Player.
// Uses the player's current server-side position as the authoritative position.
func NewValidateLocation(player *model.Player) *ValidateLocation {
	loc := player.Location()
	return &ValidateLocation{
		ObjectID: int32(player.ObjectID()),
		X:        loc.X,
		Y:        loc.Y,
		Z:        loc.Z,
		Heading:  int32(loc.Heading), // uint16 → int32
	}
}

// Write serializes the ValidateLocation packet to bytes.
//
// Packet format:
//   - opcode (byte): 0x61
//   - objectID (int32): player objectID
//   - x (int32): X coordinate
//   - y (int32): Y coordinate
//   - z (int32): Z coordinate
//   - heading (int32): heading value
//
// Returns the serialized packet bytes and any error.
func (p *ValidateLocation) Write() ([]byte, error) {
	w := packet.NewWriter(24) // opcode(1) + 5×int32(20) = 21 bytes + padding

	w.WriteByte(OpcodeValidateLocation)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)
	w.WriteInt(p.Heading)

	return w.Bytes(), nil
}
