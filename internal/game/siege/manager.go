package siege

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

// CastleInfo defines a known castle in the Interlude world.
type CastleInfo struct {
	ID             int32
	Name           string
	MaxMercenaries int32
}

// 9 castles in Lineage 2 Interlude.
var knownCastles = []CastleInfo{
	{ID: 1, Name: "Gludio", MaxMercenaries: 30},
	{ID: 2, Name: "Dion", MaxMercenaries: 30},
	{ID: 3, Name: "Giran", MaxMercenaries: 30},
	{ID: 4, Name: "Oren", MaxMercenaries: 30},
	{ID: 5, Name: "Aden", MaxMercenaries: 36},
	{ID: 6, Name: "Innadril", MaxMercenaries: 30},
	{ID: 7, Name: "Goddard", MaxMercenaries: 30},
	{ID: 8, Name: "Rune", MaxMercenaries: 30},
	{ID: 9, Name: "Schuttgart", MaxMercenaries: 30},
}

// ManagerConfig holds siege manager settings.
type ManagerConfig struct {
	SiegeCycleWeeks    int           // Siege cycle in weeks (default 2)
	SiegeLength        time.Duration // Siege duration (default 120m)
	AttackerMaxClans   int           // Max attacker clans (default 500)
	DefenderMaxClans   int           // Max defender clans (default 500)
	MaxFlags           int           // Max flags per clan (default 1)
	SiegeClanMinLevel  int32         // Min clan level to register (default 4)
}

// DefaultManagerConfig returns sensible defaults.
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		SiegeCycleWeeks:   DefaultSiegeCycle / 7,
		SiegeLength:       DefaultSiegeLength,
		AttackerMaxClans:  DefaultAttackerMax,
		DefenderMaxClans:  DefaultDefenderMax,
		MaxFlags:          DefaultMaxFlags,
		SiegeClanMinLevel: DefaultClanMinLevel,
	}
}

// Manager errors.
var (
	ErrCastleNotFound        = errors.New("castle not found")
	ErrSiegeInProgress       = errors.New("siege in progress")
	ErrRegistrationClosed    = errors.New("siege registration closed")
	ErrClanAlreadyRegistered = errors.New("clan already registered")
	ErrClanLevelTooLow       = errors.New("clan level too low for siege")
	ErrAttackerLimitReached  = errors.New("attacker limit reached")
	ErrDefenderLimitReached  = errors.New("defender limit reached")
	ErrOwnerCannotAttack     = errors.New("castle owner cannot register as attacker")
)

// Manager coordinates all 9 castles and their siege lifecycle.
// Thread-safe: protected by mu.
type Manager struct {
	mu      sync.RWMutex
	castles map[int32]*Castle // castleID → Castle
	config  ManagerConfig
}

// NewManager creates a siege manager and initializes the 9 castles.
func NewManager(cfg ManagerConfig) *Manager {
	m := &Manager{
		castles: make(map[int32]*Castle, len(knownCastles)),
		config:  cfg,
	}
	for _, info := range knownCastles {
		c := NewCastle(info.ID, info.Name, info.MaxMercenaries)
		siege := NewSiege(c)
		siege.SetState(StateRegistration)
		c.SetSiege(siege)
		m.castles[info.ID] = c
	}
	slog.Info("siege manager initialized", "castles", len(m.castles))
	return m
}

// Castle returns a castle by ID, or nil.
func (m *Manager) Castle(id int32) *Castle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.castles[id]
}

// CastleByName returns a castle by name (case-insensitive), or nil.
func (m *Manager) CastleByName(name string) *Castle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.castles {
		if equalFoldASCII(c.Name(), name) {
			return c
		}
	}
	return nil
}

// Castles returns a snapshot of all castles.
func (m *Manager) Castles() []*Castle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Castle, 0, len(m.castles))
	for _, c := range m.castles {
		result = append(result, c)
	}
	return result
}

// Config returns the manager configuration.
func (m *Manager) Config() ManagerConfig {
	return m.config
}

