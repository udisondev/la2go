package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePrivateStoreBuyManageList is the opcode for PrivateStoreBuyManageList (S2C 0xB7).
// Opens the buy store management UI for the store owner.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreBuyManageList.java
const OpcodePrivateStoreBuyManageList = 0xB7

// PrivateStoreBuyManageList opens the buy store management UI.
//
// Packet structure:
//   - opcode (byte) — 0xB7
//   - objectID (int32) — player's ObjectID
//   - adena (int32) — player's current Adena
//   - storeCount (int32) — items already in buy store
//   - for each store item: [itemID, type2, enchant, 0, 0, count, refPrice, bodyPart, type2Short, price]
type PrivateStoreBuyManageList struct {
	ObjectID    uint32
	PlayerAdena int64
	StoreItems  []*model.TradeItem
}

// Write serializes PrivateStoreBuyManageList packet.
func (p *PrivateStoreBuyManageList) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 8 + 4 + len(p.StoreItems)*40)

	w.WriteByte(OpcodePrivateStoreBuyManageList)
	w.WriteInt(int32(p.ObjectID))
	w.WriteInt(int32(p.PlayerAdena))

	// Items already in store
	w.WriteInt(int32(len(p.StoreItems)))
	for _, ti := range p.StoreItems {
		w.WriteInt(ti.ItemID)             // itemID
		w.WriteShort(int16(ti.Type2))     // type2
		w.WriteShort(int16(ti.Enchant))   // enchant (always 0 for buy store)
		w.WriteInt(ti.Count)              // max count to buy
		w.WriteInt(0)                     // referencePrice
		w.WriteInt(ti.BodyPart)           // bodyPart
		w.WriteShort(int16(ti.Type2))     // type2 again
		w.WriteInt(int32(ti.Price))       // buy price
	}

	return w.Bytes(), nil
}
