package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeBuyList is the opcode for BuyList packet (S2C 0x11).
// Sends list of items available for purchase from NPC shop.
//
// Phase 8.3: NPC Shops.
// Java reference: BuyList.java
const OpcodeBuyList = 0x11

// BuyListProduct represents a single product in a buy list.
type BuyListProduct struct {
	ItemID       int32
	Price        int64 // Final price (base * tax)
	Count        int32 // Current stock (-1 = unlimited)
	RestockDelay int32 // Seconds until restock (0 = no restock)
	// Item display info
	Type1    int16 // Item type1 (weapon=0, armor=1, jewel=2, etc=5)
	Type2    int16 // Item type2
	BodyPart int32 // Body part mask
	Weight   int32
}

// BuyList packet (S2C 0x11) sends available items for purchase.
//
// Packet structure:
//   - opcode (byte) — 0x11
//   - adena (int32) — player's current Adena
//   - listID (int32) — buy list ID
//   - itemCount (short) — number of items
//   - for each item:
//   - type1 (short)
//   - objectID (int32) — 0 for shop items
//   - itemID (int32) — template ID
//   - count (int32) — max purchasable (-1 = unlimited)
//   - type2 (short)
//   - customType1 (short) — 0
//   - bodyPart (int32) — slot mask
//   - enchant (short) — 0
//   - customType2 (short) — 0
//   - augmentation (int32) — 0
//   - mana (int32) — -1
//   - price (int32) — price per unit
//
// Phase 8.3: NPC Shops.
type BuyList struct {
	PlayerAdena int64
	ListID      int32
	Products    []BuyListProduct
}

// NewBuyList creates BuyList packet.
func NewBuyList(playerAdena int64, listID int32, products []BuyListProduct) *BuyList {
	return &BuyList{
		PlayerAdena: playerAdena,
		ListID:      listID,
		Products:    products,
	}
}

// Write serializes BuyList packet to bytes.
//
// Phase 8.3: NPC Shops.
func (p *BuyList) Write() ([]byte, error) {
	// 1 opcode + 4 adena + 4 listID + 2 count + 40 bytes per item
	w := packet.NewWriter(11 + len(p.Products)*40)

	w.WriteByte(OpcodeBuyList)
	w.WriteInt(int32(p.PlayerAdena))
	w.WriteInt(p.ListID)

	w.WriteShort(int16(len(p.Products)))

	for _, prod := range p.Products {
		w.WriteShort(prod.Type1)        // type1
		w.WriteInt(0)                   // objectID (0 for shop items)
		w.WriteInt(prod.ItemID)         // item template ID
		w.WriteInt(prod.Count)          // max purchasable (-1=unlimited)
		w.WriteShort(prod.Type2)        // type2
		w.WriteShort(0)                 // customType1
		w.WriteInt(prod.BodyPart)       // bodyPart
		w.WriteShort(0)                 // enchant
		w.WriteShort(0)                 // customType2
		w.WriteInt(0)                   // augmentation
		w.WriteInt(-1)                  // mana (shadow item)
		w.WriteInt(int32(prod.Price))   // price
	}

	return w.Bytes(), nil
}
