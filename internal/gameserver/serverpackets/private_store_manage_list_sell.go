package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodePrivateStoreManageListSell is the opcode for PrivateStoreManageListSell (S2C 0x9A).
// Opens the sell store management UI for the store owner.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreManageListSell.java
const OpcodePrivateStoreManageListSell = 0x9A

// PrivateStoreManageListSell opens the sell store management UI.
//
// Packet structure:
//   - opcode (byte) — 0x9A
//   - objectID (int32) — player's ObjectID
//   - packageSale (int32) — 1=package, 0=normal
//   - adena (int32) — player's current Adena
//   - sellableCount (int32) — number of sellable items in inventory
//   - for each sellable item: [type2, objectID, itemID, count, type2Short, enchant, type2Short2, bodyPart, price]
//   - storeCount (int32) — number of items already in store
//   - for each store item: [type2, objectID, itemID, count, type2Short, enchant, type2Short2, bodyPart, price, storePrice]
type PrivateStoreManageListSell struct {
	ObjectID      uint32
	PackageSale   bool
	PlayerAdena   int64
	SellableItems []*model.Item
	StoreItems    []*model.TradeItem
}

// Write serializes PrivateStoreManageListSell packet.
func (p *PrivateStoreManageListSell) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 12 + len(p.SellableItems)*40 + 4 + len(p.StoreItems)*44)

	w.WriteByte(OpcodePrivateStoreManageListSell)
	w.WriteInt(int32(p.ObjectID))
	if p.PackageSale {
		w.WriteInt(1)
	} else {
		w.WriteInt(0)
	}
	w.WriteInt(int32(p.PlayerAdena))

	// Sellable items from inventory
	w.WriteInt(int32(len(p.SellableItems)))
	for _, item := range p.SellableItems {
		writeItemInfo(w, item)
		// Reference price (NPC store price)
		if item.Template() != nil {
			w.WriteInt(0) // referencePrice — simplified for MVP
		} else {
			w.WriteInt(0)
		}
	}

	// Items already in store
	w.WriteInt(int32(len(p.StoreItems)))
	for _, ti := range p.StoreItems {
		w.WriteInt(int32(ti.Type2))  // type2
		w.WriteInt(int32(ti.ObjectID)) // objectID
		w.WriteInt(ti.ItemID)        // itemID
		w.WriteInt(ti.Count)         // count
		w.WriteShort(0)              // type2 short
		w.WriteShort(int16(ti.Enchant)) // enchant
		w.WriteShort(0)              // customType2
		w.WriteInt(ti.BodyPart)      // bodyPart
		w.WriteInt(int32(ti.Price))  // store price
	}

	return w.Bytes(), nil
}

// writeItemInfo writes common item display fields to packet writer.
func writeItemInfo(w *packet.Writer, item *model.Item) {
	tmpl := item.Template()
	var type2 int16
	var bodyPart int32
	if tmpl != nil {
		type2 = tmpl.Type2
		bodyPart = tmpl.BodyPartMask
	}

	w.WriteShort(type2)                // type2
	w.WriteInt(int32(item.ObjectID())) // objectID
	w.WriteInt(item.ItemID())          // itemID
	w.WriteInt(item.Count())           // count
	w.WriteShort(0)                    // type2 short
	w.WriteShort(int16(item.Enchant())) // enchant
	w.WriteShort(0)                    // customType2
	w.WriteInt(bodyPart)               // bodyPart
}
