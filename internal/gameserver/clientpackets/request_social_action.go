package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSocialAction is the opcode for RequestSocialAction (C2S 0x1B).
// Java reference: ClientPackets.REQUEST_SOCIAL_ACTION(0x1B).
const OpcodeRequestSocialAction = 0x1B

// RequestSocialAction represents a client request to perform a social action (emote).
//
// Packet structure (body after opcode):
//   - actionID (int32) — social action ID (1-8 for standard emotes)
type RequestSocialAction struct {
	ActionID int32
}

// ParseRequestSocialAction parses RequestSocialAction packet.
func ParseRequestSocialAction(data []byte) (*RequestSocialAction, error) {
	r := packet.NewReader(data)

	actionID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading actionID: %w", err)
	}

	return &RequestSocialAction{ActionID: actionID}, nil
}
