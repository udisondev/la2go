package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSocialAction is the server packet opcode for SocialAction (S2C 0x2D).
const OpcodeSocialAction = 0x2D

// Social action IDs (Interlude).
const (
	SocialActionGreeting = 2  // /bow, /greeting
	SocialActionVictory  = 3  // /victory
	SocialActionAdvance  = 4  // /advance
	SocialActionEtc      = 5  // /etc
	SocialActionYes      = 6  // /yes
	SocialActionNo       = 7  // /no
	SocialActionBow      = 8  // /bow
	SocialActionUnaware  = 9  // /unaware
	SocialActionWait     = 10 // /social wait
	SocialActionLaugh    = 11 // /laugh
	SocialActionApplaud  = 12 // /applaud
	SocialActionDance    = 13 // /dance
	SocialActionSorrow   = 14 // /sorrow
	SocialActionCharm    = 15 // /charm (heroes only)
	SocialActionShyness  = 16 // /shyness

	// SocialActionLevelUp is used internally for level-up animation.
	SocialActionLevelUp = 15

	// MinSocialActionID is the minimum valid social action ID.
	MinSocialActionID = 2
	// MaxSocialActionID is the maximum valid social action ID.
	MaxSocialActionID = 16
)

// SocialAction represents a social action packet (S2C 0x2D).
// Broadcasts a social emote/gesture animation to nearby players.
type SocialAction struct {
	ObjectID int32
	ActionID int32
}

// NewSocialAction creates a new social action packet.
func NewSocialAction(objectID, actionID int32) SocialAction {
	return SocialAction{
		ObjectID: objectID,
		ActionID: actionID,
	}
}

// Write serializes the SocialAction packet.
func (p *SocialAction) Write() ([]byte, error) {
	// opcode(1) + objectID(4) + actionID(4)
	w := packet.NewWriter(9)

	w.WriteByte(OpcodeSocialAction)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.ActionID)

	return w.Bytes(), nil
}
