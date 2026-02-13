package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeInventoryItemList is the opcode for InventoryItemList packet (S2C 0x1B).
	// Java reference: ItemList.java
	OpcodeInventoryItemList = 0x1B
)

// InventoryItemList packet (S2C 0x1B) sends list of items in character's inventory.
// Sent after UserInfo during spawn.
//
// Phase 6.0: Serializes real items from player's inventory.
// Java reference: ItemList.java, AbstractItemPacket.writeItem()
type InventoryItemList struct {
	ShowWindow bool
	Items      []*model.Item
}

// NewInventoryItemList creates InventoryItemList packet with real items.
func NewInventoryItemList(items []*model.Item) *InventoryItemList {
	return &InventoryItemList{
		Items: items,
	}
}

// Write serializes InventoryItemList packet to binary format.
//
// Packet structure:
//
//	opcode (byte) = 0x1B
//	showWindow (short)
//	itemCount (short)
//	for each item:
//	  type1 (short) — item category
//	  objectID (int32)
//	  itemID (int32) — template ID
//	  count (int32)
//	  type2 (short) — sub-type
//	  customType1 (short) — 0
//	  equipped (short) — 0 or 1
//	  bodyPart (int32) — slot mask
//	  enchant (short)
//	  customType2 (short) — 0
//	  augmentation (int32) — 0
//	  mana (int32) — -1
//
// Java reference: AbstractItemPacket.writeItem()
func (p *InventoryItemList) Write() ([]byte, error) {
	// 5 header bytes + 36 bytes per item
	w := packet.NewWriter(5 + len(p.Items)*36)

	w.WriteByte(OpcodeInventoryItemList)

	var showWindow int16
	if p.ShowWindow {
		showWindow = 1
	}
	w.WriteShort(showWindow)

	w.WriteShort(int16(len(p.Items)))

	for _, item := range p.Items {
		writeItem(w, item)
	}

	return w.Bytes(), nil
}

// writeItem сериализует один предмет в пакет.
// Java reference: AbstractItemPacket.writeItem() (36 bytes per item).
func writeItem(w *packet.Writer, item *model.Item) {
	tmpl := item.Template()

	// type1: категория предмета (weapon=0, armor=1, jewel=2, quest=3, adena=4, etc=5)
	w.WriteShort(itemType1(tmpl))

	// objectID
	w.WriteInt(int32(item.ObjectID()))

	// itemID (template/display ID)
	w.WriteInt(item.ItemID())

	// count
	w.WriteInt(item.Count())

	// type2: подтип (weapon=0, shield/armor=1, ring/earring/necklace=2, quest=3, adena=4, item=5)
	w.WriteShort(itemType2(tmpl))

	// customType1 (always 0)
	w.WriteShort(0)

	// equipped (0=no, 1=yes)
	var equipped int16
	if item.IsEquipped() {
		equipped = 1
	}
	w.WriteShort(equipped)

	// bodyPart slot mask
	w.WriteInt(bodyPartMask(tmpl))

	// enchant level
	w.WriteShort(int16(item.Enchant()))

	// customType2 (always 0)
	w.WriteShort(0)

	// augmentation (Phase 28: real augmentation ID from item)
	w.WriteInt(item.AugmentationID())

	// mana (-1 for MVP = no shadow item)
	w.WriteInt(-1)
}

// itemType1 returns type1 from item template.
// Java: Item.getType1() — pre-computed in ItemTemplate at load time.
func itemType1(tmpl *model.ItemTemplate) int16 {
	if tmpl == nil {
		return model.Type1ItemQuestItemAdena
	}
	return tmpl.Type1
}

// itemType2 returns type2 from item template.
// Java: Item.getType2() — pre-computed in ItemTemplate at load time.
func itemType2(tmpl *model.ItemTemplate) int16 {
	if tmpl == nil {
		return model.Type2Other
	}
	return tmpl.Type2
}

// bodyPartMask returns body part bitmask from item template.
// Java: BodyPart.getMask() — pre-computed in ItemTemplate at load time.
func bodyPartMask(tmpl *model.ItemTemplate) int32 {
	if tmpl == nil {
		return 0
	}
	return tmpl.BodyPartMask
}
