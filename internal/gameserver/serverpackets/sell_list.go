package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeSellList is the opcode for SellList packet (S2C 0x10).
// Sends list of player's items available for selling to NPC.
//
// Phase 8.3: NPC Shops.
// Java reference: SellList.java
const OpcodeSellList = 0x10

// SellList packet (S2C 0x10) sends player's sellable items.
//
// Packet structure:
//   - opcode (byte) — 0x10
//   - adena (int32) — player's current Adena
//   - itemCount (short) — number of items
//   - for each item:
//   - type1 (short)
//   - objectID (int32)
//   - itemID (int32)
//   - count (int32)
//   - type2 (short)
//   - customType1 (short) — 0
//   - equipped (short) — 0 (only unequipped items can be sold)
//   - bodyPart (int32)
//   - enchant (short)
//   - customType2 (short) — 0
//   - augmentation (int32) — 0
//   - mana (int32) — -1
//   - price (int32) — sell price (50% of base)
//
// Phase 8.3: NPC Shops.
type SellList struct {
	PlayerAdena int64
	Items       []SellListItem
}

// SellListItem represents one sellable item.
type SellListItem struct {
	Item      *model.Item
	SellPrice int64 // 50% of base price
}

// NewSellList creates SellList packet.
func NewSellList(playerAdena int64, items []SellListItem) *SellList {
	return &SellList{
		PlayerAdena: playerAdena,
		Items:       items,
	}
}

// Write serializes SellList packet to bytes.
//
// Phase 8.3: NPC Shops.
func (p *SellList) Write() ([]byte, error) {
	// 1 opcode + 4 adena + 2 count + 40 bytes per item
	w := packet.NewWriter(7 + len(p.Items)*40)

	w.WriteByte(OpcodeSellList)
	w.WriteInt(int32(p.PlayerAdena))

	w.WriteShort(int16(len(p.Items)))

	for _, si := range p.Items {
		item := si.Item
		tmpl := item.Template()

		w.WriteShort(itemType1(tmpl))        // type1
		w.WriteInt(int32(item.ObjectID()))    // objectID
		w.WriteInt(item.ItemID())             // item template ID
		w.WriteInt(item.Count())              // count
		w.WriteShort(itemType2(tmpl))         // type2
		w.WriteShort(0)                       // customType1
		w.WriteShort(0)                       // equipped (0 — only unequipped items)
		w.WriteInt(bodyPartMask(tmpl))        // bodyPart
		w.WriteShort(int16(item.Enchant()))   // enchant
		w.WriteShort(0)                       // customType2
		w.WriteInt(0)                         // augmentation
		w.WriteInt(-1)                        // mana
		w.WriteInt(int32(si.SellPrice))       // sell price
	}

	return w.Bytes(), nil
}
