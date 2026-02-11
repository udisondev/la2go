package gameserver

// ChatType represents a chat channel type in Lineage 2.
// Java reference: ChatType.java
type ChatType int32

const (
	ChatGeneral   ChatType = 0
	ChatShout     ChatType = 1
	ChatWhisper   ChatType = 2
	ChatParty     ChatType = 3
	ChatClan      ChatType = 4
	ChatGM        ChatType = 5
	ChatPetition  ChatType = 6
	ChatSystem    ChatType = 7
	ChatTrade     ChatType = 8
	ChatAlliance  ChatType = 9
	ChatAnnounce  ChatType = 10
	ChatBoat      ChatType = 11
	ChatMPCC      ChatType = 12
	ChatHeroVoice ChatType = 17
)

// IsValid returns true if the ChatType is a known channel ID.
func (ct ChatType) IsValid() bool {
	switch ct {
	case ChatGeneral, ChatShout, ChatWhisper, ChatParty, ChatClan,
		ChatGM, ChatPetition, ChatSystem, ChatTrade, ChatAlliance,
		ChatAnnounce, ChatBoat, ChatMPCC, ChatHeroVoice:
		return true
	}
	return false
}

// IsMVPChannel returns true if this chat type is supported in MVP (Phase 5.11).
func (ct ChatType) IsMVPChannel() bool {
	switch ct {
	case ChatGeneral, ChatShout, ChatWhisper, ChatTrade:
		return true
	}
	return false
}

// MaxMessageLength is the max text length for non-GM players.
// Java reference: Say2.java line 87.
const MaxMessageLength = 105
