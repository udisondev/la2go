package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodePetItemList is the opcode for PetItemList (S2C 0xB2).
	OpcodePetItemList = 0xB2
)

// PetItemList sends pet's inventory items to the owner.
// Phase 19: Pets/Summons System.
// Java reference: PetItemList.java
type PetItemList struct {
	Items []*model.Item
}

// NewPetItemList creates a PetItemList packet.
func NewPetItemList(items []*model.Item) PetItemList {
	return PetItemList{Items: items}
}

// Write serializes PetItemList packet.
func (p *PetItemList) Write() ([]byte, error) {
	w := packet.NewWriter(64 + len(p.Items)*36)

	w.WriteByte(OpcodePetItemList)

	// Item count
	w.WriteShort(int16(len(p.Items)))

	// Reuse writeItem helper from inventory_item_list.go
	for _, item := range p.Items {
		writeItem(w, item)
	}

	return w.Bytes(), nil
}
