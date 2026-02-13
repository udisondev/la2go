package model

// CrystalType represents item grade (crystal type) from Java CrystalType enum.
// Java reference: ItemTemplate.java CrystalType
type CrystalType int32

const (
	CrystalNone CrystalType = 0 // No grade
	CrystalD    CrystalType = 1 // D-grade
	CrystalC    CrystalType = 2 // C-grade
	CrystalB    CrystalType = 3 // B-grade
	CrystalA    CrystalType = 4 // A-grade
	CrystalS    CrystalType = 5 // S-grade
)

// String returns human-readable crystal type name.
func (ct CrystalType) String() string {
	switch ct {
	case CrystalNone:
		return "NONE"
	case CrystalD:
		return "D"
	case CrystalC:
		return "C"
	case CrystalB:
		return "B"
	case CrystalA:
		return "A"
	case CrystalS:
		return "S"
	default:
		return "NONE"
	}
}

// CrystalTypeFromString converts string crystal type to CrystalType.
func CrystalTypeFromString(s string) CrystalType {
	switch s {
	case "D":
		return CrystalD
	case "C":
		return CrystalC
	case "B":
		return CrystalB
	case "A":
		return CrystalA
	case "S":
		return CrystalS
	default:
		return CrystalNone
	}
}

// Client type1 constants (Java: ItemTemplate.TYPE1_*).
// Used in InventoryItemList, InventoryUpdate, and other client packets.
const (
	Type1WeaponRingEarringNecklace int16 = 0 // Weapons and accessories
	Type1ShieldArmor               int16 = 1 // Shield, armor
	Type1ItemQuestItemAdena        int16 = 4 // Etc items, quest items, adena
)

// Client type2 constants (Java: ItemTemplate.TYPE2_*).
const (
	Type2Weapon      int16 = 0
	Type2ShieldArmor int16 = 1
	Type2Accessory   int16 = 2
	Type2Quest       int16 = 3
	Type2Money       int16 = 4
	Type2Other       int16 = 5
)

// BodyPart bitmask constants (Java: BodyPart enum values).
// These bitmask values are sent to the client in item-related packets.
const (
	BodyPartUnderwear int32 = 0x0001
	BodyPartREar      int32 = 0x0002
	BodyPartLEar      int32 = 0x0004
	BodyPartNeck      int32 = 0x0008
	BodyPartRFinger   int32 = 0x0010
	BodyPartLFinger   int32 = 0x0020
	BodyPartHead      int32 = 0x0040
	BodyPartRHand     int32 = 0x0080
	BodyPartLHand     int32 = 0x0100
	BodyPartGloves    int32 = 0x0200
	BodyPartChest     int32 = 0x0400
	BodyPartLegs      int32 = 0x0800
	BodyPartFeet      int32 = 0x1000
	BodyPartBack      int32 = 0x2000
	BodyPartLRHand    int32 = 0x4000
	BodyPartFullArmor int32 = 0x8000
	BodyPartHair      int32 = 0x010000
	BodyPartAllDress  int32 = 0x020000
	BodyPartHair2     int32 = 0x040000
	BodyPartHairAll   int32 = 0x080000
)

// BodyPartMaskFromString converts a body part string from XML to a bitmask value.
func BodyPartMaskFromString(s string) int32 {
	switch s {
	case "underwear":
		return BodyPartUnderwear
	case "rear":
		return BodyPartREar
	case "lear":
		return BodyPartLEar
	case "neck":
		return BodyPartNeck
	case "rfinger":
		return BodyPartRFinger
	case "lfinger":
		return BodyPartLFinger
	case "head":
		return BodyPartHead
	case "rhand":
		return BodyPartRHand
	case "lhand":
		return BodyPartLHand
	case "gloves":
		return BodyPartGloves
	case "chest":
		return BodyPartChest
	case "legs":
		return BodyPartLegs
	case "feet":
		return BodyPartFeet
	case "back":
		return BodyPartBack
	case "lrhand":
		return BodyPartLRHand
	case "fullarmor":
		return BodyPartFullArmor
	case "hair":
		return BodyPartHair
	case "alldress":
		return BodyPartAllDress
	case "hair2":
		return BodyPartHair2
	case "hairall":
		return BodyPartHairAll
	default:
		return 0
	}
}

// ItemTemplate — шаблон предмета из items XML.
// Содержит базовые характеристики weapon/armor/consumable для создания конкретных Item.
//
// Phase 5.5: Weapon & Equipment System.
// Phase 19: Added CrystalType, BodyPartStr for equipment restrictions.
// Java reference: ItemTemplate.java
type ItemTemplate struct {
	ItemID int32  // Template ID (unique, from items XML)
	Name   string // Item name (e.g., "Short Sword", "Leather Shirt")
	Type   ItemType

	// Client type classification (sent in packets)
	Type1        int16 // TYPE1_* — client item category
	Type2        int16 // TYPE2_* — client sub-category
	BodyPartMask int32 // Bitmask for body part slot (sent to client)

	// Weapon stats
	PAtk         int32 // Physical attack bonus (added to base pAtk)
	AttackRange  int32 // Attack range in game units (40 for sword, 500 for bow)
	CritRate     int32 // Critical hit rate (swords=8, daggers=12, bows=12)
	RandomDamage int32 // Damage variance (swords=5, bows=10, fists=10)

	// Armor stats
	PDef     int32     // Physical defense bonus (added after slot subtraction)
	BodyPart ArmorSlot // Which body slot this armor occupies (chest, legs, etc.)

	// Equipment restrictions (Phase 19)
	CrystalType CrystalType // Item grade (NONE/D/C/B/A/S)
	BodyPartStr string      // Raw body part string from XML ("rhand","lrhand","chest", etc.)

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

// IsEquippable returns true if this item can be equipped (weapon or armor).
func (t *ItemTemplate) IsEquippable() bool {
	return t.Type == ItemTypeWeapon || t.Type == ItemTypeArmor
}

// IsTwoHanded returns true if this weapon occupies both hands (lrhand).
// Java: bodypart = "lrhand" (bow, pole, dual swords, etc.)
func (t *ItemTemplate) IsTwoHanded() bool {
	return t.BodyPartStr == "lrhand"
}
