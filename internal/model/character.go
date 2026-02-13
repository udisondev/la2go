package model

import (
	"sync"
	"sync/atomic"
)

// Character — базовый класс для живых существ (Player, NPC).
// Добавляет HP, MP, CP, level к WorldObject.
type Character struct {
	*WorldObject // embedded

	level     int32
	currentHP int32
	maxHP     int32
	currentMP int32
	maxMP     int32
	currentCP int32
	maxCP     int32

	deathOnce sync.Once // protects DoDie from double execution (race condition)

	// Cast state (Phase 5.9.4: Cast Flow).
	// Atomic flag indicates if character is currently casting a skill.
	// Used for condition checks (can't attack/move during cast).
	isCasting atomic.Bool

	// CC (crowd control) flags — atomic for lock-free concurrent access.
	// Phase 5.9+: Used by skill effects (Stun, Root, Sleep, Paralyze, Fear).
	// Java reference: Creature.java — startStunning/stopStunning etc.
	stunned   atomic.Bool
	rooted    atomic.Bool
	sleeping  atomic.Bool
	paralyzed atomic.Bool
	feared    atomic.Bool

	// Zone flags bitfield (Phase 11: ZoneId enum).
	// Each bit corresponds to a ZoneID constant (0..21).
	// Atomic for lock-free concurrent reads/writes from zone manager.
	// Java reference: Creature.java — ConcurrentHashMap<ZoneId> replaced with bitfield.
	zones atomic.Uint32
}

// NewCharacter создаёт нового персонажа с указанными максимальными значениями.
// Текущие HP/MP/CP устанавливаются равными максимальным.
func NewCharacter(objectID uint32, name string, loc Location, level, maxHP, maxMP, maxCP int32) *Character {
	return &Character{
		WorldObject: NewWorldObject(objectID, name, loc),
		level:       level,
		currentHP:   maxHP,
		maxHP:       maxHP,
		currentMP:   maxMP,
		maxMP:       maxMP,
		currentCP:   maxCP,
		maxCP:       maxCP,
	}
}

// CurrentHP возвращает текущее HP.
func (c *Character) CurrentHP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentHP
}

// MaxHP возвращает максимальное HP.
func (c *Character) MaxHP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxHP
}

// SetCurrentHP устанавливает текущее HP с валидацией (clamp 0..maxHP).
func (c *Character) SetCurrentHP(hp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hp < 0 {
		hp = 0
	}
	if hp > c.maxHP {
		hp = c.maxHP
	}
	c.currentHP = hp
}

// SetMaxHP устанавливает максимальное HP и корректирует текущее если нужно.
func (c *Character) SetMaxHP(maxHP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxHP < 1 {
		maxHP = 1
	}

	c.maxHP = maxHP

	// Если текущее HP больше нового максимума — обрезаем
	if c.currentHP > c.maxHP {
		c.currentHP = c.maxHP
	}
}

// CurrentMP возвращает текущее MP.
func (c *Character) CurrentMP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMP
}

// MaxMP возвращает максимальное MP.
func (c *Character) MaxMP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxMP
}

// SetCurrentMP устанавливает текущее MP с валидацией (clamp 0..maxMP).
func (c *Character) SetCurrentMP(mp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if mp < 0 {
		mp = 0
	}
	if mp > c.maxMP {
		mp = c.maxMP
	}
	c.currentMP = mp
}

// SetMaxMP устанавливает максимальное MP и корректирует текущее если нужно.
func (c *Character) SetMaxMP(maxMP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxMP < 1 {
		maxMP = 1
	}

	c.maxMP = maxMP

	// Если текущее MP больше нового максимума — обрезаем
	if c.currentMP > c.maxMP {
		c.currentMP = c.maxMP
	}
}

// CurrentCP возвращает текущее CP.
func (c *Character) CurrentCP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentCP
}

// MaxCP возвращает максимальное CP.
func (c *Character) MaxCP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxCP
}

// SetCurrentCP устанавливает текущее CP с валидацией (clamp 0..maxCP).
func (c *Character) SetCurrentCP(cp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cp < 0 {
		cp = 0
	}
	if cp > c.maxCP {
		cp = c.maxCP
	}
	c.currentCP = cp
}

// SetMaxCP устанавливает максимальное CP и корректирует текущее если нужно.
func (c *Character) SetMaxCP(maxCP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxCP < 1 {
		maxCP = 1
	}

	c.maxCP = maxCP

	// Если текущее CP больше нового максимума — обрезаем
	if c.currentCP > c.maxCP {
		c.currentCP = c.maxCP
	}
}

// IsDead проверяет мёртв ли персонаж (HP <= 0).
func (c *Character) IsDead() bool {
	return c.CurrentHP() <= 0
}

// HPPercentage возвращает процент текущего HP (0.0 - 1.0).
func (c *Character) HPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxHP == 0 {
		return 0.0
	}
	return float64(c.currentHP) / float64(c.maxHP)
}

// MPPercentage возвращает процент текущего MP (0.0 - 1.0).
func (c *Character) MPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxMP == 0 {
		return 0.0
	}
	return float64(c.currentMP) / float64(c.maxMP)
}

// CPPercentage возвращает процент текущего CP (0.0 - 1.0).
func (c *Character) CPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxCP == 0 {
		return 0.0
	}
	return float64(c.currentCP) / float64(c.maxCP)
}

// Level возвращает уровень персонажа.
func (c *Character) Level() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.level
}