// RegisterAttacker registers a clan as siege attacker.
func (m *Manager) RegisterAttacker(castleID, clanID int32, clanName string, clanLevel int32) error {
	c := m.Castle(castleID)
	if c == nil {
		return ErrCastleNotFound
	}

	siege := c.Siege()
	if siege == nil {
		return ErrSiegeInProgress
	}

	if siege.IsInProgress() {
		return ErrSiegeInProgress
	}

	if !siege.IsRegistration() {
		return ErrRegistrationClosed
	}

	if clanLevel < m.config.SiegeClanMinLevel {
		return ErrClanLevelTooLow
	}

	// Владелец замка не может атаковать свой замок.
	if c.OwnerClanID() == clanID {
		return ErrOwnerCannotAttack
	}

	if siege.IsClanRegistered(clanID) {
		return ErrClanAlreadyRegistered
	}

	if siege.AttackerCount() >= m.config.AttackerMaxClans {
		return ErrAttackerLimitReached
	}

	sc := NewSiegeClan(clanID, clanName, ClanTypeAttacker)
	siege.RegisterAttacker(sc)

	slog.Info("siege: attacker registered",
		"castle", c.Name(), "clan_id", clanID, "clan", clanName)
	return nil
}

// RegisterDefender registers a clan as siege defender.
// If the castle has an owner, the clan goes to pending list awaiting approval.
// If the castle has no owner, the clan is auto-approved.
func (m *Manager) RegisterDefender(castleID, clanID int32, clanName string, clanLevel int32) error {
	c := m.Castle(castleID)
	if c == nil {
		return ErrCastleNotFound
	}

	siege := c.Siege()
	if siege == nil {
		return ErrSiegeInProgress
	}

	if siege.IsInProgress() {
		return ErrSiegeInProgress
	}

	if !siege.IsRegistration() {
		return ErrRegistrationClosed
	}

	if clanLevel < m.config.SiegeClanMinLevel {
		return ErrClanLevelTooLow
	}

	if siege.IsClanRegistered(clanID) {
		return ErrClanAlreadyRegistered
	}

	if siege.DefenderCount() >= m.config.DefenderMaxClans {
		return ErrDefenderLimitReached
	}

	sc := NewSiegeClan(clanID, clanName, ClanTypeDefender)

	if c.HasOwner() {
		siege.RegisterPendingDefender(sc)
		slog.Info("siege: defender pending approval",
			"castle", c.Name(), "clan_id", clanID, "clan", clanName)
	} else {
		siege.RegisterDefender(sc)
		slog.Info("siege: defender registered",
			"castle", c.Name(), "clan_id", clanID, "clan", clanName)
	}
	return nil
}

// Unregister removes a clan from a siege.
func (m *Manager) Unregister(castleID, clanID int32) error {
	c := m.Castle(castleID)
	if c == nil {
		return ErrCastleNotFound
	}

	siege := c.Siege()
	if siege == nil {
		return ErrSiegeInProgress
	}

	if siege.IsInProgress() {
		return ErrSiegeInProgress
	}

	siege.RemoveClan(clanID)
	return nil
}

// ApproveDefender approves a pending defender clan.
func (m *Manager) ApproveDefender(castleID, clanID int32) error {
	c := m.Castle(castleID)
	if c == nil {
		return ErrCastleNotFound
	}

	siege := c.Siege()
	if siege == nil {
		return ErrSiegeInProgress
	}

	if !siege.ApprovePendingDefender(clanID) {
		return errors.New("clan not in pending list")
	}
	return nil
}

// IsClanRegistered checks if a clan is registered in any siege.
func (m *Manager) IsClanRegistered(clanID int32) (castleID int32, registered bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.castles {
		siege := c.Siege()
		if siege != nil && siege.IsClanRegistered(clanID) {
			return c.ID(), true
		}
	}
	return 0, false
}

// equalFoldASCII compares two ASCII strings case-insensitively.
func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
