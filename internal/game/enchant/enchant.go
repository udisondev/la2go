// Package enchant implements the Lineage 2 Interlude item enchantment system.
//
// Enchant flow:
//  1. Player uses an enchant scroll (UseItem) → sets activeEnchantItemID
//  2. Player drags target item → client sends RequestEnchantItem (0x58)
//  3. Server validates scroll/item compatibility, grade match, enchantability
//  4. Server calculates success rate and applies result
//
// Java reference: RequestEnchantItem.java, EnchantScroll.java, EnchantItemData.xml, EnchantItemGroups.xml
package enchant

import (
	"math/rand/v2"

	"github.com/udisondev/la2go/internal/model"
)

// ScrollType represents the enchant scroll category.
type ScrollType int

const (
	// ScrollNormal -- на провале предмет уничтожается (кристаллизуется).
	ScrollNormal ScrollType = iota
	// ScrollBlessed -- на провале enchant сбрасывается до 0.
	ScrollBlessed
	// ScrollCrystal -- всегда 100% успех (crystal/ancient scrolls).
	ScrollCrystal
)

// ScrollInfo describes an enchant scroll: target grade, target type, scroll type.
type ScrollInfo struct {
	IsWeapon   bool
	Grade      model.CrystalType
	ScrollType ScrollType
	MaxEnchant int32 // 0 = no limit (crystal scrolls); 16 default for normal/blessed
}

// Result describes the outcome of an enchant attempt.
type Result struct {
	// Success is true if enchant succeeded.
	Success bool
	// NewEnchant is the enchant level after the attempt.
	// On success: old+1. On blessed fail: 0. On normal fail: -1 (item destroyed).
	NewEnchant int32
	// Destroyed is true if the item was destroyed (normal scroll fail).
	Destroyed bool
}

// safeEnchantLevel is the max CURRENT enchant where success is guaranteed (100%).
// For regular armor/weapons: enchant 0, 1, 2 → 100% (i.e. +1, +2, +3 safe).
const safeEnchantLevel int32 = 3

// safeEnchantLevelFullArmor is the extended safe zone for full body armor.
// Full armor: enchant 0, 1, 2, 3 → 100% (i.e. +1, +2, +3, +4 safe).
const safeEnchantLevelFullArmor int32 = 4

// defaultMaxEnchant is the maximum enchant level allowed (enchant > 16 impossible).
const defaultMaxEnchant int32 = 16

// chancePercent is the success rate above safe level (66% for all types in Interlude).
// Java reference: EnchantItemGroups.xml — ARMOR_GROUP, FIGHTER_WEAPON_GROUP, MAGE_WEAPON_GROUP all use 66%.
const chancePercent = 66

