package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	OpcodeSocialAction = 0x2D
)

// Social action IDs
const (
	SocialActionLevelUp = 15
)

// SocialAction represents a social action packet (S2C 0x2D).
// Used for animations like level-up, greeting, etc.
type SocialAction struct {
	ObjectID int32
	ActionID int32
}

// NewSocialAction creates a new social action packet.
func NewSocialAction(objectID int32, actionID int32) *SocialAction {
	return &SocialAction{
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
