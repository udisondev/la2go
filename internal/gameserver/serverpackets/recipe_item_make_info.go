package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// RecipeItemMakeInfo — sends crafting result info to client.
// Java reference: RecipeItemMakeInfo.java
//
// Opcode: 0xD7
// Format:
//
//	[recipeListId:int32]
//	[isDwarvenCraft:int32]  (0=dwarven, 1=common)
//	[currentMp:int32]
//	[maxMp:int32]
//	[success:int32]  (0=fail, 1=success)
type RecipeItemMakeInfo struct {
	RecipeListID   int32
	IsDwarvenCraft bool
	CurrentMP      int32
	MaxMP          int32
	Success        bool
}

// Write serializes the RecipeItemMakeInfo packet.
func (p *RecipeItemMakeInfo) Write() ([]byte, error) {
	w := packet.NewWriter(21) // 1 + 4*5

	w.WriteByte(0xD7)
	w.WriteInt(p.RecipeListID)

	if p.IsDwarvenCraft {
		w.WriteInt(0) // 0 = dwarven
	} else {
		w.WriteInt(1) // 1 = common
	}

	w.WriteInt(p.CurrentMP)
	w.WriteInt(p.MaxMP)

	if p.Success {
		w.WriteInt(1)
	} else {
		w.WriteInt(0)
	}

	return w.Bytes(), nil
}
