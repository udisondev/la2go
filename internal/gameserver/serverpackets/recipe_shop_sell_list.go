package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// RecipeShopSellList — sends the manufacture shop listing to a buyer.
// Displayed when a buyer clicks on another player's manufacture store.
//
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: RecipeShopSellList.java
//
// Opcode: 0xD9
// Format:
//
//	[sellerObjectId:int32]
//	[mpCurrent:int32]
//	[mpMax:int32]
//	[adena:int64]
//	[count:int32]
//	for each item:
//	  [recipeId:int32]
//	  [unknown1:int32=0]  (reserved)
//	  [cost:int64]
type RecipeShopSellList struct {
	SellerObjectID uint32
	CurrentMP      int32
	MaxMP          int32
	Adena          int64
	Items          []*model.ManufactureItem
}

// Write serializes the RecipeShopSellList packet.
func (p *RecipeShopSellList) Write() ([]byte, error) {
	// 1 (opcode) + 4+4+4+8 (header) + 16*N (items: 4+4+8 per entry)
	w := packet.NewWriter(21 + 16*len(p.Items))

	w.WriteByte(0xD9)
	w.WriteInt(int32(p.SellerObjectID))
	w.WriteInt(p.CurrentMP)
	w.WriteInt(p.MaxMP)
	w.WriteLong(p.Adena)
	w.WriteInt(int32(len(p.Items)))

	for _, item := range p.Items {
		w.WriteInt(item.RecipeID)
		w.WriteInt(0) // unknown1: reserved
		w.WriteLong(item.Cost)
	}

	return w.Bytes(), nil
}
