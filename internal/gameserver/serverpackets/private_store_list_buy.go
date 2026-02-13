package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePrivateStoreListBuy is the opcode for PrivateStoreListBuy (S2C 0xB8).
// Shows a buyer's store contents to a potential seller.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreListBuy.java
const OpcodePrivateStoreListBuy = 0xB8

// PrivateStoreListBuy shows the buyer's store to potential sellers.
//
// Packet structure:
//   - opcode (byte) — 0xB8
//   - buyerObjectID (int32) — buyer's ObjectID
//   - sellerAdena (int32) — seller's current Adena (visiting player)
//   - itemCount (int32) — number of items in store
//   - for each item: [objectID, itemID, enchant, count, refPrice, 0, bodyPart, type2, price, storeCount]
type PrivateStoreListBuy struct {
	BuyerObjectID uint32
	SellerAdena   int64
	Items         []*model.TradeItem
}

// Write serializes PrivateStoreListBuy packet.
func (p *PrivateStoreListBuy) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 12 + len(p.Items)*44)

	w.WriteByte(OpcodePrivateStoreListBuy)
	w.WriteInt(int32(p.BuyerObjectID))
	w.WriteInt(int32(p.SellerAdena))

	w.WriteInt(int32(len(p.Items)))
	for _, ti := range p.Items {
		w.WriteInt(int32(ti.ObjectID))  // objectID (0 for buy stores)
		w.WriteInt(ti.ItemID)           // itemID
		w.WriteShort(int16(ti.Enchant)) // enchant
		w.WriteInt(ti.Count)            // current count remaining
		w.WriteInt(0)                   // referencePrice
		w.WriteShort(0)                 // reserved
		w.WriteInt(ti.BodyPart)         // bodyPart
		w.WriteShort(int16(ti.Type2))   // type2
		w.WriteInt(int32(ti.Price))     // buy price per unit
		w.WriteInt(ti.StoreCount)       // max buy count (original)
	}

	return w.Bytes(), nil
}
