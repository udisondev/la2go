package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePrivateStoreListSell is the opcode for PrivateStoreListSell (S2C 0x9B).
// Shows a seller's store contents to a buyer.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreListSell.java
const OpcodePrivateStoreListSell = 0x9B

// PrivateStoreListSell shows the seller's store to potential buyers.
//
// Packet structure:
//   - opcode (byte) — 0x9B
//   - sellerObjectID (int32) — seller's ObjectID
//   - packageSale (int32) — 1=package, 0=normal
//   - buyerAdena (int32) — buyer's current Adena
//   - itemCount (int32) — number of items in store
//   - for each item: [type2, objectID, itemID, count, 0, enchant, 0, bodyPart, price, refPrice]
type PrivateStoreListSell struct {
	SellerObjectID uint32
	PackageSale    bool
	BuyerAdena     int64
	Items          []*model.TradeItem
}

// Write serializes PrivateStoreListSell packet.
func (p *PrivateStoreListSell) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 16 + len(p.Items)*44)

	w.WriteByte(OpcodePrivateStoreListSell)
	w.WriteInt(int32(p.SellerObjectID))
	if p.PackageSale {
		w.WriteInt(1)
	} else {
		w.WriteInt(0)
	}
	w.WriteInt(int32(p.BuyerAdena))

	w.WriteInt(int32(len(p.Items)))
	for _, ti := range p.Items {
		w.WriteInt(int32(ti.Type2))     // type2
		w.WriteInt(int32(ti.ObjectID))  // objectID
		w.WriteInt(ti.ItemID)           // itemID
		w.WriteInt(ti.Count)            // count
		w.WriteShort(0)                 // reserved
		w.WriteShort(int16(ti.Enchant)) // enchant
		w.WriteShort(0)                 // customType2
		w.WriteInt(ti.BodyPart)         // bodyPart
		w.WriteInt(int32(ti.Price))     // store price
		w.WriteInt(0)                   // referencePrice (simplified)
	}

	return w.Bytes(), nil
}
