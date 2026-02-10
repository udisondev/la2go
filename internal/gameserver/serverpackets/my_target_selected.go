package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMyTargetSelected is the server packet opcode for MyTargetSelected.
// Server sends this when player selects a target (shows HP bar).
const OpcodeMyTargetSelected = 0xA6

// MyTargetSelected represents the server's target selection confirmation packet.
// Sent to client to display target's HP bar and highlight the selected object.
//
// Packet structure:
//   - ObjectID (int32): Target object ID
//   - Color (int32): Target color/highlight (0=normal, RGB for special)
//
// Reference: MyTargetSelected.java (L2J Mobius)
type MyTargetSelected struct {
	ObjectID int32 // Target object ID
	Color    int32 // Target highlight color (0=default)
}

// NewMyTargetSelected creates a MyTargetSelected packet for a target.
// Color defaults to 0 (normal highlight).
func NewMyTargetSelected(objectID uint32) *MyTargetSelected {
	return &MyTargetSelected{
		ObjectID: int32(objectID),
		Color:    0, // Default color (normal target)
	}
}

// NewMyTargetSelectedWithColor creates a MyTargetSelected packet with custom color.
// Used for special targets (flagged players, quest NPCs, etc).
func NewMyTargetSelectedWithColor(objectID uint32, color int32) *MyTargetSelected {
	return &MyTargetSelected{
		ObjectID: int32(objectID),
		Color:    color,
	}
}

// Write serializes the MyTargetSelected packet to bytes.
//
// Packet format:
//   - opcode (byte): 0xA6
//   - objectID (int32): target object ID
//   - color (int32): highlight color
//
// Returns the serialized packet bytes and any error.
func (p *MyTargetSelected) Write() ([]byte, error) {
	w := packet.NewWriter(16) // opcode(1) + objectID(4) + color(4) = 9 bytes + padding

	w.WriteByte(OpcodeMyTargetSelected)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.Color)

	return w.Bytes(), nil
}
