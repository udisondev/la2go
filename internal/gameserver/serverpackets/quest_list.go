package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeQuestList is the opcode for QuestList packet (S2C 0x80)
	OpcodeQuestList = 0x80
)

// QuestList packet (S2C 0x80) sends list of active quests.
// Sent after UserInfo during spawn.
type QuestList struct {
	// Quests []Quest // TODO Phase 5.5: implement quest system
}

// NewQuestList creates empty QuestList packet.
// TODO Phase 5.5: Load quests from database.
func NewQuestList() *QuestList {
	return &QuestList{}
}

// Write serializes QuestList packet to binary format.
func (p *QuestList) Write() ([]byte, error) {
	// Empty quest list for now
	w := packet.NewWriter(16)

	w.WriteByte(OpcodeQuestList)
	w.WriteShort(0) // Quest count = 0

	return w.Bytes(), nil
}
