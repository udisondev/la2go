package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeUseItem is the client packet opcode for UseItem (C2S 0x14).
// Client sends this when player double-clicks an item in inventory.
// Java reference: UseItem.java
const OpcodeUseItem = 0x14

// UseItem represents the client's use item request packet.
// Sent when player double-clicks an inventory item (equip/unequip or consume).
//
// Packet structure:
//   - ObjectID (int32): Item's object ID in inventory
//   - CtrlPressed (int32): 1 if Ctrl was held (force attack in PvP)
//
// Java reference: UseItem.java, opcode 0x14
type UseItem struct {
	ObjectID    int32 // Item's unique object ID
	CtrlPressed bool  // True if Ctrl key was held during use
}

// ParseUseItem parses a UseItem packet from raw bytes.
func ParseUseItem(data []byte) (*UseItem, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ObjectID: %w", err)
	}

	ctrlPressed, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading CtrlPressed: %w", err)
	}

	return &UseItem{
		ObjectID:    objectID,
		CtrlPressed: ctrlPressed != 0,
	}, nil
}
