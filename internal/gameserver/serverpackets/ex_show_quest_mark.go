package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeExShowQuestMark is the extended opcode for ExShowQuestMark (S2C 0xFE subop 0x1A).
	OpcodeExShowQuestMark byte = 0xFE
	// SubOpcodeExShowQuestMark is the sub-opcode (0x1A).
	SubOpcodeExShowQuestMark int16 = 0x1A
)

// ExShowQuestMark packet (S2C 0xFE:0x1A) shows a quest mark (!) over the NPC.
// Used to indicate quest availability or progress.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x1A
//   - questID (int32)
//   - questState (int32) — 0=available, 1=in progress, 2=completed
//
// Phase 16: Quest System Framework.
type ExShowQuestMark struct {
	QuestID    int32
	QuestState int32
}

// NewExShowQuestMark creates an ExShowQuestMark packet.
func NewExShowQuestMark(questID, questState int32) ExShowQuestMark {
	return ExShowQuestMark{
		QuestID:    questID,
		QuestState: questState,
	}
}

// Write serializes ExShowQuestMark packet to binary format.
func (p ExShowQuestMark) Write() ([]byte, error) {
	// 1 opcode + 2 subop + 4 questID + 4 state = 11
	w := packet.NewWriter(11)

	w.WriteByte(OpcodeExShowQuestMark)
	w.WriteShort(SubOpcodeExShowQuestMark)
	w.WriteInt(p.QuestID)
	w.WriteInt(p.QuestState)

	return w.Bytes(), nil
}
