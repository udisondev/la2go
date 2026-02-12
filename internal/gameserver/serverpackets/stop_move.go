package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeStopMove is the server packet opcode for StopMove.
// Server broadcasts this when a player stops moving.
const OpcodeStopMove = 0x47

// StopMove represents the server's movement stop broadcast packet.
// Sent to visible players when a player stops moving (including forced stops
// due to validation failures).
//
// Packet structure:
//   - ObjectID (int32): Player's objectID
//   - X (int32): Final X coordinate
//   - Y (int32): Final Y coordinate
//   - Z (int32): Final Z coordinate
//   - Heading (int32): Final heading (uint16 stored as int32)
//
// Reference: StopMove.java (L2J Mobius)
type StopMove struct {
	ObjectID int32 // Player objectID
	X        int32 // Final X coordinate
	Y        int32 // Final Y coordinate
	Z        int32 // Final Z coordinate
	Heading  int32 // Final heading (uint16 → int32)
}

// NewStopMove creates a StopMove packet from a Player.
// Uses the player's current position as the final stopped position.
func NewStopMove(player *model.Player) StopMove {
	loc := player.Location()
	return StopMove{
		ObjectID: int32(player.ObjectID()),
		X:        loc.X,
		Y:        loc.Y,
		Z:        loc.Z,
		Heading:  int32(loc.Heading), // uint16 → int32
	}
}

// Write serializes the StopMove packet to bytes.
//
// Packet format:
//   - opcode (byte): 0x47
//   - objectID (int32): player objectID
//   - x (int32): X coordinate
//   - y (int32): Y coordinate
//   - z (int32): Z coordinate
//   - heading (int32): heading value
//
// Returns the serialized packet bytes and any error.
func (p *StopMove) Write() ([]byte, error) {
	w := packet.NewWriter(24) // opcode(1) + 5×int32(20) = 21 bytes + padding

	w.WriteByte(OpcodeStopMove)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)
	w.WriteInt(p.Heading)

	return w.Bytes(), nil
}
