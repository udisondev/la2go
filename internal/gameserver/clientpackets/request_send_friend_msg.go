package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSendFriendMsg is the C2S opcode 0xCC.
// Client sends this to deliver a private message to a friend.
const OpcodeRequestSendFriendMsg byte = 0xCC

// RequestSendFriendMsg represents a friend PM request.
type RequestSendFriendMsg struct {
	Message  string // message text (max 300 chars)
	Receiver string // target player name
}

// ParseRequestSendFriendMsg parses the packet from raw bytes.
func ParseRequestSendFriendMsg(data []byte) (*RequestSendFriendMsg, error) {
	r := packet.NewReader(data)

	msg, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading Message: %w", err)
	}

	receiver, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading Receiver: %w", err)
	}

	return &RequestSendFriendMsg{
		Message:  msg,
		Receiver: receiver,
	}, nil
}
