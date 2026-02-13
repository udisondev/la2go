package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAnswerTradeRequest is the opcode for AnswerTradeRequest packet (C2S 0x44).
// Player responds to an incoming trade request.
//
// Java reference: AnswerTradeRequest.java
const OpcodeAnswerTradeRequest = 0x44

// AnswerTradeRequest packet (C2S 0x44) responds to a trade request.
//
// Packet structure:
//   - response (int32) — 1=accept, 0=reject
type AnswerTradeRequest struct {
	Response int32
}

// ParseAnswerTradeRequest parses AnswerTradeRequest packet from raw bytes.
func ParseAnswerTradeRequest(data []byte) (*AnswerTradeRequest, error) {
	r := packet.NewReader(data)

	response, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &AnswerTradeRequest{Response: response}, nil
}
