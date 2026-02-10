package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAutoAttackStop is the server packet opcode for auto-attack stop notification.
// Sent when player exits combat stance (15 seconds after last attack).
//
// Phase 5.3: Basic Combat System.
const OpcodeAutoAttackStop = 0x2C

// AutoAttackStop represents auto-attack stop packet (S2C 0x2C).
// Notifies client to stop auto-attack animation.
//
// Packet structure:
//   - opcode (byte) — 0x2C
//   - objectID (int32) — player objectID who stopped attacking
//
// Phase 5.3: Basic Combat System.
// Java reference: AutoAttackStop.java (opcode 0x2C).
type AutoAttackStop struct {
	ObjectID uint32 // Player objectID
}

// NewAutoAttackStop creates new AutoAttackStop packet.
// Used by AttackStanceManager when combat expires (15 seconds timeout).
//
// Returns value (not pointer) to avoid heap allocation.
//
// Phase 5.3: Basic Combat System.
func NewAutoAttackStop(objectID uint32) AutoAttackStop {
	return AutoAttackStop{ObjectID: objectID}
}

// Write serializes AutoAttackStop packet to bytes.
//
// Packet size: 5 bytes (1 + 4).
//
// Phase 5.3: Basic Combat System.
func (p *AutoAttackStop) Write() ([]byte, error) {
	w := packet.NewWriter(5) // 1 byte opcode + 4 bytes int32

	w.WriteByte(OpcodeAutoAttackStop)
	w.WriteInt(int32(p.ObjectID))

	return w.Bytes(), nil
}
