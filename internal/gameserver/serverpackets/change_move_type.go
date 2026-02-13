package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeChangeMoveType is the opcode for ChangeMoveType (S2C 0x2E).
// Java reference: ServerPackets.CHANGE_MOVE_TYPE(0x2E).
const OpcodeChangeMoveType = 0x2E

// ChangeMoveType notifies clients about a character's walk/run mode change.
//
// Packet structure:
//   - opcode    (byte)  — 0x2E
//   - objectID  (int32) — character object ID
//   - moveType  (int32) — 0=walk, 1=run
//   - zero      (int32) — reserved (always 0)
type ChangeMoveType struct {
	ObjectID int32
	MoveType int32 // 0=walk, 1=run
}

// Write serializes ChangeMoveType packet.
func (p *ChangeMoveType) Write() ([]byte, error) {
	w := packet.NewWriter(16)
	w.WriteByte(OpcodeChangeMoveType)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.MoveType)
	w.WriteInt(0) // reserved
	return w.Bytes(), nil
}