// SetLevel устанавливает уровень персонажа (clamp 1..100).
func (c *Character) SetLevel(level int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}
	c.level = level
}

// GetBasePDef returns base physical defense.
// Uses level modifier formula (для NPCs и fallback).
//
// Formula: basePDef × levelMod
// where basePDef = 80 + level×3 (NPC formula)
//       levelMod = (level + 89) / 100.0
//
// NOTE: Players should override this method with template-based calculation.
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: FuncPDefMod.java:45-87
func (c *Character) GetBasePDef() int32 {
	c.mu.RLock()
	level := c.level
	c.mu.RUnlock()

	// NPC fallback formula
	basePDef := float64(80 + level*3)
	levelMod := float64(level+89) / 100.0

	finalPDef := basePDef * levelMod
	return int32(finalPDef)
}

// ReduceCurrentHP reduces HP by specified amount (minimum 0).
// Does NOT send StatusUpdate packet (caller's responsibility).
//
// Thread-safe: acquires write lock.
//
// Phase 5.3: Basic Combat System (simplified, no DOT/reflect damage).
func (c *Character) ReduceCurrentHP(damage int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentHP = max(c.currentHP-damage, 0)
	// AI onDamage events are triggered by the combat layer (handler/combat package),
	// not here — model package has no AI dependency to avoid import cycles.
}

// DoDie handles character death. Returns true if this call performed the death
// (first caller wins). Subsequent calls return false.
// Uses sync.Once to prevent double death (race condition from concurrent damage).
//
// Thread-safe: sync.Once + write lock.
func (c *Character) DoDie(killer *Player) bool {
	executed := false
	c.deathOnce.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.currentHP > 0 {
			c.currentHP = 0
		}
		executed = true
	})
	return executed
}

// ResetDeathOnce resets the death guard (for respawn).
// Must be called when character is revived/respawned.
func (c *Character) ResetDeathOnce() {
	c.deathOnce = sync.Once{}
}

// --- Cast state (Phase 5.9.4: Cast Flow) ---

// IsCasting returns true if the character is currently casting a skill.
func (c *Character) IsCasting() bool {
	return c.isCasting.Load()
}

// SetCasting sets or clears the casting flag.
func (c *Character) SetCasting(casting bool) {
	c.isCasting.Store(casting)
}

// --- Zone flags (Phase 11: ZoneId enum) ---

// SetInsideZone sets or clears a zone flag on this character.
// Uses CAS loop for lock-free thread safety.
// Java reference: Creature.setInsideZone(ZoneId, boolean)
func (c *Character) SetInsideZone(id ZoneID, inside bool) {
	mask := uint32(1) << id
	for {
		old := c.zones.Load()
		var updated uint32
		if inside {
			updated = old | mask
		} else {
			updated = old &^ mask
		}
		if c.zones.CompareAndSwap(old, updated) {
			return
		}
	}
}

// IsInsideZone checks if a zone flag is set on this character.
// Lock-free: single atomic load.
// Java reference: Creature.isInsideZone(ZoneId)
func (c *Character) IsInsideZone(id ZoneID) bool {
	return (c.zones.Load() & (1 << id)) != 0
}

// ClearAllZoneFlags resets all zone flags to zero.
// Called on teleport/respawn to prevent stale zone state.
func (c *Character) ClearAllZoneFlags() {
	c.zones.Store(0)
}

// --- CC (Crowd Control) flags (Phase 5.9+) ---

// IsStunned returns true if the character is stunned (no actions/movement).
func (c *Character) IsStunned() bool { return c.stunned.Load() }

// SetStunned sets or clears the stun flag.
func (c *Character) SetStunned(v bool) { c.stunned.Store(v) }

// IsRooted returns true if the character is rooted (no movement, can still act).
func (c *Character) IsRooted() bool { return c.rooted.Load() }

// SetRooted sets or clears the root flag.
func (c *Character) SetRooted(v bool) { c.rooted.Store(v) }

// IsSleeping returns true if the character is sleeping (breaks on damage).
func (c *Character) IsSleeping() bool { return c.sleeping.Load() }

// SetSleeping sets or clears the sleep flag.
func (c *Character) SetSleeping(v bool) { c.sleeping.Store(v) }

// IsParalyzed returns true if the character is paralyzed (no actions/movement).
func (c *Character) IsParalyzed() bool { return c.paralyzed.Load() }

// SetParalyzed sets or clears the paralysis flag.
func (c *Character) SetParalyzed(v bool) { c.paralyzed.Store(v) }

// IsFeared returns true if the character is feared (forced movement away from caster).
func (c *Character) IsFeared() bool { return c.feared.Load() }

// SetFeared sets or clears the fear flag.
func (c *Character) SetFeared(v bool) { c.feared.Store(v) }

// IsImmobilized returns true if the character cannot move (stunned, rooted, sleeping, or paralyzed).
func (c *Character) IsImmobilized() bool {
	return c.stunned.Load() || c.rooted.Load() || c.sleeping.Load() || c.paralyzed.Load()
}

// IsDisabled returns true if the character cannot take any action (stunned, sleeping, or paralyzed).
func (c *Character) IsDisabled() bool {
	return c.stunned.Load() || c.sleeping.Load() || c.paralyzed.Load()
}

// ClearAllCCFlags resets all crowd control flags.
// Called on death/respawn.
func (c *Character) ClearAllCCFlags() {
	c.stunned.Store(false)
	c.rooted.Store(false)
	c.sleeping.Store(false)
	c.paralyzed.Store(false)
	c.feared.Store(false)
}
