package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestFriendInvite is the opcode for friend invite request (C2S 0x5E).
const OpcodeRequestFriendInvite = 0x5E

// RequestFriendInvite represents a client request to invite another player as friend.
type RequestFriendInvite struct {
	Name string
}

// ParseRequestFriendInvite parses a friend invite request packet.
func ParseRequestFriendInvite(data []byte) (*RequestFriendInvite, error) {
	r := packet.NewReader(data)
	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}
	return &RequestFriendInvite{Name: name}, nil
}