// scrollTable maps scroll itemID to its properties.
// Data from EnchantItemData.xml (L2J Mobius CT 0 Interlude).
var scrollTable = map[int32]ScrollInfo{
	// Normal weapon scrolls (SCRL_ENCHANT_WP)
	729: {IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	947: {IsWeapon: true, Grade: model.CrystalB, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	951: {IsWeapon: true, Grade: model.CrystalC, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	955: {IsWeapon: true, Grade: model.CrystalD, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	959: {IsWeapon: true, Grade: model.CrystalS, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},

	// Normal armor scrolls (SCRL_ENCHANT_AM)
	730: {IsWeapon: false, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	948: {IsWeapon: false, Grade: model.CrystalB, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	952: {IsWeapon: false, Grade: model.CrystalC, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	956: {IsWeapon: false, Grade: model.CrystalD, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},
	960: {IsWeapon: false, Grade: model.CrystalS, ScrollType: ScrollNormal, MaxEnchant: defaultMaxEnchant},

	// Blessed weapon scrolls (BLESS_SCRL_ENCHANT_WP)
	6569: {IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6571: {IsWeapon: true, Grade: model.CrystalB, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6573: {IsWeapon: true, Grade: model.CrystalC, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6575: {IsWeapon: true, Grade: model.CrystalD, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6577: {IsWeapon: true, Grade: model.CrystalS, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},

	// Blessed armor scrolls (BLESS_SCRL_ENCHANT_AM)
	6570: {IsWeapon: false, Grade: model.CrystalA, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6572: {IsWeapon: false, Grade: model.CrystalB, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6574: {IsWeapon: false, Grade: model.CrystalC, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6576: {IsWeapon: false, Grade: model.CrystalD, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},
	6578: {IsWeapon: false, Grade: model.CrystalS, ScrollType: ScrollBlessed, MaxEnchant: defaultMaxEnchant},

	// Crystal (safe/ancient) weapon scrolls (ANCIENT_CRYSTAL_ENCHANT_WP)
	// Crystal scrolls have no maxEnchant limit and bonusRate=100 → always succeed.
	731: {IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollCrystal},
	949: {IsWeapon: true, Grade: model.CrystalB, ScrollType: ScrollCrystal},
	953: {IsWeapon: true, Grade: model.CrystalC, ScrollType: ScrollCrystal},
	957: {IsWeapon: true, Grade: model.CrystalD, ScrollType: ScrollCrystal},
	961: {IsWeapon: true, Grade: model.CrystalS, ScrollType: ScrollCrystal},

	// Crystal (safe/ancient) armor scrolls (ANCIENT_CRYSTAL_ENCHANT_AM)
	732: {IsWeapon: false, Grade: model.CrystalA, ScrollType: ScrollCrystal},
	950: {IsWeapon: false, Grade: model.CrystalB, ScrollType: ScrollCrystal},
	954: {IsWeapon: false, Grade: model.CrystalC, ScrollType: ScrollCrystal},
	958: {IsWeapon: false, Grade: model.CrystalD, ScrollType: ScrollCrystal},
	962: {IsWeapon: false, Grade: model.CrystalS, ScrollType: ScrollCrystal},
}

// IsScroll returns scroll info if the itemID is a known enchant scroll.
func IsScroll(itemID int32) (ScrollInfo, bool) {
	info, ok := scrollTable[itemID]
	return info, ok
}

// IsEnchantable returns true if the item can be enchanted.
// Only weapons and armor (not consumables, quest items, etc.) are enchantable.
func IsEnchantable(item *model.Item) bool {
	if item == nil {
		return false
	}
	tmpl := item.Template()
	if tmpl == nil {
		return false
	}
	return tmpl.IsWeapon() || tmpl.IsArmor()
}

// IsFullArmor returns true if the item occupies the "fullarmor" body part slot.
// Full armor has an extended safe enchant zone (+4 instead of +3).
func IsFullArmor(item *model.Item) bool {
	if item == nil {
		return false
	}
	tmpl := item.Template()
	if tmpl == nil {
		return false
	}
	return tmpl.BodyPartStr == "fullarmor"
}

// Validate checks if the scroll can be used on the given item.
// Returns a non-empty error string describing the problem, or empty string if valid.
func Validate(scroll ScrollInfo, item *model.Item) string {
	if !IsEnchantable(item) {
		return "item not enchantable"
	}

	tmpl := item.Template()

	// Проверка типа: weapon scroll → weapon item, armor scroll → armor item
	if scroll.IsWeapon && !tmpl.IsWeapon() {
		return "weapon scroll on non-weapon item"
	}
	if !scroll.IsWeapon && !tmpl.IsArmor() {
		return "armor scroll on non-armor item"
	}

	// Проверка грейда: скролл должен соответствовать грейду предмета
	if scroll.Grade != tmpl.CrystalType {
		return "scroll grade mismatch"
	}

	// Проверка максимального уровня
	maxEnchant := scroll.MaxEnchant
	if maxEnchant == 0 {
		maxEnchant = defaultMaxEnchant
	}
	if item.Enchant() >= maxEnchant {
		return "max enchant level reached"
	}

	return ""
}

// SuccessChance returns the enchant success chance (0-100) for the given item at its current enchant level.
// Crystal scrolls always return 100.
func SuccessChance(scroll ScrollInfo, item *model.Item) int {
	if scroll.ScrollType == ScrollCrystal {
		return 100
	}

	currentEnchant := item.Enchant()

	// Определяем safe level на основе типа предмета
	safe := safeEnchantLevel
	if IsFullArmor(item) {
		safe = safeEnchantLevelFullArmor
	}

	if currentEnchant < safe {
		return 100
	}

	return chancePercent
}

// TryEnchant performs the enchant attempt.
// Rolls RNG, returns the result. Does NOT modify item state.
// Caller is responsible for consuming the scroll and applying the result.
func TryEnchant(scroll ScrollInfo, item *model.Item) Result {
	chance := SuccessChance(scroll, item)
	currentEnchant := item.Enchant()

	if chance >= 100 || rand.IntN(100) < chance {
		return Result{
			Success:    true,
			NewEnchant: currentEnchant + 1,
		}
	}

	// Провал
	switch scroll.ScrollType {
	case ScrollCrystal:
		// Crystal scroll: предмет не меняется (safe fail).
		// Этого не должно произойти т.к. chance=100, но на всякий случай.
		return Result{
			Success:    false,
			NewEnchant: currentEnchant,
		}
	case ScrollBlessed:
		// Blessed scroll: enchant сбрасывается до 0.
		return Result{
			Success:    false,
			NewEnchant: 0,
		}
	default:
		// Normal scroll: предмет уничтожается.
		return Result{
			Success:   false,
			Destroyed: true,
		}
	}
}

// TryEnchantWithRoll performs the enchant attempt with an explicit roll value (0-99).
// Used for deterministic testing.
func TryEnchantWithRoll(scroll ScrollInfo, item *model.Item, roll int) Result {
	chance := SuccessChance(scroll, item)
	currentEnchant := item.Enchant()

	if chance >= 100 || roll < chance {
		return Result{
			Success:    true,
			NewEnchant: currentEnchant + 1,
		}
	}

	switch scroll.ScrollType {
	case ScrollCrystal:
		return Result{
			Success:    false,
			NewEnchant: currentEnchant,
		}
	case ScrollBlessed:
		return Result{
			Success:    false,
			NewEnchant: 0,
		}
	default:
		return Result{
			Success:   false,
			Destroyed: true,
		}
	}
}
