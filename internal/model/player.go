package model

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/data"
)

// StatBonusProvider provides stat bonuses from active effects (buffs/debuffs).
// Interface to avoid import cycle between model ↔ skill packages.
// Phase 5.9.3: Effect Framework.
type StatBonusProvider interface {
	GetStatBonus(stat string) float64
}

// Player — игровой персонаж.
// Добавляет player-specific данные к Character.
type Player struct {
	*Character // embedded

	characterID int64
	accountID   int64
	level       int32
	raceID      int32
	classID     int32
	experience  int64
	sp          int64
	createdAt   time.Time
	lastLogin   time.Time

	playerMu sync.RWMutex // отдельный mutex для player data

	// Visibility cache (Phase 4.5 PR3)
	// Stores *VisibilityCache — updated by VisibilityManager every 100ms
	// atomic.Value allows lock-free concurrent reads
	visibilityCache atomic.Value // *VisibilityCache (defined in internal/world)

	// Movement tracking (Phase 5.1)
	// Tracks client and server positions separately for desync detection
	movement *PlayerMovement

	// Target tracking (Phase 5.2)
	// Currently selected target (player, NPC, or item)
	// Protected by playerMu for thread-safe access
	target *WorldObject

	// Inventory (Phase 5.5)
	// Player's inventory and equipped items (paperdoll)
	inventory *Inventory

	// Skills (Phase 5.9.2)
	// Learned skills: map[skillID]*SkillInfo
	// Protected by playerMu for thread-safe access
	skills map[int32]*SkillInfo

	// Effect manager (Phase 5.9.3)
	// Tracks active buffs/debuffs and provides stat bonuses
	// Interface to avoid import cycle between model ↔ skill packages
	effectManager StatBonusProvider
}

// NewPlayer создаёт нового игрока с валидацией.
// Phase 4.15: Added objectID parameter to link Player with WorldObject.
// objectID must be unique across all world objects (players, NPCs, items).
// Use 0 for testing/mock objects only (not suitable for production).
func NewPlayer(objectID uint32, characterID, accountID int64, name string, level, raceID, classID int32) (*Player, error) {
	if name == "" || len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters, got %q", name)
	}
	if level < 1 || level > 80 {
		return nil, fmt.Errorf("level must be between 1 and 80, got %d", level)
	}

	// Default spawn location (будет из config/DB в Phase 4.3+)
	loc := NewLocation(0, 0, 0, 0)

	// Default stats (будет из PlayerTemplate в Phase 4.3+)
	// Для MVP используем simple hardcoded values
	maxHP := int32(1000 + level*50)  // Linear scaling для тестов
	maxMP := int32(500 + level*25)
	maxCP := int32(800 + level*40)

	p := &Player{
		Character:   NewCharacter(objectID, name, loc, level, maxHP, maxMP, maxCP),
		characterID: characterID,
		accountID:   accountID,
		level:       level, // FIXME: level duplicated in Character and Player
		raceID:      raceID,
		classID:     classID,
		experience:  0,
		createdAt:   time.Now(),
		movement:    NewPlayerMovement(loc.X, loc.Y, loc.Z), // Phase 5.1
		inventory:   NewInventory(characterID),              // Phase 5.5
	}

	// Initialize visibility cache (Phase 4.5 PR3)
	p.visibilityCache.Store((*VisibilityCache)(nil))

	// Phase 5.6: Set WorldObject.Data reference for PvE combat
	p.WorldObject.Data = p

	return p, nil
}

// CharacterID возвращает DB ID персонажа (immutable).
func (p *Player) CharacterID() int64 {
	return p.characterID
}

// AccountID возвращает ID аккаунта (immutable).
func (p *Player) AccountID() int64 {
	return p.accountID
}

// Level возвращает уровень персонажа.
func (p *Player) Level() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.level
}

// SetLevel устанавливает уровень с валидацией.
func (p *Player) SetLevel(level int32) error {
	if level < 1 || level > 80 {
		return fmt.Errorf("level must be between 1 and 80, got %d", level)
	}

	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.level = level
	return nil
}

// RaceID возвращает ID расы.
func (p *Player) RaceID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.raceID
}

// ClassID возвращает ID класса.
func (p *Player) ClassID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.classID
}

// Experience возвращает текущий опыт.
func (p *Player) Experience() int64 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.experience
}

