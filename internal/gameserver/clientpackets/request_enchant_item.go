package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestEnchantItem is the opcode for RequestEnchantItem packet (C2S 0x58).
// Player attempts to enchant an item using the active enchant scroll.
//
// Java reference: RequestEnchantItem.java
const OpcodeRequestEnchantItem = 0x58

// RequestEnchantItem packet (C2S 0x58) enchants an item.
//
// Packet structure:
//   - objectID (int32) — ObjectID of item to enchant
type RequestEnchantItem struct {
	ObjectID int32
}

// ParseRequestEnchantItem parses RequestEnchantItem packet from raw bytes.
func ParseRequestEnchantItem(data []byte) (*RequestEnchantItem, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	return &RequestEnchantItem{ObjectID: objectID}, nil
}
