package model

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

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
	}

	// Initialize visibility cache (Phase 4.5 PR3)
	p.visibilityCache.Store((*VisibilityCache)(nil))

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

// GetBasePAtk returns base physical attack power.
// MVP: hardcoded formula (100 + level × 5).
//
// TODO Phase 5.4: load from character template + weapon stats.
//
// Phase 5.3: Basic Combat System.
func (p *Player) GetBasePAtk() int32 {
	// MVP: simple linear scaling
	// Level 1: 105, Level 80: 500
	return 100 + p.Level()*5
}

// GetPAtkSpd returns physical attack speed.
// MVP: fixed value 300 (typical fighter).
//
// TODO Phase 5.4: load from template + buffs + weapon speed.
//
// Phase 5.3: Basic Combat System.
func (p *Player) GetPAtkSpd() float64 {
	// MVP: fixed attack speed (300 → ~1.66 attacks per second)
	return 300.0
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

// DoAttack выполняет физическую атаку на target.
// NOTE: This is a no-op stub to satisfy signature requirements.
// Actual combat logic is in combat.CombatMgr.ExecuteAttack() (called from handler).
//
// Phase 5.3: Basic Combat System (stub to avoid import cycle).
func (p *Player) DoAttack(target *WorldObject) {
	// No-op: Combat logic delegated to combat.CombatMgr to avoid import cycle.
	// Handler calls combat.CombatMgr.ExecuteAttack() directly.
}
