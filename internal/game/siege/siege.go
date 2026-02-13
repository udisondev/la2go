package siege

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// SiegeState represents the current state of a siege.
type SiegeState int32

const (
	StateInactive     SiegeState = 0 // No siege active
	StateRegistration SiegeState = 1 // Clans can register
	StateCountdown    SiegeState = 2 // Countdown before siege start
	StateRunning      SiegeState = 3 // Siege in progress
)

// Default siege configuration.
const (
	DefaultSiegeLength     = 120 * time.Minute // 2 hours
	DefaultSiegeCycle      = 14                // 2 weeks in days
	DefaultMaxFlags        = 1
	DefaultClanMinLevel    = 4
	DefaultAttackerMax     = 500
	DefaultDefenderMax     = 500
)

// TowerSpawn holds tower spawn data.
type TowerSpawn struct {
	X, Y, Z int32
	NpcID   int32
	ZoneIDs []int32 // FlameTower: associated zone IDs
}

// Siege represents an active castle siege.
// Thread-safe: protected by mu.
type Siege struct {
	mu sync.RWMutex

	castle       *Castle
	state        atomic.Int32 // SiegeState
	isNormalSide bool         // true = normal sides, false = swapped (mid-victory)

	attackerClans map[int32]*SiegeClan // clanID → SiegeClan
	defenderClans map[int32]*SiegeClan // clanID → SiegeClan
	pendingClans  map[int32]*SiegeClan // clanID → SiegeClan (unapproved defenders)

	controlTowers    []*TowerSpawn // Active control towers
	flameTowers      []*TowerSpawn // Active flame towers
	controlTowerAlive int32        // Remaining alive control towers

	startTime time.Time
	endTime   time.Time

	// Таймер следующего этапа (nil если не запущен).
	timer *time.Timer
}

// NewSiege creates a new siege for the given castle.
func NewSiege(castle *Castle) *Siege {
	return &Siege{
		castle:        castle,
		isNormalSide:  true,
		attackerClans: make(map[int32]*SiegeClan, 16),
		defenderClans: make(map[int32]*SiegeClan, 16),
		pendingClans:  make(map[int32]*SiegeClan, 8),
	}
}

// Castle returns the castle under siege.
func (s *Siege) Castle() *Castle { return s.castle }

// State returns the current siege state.
func (s *Siege) State() SiegeState {
	return SiegeState(s.state.Load())
}

// SetState sets the siege state.
func (s *Siege) SetState(st SiegeState) {
	s.state.Store(int32(st))
}

// IsInProgress returns true if the siege is running.
func (s *Siege) IsInProgress() bool {
	return s.State() == StateRunning
}

// IsRegistration returns true if registration is open.
func (s *Siege) IsRegistration() bool {
	st := s.State()
	return st == StateRegistration || st == StateInactive
}

// --- Clan registration ---

// RegisterAttacker adds a clan as an attacker.
func (s *Siege) RegisterAttacker(sc *SiegeClan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc.Type = ClanTypeAttacker
	s.attackerClans[sc.ClanID] = sc
}

// RegisterDefender adds a clan as a defender.
func (s *Siege) RegisterDefender(sc *SiegeClan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc.Type = ClanTypeDefender
	s.defenderClans[sc.ClanID] = sc
}

// RegisterPendingDefender adds a clan as unapproved defender.
func (s *Siege) RegisterPendingDefender(sc *SiegeClan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc.Type = ClanTypeDefenderNotApproved
	s.pendingClans[sc.ClanID] = sc
}

// ApprovePendingDefender moves a pending defender to approved.
func (s *Siege) ApprovePendingDefender(clanID int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	sc, ok := s.pendingClans[clanID]
	if !ok {
		return false
	}
	delete(s.pendingClans, clanID)
	sc.Type = ClanTypeDefender
	s.defenderClans[clanID] = sc
	return true
}

// RemoveClan removes a clan from all lists.
func (s *Siege) RemoveClan(clanID int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.attackerClans, clanID)
	delete(s.defenderClans, clanID)
	delete(s.pendingClans, clanID)
}

// AttackerClan returns an attacker clan by ID, or nil.
func (s *Siege) AttackerClan(clanID int32) *SiegeClan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.attackerClans[clanID]
}

// DefenderClan returns a defender clan by ID, or nil.
func (s *Siege) DefenderClan(clanID int32) *SiegeClan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defenderClans[clanID]
}

// IsAttacker returns true if the clan is registered as attacker.
func (s *Siege) IsAttacker(clanID int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.attackerClans[clanID]
	return ok
}

// IsDefender returns true if the clan is registered as defender (or owner).
func (s *Siege) IsDefender(clanID int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.defenderClans[clanID]
	return ok
}

// IsClanRegistered returns true if the clan is registered in any role.
func (s *Siege) IsClanRegistered(clanID int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.attackerClans[clanID]; ok {
		return true
	}
	if _, ok := s.defenderClans[clanID]; ok {
		return true
	}
	_, ok := s.pendingClans[clanID]
	return ok
}

