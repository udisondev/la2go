package party

import (
	"sync"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// Manager manages all active parties on the server.
// Thread-safe: uses RWMutex for party map and atomic for ID generation.
type Manager struct {
	mu      sync.RWMutex
	parties map[int32]*model.Party
	nextID  atomic.Int32
}

// NewManager creates a new party manager.
func NewManager() *Manager {
	m := &Manager{
		parties: make(map[int32]*model.Party),
	}
	return m
}

// CreateParty creates a new party with the given leader and loot rule.
// Returns the created party.
func (m *Manager) CreateParty(leader *model.Player, lootRule int32) *model.Party {
	id := m.nextID.Add(1)
	party := model.NewParty(id, leader, lootRule)

	m.mu.Lock()
	m.parties[id] = party
	m.mu.Unlock()

	return party
}

// DisbandParty removes a party by ID.
// Does NOT notify party members -- caller is responsible for sending packets.
func (m *Manager) DisbandParty(partyID int32) {
	m.mu.Lock()
	delete(m.parties, partyID)
	m.mu.Unlock()
}

// GetParty returns a party by ID, or nil if not found.
func (m *Manager) GetParty(partyID int32) *model.Party {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parties[partyID]
}

// PartyCount returns the number of active parties.
func (m *Manager) PartyCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.parties)
}
