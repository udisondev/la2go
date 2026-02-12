package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

const (
	// OpcodeLeaveWorld is the opcode for LeaveWorld packet (S2C 0x7E)
	OpcodeLeaveWorld = 0x7E
)

// LeaveWorld notifies client that server is closing the connection.
// Used by Logout, RequestRestart, and Disconnection handlers.
//
// Packet structure:
//   - opcode: byte (0x7E)
//   - No payload
//
// Reference: L2J_Mobius LeaveWorld.java
type LeaveWorld struct{}

// NewLeaveWorld creates a LeaveWorld packet.
func NewLeaveWorld() LeaveWorld {
	return LeaveWorld{}
}

// Write serializes the LeaveWorld packet to bytes.
func (p *LeaveWorld) Write() ([]byte, error) {
	w := packet.NewWriter(1) // Only opcode

	if err := w.WriteByte(OpcodeLeaveWorld); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
