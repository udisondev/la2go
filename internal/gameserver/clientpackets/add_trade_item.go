package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAddTradeItem is the opcode for AddTradeItem packet (C2S 0x16).
// Player adds an item to the active trade window.
//
// Java reference: AddTradeItem.java
const OpcodeAddTradeItem = 0x16

// AddTradeItem packet (C2S 0x16) adds an item to the trade.
//
// Packet structure:
//   - tradeID (int32) — trade session ID (not used in logic)
//   - objectID (int32) — item ObjectID in inventory
//   - count (int32) — quantity to trade
type AddTradeItem struct {
	TradeID  int32
	ObjectID int32
	Count    int32
}

// ParseAddTradeItem parses AddTradeItem packet from raw bytes.
func ParseAddTradeItem(data []byte) (*AddTradeItem, error) {
	r := packet.NewReader(data)

	tradeID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading tradeID: %w", err)
	}

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	return &AddTradeItem{TradeID: tradeID, ObjectID: objectID, Count: count}, nil
}
