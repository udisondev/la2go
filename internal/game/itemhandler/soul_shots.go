package itemhandler

import (
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// SoulShot skill IDs (Java reference: SoulShots.java)
const (
	SkillSoulShotNone = 2039
	SkillSoulShotD    = 2150
	SkillSoulShotC    = 2151
	SkillSoulShotB    = 2152
	SkillSoulShotA    = 2153
	SkillSoulShotS    = 2154
)

// soulShotsHandler processes Soul Shot items.
// Charges the equipped weapon with soul shots for enhanced physical attacks.
//
// Java reference: SoulShots.java
type soulShotsHandler struct{}

func (h *soulShotsHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	weapon := player.GetEquippedWeapon()
	if weapon == nil {
		slog.Debug("SoulShots: no weapon equipped",
			"player", player.Name())
		return nil
	}

	weaponDef := data.GetItemDef(weapon.ItemID())
	if weaponDef == nil {
		return nil
	}

	// Check if weapon already charged with soul shot
	if weapon.IsChargedSoulShot() {
		return nil
	}

	// Soul shot count = weapon.soulshots
	shotCount := int64(weaponDef.SoulShots())
	if shotCount == 0 {
		shotCount = 1
	}

	// Check grade match: shot grade must match weapon grade
	shotDef := data.GetItemDef(item.ItemID())
	if shotDef == nil {
		return nil
	}
	if !gradeMatch(shotDef.CrystalType(), weaponDef.CrystalType()) {
		slog.Debug("SoulShots: grade mismatch",
			"player", player.Name(),
			"shotGrade", shotDef.CrystalType(),
			"weaponGrade", weaponDef.CrystalType())
		return nil
	}

	// Check enough items
	inv := player.Inventory()
	available := inv.CountItemsByID(item.ItemID())
	if available < shotCount {
		slog.Debug("SoulShots: not enough shots",
			"player", player.Name(),
			"need", shotCount,
			"have", available)
		return nil
	}

	// Charge weapon
	weapon.SetChargedSoulShot(true)

	// Determine animation skill ID based on weapon grade
	skillID := soulShotSkillForGrade(weaponDef.CrystalType())

	return &UseResult{
		ConsumeCount: shotCount,
		SkillID:      skillID,
		SkillLevel:   1,
		Broadcast:    true,
	}
}

func soulShotSkillForGrade(grade string) int32 {
	switch grade {
	case "D":
		return SkillSoulShotD
	case "C":
		return SkillSoulShotC
	case "B":
		return SkillSoulShotB
	case "A":
		return SkillSoulShotA
	case "S":
		return SkillSoulShotS
	default:
		return SkillSoulShotNone
	}
}

// spiritShotHandler processes Spirit Shot items.
// Charges the equipped weapon for enhanced magical attacks.
//
// Java reference: SpiritShot.java
type spiritShotHandler struct{}

// Spirit Shot skill IDs
const (
	SkillSpiritShotNone = 2061
	SkillSpiritShotD    = 2155
	SkillSpiritShotC    = 2156
	SkillSpiritShotB    = 2157
	SkillSpiritShotA    = 2158
	SkillSpiritShotS    = 2159
)

func (h *spiritShotHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	weapon := player.GetEquippedWeapon()
	if weapon == nil {
		return nil
	}

	weaponDef := data.GetItemDef(weapon.ItemID())
	if weaponDef == nil {
		return nil
	}

	if weapon.IsChargedSpiritShot() {
		return nil
	}

	shotCount := int64(weaponDef.SpiritShots())
	if shotCount == 0 {
		shotCount = 1
	}

	shotDef := data.GetItemDef(item.ItemID())
	if shotDef == nil {
		return nil
	}
	if !gradeMatch(shotDef.CrystalType(), weaponDef.CrystalType()) {
		return nil
	}

	inv := player.Inventory()
	available := inv.CountItemsByID(item.ItemID())
	if available < shotCount {
		return nil
	}

	weapon.SetChargedSpiritShot(true)
	skillID := spiritShotSkillForGrade(weaponDef.CrystalType())

	return &UseResult{
		ConsumeCount: shotCount,
		SkillID:      skillID,
		SkillLevel:   1,
		Broadcast:    true,
	}
}

func spiritShotSkillForGrade(grade string) int32 {
	switch grade {
	case "D":
		return SkillSpiritShotD
	case "C":
		return SkillSpiritShotC
	case "B":
		return SkillSpiritShotB
	case "A":
		return SkillSpiritShotA
	case "S":
		return SkillSpiritShotS
	default:
		return SkillSpiritShotNone
	}
}

// blessedSpiritShotHandler processes Blessed Spirit Shot items.
// Same as SpiritShot but blessed (stronger magic charge).
//
// Java reference: BlessedSpiritShot.java
type blessedSpiritShotHandler struct{}

// Blessed Spirit Shot skill IDs
const (
	SkillBlessedSpiritShotNone = 2061 // same base as spirit shot
	SkillBlessedSpiritShotD    = 2160
	SkillBlessedSpiritShotC    = 2161
	SkillBlessedSpiritShotB    = 2162
	SkillBlessedSpiritShotA    = 2163
	SkillBlessedSpiritShotS    = 2164
)

func (h *blessedSpiritShotHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	weapon := player.GetEquippedWeapon()
	if weapon == nil {
		return nil
	}

	weaponDef := data.GetItemDef(weapon.ItemID())
	if weaponDef == nil {
		return nil
	}

	if weapon.IsChargedBlessedSpiritShot() {
		return nil
	}

	shotCount := int64(weaponDef.SpiritShots())
	if shotCount == 0 {
		shotCount = 1
	}

	shotDef := data.GetItemDef(item.ItemID())
	if shotDef == nil {
		return nil
	}
	if !gradeMatch(shotDef.CrystalType(), weaponDef.CrystalType()) {
		return nil
	}

	inv := player.Inventory()
	available := inv.CountItemsByID(item.ItemID())
	if available < shotCount {
		return nil
	}

	weapon.SetChargedBlessedSpiritShot(true)
	skillID := blessedSpiritShotSkillForGrade(weaponDef.CrystalType())

	return &UseResult{
		ConsumeCount: shotCount,
		SkillID:      skillID,
		SkillLevel:   1,
		Broadcast:    true,
	}
}

func blessedSpiritShotSkillForGrade(grade string) int32 {
	switch grade {
	case "D":
		return SkillBlessedSpiritShotD
	case "C":
		return SkillBlessedSpiritShotC
	case "B":
		return SkillBlessedSpiritShotB
	case "A":
		return SkillBlessedSpiritShotA
	case "S":
		return SkillBlessedSpiritShotS
	default:
		return SkillBlessedSpiritShotNone
	}
}

// fishShotsHandler processes Fish Shot items for fishing.
//
// Java reference: FishShots.java
type fishShotsHandler struct{}

func (h *fishShotsHandler) UseItem(player *model.Player, item *model.Item, skillID, skillLevel int32) *UseResult {
	// Fish shots work like spirit shots but for fishing rods
	weapon := player.GetEquippedWeapon()
	if weapon == nil {
		return nil
	}

	weaponDef := data.GetItemDef(weapon.ItemID())
	if weaponDef == nil || weaponDef.WeaponType() != "FISHINGROD" {
		return nil
	}

	if weapon.IsChargedSpiritShot() {
		return nil
	}

	weapon.SetChargedSpiritShot(true)

	return &UseResult{
		ConsumeCount: 1,
		SkillID:      skillID,
		SkillLevel:   skillLevel,
		Broadcast:    true,
	}
}

// gradeMatch checks if shot grade matches weapon grade.
// No-grade shots can be used with any weapon. Otherwise grades must match.
func gradeMatch(shotGrade, weaponGrade string) bool {
	if shotGrade == "" || shotGrade == "NONE" {
		return true
	}
	return shotGrade == weaponGrade
}
