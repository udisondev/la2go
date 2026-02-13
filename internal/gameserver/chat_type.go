package gameserver

// ChatType represents a chat channel type in Lineage 2.
// Java reference: ChatType.java
type ChatType int32

// Chat type IDs matching Java ChatType enum clientId values.
// Java reference: ChatType.java
const (
	ChatGeneral          ChatType = 0
	ChatShout            ChatType = 1
	ChatWhisper          ChatType = 2
	ChatParty            ChatType = 3
	ChatClan             ChatType = 4
	ChatGM               ChatType = 5
	ChatPetitionPlayer   ChatType = 6
	ChatPetitionGM       ChatType = 7
	ChatTrade            ChatType = 8
	ChatAlliance         ChatType = 9
	ChatAnnounce         ChatType = 10
	ChatBoat             ChatType = 11
	ChatFriend           ChatType = 12 // Java: FRIEND=12 (NOT MPCC!)
	ChatMSNChat          ChatType = 13
	ChatPartyMatchRoom   ChatType = 14
	ChatPartyRoomCmd     ChatType = 15 // PARTYROOM_COMMANDER
	ChatPartyRoomAll     ChatType = 16 // PARTYROOM_ALL
	ChatHeroVoice        ChatType = 17
	ChatCriticalAnnounce ChatType = 18
	ChatScreenAnnounce   ChatType = 19
	ChatBattlefield      ChatType = 20
	ChatMPCC             ChatType = 21 // Java: MPCC_ROOM=21 (was 12 in Go!)
)

// IsValid returns true if the ChatType is a known channel ID.
func (ct ChatType) IsValid() bool {
	switch ct {
	case ChatGeneral, ChatShout, ChatWhisper, ChatParty, ChatClan,
		ChatGM, ChatPetitionPlayer, ChatPetitionGM, ChatTrade, ChatAlliance,
		ChatAnnounce, ChatBoat, ChatFriend, ChatMSNChat,
		ChatPartyMatchRoom, ChatPartyRoomCmd, ChatPartyRoomAll,
		ChatHeroVoice, ChatCriticalAnnounce, ChatScreenAnnounce,
		ChatBattlefield, ChatMPCC:
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
