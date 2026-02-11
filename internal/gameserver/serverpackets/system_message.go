package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	OpcodeSystemMessage = 0x64
)

// SystemMessage parameter types
// Java reference: SystemMessage.java
const (
	ParamTypeText       = 0
	ParamTypeNumber     = 1
	ParamTypeNpcName    = 2
	ParamTypeItemName   = 3
	ParamTypeSkill      = 4
	ParamTypeLong       = 6
	ParamTypePlayerName = 12
)

// SystemMessage IDs (Interlude)
// Java reference: SystemMessageId.java
const (
	SysMsgTargetIsNotFound               = 3    // "$s1 does not exist."
	SysMsgChattingProhibited             = 27   // "Chatting is currently prohibited."
	SysMsgYouHaveObtainedS1Adena         = 28   // "You have obtained $s1 adena."
	SysMsgYouHaveObtainedS2S1            = 29   // "You have obtained $s2 $s1."
	SysMsgYouHitForS1Damage              = 35   // "You hit for $s1 damage."
	SysMsgNotEnoughMP                    = 24   // "Not enough MP."
	SysMsgNotEnoughHP                    = 23   // "Not enough HP."
	SysMsgYouEarnedS1Exp                 = 186  // "You have earned $s1 experience."
	SysMsgYouEarnedS1ExpAndS2SP          = 336  // "You have earned $s1 experience and $s2 SP."
	SysMsgYourLevelHasIncreased          = 339  // "Your level has increased!"
	SysMsgYouAcquiredS1SP                = 1044 // "You have acquired $s1 SP."
	SysMsgYouHaveExceededTheChatTextLimit = 78   // "You have exceeded the length limit for chat messages."
)

// SystemMessage represents a system message packet (S2C 0x64).
// System messages use predefined message IDs with optional parameters.
type SystemMessage struct {
	MessageID int32
	Params    []systemMessageParam
}

type systemMessageParam struct {
	Type      int32
	IntValue  int32
	LongValue int64
	StrValue  string
	// SkillLevel is used for ParamTypeSkill (second int)
	SkillLevel int32
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
		Type:     ParamTypeNumber,
		IntValue: int32(value),
	})
	return m
}

// AddString adds a string parameter ($s1, etc.).
func (m *SystemMessage) AddString(text string) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:     ParamTypeText,
		StrValue: text,
	})
	return m
}

// AddPlayerName adds a player name parameter.
func (m *SystemMessage) AddPlayerName(name string) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:     ParamTypePlayerName,
		StrValue: name,
	})
	return m
}

// AddItemName adds an item name parameter by item ID.
// Client resolves the name from its local data.
func (m *SystemMessage) AddItemName(itemID int32) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:     ParamTypeItemName,
		IntValue: itemID,
	})
	return m
}

// AddNpcName adds an NPC name parameter by NPC ID.
// Client resolves the name from its local data.
func (m *SystemMessage) AddNpcName(npcID int32) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:     ParamTypeNpcName,
		IntValue: npcID,
	})
	return m
}

// AddSkillName adds a skill name parameter by skill ID and level.
// Client resolves the name from its local data.
func (m *SystemMessage) AddSkillName(skillID, level int32) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:       ParamTypeSkill,
		IntValue:   skillID,
		SkillLevel: level,
	})
	return m
}

// AddLong adds a long (int64) parameter.
func (m *SystemMessage) AddLong(value int64) *SystemMessage {
	m.Params = append(m.Params, systemMessageParam{
		Type:      ParamTypeLong,
		LongValue: value,
	})
	return m
}

// Write serializes the SystemMessage packet.
// Handles all parameter types: text, number, npc, item, skill, long, playerName.
func (m *SystemMessage) Write() ([]byte, error) {
	w := packet.NewWriter(64 + len(m.Params)*16)

	w.WriteByte(OpcodeSystemMessage)
	w.WriteInt(m.MessageID)
	w.WriteInt(int32(len(m.Params)))

	for _, p := range m.Params {
		w.WriteInt(p.Type)
		switch p.Type {
		case ParamTypeText, ParamTypePlayerName:
			w.WriteString(p.StrValue)
		case ParamTypeNumber, ParamTypeNpcName, ParamTypeItemName:
			w.WriteInt(p.IntValue)
		case ParamTypeSkill:
			w.WriteInt(p.IntValue)
			w.WriteInt(p.SkillLevel)
		case ParamTypeLong:
			w.WriteLong(p.LongValue)
		default:
			w.WriteInt(p.IntValue)
		}
	}

	return w.Bytes(), nil
}
