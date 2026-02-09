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
}

// NewPlayer создаёт нового игрока с валидацией.
func NewPlayer(characterID, accountID int64, name string, level, raceID, classID int32) (*Player, error) {
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
		Character:   NewCharacter(0, name, loc, level, maxHP, maxMP, maxCP),
		characterID: characterID,
		accountID:   accountID,
		level:       level, // FIXME: level duplicated in Character and Player
		raceID:      raceID,
		classID:     classID,
		experience:  0,
		createdAt:   time.Now(),
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
