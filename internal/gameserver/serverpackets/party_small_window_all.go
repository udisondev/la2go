package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePartySmallWindowAll is the server packet opcode for full party window (S2C 0x4E).
// Sent when a player joins a party to display all existing members.
//
// Java reference: PartySmallWindowAll.java (opcode 0x4E).
const OpcodePartySmallWindowAll = 0x4E

// PartySmallWindowAll represents the full party member list sent to the client.
//
// Packet structure (S2C 0x4E):
//   - opcode      byte    0x4E
//   - leaderObjID int32   party leader's objectID
//   - lootRule    int32   loot distribution rule
//   - memberCount int32   number of members (excluding the player receiving this packet)
//   - for each member:
//   - objectID  int32   member's objectID
//   - name      string  member's name (UTF-16LE null-terminated)
//   - curHP     int32   current HP
//   - maxHP     int32   maximum HP
//   - curMP     int32   current MP
//   - maxMP     int32   maximum MP
//   - level     int32   member's level
//   - classID   int32   member's class ID
//   - padding   int32   0 (reserved, L2 client expects this)
type PartySmallWindowAll struct {
	LeaderObjID int32
	LootRule    int32
	Members     []*model.Player // все члены группы кроме получателя пакета
}

// NewPartySmallWindowAll creates a PartySmallWindowAll packet.
// excludeObjID is the objectID of the player receiving this packet (excluded from list).
func NewPartySmallWindowAll(party *model.Party, excludeObjID uint32) PartySmallWindowAll {
	members := party.Members()
	filtered := make([]*model.Player, 0, len(members)-1)
	for _, m := range members {
		if m.ObjectID() != excludeObjID {
			filtered = append(filtered, m)
		}
	}

	return PartySmallWindowAll{
		LeaderObjID: int32(party.Leader().ObjectID()),
		LootRule:    party.LootRule(),
		Members:     filtered,
	}
}

// Write serializes the PartySmallWindowAll packet to bytes.
func (p *PartySmallWindowAll) Write() ([]byte, error) {
	// opcode(1) + leaderObjID(4) + lootRule(4) + memberCount(4)
	// + per member: objectID(4) + name(~32) + HP/MP/level/classID/padding(7*4=28)
	estimatedSize := 13 + len(p.Members)*96
	w := packet.NewWriter(estimatedSize)

	w.WriteByte(OpcodePartySmallWindowAll)
	w.WriteInt(p.LeaderObjID)
	w.WriteInt(p.LootRule)
	w.WriteInt(int32(len(p.Members)))

	for _, m := range p.Members {
		w.WriteInt(int32(m.ObjectID()))
		w.WriteString(m.Name())
		w.WriteInt(m.CurrentHP())
		w.WriteInt(m.MaxHP())
		w.WriteInt(m.CurrentMP())
		w.WriteInt(m.MaxMP())
		w.WriteInt(m.Level())
		w.WriteInt(m.ClassID())
		w.WriteInt(0) // padding / reserved
	}

	return w.Bytes(), nil
}
