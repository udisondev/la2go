package data

import (
	"log/slog"
)

// LoadPlayerTemplates строит PlayerTemplates из Go-литералов (playerTemplateDefs).
// Вызывается при старте сервера (cmd/gameserver/main.go).
func LoadPlayerTemplates() error {
	PlayerTemplates = make(map[uint8]*PlayerTemplate, len(playerTemplateDefs))

	for i := range playerTemplateDefs {
		def := &playerTemplateDefs[i]
		template := convertTemplateDef(def)
		PlayerTemplates[template.ClassID] = template
	}

	slog.Info("loaded player templates", "count", len(PlayerTemplates))
	return nil
}

// convertTemplateDef конвертирует playerTemplateDef → PlayerTemplate.
func convertTemplateDef(def *playerTemplateDef) *PlayerTemplate {
	pDefSum := def.pDefChest + def.pDefLegs + def.pDefHead + def.pDefFeet +
		def.pDefGloves + def.pDefUnderwear + def.pDefCloak
	mDefSum := def.mDefRear + def.mDefLear + def.mDefRFinger + def.mDefLFinger + def.mDefNeck

	t := &PlayerTemplate{
		ClassID: def.classID,

		BaseSTR: def.baseSTR,
		BaseCON: def.baseCON,
		BaseDEX: def.baseDEX,
		BaseINT: def.baseINT,
		BaseWIT: def.baseWIT,
		BaseMEN: def.baseMEN,

		BasePAtk:     def.basePAtk,
		BaseMAtk:     def.baseMAtk,
		BasePDef:     pDefSum,
		BaseMDef:     mDefSum,
		BasePAtkSpd:  def.basePAtkSpd,
		BaseMAtkSpd:  def.baseMAtkSpd,
		BaseCritRate: def.baseCritRate,
		BaseAtkRange: def.baseAtkRange,
		RandomDamage: def.randomDamage,

		CollisionRadiusMale:   def.collisionRadiusMale,
		CollisionHeightMale:   def.collisionHeightMale,
		CollisionRadiusFemale: def.collisionRadiusFemale,
		CollisionHeightFemale: def.collisionHeightFemale,

		HPByLevel:      def.hp,
		MPByLevel:      def.mp,
		CPByLevel:      def.cp,
		HPRegenByLevel: def.hpRegen,
		MPRegenByLevel: def.mpRegen,
		CPRegenByLevel: def.cpRegen,

		SlotDef:        make(map[uint8]int32),
		CreationPoints: def.creationPoints,
	}

	// PDef slots
	t.SlotDef[SlotChest] = def.pDefChest
	t.SlotDef[SlotLegs] = def.pDefLegs
	t.SlotDef[SlotHead] = def.pDefHead
	t.SlotDef[SlotFeet] = def.pDefFeet
	t.SlotDef[SlotGloves] = def.pDefGloves
	t.SlotDef[SlotUnderwear] = def.pDefUnderwear
	t.SlotDef[SlotCloak] = def.pDefCloak

	// MDef slots
	t.SlotDef[SlotRightEar] = def.mDefRear
	t.SlotDef[SlotLeftEar] = def.mDefLear
	t.SlotDef[SlotRightFinger] = def.mDefRFinger
	t.SlotDef[SlotLeftFinger] = def.mDefLFinger
	t.SlotDef[SlotNeck] = def.mDefNeck

	return t
}
