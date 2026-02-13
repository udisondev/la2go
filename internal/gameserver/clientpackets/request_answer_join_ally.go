package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAnswerJoinAlly is the C2S opcode 0x83.
// Client sends this to accept or decline an alliance invitation.
const OpcodeRequestAnswerJoinAlly byte = 0x83

// RequestAnswerJoinAlly represents the response to an alliance invitation.
//
// Packet structure:
//   - Response (int32): 0 = decline, 1 = accept
type RequestAnswerJoinAlly struct {
	Response int32
}

// ParseRequestAnswerJoinAlly parses the C2S RequestAnswerJoinAlly packet.
func ParseRequestAnswerJoinAlly(data []byte) (*RequestAnswerJoinAlly, error) {
	r := packet.NewReader(data)

	response, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &RequestAnswerJoinAlly{Response: response}, nil
}