// AddExperience добавляет опыт (может быть отрицательным для penalty).
func (p *Player) AddExperience(exp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.experience += exp
	if p.experience < 0 {
		p.experience = 0
	}
}

// SetExperience устанавливает точное значение опыта.
func (p *Player) SetExperience(exp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if exp < 0 {
		exp = 0
	}
	p.experience = exp
}

// SP возвращает текущие skill points.
func (p *Player) SP() int64 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.sp
}

// AddSP добавляет skill points.
func (p *Player) AddSP(sp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.sp += sp
	if p.sp < 0 {
		p.sp = 0
	}
}

// SetSP устанавливает точное значение SP (для загрузки из DB).
func (p *Player) SetSP(sp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if sp < 0 {
		sp = 0
	}
	p.sp = sp
}

// CreatedAt возвращает время создания персонажа.
func (p *Player) CreatedAt() time.Time {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.createdAt
}

// LastLogin возвращает время последнего входа.
func (p *Player) LastLogin() time.Time {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.lastLogin
}

// UpdateLastLogin обновляет время последнего входа на текущее.
func (p *Player) UpdateLastLogin() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.lastLogin = time.Now()
}

// SetLastLogin устанавливает время последнего входа (для загрузки из DB).
func (p *Player) SetLastLogin(t time.Time) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.lastLogin = t
}

// SetCreatedAt устанавливает время создания (для загрузки из DB).
func (p *Player) SetCreatedAt(t time.Time) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.createdAt = t
}

// SetCharacterID устанавливает DB ID после создания в БД.
func (p *Player) SetCharacterID(id int64) {
	// characterID immutable, но setter нужен для repository.Create
	p.characterID = id
}

// SetRaceID устанавливает ID расы (для загрузки из DB).
func (p *Player) SetRaceID(raceID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.raceID = raceID
}

// SetClassID устанавливает ID класса (для изменения профессии).
func (p *Player) SetClassID(classID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.classID = classID
}

// GetVisibilityCache returns current visibility cache (may be nil if not initialized).
// Safe for concurrent reads (atomic.Value.Load).
// Phase 4.5 PR3: Visibility Cache optimization.
func (p *Player) GetVisibilityCache() *VisibilityCache {
	v := p.visibilityCache.Load()
	if v == nil {
		return nil
	}
	return v.(*VisibilityCache)
}

// SetVisibilityCache updates visibility cache atomically.
// Safe for concurrent writes (atomic.Value.Store is thread-safe).
// Phase 4.5 PR3: Called by VisibilityManager every 100ms.
func (p *Player) SetVisibilityCache(cache *VisibilityCache) {
	p.visibilityCache.Store(cache)
}

// InvalidateVisibilityCache clears visibility cache (sets to nil).
// Called when player moves to different region, teleports, or logs out.
// Phase 4.5 PR3: Forces fresh query on next visibility check.
func (p *Player) InvalidateVisibilityCache() {
	p.visibilityCache.Store((*VisibilityCache)(nil))
}

// Movement returns the movement tracking state.
// Thread-safe: PlayerMovement uses RWMutex internally.
// Phase 5.1: Movement validation and desync detection.
func (p *Player) Movement() *PlayerMovement {
	return p.movement
}

// CanLogout returns true if player can logout/restart safely.
// Checks multiple conditions: attack stance, trading, enchanting, events, festivals.
//
// Phase 4.17.5: MVP implementation with basic checks.
// TODO Phase 5.x: Add full checks (subclass lock, enchant, event registration, festivals).
//
// Reference: L2J_Mobius Player.canLogout() (8270-8313)
func (p *Player) CanLogout() bool {
	// TODO Phase 5.x: Add subclass lock check
	// if p.subclassLock.Load() {
	//     return false
	// }

	// TODO Phase 5.x: Add active enchant check
	// if p.activeEnchantItemID.Load() != IDNone {
	//     return false
	// }

	// Check attack stance (combat)
	// TODO Phase 5.x: Implement AttackStanceTaskManager
	// For MVP, use stub method that always returns false (no combat system yet)
	if p.HasAttackStance() {
		// TODO: Send system message "YOU_CANNOT_EXIT_WHILE_IN_COMBAT"
		return false
	}

	// TODO Phase 5.x: Add event registration check
	// if p.IsRegisteredOnEvent() {
	//     return false
	// }

	// TODO Phase 5.x: Add festival participant check
	// if p.IsFestivalParticipant() {
	//     return false
	// }

	return true
}

