package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	OpcodeSystemMessage = 0x64
)

// SystemMessage parameter types
const (
	ParamTypeText     = 0
	ParamTypeNumber   = 1
	ParamTypeNpcName  = 2
	ParamTypeItemName = 3
	ParamTypeSkill    = 4
)

// SystemMessage IDs (Interlude)
const (
	SysMsgYouEarnedS1Exp         = 186  // "You have earned $s1 experience."
	SysMsgYouEarnedS1ExpAndS2SP  = 336  // "You have earned $s1 experience and $s2 SP."
	SysMsgYourLevelHasIncreased  = 339  // "Your level has increased!"
	SysMsgYouAcquiredS1SP        = 1044 // "You have acquired $s1 SP."
)

// SystemMessage represents a system message packet (S2C 0x64).
// System messages use predefined message IDs with optional parameters.
type SystemMessage struct {
	MessageID int32
	Params    []systemMessageParam
}

type systemMessageParam struct {
	Type  int32
	Value int64
}

// NewSystemMessage creates a new system message with the given ID.
func NewSystemMessage(messageID int32) *SystemMessage {
	return &SystemMessage{
		MessageID: messageID,
	}
}

// AddNumber adds a number parameter ($s1, $s2, etc.).
func (m *SystemMessage) AddNumber(value int64) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:  ParamTypeNumber,
		Value: value,
	})
	return m
}

// Write serializes the SystemMessage packet.
func (m *SystemMessage) Write() ([]byte, error) {
	// opcode(1) + messageID(4) + paramCount(4) + params(4+4 each)
	w := packet.NewWriter(9 + len(m.Params)*8)

	w.WriteByte(OpcodeSystemMessage)
	w.WriteInt(m.MessageID)
	w.WriteInt(int32(len(m.Params)))

	for _, p := range m.Params {
		w.WriteInt(p.Type)
		w.WriteInt(int32(p.Value))
	}

	return w.Bytes(), nil
}
