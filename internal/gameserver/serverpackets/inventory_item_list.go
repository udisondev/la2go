package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeInventoryItemList is the opcode for InventoryItemList packet (S2C 0x27)
	OpcodeInventoryItemList = 0x27
)

// InventoryItemList packet (S2C 0x27) sends list of items in character's inventory.
// Sent after UserInfo during spawn.
type InventoryItemList struct {
	// Items []InventoryItem // TODO Phase 4.8: implement item system
}

// NewInventoryItemList creates empty InventoryItemList packet.
// TODO Phase 4.8: Load items from database.
func NewInventoryItemList() *InventoryItemList {
	return &InventoryItemList{}
}

// Write serializes InventoryItemList packet to binary format.
func (p *InventoryItemList) Write() ([]byte, error) {
	// Empty inventory for now (no items)
	w := packet.NewWriter(16)

	w.WriteByte(OpcodeInventoryItemList)
	w.WriteShort(0) // Item count = 0

	return w.Bytes(), nil
}
