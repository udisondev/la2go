package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAnswerJoinParty is the client packet opcode for party invite answer (C2S 0x2A).
//
// Packet structure (C2S 0x2A):
//   - answer int32  0 = decline, 1 = accept
//
// Java reference: RequestAnswerJoinParty.java (opcode 0x2A).
const OpcodeRequestAnswerJoinParty = 0x2A

// RequestAnswerJoinParty represents a client response to a party invite.
type RequestAnswerJoinParty struct {
	Answer int32 // 0 = decline, 1 = accept
}

// ParseRequestAnswerJoinParty parses RequestAnswerJoinParty packet from raw bytes.
// Opcode already stripped by HandlePacket.
func ParseRequestAnswerJoinParty(data []byte) (*RequestAnswerJoinParty, error) {
	r := packet.NewReader(data)

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading answer: %w", err)
	}

	return &RequestAnswerJoinParty{
		Answer: answer,
	}, nil
}

// IsAccepted returns true if the invite was accepted.
func (p *RequestAnswerJoinParty) IsAccepted() bool {
	return p.Answer == 1
}
