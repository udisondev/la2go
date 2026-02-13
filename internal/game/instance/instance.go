// Package instance implements the Lineage 2 instance zone system.
// Instances are private copies of game zones for individual players or parties,
// used by Dimensional Rift, quest dungeons, and other instanced content.
//
// Phase 26: Instance Zones.
package instance

import (
	"sync"
	"sync/atomic"
	"time"
)

// State represents the lifecycle state of an instance.
type State int32

const (
	StateCreated    State = iota // Instance created, not yet active
	StateActive                 // Instance is active, players can enter
	StateDestroying             // Destroying (cleanup in progress)
	StateDestroyed              // Fully destroyed
)

// String returns a human-readable state name.
func (s State) String() string {
	switch s {
	case StateCreated:
		return "CREATED"
	case StateActive:
		return "ACTIVE"
	case StateDestroying:
		return "DESTROYING"
	case StateDestroyed:
		return "DESTROYED"
	default:
		return "UNKNOWN"
	}
}

// Instance represents a private copy of a game zone.
// Thread-safe for concurrent access.
type Instance struct {
	mu sync.RWMutex

	id         int32
	templateID int32
	ownerID    uint32 // objectID of the creator
	createdAt  time.Time
	duration   time.Duration // max lifetime (0 = no limit)
	state      atomic.Int32  // State

	// Players currently inside (objectID set).
	players map[uint32]struct{}

	// NPCs spawned in this instance (objectID set).
	npcs map[uint32]struct{}

	// Destroy timer — fires when instance is empty for emptyTimeout.
	emptyTimer  *time.Timer
	emptyDelay  time.Duration // delay before destroying empty instance
	cancelEmpty chan struct{} // signals to cancel empty timer
}

// DefaultEmptyDelay is the default time before an empty instance is destroyed.
const DefaultEmptyDelay = 5 * time.Minute

// NewInstance creates a new instance from a template.
func NewInstance(id, templateID int32, ownerID uint32, duration time.Duration) *Instance {
	inst := &Instance{
		id:          id,
		templateID:  templateID,
		ownerID:     ownerID,
		createdAt:   time.Now(),
		duration:    duration,
		players:     make(map[uint32]struct{}, 8),
		npcs:        make(map[uint32]struct{}, 32),
		emptyDelay:  DefaultEmptyDelay,
		cancelEmpty: make(chan struct{}),
	}
	inst.state.Store(int32(StateCreated))
	return inst
}

// ID returns the unique instance identifier.
func (i *Instance) ID() int32 { return i.id }

// TemplateID returns the template this instance was created from.
func (i *Instance) TemplateID() int32 { return i.templateID }

// OwnerID returns the objectID of the instance creator.
func (i *Instance) OwnerID() uint32 { return i.ownerID }

// CreatedAt returns when the instance was created.
func (i *Instance) CreatedAt() time.Time { return i.createdAt }

// Duration returns the max lifetime of this instance.
func (i *Instance) Duration() time.Duration { return i.duration }

// State returns the current lifecycle state.
func (i *Instance) State() State { return State(i.state.Load()) }

// SetState sets the lifecycle state.
func (i *Instance) SetState(s State) { i.state.Store(int32(s)) }

// EmptyDelay returns the delay before destroying an empty instance.
func (i *Instance) EmptyDelay() time.Duration {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.emptyDelay
}

// SetEmptyDelay configures the delay before destroying an empty instance.
func (i *Instance) SetEmptyDelay(d time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.emptyDelay = d
}

// IsExpired returns true if the instance has exceeded its max lifetime.
func (i *Instance) IsExpired() bool {
	if i.duration <= 0 {
		return false
	}
	return time.Since(i.createdAt) > i.duration
}

// AddPlayer adds a player to this instance.
// Returns false if the player is already inside.
func (i *Instance) AddPlayer(objectID uint32) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.players[objectID]; ok {
		return false
	}
	i.players[objectID] = struct{}{}

	// Отменяем таймер уничтожения пустого инстанса.
	i.cancelEmptyTimerLocked()

	return true
}

// RemovePlayer removes a player from this instance.
// Returns true if the instance is now empty (should be scheduled for destruction).
func (i *Instance) RemovePlayer(objectID uint32) (removed bool, empty bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.players[objectID]; !ok {
		return false, false
	}
	delete(i.players, objectID)
	return true, len(i.players) == 0
}

// HasPlayer returns true if the player is in this instance.
func (i *Instance) HasPlayer(objectID uint32) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	_, ok := i.players[objectID]
	return ok
}

// PlayerCount returns the number of players in this instance.
func (i *Instance) PlayerCount() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.players)
}

// Players returns a copy of all player objectIDs.
func (i *Instance) Players() []uint32 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ids := make([]uint32, 0, len(i.players))
	for id := range i.players {
		ids = append(ids, id)
	}
	return ids
}

// AddNpc registers an NPC objectID in this instance.
func (i *Instance) AddNpc(objectID uint32) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.npcs[objectID] = struct{}{}
}

// RemoveNpc removes an NPC from this instance.
func (i *Instance) RemoveNpc(objectID uint32) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.npcs, objectID)
}

// HasNpc returns true if the NPC is in this instance.
func (i *Instance) HasNpc(objectID uint32) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	_, ok := i.npcs[objectID]
	return ok
}

// NpcCount returns the number of NPCs in this instance.
func (i *Instance) NpcCount() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.npcs)
}

// Npcs returns a copy of all NPC objectIDs.
func (i *Instance) Npcs() []uint32 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ids := make([]uint32, 0, len(i.npcs))
	for id := range i.npcs {
		ids = append(ids, id)
	}
	return ids
}

// cancelEmptyTimerLocked cancels the empty-instance destroy timer.
// Must be called with mu held.
func (i *Instance) cancelEmptyTimerLocked() {
	select {
	case <-i.cancelEmpty:
		// Уже закрыт — пересоздаём.
		i.cancelEmpty = make(chan struct{})
	default:
		close(i.cancelEmpty)
		i.cancelEmpty = make(chan struct{})
	}
	if i.emptyTimer != nil {
		i.emptyTimer.Stop()
		i.emptyTimer = nil
	}
}

// CancelEmpty returns the cancel channel for the empty timer.
// Used by Manager to coordinate destruction.
func (i *Instance) CancelEmpty() <-chan struct{} {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.cancelEmpty
}

// SetEmptyTimer stores a reference to the empty-destroy timer.
func (i *Instance) SetEmptyTimer(t *time.Timer) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.emptyTimer = t
}