// HasAttackStance returns true if player is in combat (attacked or was attacked recently).
// Combat state persists for 15 seconds after last attack (COMBAT_TIME).
//
// Phase 4.17.5: Stub implementation (always returns false).
// TODO Phase 4.18: Implement AttackStanceTaskManager with 15-second cooldown.
//
// Reference: L2J_Mobius AttackStanceTaskManager
func (p *Player) HasAttackStance() bool {
	// TODO Phase 4.18: Track last attack time
	// return time.Since(p.lastAttackTime) < 15*time.Second
	return false
}

// IsTrading returns true if player is in trade mode (private store, manufacture).
//
// Phase 4.17.5: Stub implementation (always returns false).
// TODO Phase 5.x: Implement PrivateStoreType tracking (SELL, BUY, MANUFACTURE, etc.).
//
// Reference: L2J_Mobius Player.getPrivateStoreType()
func (p *Player) IsTrading() bool {
	// TODO Phase 5.x: Track private store state
	// return p.privateStoreType.Load() != PrivateStoreTypeNone
	return false
}

// Target returns the currently selected target.
// Returns nil if no target is selected.
// Thread-safe: acquires read lock.
//
// Phase 5.2: Target System.
func (p *Player) Target() *WorldObject {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.target
}

// SetTarget sets the currently selected target.
// Pass nil to clear the target.
// Thread-safe: acquires write lock.
//
// Phase 5.2: Target System.
func (p *Player) SetTarget(target *WorldObject) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.target = target
}

// ClearTarget clears the currently selected target.
// Convenience method equivalent to SetTarget(nil).
// Thread-safe: acquires write lock.
//
// Phase 5.2: Target System.
func (p *Player) ClearTarget() {
	p.SetTarget(nil)
}

// HasTarget returns true if player has a target selected.
// Thread-safe: acquires read lock.
//
// Phase 5.2: Target System.
func (p *Player) HasTarget() bool {
	return p.Target() != nil
}

// Inventory returns player's inventory (inventory + equipped items).
// Thread-safe: Inventory methods use internal mutex.
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) Inventory() *Inventory {
	return p.inventory
}

// GetEquippedWeapon returns equipped weapon (может быть nil).
// Convenience method для Inventory().GetPaperdollItem(PaperdollRHand).
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) GetEquippedWeapon() *Item {
	if p.inventory == nil {
		return nil
	}
	return p.inventory.GetPaperdollItem(PaperdollRHand)
}

// GetEquippedArmor returns equipped armor для указанного slot (может быть nil).
// Convenience method для Inventory().GetPaperdollItem(slot).
//
// Parameters:
//   - slot: paperdoll slot index (PaperdollChest, PaperdollLegs, etc.)
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) GetEquippedArmor(slot int32) *Item {
	if p.inventory == nil {
		return nil
	}
	return p.inventory.GetPaperdollItem(slot)
}

// GetBasePAtk returns base physical attack power from character template (no weapon).
// This is the "nude" pAtk WITHOUT weapon bonus, but WITH STR bonus and level modifier.
//
// Formula: basePAtk × STRBonus[STR] × levelMod
// where levelMod = (level + 89) / 100.0
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: FuncPAtkMod.java:45
func (p *Player) GetBasePAtk() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 100 // Fallback для неизвестного класса
	}

	basePAtk := float64(template.BasePAtk)
	strBonus := data.GetSTRBonus(p.GetSTR())
	levelMod := p.GetLevelMod()

	finalPAtk := basePAtk * strBonus * levelMod
	return int32(finalPAtk)
}

// GetPAtk returns final physical attack power WITH weapon bonus.
// Weapon pAtk is added BEFORE applying STR bonus and level modifier.
//
// Formula: (basePAtk + weaponPAtk) × STRBonus[STR] × levelMod
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: CreatureStat.java:539, FuncAdd + FuncPAtkMod
func (p *Player) GetPAtk() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 100 // Fallback
	}

	// Base pAtk from character template
	basePAtk := float64(template.BasePAtk)

	// Add weapon pAtk
	weaponPAtk := int32(0)
	if weapon := p.GetEquippedWeapon(); weapon != nil {
		weaponPAtk = weapon.Template().PAtk
	}

	// Apply STR bonus and level modifier
	strBonus := data.GetSTRBonus(p.GetSTR())
	levelMod := p.GetLevelMod()

	finalPAtk := (basePAtk + float64(weaponPAtk)) * strBonus * levelMod
	return int32(finalPAtk)
}

