package clan

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
)

// Clan creation limits.
const (
	MinClanNameLen = 2
	MaxClanNameLen = 16
	MinCreateLevel = 10 // Player must be level 10+ to create a clan
)

// Table errors.
var (
	ErrClanNameTaken   = errors.New("clan name already taken")
	ErrClanNameInvalid = errors.New("invalid clan name")
	ErrClanNotFound    = errors.New("clan not found")
)

// Table manages all clans on the server.
// Thread-safe: protected by RWMutex.
type Table struct {
	mu sync.RWMutex

	// Clans by ID.
	clans map[int32]*Clan

	// Clan name -> ID index (lowercase for case-insensitive lookup).
	nameIndex map[string]int32

	// Next clan ID counter.
	nextID atomic.Int32
}

// NewTable creates a new clan table.
func NewTable() *Table {
	return &Table{
		clans:     make(map[int32]*Clan, 128),
		nameIndex: make(map[string]int32, 128),
	}
}

// SetNextID sets the next clan ID counter (used on load from DB).
func (t *Table) SetNextID(id int32) {
	t.nextID.Store(id)
}

// Create creates a new clan and adds it to the table.
// Returns the new clan or an error if the name is invalid/taken.
func (t *Table) Create(name string, leaderID int64) (*Clan, error) {
	if err := validateClanName(name); err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	lowerName := strings.ToLower(name)
	if _, ok := t.nameIndex[lowerName]; ok {
		return nil, ErrClanNameTaken
	}

	id := t.nextID.Add(1)
	c := New(id, name, leaderID)

	t.clans[id] = c
	t.nameIndex[lowerName] = id

	slog.Info("clan created", "clan_id", id, "name", name, "leader_id", leaderID)
	return c, nil
}

// Register adds an existing clan to the table (used when loading from DB).
func (t *Table) Register(c *Clan) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	lowerName := strings.ToLower(c.Name())
	if _, ok := t.nameIndex[lowerName]; ok {
		return fmt.Errorf("register clan %q: %w", c.Name(), ErrClanNameTaken)
	}

	t.clans[c.ID()] = c
	t.nameIndex[lowerName] = c.ID()
	return nil
}

// Disband removes a clan from the table.
// Returns ErrClanNotFound if the clan doesn't exist.
func (t *Table) Disband(clanID int32) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	c, ok := t.clans[clanID]
	if !ok {
		return ErrClanNotFound
	}

	lowerName := strings.ToLower(c.Name())
	delete(t.clans, clanID)
	delete(t.nameIndex, lowerName)

	slog.Info("clan disbanded", "clan_id", clanID, "name", c.Name())
	return nil
}

// Clan returns a clan by ID, or nil if not found.
func (t *Table) Clan(id int32) *Clan {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.clans[id]
}

// ClanByName returns a clan by name (case-insensitive), or nil if not found.
func (t *Table) ClanByName(name string) *Clan {
	t.mu.RLock()
	defer t.mu.RUnlock()

	id, ok := t.nameIndex[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return t.clans[id]
}

// Count returns the number of registered clans.
func (t *Table) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.clans)
}

// ForEach iterates over all clans.
// Return false from fn to stop iteration.
func (t *Table) ForEach(fn func(*Clan) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, c := range t.clans {
		if !fn(c) {
			return
		}
	}
}

// AllyExists returns true if an alliance with the given name already exists (case-insensitive).
func (t *Table) AllyExists(allyName string) bool {
	lowerName := strings.ToLower(allyName)
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, c := range t.clans {
		if c.AllyID() != 0 && c.IsAllyLeader() && strings.ToLower(c.AllyName()) == lowerName {
			return true
		}
	}
	return false
}

// ClanAllies returns all clans belonging to the given alliance ID.
func (t *Table) ClanAllies(allyID int32) []*Clan {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*Clan
	for _, c := range t.clans {
		if c.AllyID() == allyID {
			result = append(result, c)
		}
	}
	return result
}

// ClanAllyCount returns the number of clans in the given alliance.
func (t *Table) ClanAllyCount(allyID int32) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, c := range t.clans {
		if c.AllyID() == allyID {
			count++
		}
	}
	return count
}

// validateClanName checks clan name constraints.
func validateClanName(name string) error {
	if len(name) < MinClanNameLen || len(name) > MaxClanNameLen {
		return fmt.Errorf("%w: length must be %d-%d", ErrClanNameInvalid, MinClanNameLen, MaxClanNameLen)
	}
	for _, r := range name {
		if !isValidClanNameChar(r) {
			return fmt.Errorf("%w: invalid character %q", ErrClanNameInvalid, r)
		}
	}
	return nil
}

// isValidClanNameChar returns true if the rune is allowed in a clan name.
// Allows: A-Z, a-z, 0-9.
func isValidClanNameChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
