package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeTeleportToLocation is the opcode for TeleportToLocation packet (S2C 0x28).
// Tells the client to teleport a character to new coordinates.
//
// Phase 12: Teleporter System.
// Java reference: TeleportToLocation.java
const OpcodeTeleportToLocation = 0x28

// TeleportToLocation sends teleportation coordinates to the client.
//
// Packet structure:
//   - opcode (byte) — 0x28
//   - objectID (int32) — target character object ID
//   - x (int32) — destination X coordinate
//   - y (int32) — destination Y coordinate
//   - z (int32) — destination Z coordinate
//   - fade (int32) — 0 = fade effect, 1 = instant
//   - heading (int32) — facing direction (0 = keep current)
//
// Phase 12: Teleporter System.
type TeleportToLocation struct {
	ObjectID int32
	X, Y, Z  int32
	Fade     int32 // 0 = fade out/in, 1 = instant
	Heading  int32
}

// NewTeleportToLocation creates TeleportToLocation with fade effect (default).
func NewTeleportToLocation(objectID, x, y, z int32) TeleportToLocation {
	return TeleportToLocation{
		ObjectID: objectID,
		X:        x,
		Y:        y,
		Z:        z,
		Fade:     0,
		Heading:  0,
	}
}

// Write serializes TeleportToLocation packet to bytes.
//
// Phase 12: Teleporter System.
func (p TeleportToLocation) Write() ([]byte, error) {
	// 1 opcode + 4*6 fields = 25 bytes
	w := packet.NewWriter(25)

	w.WriteByte(OpcodeTeleportToLocation)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)
	w.WriteInt(p.Fade)
	w.WriteInt(p.Heading)

	return w.Bytes(), nil
}
