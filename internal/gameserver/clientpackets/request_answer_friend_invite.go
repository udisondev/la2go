package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAnswerFriendInvite is the opcode for friend invite response (C2S 0x5F).
const OpcodeRequestAnswerFriendInvite = 0x5F

// RequestAnswerFriendInvite represents a client response to a friend invite.
type RequestAnswerFriendInvite struct {
	Response int32 // 1=accept, 0=decline
}

// ParseRequestAnswerFriendInvite parses a friend invite answer packet.
func ParseRequestAnswerFriendInvite(data []byte) (*RequestAnswerFriendInvite, error) {
	r := packet.NewReader(data)
	resp, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return &RequestAnswerFriendInvite{Response: resp}, nil
}
