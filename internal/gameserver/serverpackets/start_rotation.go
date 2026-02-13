package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeStartRotation is the opcode for StartRotation (S2C 0x62).
// Java reference: ServerPackets.BEGIN_ROTATION(0x62).
const OpcodeStartRotation = 0x62

// StartRotation notifies clients that a character started rotating.
//
// Packet structure:
//   - opcode   (byte)  — 0x62
//   - objectID (int32) — character object ID
//   - degree   (int32) — rotation angle
//   - side     (int32) — rotation direction (1=left, -1=right)
//   - speed    (int32) — rotation speed (0 for default)
type StartRotation struct {
	ObjectID int32
	Degree   int32
	Side     int32
	Speed    int32
}

// Write serializes StartRotation packet.
func (p *StartRotation) Write() ([]byte, error) {
	w := packet.NewWriter(20)
	w.WriteByte(OpcodeStartRotation)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.Degree)
	w.WriteInt(p.Side)
	w.WriteInt(p.Speed)
	return w.Bytes(), nil
}
