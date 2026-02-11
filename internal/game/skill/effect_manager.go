package skill

import (
	"log/slog"
	"sync"
)

const (
	maxBuffs   = 24
	maxDebuffs = 8
)

// EffectManager tracks active buffs/debuffs per character.
// Implements model.StatBonusProvider interface.
//
// Thread-safe: all methods are protected by sync.RWMutex.
//
// Phase 5.9.3: Effect Framework.
// Java reference: EffectList.java
type EffectManager struct {
	mu       sync.RWMutex
	buffs    []*ActiveEffect
	debuffs  []*ActiveEffect
	passives []*ActiveEffect

	// Stat modifiers from all active effects
	modifiers []StatModifier
}

// NewEffectManager creates a new empty EffectManager.
func NewEffectManager() *EffectManager {
	return &EffectManager{
		buffs:     make([]*ActiveEffect, 0, maxBuffs),
		debuffs:   make([]*ActiveEffect, 0, maxDebuffs),
		passives:  make([]*ActiveEffect, 0, 8),
		modifiers: make([]StatModifier, 0, 16),
	}
}

// AddBuff adds a buff effect with stacking check.
// Returns true if the effect was added/replaced, false if rejected.
//
// Stacking rules (same AbnormalType):
//   - Higher AbnormalLevel → replaces existing
//   - Same AbnormalLevel → refreshes duration
//   - Lower AbnormalLevel → rejected
//
// If buff limit (24) is reached, oldest buff is removed.
func (m *EffectManager) AddBuff(ae *ActiveEffect) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check stacking
	if ae.AbnormalType != "" {
		for i, existing := range m.buffs {
			if existing.AbnormalType == ae.AbnormalType {
				if ae.AbnormalLevel > existing.AbnormalLevel {
					// Higher level replaces
					existing.Effect.OnExit(existing.CasterObjID, existing.TargetObjID)
					m.buffs[i] = ae
					ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
					m.rebuildModifiers()
					return true
				}
				if ae.AbnormalLevel == existing.AbnormalLevel {
					// Same level refreshes
					existing.RemainingMs = ae.RemainingMs
					return true
				}
				// Lower level rejected
				return false
			}
		}
	}

	// Buff limit check
	if len(m.buffs) >= maxBuffs {
		// Remove oldest buff
		oldest := m.buffs[0]
		oldest.Effect.OnExit(oldest.CasterObjID, oldest.TargetObjID)
		m.buffs = m.buffs[1:]

		slog.Debug("buff limit reached, removed oldest",
			"removedSkill", oldest.SkillID,
			"target", ae.TargetObjID)
	}

	m.buffs = append(m.buffs, ae)
	ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
	m.rebuildModifiers()
	return true
}

// AddDebuff adds a debuff effect with stacking check.
// Returns true if the effect was added/replaced, false if rejected.
// Same stacking rules as AddBuff, with 8 debuff limit.
func (m *EffectManager) AddDebuff(ae *ActiveEffect) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check stacking
	if ae.AbnormalType != "" {
		for i, existing := range m.debuffs {
			if existing.AbnormalType == ae.AbnormalType {
				if ae.AbnormalLevel > existing.AbnormalLevel {
					existing.Effect.OnExit(existing.CasterObjID, existing.TargetObjID)
					m.debuffs[i] = ae
					ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
					m.rebuildModifiers()
					return true
				}
				if ae.AbnormalLevel == existing.AbnormalLevel {
					existing.RemainingMs = ae.RemainingMs
					return true
				}
				return false
			}
		}
	}

	if len(m.debuffs) >= maxDebuffs {
		oldest := m.debuffs[0]
		oldest.Effect.OnExit(oldest.CasterObjID, oldest.TargetObjID)
		m.debuffs = m.debuffs[1:]
	}

	m.debuffs = append(m.debuffs, ae)
	ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
	m.rebuildModifiers()
	return true
}

// AddPassive adds a passive effect (no stacking or limit checks).
func (m *EffectManager) AddPassive(ae *ActiveEffect) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Replace existing passive with same skill ID
	for i, existing := range m.passives {
		if existing.SkillID == ae.SkillID {
			existing.Effect.OnExit(existing.CasterObjID, existing.TargetObjID)
			m.passives[i] = ae
			ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
			m.rebuildModifiers()
			return
		}
	}

	m.passives = append(m.passives, ae)
	ae.Effect.OnStart(ae.CasterObjID, ae.TargetObjID)
	m.rebuildModifiers()
}