// AttackerClans returns a snapshot of all attacker clans.
func (s *Siege) AttackerClans() []*SiegeClan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*SiegeClan, 0, len(s.attackerClans))
	for _, sc := range s.attackerClans {
		result = append(result, sc)
	}
	return result
}

// DefenderClans returns a snapshot of all defender clans.
func (s *Siege) DefenderClans() []*SiegeClan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*SiegeClan, 0, len(s.defenderClans))
	for _, sc := range s.defenderClans {
		result = append(result, sc)
	}
	return result
}

// PendingClans returns a snapshot of all pending defender clans.
func (s *Siege) PendingClans() []*SiegeClan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*SiegeClan, 0, len(s.pendingClans))
	for _, sc := range s.pendingClans {
		result = append(result, sc)
	}
	return result
}

// AttackerCount returns the number of attacker clans.
func (s *Siege) AttackerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.attackerClans)
}

// DefenderCount returns the number of defender clans.
func (s *Siege) DefenderCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.defenderClans)
}

// --- Tower management ---

// SetControlTowers sets control tower spawns.
func (s *Siege) SetControlTowers(towers []*TowerSpawn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controlTowers = towers
	s.controlTowerAlive = int32(len(towers))
}

// SetFlameTowers sets flame tower spawns.
func (s *Siege) SetFlameTowers(towers []*TowerSpawn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flameTowers = towers
}

// ControlTowerDestroyed decrements the alive control tower count.
// Returns true if all towers are destroyed (castle can be captured).
func (s *Siege) ControlTowerDestroyed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controlTowerAlive--
	if s.controlTowerAlive < 0 {
		s.controlTowerAlive = 0
	}
	return s.controlTowerAlive <= 0
}

// ControlTowersAlive returns remaining alive control towers.
func (s *Siege) ControlTowersAlive() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controlTowerAlive
}

// ControlTowers returns control tower spawns.
func (s *Siege) ControlTowers() []*TowerSpawn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controlTowers
}

// FlameTowers returns flame tower spawns.
func (s *Siege) FlameTowers() []*TowerSpawn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.flameTowers
}

// --- Siege lifecycle ---

// StartSiege transitions to running state.
func (s *Siege) StartSiege() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.startTime = time.Now()
	s.state.Store(int32(StateRunning))

	// Владелец замка автоматически становится защитником.
	ownerID := s.castle.OwnerClanID()
	if ownerID > 0 {
		if _, ok := s.defenderClans[ownerID]; !ok {
			s.defenderClans[ownerID] = &SiegeClan{
				ClanID: ownerID,
				Type:   ClanTypeOwner,
			}
		}
	}

	// Перемещаем одобренных pending в defenders.
	for id, sc := range s.pendingClans {
		sc.Type = ClanTypeDefender
		s.defenderClans[id] = sc
	}
	s.pendingClans = make(map[int32]*SiegeClan, 8)

	slog.Info("siege started",
		"castle", s.castle.Name(),
		"attackers", len(s.attackerClans),
		"defenders", len(s.defenderClans))
}

// EndSiege transitions to inactive state and cleans up.
func (s *Siege) EndSiege() {
	s.state.Store(int32(StateInactive))

	s.mu.Lock()
	endTime := time.Now()
	s.endTime = endTime

	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	attackerCount := len(s.attackerClans)
	defenderCount := len(s.defenderClans)
	s.mu.Unlock()

	slog.Info("siege ended",
		"castle", s.castle.Name(),
		"duration", endTime.Sub(s.startTime).Round(time.Second),
		"attackers", attackerCount,
		"defenders", defenderCount)
}

// ClearClans removes all registered clans.
func (s *Siege) ClearClans() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attackerClans = make(map[int32]*SiegeClan, 16)
	s.defenderClans = make(map[int32]*SiegeClan, 16)
	s.pendingClans = make(map[int32]*SiegeClan, 8)
}

// StartTime returns when the siege started.
func (s *Siege) StartTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

// EndTime returns when the siege ended.
func (s *Siege) EndTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endTime
}

// SetTimer sets the siege countdown timer.
func (s *Siege) SetTimer(t *time.Timer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = t
}

// StopTimer stops the current timer.
func (s *Siege) StopTimer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

// MidVictory handles the mid-siege castle capture when all control towers are destroyed.
// The attacking clan takes ownership, and sides swap.
func (s *Siege) MidVictory(newOwnerClanID int32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Меняем владельца.
	s.castle.SetOwnerClanID(newOwnerClanID)

	// Меняем стороны: бывший владелец → атакующий, новый владелец → защитник.
	s.isNormalSide = !s.isNormalSide

	// Перемещаем нового владельца из атакующих в защитники.
	if sc, ok := s.attackerClans[newOwnerClanID]; ok {
		delete(s.attackerClans, newOwnerClanID)
		sc.Type = ClanTypeOwner
		s.defenderClans[newOwnerClanID] = sc
	}

	slog.Info("mid-victory: castle captured",
		"castle", s.castle.Name(),
		"new_owner_clan", newOwnerClanID)
}
