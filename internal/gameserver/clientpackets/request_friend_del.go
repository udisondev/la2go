package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestFriendDel is the opcode for friend deletion request (C2S 0x61).
const OpcodeRequestFriendDel = 0x61

// RequestFriendDel represents a client request to remove a friend.
type RequestFriendDel struct {
	Name string
}

// ParseRequestFriendDel parses a friend deletion request packet.
func ParseRequestFriendDel(data []byte) (*RequestFriendDel, error) {
	r := packet.NewReader(data)
	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}
	return &RequestFriendDel{Name: name}, nil
}
