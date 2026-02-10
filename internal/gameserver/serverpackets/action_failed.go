package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeActionFailed is the server packet opcode for action failed notification.
// Generic failure response for various client actions.
//
// Phase 5.3: Basic Combat System.
const OpcodeActionFailed = 0x25

// ActionFailed represents generic action failure packet (S2C 0x25).
// No payload — just opcode.
//
// Packet structure:
//   - opcode (byte) — 0x25
//
// Phase 5.3: Basic Combat System.
// Java reference: ActionFailed.java (opcode 0x25, no payload).
type ActionFailed struct {
	// No fields — static packet
}

// NewActionFailed creates new ActionFailed packet.
// Returns value (not pointer) to avoid heap allocation.
//
// Phase 5.3: Basic Combat System.
func NewActionFailed() ActionFailed {
	return ActionFailed{}
}

// Write serializes ActionFailed packet to bytes.
//
// Packet size: 1 byte (opcode only).
//
// Phase 5.3: Basic Combat System.
func (p ActionFailed) Write() ([]byte, error) {
	w := packet.NewWriter(1)
	w.WriteByte(OpcodeActionFailed)
	return w.Bytes(), nil
}