// GetPAtkSpd returns physical attack speed.
// Uses real character template base speed (default: 300).
//
// TODO Phase 5.5: add weapon speed modifier + DEX bonus + buffs.
//
// Phase 5.4: Character Templates & Stats System.
func (p *Player) GetPAtkSpd() float64 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 300.0 // Fallback
	}

	// For MVP: return template base speed (no weapon/DEX/buff modifiers)
	return float64(template.BasePAtkSpd)
}

// GetAttackRange returns physical attack range in game units.
// Weapon overrides template base range (fists=20 → sword=40 → bow=500).
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: CreatureStat.java:591-605
func (p *Player) GetAttackRange() int32 {
	// If weapon equipped, use weapon range
	if weapon := p.GetEquippedWeapon(); weapon != nil {
		return weapon.Template().AttackRange
	}

	// Fists: use template base range
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 20 // Fallback (typical fists range)
	}
	return template.BaseAtkRange
}

// GetAttackDelay returns delay between attacks (attack speed).
// MVP: simplified formula for unarmed player.
//
// Formula: 500000 / PAtkSpd (in milliseconds).
// Example: 300 PAtkSpd → 1666ms delay (~0.6 attacks/sec).
//
// Phase 5.3: Basic Combat System.
// Java reference: Creature.getAttackEndTime() (line 5419, 5433).
func (p *Player) GetAttackDelay() time.Duration {
	pAtkSpd := p.GetPAtkSpd()

	// Formula: 500000 / PAtkSpd (в миллисекундах)
	// Typical value: 300 PAtkSpd → 1666ms delay
	delayMs := int(500000 / pAtkSpd)

	return time.Duration(delayMs) * time.Millisecond
}

// GetLevelMod returns level modifier для stat scaling.
// Formula: (level + 89) / 100.0
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: Creature.getLevelMod() (CreatureTemplate.java)
func (p *Player) GetLevelMod() float64 {
	return float64(p.Level()+89) / 100.0
}

// GetSTR returns current STR attribute.
// For MVP: returns base STR from template (no equipment/buffs modifiers).
//
// TODO Phase 5.5: add equipment STR bonus + buff modifiers.
//
// Phase 5.4: Character Templates & Stats System.
func (p *Player) GetSTR() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 40 // Default fallback (Human Fighter base STR)
	}
	return template.BaseSTR
}

// GetBasePDef returns base physical defense (overrides Character.GetBasePDef).
// Uses real character template stats + level modifier (nude).
//
// Formula: basePDef × levelMod
// where levelMod = (level + 89) / 100.0
//
// For MVP: assumes nude (no equipped items).
// TODO Phase 5.5: subtract equipped slot defs.
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: FuncPDefMod.java:45-87
func (p *Player) GetBasePDef() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		// Fallback to Character.GetBasePDef
		return p.Character.GetBasePDef()
	}

	basePDef := float64(template.BasePDef) // Nude defense (sum of all slots)
	levelMod := p.GetLevelMod()

	finalPDef := basePDef * levelMod
	return int32(finalPDef)
}

