package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// RecipeBookItemList — sends player's recipe book to client.
// Java reference: RecipeBookItemList.java
//
// Opcode: 0xD6
// Format:
//
//	[isDwarvenCraft:int32]  (0=dwarven, 1=common)
//	[maxMp:int32]
//	[recipeCount:int32]
//	  [recipeListId:int32]
//	  [displayIndex:int32]  (1-based)
//	...
type RecipeBookItemList struct {
	IsDwarvenCraft bool
	MaxMP          int32
	RecipeIDs      []int32
}

// Write serializes the RecipeBookItemList packet.
func (p *RecipeBookItemList) Write() ([]byte, error) {
	// 1 (opcode) + 4+4+4 (header) + 8*N (recipes)
	w := packet.NewWriter(13 + 8*len(p.RecipeIDs))

	w.WriteByte(0xD6)

	if p.IsDwarvenCraft {
		w.WriteInt(0) // 0 = dwarven
	} else {
		w.WriteInt(1) // 1 = common
	}

	w.WriteInt(p.MaxMP)
	w.WriteInt(int32(len(p.RecipeIDs)))

	for i, id := range p.RecipeIDs {
		w.WriteInt(id)
		w.WriteInt(int32(i + 1)) // 1-based display index
	}

	return w.Bytes(), nil
}
