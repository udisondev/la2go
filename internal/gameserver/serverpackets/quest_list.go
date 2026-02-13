package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeQuestList is the opcode for QuestList packet (S2C 0x80).
	OpcodeQuestList = 0x80
)

// QuestEntry represents a single quest in the QuestList packet.
type QuestEntry struct {
	QuestID int32 // Quest identifier
	State   int32 // 0=created, 1=started, 2=completed
}

// QuestList packet (S2C 0x80) sends list of active quests.
// Sent after UserInfo during spawn.
//
// Packet structure:
//   - opcode (byte) — 0x80
//   - count (short) — number of quests
//   - for each quest:
//   - questID (int32)
//   - state (int32)
//
// Phase 16: Quest System Framework.
type QuestList struct {
	Quests []QuestEntry
}

// NewQuestList creates empty QuestList packet.
func NewQuestList() QuestList {
	return QuestList{}
}

// NewQuestListWithEntries creates QuestList with quest entries.
func NewQuestListWithEntries(quests []QuestEntry) QuestList {
	return QuestList{Quests: quests}
}

// Write serializes QuestList packet to binary format.
func (p *QuestList) Write() ([]byte, error) {
	// 1 opcode + 2 count + quests * (4+4)
	w := packet.NewWriter(3 + len(p.Quests)*8)

	w.WriteByte(OpcodeQuestList)
	w.WriteShort(int16(len(p.Quests)))

	for _, q := range p.Quests {
		w.WriteInt(q.QuestID)
		w.WriteInt(q.State)
	}

	return w.Bytes(), nil
}