// GetPDef returns final physical defense WITH armor (slot subtraction!).
// Armor pDef uses slot-based subtraction: subtract base slot def for equipped slots,
// then add armor pDef, then apply level modifier.
//
// Formula: (basePDef - equippedSlotsDef + armorPDef) × levelMod
//
// Example (Human Fighter level 10):
//   - Nude: 80 × 0.99 = 79.2 ≈ 79
//   - Leather Shirt (chest pDef=43): (80 - 31 + 43) × 0.99 = 92 × 0.99 = 91.08 ≈ 91
//   - Full Plate (chest=100, legs=50): (80 - 31 - 18 + 100 + 50) × 0.99 = 181 × 0.99 = 179.19 ≈ 179
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: FuncPDefMod.java:44-88
func (p *Player) GetPDef() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		// Fallback to GetBasePDef (nude)
		return p.GetBasePDef()
	}

	// Start with nude pDef (sum of all empty slot defs)
	basePDef := float64(template.BasePDef)

	// Armor slots that contribute to pDef
	slots := []int32{
		PaperdollChest, PaperdollLegs, PaperdollHead,
		PaperdollFeet, PaperdollGloves, PaperdollUnder, PaperdollCloak,
	}

	// Subtract equipped slot base defs
	for _, slot := range slots {
		if p.GetEquippedArmor(slot) != nil {
			// Slot occupied → subtract base slot def
			templateSlot := paperdollSlotToTemplateSlot(slot)
			if slotDef, exists := template.SlotDef[templateSlot]; exists {
				basePDef -= float64(slotDef)
			}
		}
	}

	// Add armor pDef
	armorPDef := float64(0)
	for _, slot := range slots {
		if armor := p.GetEquippedArmor(slot); armor != nil {
			armorPDef += float64(armor.Template().PDef)
		}
	}

	// Apply level modifier
	levelMod := p.GetLevelMod()
	finalPDef := (basePDef + armorPDef) * levelMod

	return int32(finalPDef)
}

// paperdollSlotToTemplateSlot converts paperdoll slot index to template SlotDef key.
// Used for armor slot subtraction in GetPDef().
//
// Phase 5.5: Weapon & Equipment System.
func paperdollSlotToTemplateSlot(paperdollSlot int32) uint8 {
	switch paperdollSlot {
	case PaperdollChest:
		return data.SlotChest
	case PaperdollLegs:
		return data.SlotLegs
	case PaperdollHead:
		return data.SlotHead
	case PaperdollFeet:
		return data.SlotFeet
	case PaperdollGloves:
		return data.SlotGloves
	case PaperdollUnder:
		return data.SlotUnderwear
	case PaperdollCloak:
		return data.SlotCloak
	default:
		return 0 // Unknown slot
	}
}

// DoAttack выполняет физическую атаку на target.
// NOTE: This is a no-op stub to satisfy signature requirements.
// Actual combat logic is in combat.CombatMgr.ExecuteAttack() (called from handler).
//
// Phase 5.3: Basic Combat System (stub to avoid import cycle).
func (p *Player) DoAttack(target *WorldObject) {
	// No-op: Combat logic delegated to combat.CombatMgr to avoid import cycle.
	// Handler calls combat.CombatMgr.ExecuteAttack() directly.
}

// --- Skills (Phase 5.9.2) ---

// AddSkill adds or updates a skill in player's collection.
// Thread-safe: acquires write lock.
func (p *Player) AddSkill(skillID, level int32, passive bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.skills == nil {
		p.skills = make(map[int32]*SkillInfo)
	}
	p.skills[skillID] = &SkillInfo{
		SkillID: skillID,
		Level:   level,
		Passive: passive,
	}
}

// GetSkill returns SkillInfo for given skill ID.
// Returns nil if skill not learned.
// Thread-safe: acquires read lock.
func (p *Player) GetSkill(skillID int32) *SkillInfo {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.skills == nil {
		return nil
	}
	return p.skills[skillID]
}

// HasSkill returns true if player has learned the given skill.
// Thread-safe: acquires read lock.
func (p *Player) HasSkill(skillID int32) bool {
	return p.GetSkill(skillID) != nil
}

// Skills returns a copy of all learned skills.
// Thread-safe: acquires read lock.
func (p *Player) Skills() []*SkillInfo {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.skills == nil {
		return nil
	}

	result := make([]*SkillInfo, 0, len(p.skills))
	for _, s := range p.skills {
		result = append(result, s)
	}
	return result
}

// SkillCount returns the number of learned skills.
// Thread-safe: acquires read lock.
func (p *Player) SkillCount() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	return len(p.skills)
}

// RemoveSkill removes a skill from player's collection.
// Thread-safe: acquires write lock.
func (p *Player) RemoveSkill(skillID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	delete(p.skills, skillID)
}

// --- Effect Manager (Phase 5.9.3) ---

// SetEffectManager sets the effect manager for this player.
// Called during player initialization.
func (p *Player) SetEffectManager(em StatBonusProvider) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.effectManager = em
}

// EffectManager returns the effect manager.
func (p *Player) EffectManager() StatBonusProvider {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.effectManager
}
