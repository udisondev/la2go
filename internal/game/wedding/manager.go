// Package wedding implements the L2 Interlude marriage system.
//
// Players can get engaged via the .engage voiced command, then married
// via NPC 50007 (Wedding Minister). Married couples share a special
// social link, can teleport to each other (.gotolove), and pay adena
// penalties on divorce.
//
// Phase 33: Marriage System.
package wedding

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// Wedding system configuration constants.
// Based on L2J WeddingConfig defaults.
const (
	DefaultWeddingPrice int64 = 250_000_000 // 250M adena to marry
	DefaultDivorceCost  int64 = 50_000_000  // base cost, 20% of partner's adena
	DefaultTeleportCost int64 = 50_000      // .gotolove teleport cost
	DivorcePenaltyPct         = 20          // percent of adena transferred on divorce
)

// Sentinel errors.
var (
	ErrAlreadyEngaged    = errors.New("already engaged or married")
	ErrSelfEngage        = errors.New("cannot engage yourself")
	ErrNotEngaged        = errors.New("not engaged or married")
	ErrNotMarried        = errors.New("not married")
	ErrCoupleNotFound    = errors.New("couple not found")
	ErrPartnerOffline    = errors.New("partner is offline")
	ErrInsufficientAdena = errors.New("not enough adena")
)

// CoupleStore persists couple data to the database.
// Implemented by db.CoupleRepository.
type CoupleStore interface {
	LoadAll(ctx context.Context) ([]CoupleRow, error)
	Create(ctx context.Context, p1ID, p2ID int32) (int32, error)
	UpdateMarried(ctx context.Context, coupleID int32) error
	Delete(ctx context.Context, coupleID int32) error
	FindByPlayer(ctx context.Context, playerID int32) (*CoupleRow, error)
}

// CoupleRow mirrors the database row for couples.
type CoupleRow struct {
	ID        int32
	Player1ID int32
	Player2ID int32
	Married   bool
}

// Manager manages all active couples on the server.
// Thread-safe: all mutable state protected by mu.
type Manager struct {
	mu sync.RWMutex

	// couples: coupleID → *model.Couple
	couples map[int32]*model.Couple

	// playerIndex: playerObjectID → coupleID (for fast lookup)
	playerIndex map[int32]int32

	store CoupleStore
}

// NewManager creates a new wedding manager.
func NewManager(store CoupleStore) *Manager {
	return &Manager{
		couples:     make(map[int32]*model.Couple),
		playerIndex: make(map[int32]int32),
		store:       store,
	}
}

// Init loads all couples from the database into memory.
func (m *Manager) Init(ctx context.Context) error {
	rows, err := m.store.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("load couples: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, row := range rows {
		c := &model.Couple{
			ID:        row.ID,
			Player1ID: row.Player1ID,
			Player2ID: row.Player2ID,
			Married:   row.Married,
		}
		m.couples[c.ID] = c
		m.playerIndex[c.Player1ID] = c.ID
		m.playerIndex[c.Player2ID] = c.ID
	}

	slog.Info("wedding manager initialized", "couples", len(m.couples))
	return nil
}

