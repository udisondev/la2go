package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestUnEquipItem is the opcode for RequestUnEquipItem (C2S 0x11).
// Java reference: ClientPackets.REQUEST_UNEQUIP_ITEM(0x11).
const OpcodeRequestUnEquipItem = 0x11

// RequestUnEquipItem represents a client request to unequip an item from a slot.
//
// Packet structure (body after opcode):
//   - slot (int32) — paperdoll slot to unequip from
type RequestUnEquipItem struct {
	Slot int32
}

// ParseRequestUnEquipItem parses RequestUnEquipItem packet.
func ParseRequestUnEquipItem(data []byte) (*RequestUnEquipItem, error) {
	r := packet.NewReader(data)

	slot, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading slot: %w", err)
	}

	return &RequestUnEquipItem{Slot: slot}, nil
}
