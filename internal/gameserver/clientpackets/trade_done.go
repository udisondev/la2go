package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeTradeDone is the opcode for TradeDone packet (C2S 0x17).
// Player confirms or cancels the active trade.
//
// Java reference: TradeDone.java
const OpcodeTradeDone = 0x17

// TradeDone packet (C2S 0x17) confirms or cancels a trade.
//
// Packet structure:
//   - response (int32) — 1=confirm, 0=cancel
type TradeDone struct {
	Response int32
}

// ParseTradeDone parses TradeDone packet from raw bytes.
func ParseTradeDone(data []byte) (*TradeDone, error) {
	r := packet.NewReader(data)

	response, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &TradeDone{Response: response}, nil
}
