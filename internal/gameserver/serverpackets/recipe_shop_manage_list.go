package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// RecipeShopManageList — sends the manufacture management window to the crafter.
// Shows their known recipes so they can select which to offer in the shop.
//
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: RecipeShopManageList.java
//
// Opcode: 0xD8
// Format:
//
//	[playerObjectId:int32]
//	[mpCurrent:int32]
//	[mpMax:int32]
//	[count:int32]
//	for each recipe:
//	  [recipeId:int32]
//	  [unknown1:int32=1]  (status flag; 1=available)
//	  [unknown2:int32=0]  (reserved)
type RecipeShopManageList struct {
	PlayerObjectID uint32
	CurrentMP      int32
	MaxMP          int32
	RecipeIDs      []int32 // Recipe list IDs the player knows
}

// Write serializes the RecipeShopManageList packet.
func (p *RecipeShopManageList) Write() ([]byte, error) {
	// 1 (opcode) + 4+4+4 (header) + 12*N (recipes: 4+4+4 per entry)
	w := packet.NewWriter(13 + 12*len(p.RecipeIDs))

	w.WriteByte(0xD8)
	w.WriteInt(int32(p.PlayerObjectID))
	w.WriteInt(p.CurrentMP)
	w.WriteInt(p.MaxMP)
	w.WriteInt(int32(len(p.RecipeIDs)))

	for _, id := range p.RecipeIDs {
		w.WriteInt(id)
		w.WriteInt(1) // unknown1: status=available
		w.WriteInt(0) // unknown2: reserved
	}

	return w.Bytes(), nil
}
