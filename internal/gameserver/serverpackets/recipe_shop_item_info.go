package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// RecipeShopItemInfo — sends manufacture shop recipe info to client.
// Java reference: RecipeShopItemInfo.java
//
// Opcode: 0xDA
// Format:
//
//	[crafterObjectId:int32]
//	[recipeId:int32]  (item ID of recipe book, NOT recipeListId!)
//	[currentMp:int32]
//	[maxMp:int32]
//	[unknown:int32]  (always 0xFFFFFFFF)
type RecipeShopItemInfo struct {
	CrafterObjectID uint32
	RecipeID        int32 // item ID of recipe book
	CurrentMP       int32
	MaxMP           int32
}

// Write serializes the RecipeShopItemInfo packet.
func (p *RecipeShopItemInfo) Write() ([]byte, error) {
	w := packet.NewWriter(21) // 1 + 4*5

	w.WriteByte(0xDA)
	w.WriteInt(int32(p.CrafterObjectID))
	w.WriteInt(p.RecipeID)
	w.WriteInt(p.CurrentMP)
	w.WriteInt(p.MaxMP)
	w.WriteInt(-1) // unknown, always 0xFFFFFFFF

	return w.Bytes(), nil
}
