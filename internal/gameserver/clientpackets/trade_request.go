package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeTradeRequest is the opcode for TradeRequest packet (C2S 0x15).
// Player initiates a trade with target player.
//
// Java reference: TradeRequest.java
const OpcodeTradeRequest = 0x15

// TradeRequest packet (C2S 0x15) initiates player-to-player trade.
//
// Packet structure:
//   - objectID (int32) — target player's ObjectID
type TradeRequest struct {
	TargetObjectID int32
}

// ParseTradeRequest parses TradeRequest packet from raw bytes.
func ParseTradeRequest(data []byte) (*TradeRequest, error) {
	r := packet.NewReader(data)

	targetObjectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading targetObjectID: %w", err)
	}

	return &TradeRequest{TargetObjectID: targetObjectID}, nil
}
