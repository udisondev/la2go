package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeStopRotation is the opcode for StopRotation (S2C 0x63).
// Java reference: ServerPackets.FINISH_ROTATION(0x63).
const OpcodeStopRotation = 0x63

// StopRotation notifies clients that a character finished rotating.
//
// Packet structure:
//   - opcode   (byte)  — 0x63
//   - objectID (int32) — character object ID
//   - degree   (int32) — final rotation angle
//   - speed    (int32) — rotation speed (0 for default)
type StopRotation struct {
	ObjectID int32
	Degree   int32
	Speed    int32
}

// Write serializes StopRotation packet.
func (p *StopRotation) Write() ([]byte, error) {
	w := packet.NewWriter(16)
	w.WriteByte(OpcodeStopRotation)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.Degree)
	w.WriteInt(p.Speed)
	return w.Bytes(), nil
}
