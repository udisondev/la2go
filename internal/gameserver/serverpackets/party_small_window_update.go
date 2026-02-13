package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePartySmallWindowUpdate is the server packet opcode for updating party member info (S2C 0x52).
// Sent to party members when a member's HP/MP/level changes.
//
// Java reference: PartySmallWindowUpdate.java (opcode 0x52).
const OpcodePartySmallWindowUpdate = 0x52

// PartySmallWindowUpdate represents an update of a party member's stats.
//
// Packet structure (S2C 0x52):
//   - opcode   byte   0x52
//   - objectID int32  member's objectID
//   - curHP    int32  current HP
//   - maxHP    int32  maximum HP
//   - curMP    int32  current MP
//   - maxMP    int32  maximum MP
//   - level    int32  member's level
//   - classID  int32  member's class ID
type PartySmallWindowUpdate struct {
	ObjectID int32
	CurHP    int32
	MaxHP    int32
	CurMP    int32
	MaxMP    int32
	Level    int32
	ClassID  int32
}

// NewPartySmallWindowUpdate creates a PartySmallWindowUpdate packet from player data.
func NewPartySmallWindowUpdate(player *model.Player) PartySmallWindowUpdate {
	return PartySmallWindowUpdate{
		ObjectID: int32(player.ObjectID()),
		CurHP:    player.CurrentHP(),
		MaxHP:    player.MaxHP(),
		CurMP:    player.CurrentMP(),
		MaxMP:    player.MaxMP(),
		Level:    player.Level(),
		ClassID:  player.ClassID(),
	}
}

// Write serializes the PartySmallWindowUpdate packet to bytes.
func (p *PartySmallWindowUpdate) Write() ([]byte, error) {
	// opcode(1) + 7 * int32(4) = 29 bytes
	w := packet.NewWriter(32)

	w.WriteByte(OpcodePartySmallWindowUpdate)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.CurHP)
	w.WriteInt(p.MaxHP)
	w.WriteInt(p.CurMP)
	w.WriteInt(p.MaxMP)
	w.WriteInt(p.Level)
	w.WriteInt(p.ClassID)

	return w.Bytes(), nil
}