// Engage creates a new couple (engagement) between two players.
// Both players must not be already engaged or married.
// Returns the created couple.
func (m *Manager) Engage(ctx context.Context, p1, p2 *model.Player) (*model.Couple, error) {
	p1ID := int32(p1.ObjectID())
	p2ID := int32(p2.ObjectID())

	if p1ID == p2ID {
		return nil, ErrSelfEngage
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.playerIndex[p1ID]; ok {
		return nil, ErrAlreadyEngaged
	}
	if _, ok := m.playerIndex[p2ID]; ok {
		return nil, ErrAlreadyEngaged
	}

	coupleID, err := m.store.Create(ctx, p1ID, p2ID)
	if err != nil {
		return nil, fmt.Errorf("create couple: %w", err)
	}

	// Normalize order (DB CHECK ensures player1 < player2).
	minID, maxID := p1ID, p2ID
	if minID > maxID {
		minID, maxID = maxID, minID
	}

	c := &model.Couple{
		ID:        coupleID,
		Player1ID: minID,
		Player2ID: maxID,
		Married:   false,
	}

	m.couples[coupleID] = c
	m.playerIndex[minID] = coupleID
	m.playerIndex[maxID] = coupleID

	// Update player state.
	p1.SetPartnerID(p2ID)
	p1.SetCoupleID(coupleID)
	p2.SetPartnerID(p1ID)
	p2.SetCoupleID(coupleID)

	slog.Info("couple engaged",
		"coupleID", coupleID,
		"player1", minID,
		"player2", maxID)

	return c, nil
}

// Marry marks an existing couple as married (ceremony complete).
func (m *Manager) Marry(ctx context.Context, coupleID int32) error {
	m.mu.Lock()
	c, ok := m.couples[coupleID]
	if !ok {
		m.mu.Unlock()
		return ErrCoupleNotFound
	}
	if c.Married {
		m.mu.Unlock()
		return nil // already married, idempotent
	}
	c.Married = true
	m.mu.Unlock()

	if err := m.store.UpdateMarried(ctx, coupleID); err != nil {
		// Rollback in-memory state.
		m.mu.Lock()
		c.Married = false
		m.mu.Unlock()
		return fmt.Errorf("update married: %w", err)
	}

	slog.Info("couple married", "coupleID", coupleID)
	return nil
}

// Divorce removes a couple and clears both players' marriage state.
// p1 and p2 may be nil if one or both are offline — only online
// players will have their in-memory state cleared.
func (m *Manager) Divorce(ctx context.Context, coupleID int32, p1, p2 *model.Player) error {
	m.mu.Lock()
	c, ok := m.couples[coupleID]
	if !ok {
		m.mu.Unlock()
		return ErrCoupleNotFound
	}

	delete(m.couples, coupleID)
	delete(m.playerIndex, c.Player1ID)
	delete(m.playerIndex, c.Player2ID)
	m.mu.Unlock()

	if err := m.store.Delete(ctx, coupleID); err != nil {
		// Re-add to memory on failure.
		m.mu.Lock()
		m.couples[coupleID] = c
		m.playerIndex[c.Player1ID] = coupleID
		m.playerIndex[c.Player2ID] = coupleID
		m.mu.Unlock()
		return fmt.Errorf("delete couple: %w", err)
	}

	// Clear in-memory player state for online players.
	if p1 != nil {
		p1.ClearMarriageState()
	}
	if p2 != nil {
		p2.ClearMarriageState()
	}

	slog.Info("couple divorced",
		"coupleID", coupleID,
		"player1", c.Player1ID,
		"player2", c.Player2ID)

	return nil
}

// CoupleByPlayer returns the couple for a player, or nil if not found.
func (m *Manager) CoupleByPlayer(playerID int32) *model.Couple {
	m.mu.RLock()
	defer m.mu.RUnlock()

	coupleID, ok := m.playerIndex[playerID]
	if !ok {
		return nil
	}
	return m.couples[coupleID]
}

// Couple returns a couple by ID, or nil if not found.
func (m *Manager) Couple(coupleID int32) *model.Couple {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.couples[coupleID]
}

// CoupleCount returns the number of active couples.
func (m *Manager) CoupleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.couples)
}

// RestorePlayerState sets marriage fields on a player after login.
// Called during character loading when the player has a couple record.
func (m *Manager) RestorePlayerState(player *model.Player) {
	objectID := int32(player.ObjectID())

	m.mu.RLock()
	coupleID, ok := m.playerIndex[objectID]
	if !ok {
		m.mu.RUnlock()
		return
	}
	c := m.couples[coupleID]
	m.mu.RUnlock()

	if c == nil {
		return
	}

	player.SetCoupleID(c.ID)
	player.SetPartnerID(c.PartnerOf(objectID))
	player.SetMarried(c.Married)
}
