package data

import (
	"log/slog"
	"strconv"
	"strings"
)

// SkillTable — глобальный registry всех skill templates.
// map[skillID]map[level]*SkillTemplate
// Загружается через LoadSkills() при старте сервера.
var SkillTable map[int32]map[int32]*SkillTemplate

// skillMaxLevel — precomputed max level per skill ID.
// O(1) lookup instead of iterating map keys.
var skillMaxLevel map[int32]int32

// GetSkillTemplate возвращает SkillTemplate по ID и Level.
// Returns nil если скилл не найден.
func GetSkillTemplate(skillID, level int32) *SkillTemplate {
	if SkillTable == nil {
		return nil
	}
	levels, ok := SkillTable[skillID]
	if !ok {
		return nil
	}
	return levels[level]
}

// GetSkillMaxLevel возвращает максимальный уровень скилла.
// O(1) lookup via precomputed map (populated during LoadSkills).
// Returns 0 если скилл не найден.
func GetSkillMaxLevel(skillID int32) int32 {
	if skillMaxLevel == nil {
		return 0
	}
	return skillMaxLevel[skillID]
}

// LoadSkills строит SkillTable из Go-литералов (skillDefs).
// Вызывается при старте сервера.
func LoadSkills() error {
	SkillTable = make(map[int32]map[int32]*SkillTemplate)

	for i := range skillDefs {
		buildSkillTemplates(&skillDefs[i])
	}

	// Precompute max levels for O(1) GetSkillMaxLevel()
	skillMaxLevel = make(map[int32]int32, len(SkillTable))
	var totalSkills int
	for skillID, levels := range SkillTable {
		totalSkills += len(levels)
		var maxLvl int32
		for lvl := range levels {
			if lvl > maxLvl {
				maxLvl = lvl
			}
		}
		skillMaxLevel[skillID] = maxLvl
	}

	slog.Info("loaded skills", "skill_ids", len(SkillTable), "total_entries", totalSkills)
	return nil
}

// buildSkillTemplates создаёт SkillTemplate для каждого уровня скилла из определения.
func buildSkillTemplates(def *skillDef) {
	levels := max(def.levels, 1)

	if SkillTable[def.id] == nil {
		SkillTable[def.id] = make(map[int32]*SkillTemplate, int(levels))
	}

	for levelIdx := range levels {
		level := levelIdx + 1

		template := &SkillTemplate{
			ID:            def.id,
			Level:         level,
			Name:          def.name,
			OperateType:   ParseOperateType(def.operateType),
			TargetType:    ParseTargetType(def.targetType),
			IsMagic:       def.isMagic,
			Power:         perLevelFloat(def.power, int(levelIdx)),
			MpConsume:     perLevelInt32(def.mpConsume, int(levelIdx)),
			HpConsume:     perLevelInt32(def.hpConsume, int(levelIdx)),
			HitTime:       def.hitTime,
			CoolTime:      def.coolTime,
			ReuseDelay:    def.reuseDelay,
			CastRange:     def.castRange,
			EffectRange:   def.effectRange,
			MagicLevel:    resolveInt32OrArray(def.magicLevel, def.magicLevelByLvl, int(levelIdx)),
			AbnormalType:  def.abnormalType,
			AbnormalLevel: resolveInt32OrArray(def.abnormalLevel, def.abnormalLevelTbl, int(levelIdx)),
			AbnormalTime:  perLevelInt32(def.abnormalTime, int(levelIdx)),
			Effects:       buildEffects(def.effects, int(levelIdx)),

			// Extended attributes
			IsDebuff:         def.isDebuff,
			OverHit:          def.overHit,
			IgnoreShield:     def.ignoreShld,
			NextActionAttack: def.nextActionAttack,
			Trait:            def.trait,
			BasicProperty:    def.basicProperty,
			Element:          def.element,
			ElementPower:     def.elementPower,
			AffectRange:      def.affectRange,
			ActivateRate:     def.activateRate,
			LvlBonusRate:     def.lvlBonusRate,
			BlowChance:       def.blowChance,
			ChargeConsume:    def.chargeConsume,
			MpInitialConsume: perLevelInt32(def.mpInitialConsume, int(levelIdx)),
			ItemConsumeId:    def.itemConsumeId,
			ItemConsumeCount: def.itemConsumeCount,
			EffectPoint:      perLevelInt32(def.effectPoint, int(levelIdx)),
			BaseCritRate:     def.baseCritRate,
			EnchantGroup1:    def.enchantGroup1,
			EnchantGroup2:    def.enchantGroup2,
		}

		SkillTable[def.id][level] = template
	}

	// Build enchant levels (101-130 for route 1, 141-170 for route 2)
	buildEnchantLevels(def, 1, def.enchant1, def.enchant1Effects, 101)
	buildEnchantLevels(def, 2, def.enchant2, def.enchant2Effects, 141)
}

