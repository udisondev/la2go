package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// RecipeShopMsg — broadcasts the manufacture store message (title) to nearby players.
// This is the text shown above the crafter's head when they are in manufacture mode.
//
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: RecipeShopMsg.java
//
// Opcode: 0xDB
// Format:
//
//	[objectId:int32]
//	[message:string]  (UTF-16LE null-terminated)
type RecipeShopMsg struct {
	ObjectID uint32
	Message  string
}

// Write serializes the RecipeShopMsg packet.
func (p *RecipeShopMsg) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 4 + len(p.Message)*2 + 2)

	w.WriteByte(0xDB)
	w.WriteInt(int32(p.ObjectID))
	w.WriteString(p.Message)

	return w.Bytes(), nil
}
