package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeInventoryItemList is the opcode for InventoryItemList packet (S2C 0x27)
	OpcodeInventoryItemList = 0x27
)

// InventoryItemList packet (S2C 0x27) sends list of items in character's inventory.
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
//	opcode (byte) = 0x27
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

	// augmentation (0 for MVP)
	w.WriteInt(0)

	// mana (-1 for MVP = no shadow item)
	w.WriteInt(-1)
}

// itemType1 возвращает type1 для item.
// Java: Item.getType1() — weapon=0, shield/armor=1, ring/earring/necklace=2, etc=5.
func itemType1(tmpl *model.ItemTemplate) int16 {
	switch tmpl.Type {
	case model.ItemTypeWeapon:
		return 0 // weapon
	case model.ItemTypeArmor:
		if tmpl.BodyPart == model.ArmorSlotNeck ||
			tmpl.BodyPart == model.ArmorSlotEar ||
			tmpl.BodyPart == model.ArmorSlotFinger {
			return 2 // jewel
		}
		return 1 // shield/armor
	default:
		return 5 // etc
	}
}

// itemType2 возвращает type2 для item.
// Java: Item.getType2() — weapon=0, shield/armor=1, ring/earring/necklace=2, quest=3, adena=4, item=5.
func itemType2(tmpl *model.ItemTemplate) int16 {
	switch tmpl.Type {
	case model.ItemTypeWeapon:
		return 0
	case model.ItemTypeArmor:
		if tmpl.BodyPart == model.ArmorSlotNeck ||
			tmpl.BodyPart == model.ArmorSlotEar ||
			tmpl.BodyPart == model.ArmorSlotFinger {
			return 2
		}
		return 1
	case model.ItemTypeQuestItem:
		return 3
	default:
		return 5
	}
}

// bodyPartMask возвращает mask слота для клиента.
// Java: BodyPart.getMask() — e.g. chest=0x0400, rhand=0x4000.
func bodyPartMask(tmpl *model.ItemTemplate) int32 {
	if tmpl.Type == model.ItemTypeWeapon {
		return 0x4000 // rhand
	}

	switch tmpl.BodyPart {
	case model.ArmorSlotChest:
		return 0x0400
	case model.ArmorSlotLegs:
		return 0x0800
	case model.ArmorSlotHead:
		return 0x0040
	case model.ArmorSlotFeet:
		return 0x1000
	case model.ArmorSlotGloves:
		return 0x0200
	case model.ArmorSlotUnder:
		return 0x0001
	case model.ArmorSlotCloak:
		return 0x4000 // back
	case model.ArmorSlotNeck:
		return 0x0008
	case model.ArmorSlotEar:
		return 0x0006 // lr.ear
	case model.ArmorSlotFinger:
		return 0x0030 // lr.finger
	default:
		return 0
	}
}