// RemoveEffect removes all effects with the given AbnormalType.
// Calls OnExit for each removed effect.
func (m *EffectManager) RemoveEffect(abnormalType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buffs = removeByAbnormalType(m.buffs, abnormalType)
	m.debuffs = removeByAbnormalType(m.debuffs, abnormalType)
	m.rebuildModifiers()
}

// RemoveBySkillID removes effect with specific skill ID.
func (m *EffectManager) RemoveBySkillID(skillID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buffs = removeBySkillID(m.buffs, skillID)
	m.debuffs = removeBySkillID(m.debuffs, skillID)
	m.passives = removeBySkillID(m.passives, skillID)
	m.rebuildModifiers()
}

// GetStatBonus returns the total stat bonus from all active effects.
// Implements model.StatBonusProvider interface.
//
// Additive bonuses are summed first, then multiplicative bonuses are applied.
// Returns 0.0 if no modifiers affect the stat.
func (m *EffectManager) GetStatBonus(stat string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	addBonus := 0.0
	mulBonus := 1.0
	hasMul := false

	for _, mod := range m.modifiers {
		if mod.Stat == stat {
			switch mod.Type {
			case StatModAdd:
				addBonus += mod.Value
			case StatModMul:
				mulBonus *= mod.Value
				hasMul = true
			}
		}
	}

	if hasMul {
		return addBonus * mulBonus
	}
	return addBonus
}

// Tick decrements timers on all active effects by deltaMs.
// Removes expired effects and calls OnExit for them.
func (m *EffectManager) Tick(deltaMs int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	changed := false
	m.buffs, changed = tickEffects(m.buffs, deltaMs)
	debuffChanged := false
	m.debuffs, debuffChanged = tickEffects(m.debuffs, deltaMs)

	if changed || debuffChanged {
		m.rebuildModifiers()
	}
}

// ActiveBuffs returns a copy of active buff effects.
func (m *EffectManager) ActiveBuffs() []*ActiveEffect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ActiveEffect, len(m.buffs))
	copy(result, m.buffs)
	return result
}

// ActiveDebuffs returns a copy of active debuff effects.
func (m *EffectManager) ActiveDebuffs() []*ActiveEffect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ActiveEffect, len(m.debuffs))
	copy(result, m.debuffs)
	return result
}

// BuffCount returns current number of active buffs.
func (m *EffectManager) BuffCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.buffs)
}

// DebuffCount returns current number of active debuffs.
func (m *EffectManager) DebuffCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.debuffs)
}

// rebuildModifiers recalculates stat modifiers from all active effects.
// Must be called with mu held.
func (m *EffectManager) rebuildModifiers() {
	m.modifiers = m.modifiers[:0]

	collectModifiers := func(effects []*ActiveEffect) {
		for _, ae := range effects {
			if provider, ok := ae.Effect.(StatModifierProvider); ok {
				m.modifiers = append(m.modifiers, provider.StatModifiers()...)
			}
		}
	}

	collectModifiers(m.buffs)
	collectModifiers(m.debuffs)
	collectModifiers(m.passives)
}

// StatModifierProvider is optionally implemented by Effect types that modify stats.
type StatModifierProvider interface {
	StatModifiers() []StatModifier
}

// removeByAbnormalType removes effects with matching AbnormalType, calling OnExit.
func removeByAbnormalType(effects []*ActiveEffect, abnormalType string) []*ActiveEffect {
	n := 0
	for _, ae := range effects {
		if ae.AbnormalType == abnormalType {
			ae.Effect.OnExit(ae.CasterObjID, ae.TargetObjID)
		} else {
			effects[n] = ae
			n++
		}
	}
	return effects[:n]
}

// removeBySkillID removes effects with matching SkillID, calling OnExit.
func removeBySkillID(effects []*ActiveEffect, skillID int32) []*ActiveEffect {
	n := 0
	for _, ae := range effects {
		if ae.SkillID == skillID {
			ae.Effect.OnExit(ae.CasterObjID, ae.TargetObjID)
		} else {
			effects[n] = ae
			n++
		}
	}
	return effects[:n]
}

// tickEffects decrements timers and removes expired effects.
// Returns the updated slice and whether any effects were removed.
func tickEffects(effects []*ActiveEffect, deltaMs int32) ([]*ActiveEffect, bool) {
	changed := false
	n := 0
	for _, ae := range effects {
		if !ae.Tick(deltaMs) {
			ae.Effect.OnExit(ae.CasterObjID, ae.TargetObjID)
			changed = true
		} else {
			effects[n] = ae
			n++
		}
	}
	return effects[:n], changed
}