// buildEnchantLevels creates SkillTemplate entries for enchant levels.
func buildEnchantLevels(def *skillDef, _ int32, overrides []enchantOverride, enchEffects []effectDef, startLevel int32) {
	if len(overrides) == 0 {
		return
	}

	// Determine number of enchant levels from override values
	enchLevels := int32(0)
	for _, o := range overrides {
		if int32(len(o.values)) > enchLevels {
			enchLevels = int32(len(o.values))
		}
	}
	if enchLevels == 0 {
		return
	}

	// Base template is the max-level version
	maxLevel := max(def.levels, 1)
	baseIdx := int(maxLevel - 1)

	for enchIdx := range enchLevels {
		level := startLevel + enchIdx

		template := &SkillTemplate{
			ID:            def.id,
			Level:         level,
			Name:          def.name,
			OperateType:   ParseOperateType(def.operateType),
			TargetType:    ParseTargetType(def.targetType),
			IsMagic:       def.isMagic,
			Power:         perLevelFloat(def.power, baseIdx),
			MpConsume:     perLevelInt32(def.mpConsume, baseIdx),
			HpConsume:     perLevelInt32(def.hpConsume, baseIdx),
			HitTime:       def.hitTime,
			CoolTime:      def.coolTime,
			ReuseDelay:    def.reuseDelay,
			CastRange:     def.castRange,
			EffectRange:   def.effectRange,
			MagicLevel:    resolveInt32OrArray(def.magicLevel, def.magicLevelByLvl, baseIdx),
			AbnormalType:  def.abnormalType,
			AbnormalLevel: resolveInt32OrArray(def.abnormalLevel, def.abnormalLevelTbl, baseIdx),
			AbnormalTime:  perLevelInt32(def.abnormalTime, baseIdx),

			IsDebuff:         def.isDebuff,
			OverHit:          def.overHit,
			IgnoreShield:     def.ignoreShld,
			NextActionAttack: def.nextActionAttack,
			Trait:            def.trait,
			BasicProperty:    def.basicProperty,
			Element:          def.element,
			ElementPower:     def.elementPower,
			AffectRange:      def.affectRange,
			ActivateRate:     def.activateRate,
			LvlBonusRate:     def.lvlBonusRate,
			BlowChance:       def.blowChance,
			ChargeConsume:    def.chargeConsume,
			MpInitialConsume: perLevelInt32(def.mpInitialConsume, baseIdx),
			ItemConsumeId:    def.itemConsumeId,
			ItemConsumeCount: def.itemConsumeCount,
			EffectPoint:      perLevelInt32(def.effectPoint, baseIdx),
			BaseCritRate:     def.baseCritRate,
			EnchantGroup1:    def.enchantGroup1,
			EnchantGroup2:    def.enchantGroup2,
		}

		// Apply enchant overrides
		for _, o := range overrides {
			val := perLevelStr(o.values, int(enchIdx))
			applyEnchantOverride(template, o.attr, val)
		}

		// Use enchant-specific effects if provided, otherwise base effects
		if len(enchEffects) > 0 {
			template.Effects = buildEffects(enchEffects, int(enchIdx))
		} else {
			template.Effects = buildEffects(def.effects, baseIdx)
		}

		SkillTable[def.id][level] = template
	}
}

// applyEnchantOverride applies a single enchant attribute override to a template.
func applyEnchantOverride(t *SkillTemplate, attr, val string) {
	switch attr {
	case "power":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			t.Power = f
		}
	case "mpConsume":
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			t.MpConsume = int32(n)
		}
	case "magicLevel", "magicLvl":
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			t.MagicLevel = int32(n)
		}
	case "activateRate":
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			t.ActivateRate = int32(n)
		}
	case "effectPoint":
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			t.EffectPoint = int32(n)
		}
	case "abnormalTime":
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			t.AbnormalTime = int32(n)
		}
	}
}

// resolveInt32OrArray picks scalar or per-level value.
func resolveInt32OrArray(scalar int32, array []int32, levelIdx int) int32 {
	if len(array) > 0 {
		return perLevelInt32(array, levelIdx)
	}
	return scalar
}

// buildEffects разрешает per-level параметры эффектов для конкретного уровня.
func buildEffects(defs []effectDef, levelIdx int) []EffectDef {
	result := make([]EffectDef, len(defs))
	for i, def := range defs {
		params := make(map[string]string, len(def.params)+len(def.perLvl))
		for k, v := range def.params {
			params[k] = v
		}
		for k, vals := range def.perLvl {
			params[k] = perLevelStr(vals, levelIdx)
		}

		// Build stat mods
		var statMods []StatMod
		for _, sm := range def.statMods {
			val := resolveStatModVal(sm.val, levelIdx)
			statMods = append(statMods, StatMod{
				Op:   sm.op,
				Stat: sm.stat,
				Val:  val,
			})
		}

		result[i] = EffectDef{
			Name:     def.name,
			Params:   params,
			StatMods: statMods,
		}
	}
	return result
}

// resolveStatModVal resolves stat mod value (may be space-separated per-level).
func resolveStatModVal(val string, levelIdx int) float64 {
	parts := strings.Fields(val)
	if len(parts) == 0 {
		return 0
	}
	idx := min(levelIdx, len(parts)-1)
	f, _ := strconv.ParseFloat(parts[idx], 64)
	return f
}

// perLevelFloat возвращает значение для уровня из per-level массива.
// Если массив короче — берётся последний элемент.
func perLevelFloat(vals []float64, levelIdx int) float64 {
	if len(vals) == 0 {
		return 0
	}
	if levelIdx < len(vals) {
		return vals[levelIdx]
	}
	return vals[len(vals)-1]
}

// perLevelInt32 возвращает значение для уровня из per-level массива.
func perLevelInt32(vals []int32, levelIdx int) int32 {
	if len(vals) == 0 {
		return 0
	}
	if levelIdx < len(vals) {
		return vals[levelIdx]
	}
	return vals[len(vals)-1]
}

// perLevelStr возвращает значение для уровня из per-level массива строк.
func perLevelStr(vals []string, levelIdx int) string {
	if len(vals) == 0 {
		return ""
	}
	if levelIdx < len(vals) {
		return vals[levelIdx]
	}
	return vals[len(vals)-1]
}
