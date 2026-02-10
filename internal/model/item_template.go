package model

// ItemTemplate — шаблон предмета из items XML.
// Содержит базовые характеристики weapon/armor/consumable для создания конкретных Item.
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: ItemTemplate.java
type ItemTemplate struct {
	ItemID int32  // Template ID (unique, from items XML)
	Name   string // Item name (e.g., "Short Sword", "Leather Shirt")
	Type   ItemType

	// Weapon stats
	PAtk        int32 // Physical attack bonus (added to base pAtk)
	AttackRange int32 // Attack range in game units (40 for sword, 500 for bow)

	// Armor stats
	PDef     int32     // Physical defense bonus (added after slot subtraction)
	BodyPart ArmorSlot // Which body slot this armor occupies (chest, legs, etc.)

	// Common stats
	Weight    int32 // Item weight (affects inventory capacity)
	Stackable bool  // Can stack multiple items (arrows, potions)
	Tradeable bool  // Can trade with other players
}

// ItemType определяет категорию предмета.
type ItemType int32

const (
	ItemTypeWeapon ItemType = iota
	ItemTypeArmor
	ItemTypeConsumable
	ItemTypeQuestItem
	ItemTypeEtcItem
)

// String returns human-readable item type name.
func (it ItemType) String() string {
	switch it {
	case ItemTypeWeapon:
		return "Weapon"
	case ItemTypeArmor:
		return "Armor"
	case ItemTypeConsumable:
		return "Consumable"
	case ItemTypeQuestItem:
		return "QuestItem"
	case ItemTypeEtcItem:
		return "EtcItem"
	default:
		return "Unknown"
	}
}

// ArmorSlot определяет слот брони (соответствует paperdoll slots для armor).
type ArmorSlot int32

const (
	ArmorSlotNone ArmorSlot = iota
	ArmorSlotChest
	ArmorSlotLegs
	ArmorSlotHead
	ArmorSlotFeet
	ArmorSlotGloves
	ArmorSlotUnder // Underwear
	ArmorSlotCloak
	ArmorSlotNeck  // Necklace
	ArmorSlotEar   // Earrings
	ArmorSlotFinger // Rings
)

// String returns human-readable armor slot name.
func (as ArmorSlot) String() string {
	switch as {
	case ArmorSlotNone:
		return "None"
	case ArmorSlotChest:
		return "Chest"
	case ArmorSlotLegs:
		return "Legs"
	case ArmorSlotHead:
		return "Head"
	case ArmorSlotFeet:
		return "Feet"
	case ArmorSlotGloves:
		return "Gloves"
	case ArmorSlotUnder:
		return "Underwear"
	case ArmorSlotCloak:
		return "Cloak"
	case ArmorSlotNeck:
		return "Necklace"
	case ArmorSlotEar:
		return "Earring"
	case ArmorSlotFinger:
		return "Ring"
	default:
		return "Unknown"
	}
}

// IsWeapon returns true if this template is a weapon.
func (t *ItemTemplate) IsWeapon() bool {
	return t.Type == ItemTypeWeapon
}

// IsArmor returns true if this template is armor.
func (t *ItemTemplate) IsArmor() bool {
	return t.Type == ItemTypeArmor
}

// IsConsumable returns true if this template is consumable (potion, scroll).
func (t *ItemTemplate) IsConsumable() bool {
	return t.Type == ItemTypeConsumable
}
