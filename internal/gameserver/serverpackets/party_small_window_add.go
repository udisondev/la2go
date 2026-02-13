package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePartySmallWindowAdd is the server packet opcode for adding a party member (S2C 0x4F).
// Sent to existing party members when a new player joins.
//
// Java reference: PartySmallWindowAdd.java (opcode 0x4F).
const OpcodePartySmallWindowAdd = 0x4F

// PartySmallWindowAdd represents adding a single member to the party window.
//
// Packet structure (S2C 0x4F):
//   - opcode      byte    0x4F
//   - leaderObjID int32   party leader's objectID
//   - lootRule    int32   loot distribution rule
//   - objectID    int32   new member's objectID
//   - name        string  new member's name (UTF-16LE null-terminated)
//   - curHP       int32   current HP
//   - maxHP       int32   maximum HP
//   - curMP       int32   current MP
//   - maxMP       int32   maximum MP
//   - level       int32   member's level
//   - classID     int32   member's class ID
//   - padding     int32   0 (reserved)
type PartySmallWindowAdd struct {
	LeaderObjID int32
	LootRule    int32
	Member      *model.Player
}

// NewPartySmallWindowAdd creates a PartySmallWindowAdd packet.
func NewPartySmallWindowAdd(party *model.Party, member *model.Player) PartySmallWindowAdd {
	return PartySmallWindowAdd{
		LeaderObjID: int32(party.Leader().ObjectID()),
		LootRule:    party.LootRule(),
		Member:      member,
	}
}

// Write serializes the PartySmallWindowAdd packet to bytes.
func (p *PartySmallWindowAdd) Write() ([]byte, error) {
	// opcode(1) + leaderObjID(4) + lootRule(4) + member data(~64)
	w := packet.NewWriter(96)

	w.WriteByte(OpcodePartySmallWindowAdd)
	w.WriteInt(p.LeaderObjID)
	w.WriteInt(p.LootRule)
	w.WriteInt(int32(p.Member.ObjectID()))
	w.WriteString(p.Member.Name())
	w.WriteInt(p.Member.CurrentHP())
	w.WriteInt(p.Member.MaxHP())
	w.WriteInt(p.Member.CurrentMP())
	w.WriteInt(p.Member.MaxMP())
	w.WriteInt(p.Member.Level())
	w.WriteInt(p.Member.ClassID())
	w.WriteInt(0) // padding / reserved

	return w.Bytes(), nil
}
