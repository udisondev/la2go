package offlinetrade

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// TradeEntry represents a single item in an offline trader's store.
type TradeEntry struct {
	ItemIdentifier int32 // itemID for BUY, objectID for SELL
	Count          int64
	Price          int64
}

// Trader represents a player in offline trade mode.
// The player's WorldObject remains in the world, but the TCP connection is closed.
type Trader struct {
	CharacterID int64
	ObjectID    uint32
	AccountName string
	StoreType   model.PrivateStoreType
	Title       string
	Items       []TradeEntry
	StartedAt   time.Time
	expireTimer *time.Timer
}

// Table manages offline traders — players who disconnected while
// having an active private store (sell/buy/package).
//
// Java reference: OfflineTraderTable.java
type Table struct {
	mu      sync.RWMutex
	traders map[uint32]*Trader // objectID → trader
	byChar  map[int64]uint32   // characterID → objectID (for reconnection lookup)
	count   atomic.Int32
	maxDur  time.Duration
	onExpire func(objectID uint32) // callback when trader expires
}

// NewTable creates a new offline trader table.
// maxDuration controls how long a trader can stay offline (0 = unlimited).
func NewTable(maxDuration time.Duration) *Table {
	return &Table{
		traders: make(map[uint32]*Trader, 16),
		byChar:  make(map[int64]uint32, 16),
		maxDur:  maxDuration,
	}
}

// SetExpireCallback sets the function called when a trader's time expires.
// The callback receives the objectID of the expired trader.
func (t *Table) SetExpireCallback(fn func(objectID uint32)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onExpire = fn
}

// Add registers a player as an offline trader.
// Returns error if the player is already registered.
func (t *Table) Add(trader *Trader) error {
	if trader == nil {
		return fmt.Errorf("trader is nil")
	}
	if trader.ObjectID == 0 {
		return fmt.Errorf("trader objectID is zero")
	}
	if !trader.StoreType.IsInStoreMode() {
		return fmt.Errorf("invalid store type for offline trade: %v", trader.StoreType)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.traders[trader.ObjectID]; exists {
		return fmt.Errorf("trader %d already registered", trader.ObjectID)
	}

	t.traders[trader.ObjectID] = trader
	t.byChar[trader.CharacterID] = trader.ObjectID
	t.count.Add(1)

	// Запускаем таймер автоудаления (если настроен maxDuration)
	if t.maxDur > 0 {
		remaining := t.maxDur - time.Since(trader.StartedAt)
		if remaining <= 0 {
			// Уже истёк — удаляем сразу
			delete(t.traders, trader.ObjectID)
			delete(t.byChar, trader.CharacterID)
			t.count.Add(-1)
			return fmt.Errorf("trader %d already expired", trader.ObjectID)
		}
		trader.expireTimer = time.AfterFunc(remaining, func() {
			t.handleExpire(trader.ObjectID)
		})
	}

	slog.Info("offline trader registered",
		"objectID", trader.ObjectID,
		"characterID", trader.CharacterID,
		"storeType", trader.StoreType,
		"title", trader.Title)

	return nil
}

// Remove unregisters an offline trader.
// Returns the removed Trader or nil if not found.
func (t *Table) Remove(objectID uint32) *Trader {
	t.mu.Lock()
	defer t.mu.Unlock()

	trader, exists := t.traders[objectID]
	if !exists {
		return nil
	}

	if trader.expireTimer != nil {
		trader.expireTimer.Stop()
	}

	delete(t.traders, objectID)
	delete(t.byChar, trader.CharacterID)
	t.count.Add(-1)

	slog.Info("offline trader removed",
		"objectID", objectID,
		"characterID", trader.CharacterID)

	return trader
}

// RemoveByCharacter unregisters an offline trader by character ID.
// Used when the same account reconnects.
func (t *Table) RemoveByCharacter(characterID int64) *Trader {
	t.mu.RLock()
	objectID, exists := t.byChar[characterID]
	t.mu.RUnlock()

	if !exists {
		return nil
	}
	return t.Remove(objectID)
}

// Get returns the offline trader for the given objectID, or nil.
func (t *Table) Get(objectID uint32) *Trader {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.traders[objectID]
}

// IsOfflineTrader returns true if the objectID belongs to an offline trader.
func (t *Table) IsOfflineTrader(objectID uint32) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.traders[objectID]
	return exists
}

// IsCharacterOffline returns true if the character is offline trading.
func (t *Table) IsCharacterOffline(characterID int64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.byChar[characterID]
	return exists
}

// Count returns the number of active offline traders.
func (t *Table) Count() int {
	return int(t.count.Load())
}

// ForEach iterates over all offline traders.
// The callback receives a copy of the Trader.
// Return false from fn to stop iteration.
func (t *Table) ForEach(fn func(trader *Trader) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, trader := range t.traders {
		if !fn(trader) {
			return
		}
	}
}

// ExportAll returns a snapshot of all offline traders for DB persistence.
func (t *Table) ExportAll() []*Trader {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*Trader, 0, len(t.traders))
	for _, trader := range t.traders {
		result = append(result, trader)
	}
	return result
}

// UpdateTraderItems updates the items list for an offline trader.
// Called after a transaction (someone bought/sold from the offline store).
func (t *Table) UpdateTraderItems(objectID uint32, items []TradeEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	trader, exists := t.traders[objectID]
	if !exists {
		return
	}
	trader.Items = items
}

// RemoveIfEmpty removes the trader if their item list is empty.
// Returns true if the trader was removed.
func (t *Table) RemoveIfEmpty(objectID uint32) bool {
	t.mu.Lock()

	trader, exists := t.traders[objectID]
	if !exists {
		t.mu.Unlock()
		return false
	}

	if len(trader.Items) > 0 {
		t.mu.Unlock()
		return false
	}

	if trader.expireTimer != nil {
		trader.expireTimer.Stop()
	}
	delete(t.traders, objectID)
	delete(t.byChar, trader.CharacterID)
	t.count.Add(-1)
	t.mu.Unlock()

	slog.Info("offline trader removed (store empty)",
		"objectID", objectID,
		"characterID", trader.CharacterID)

	return true
}

// StopAll stops all expire timers and clears the table.
// Called during graceful shutdown after saving to DB.
func (t *Table) StopAll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, trader := range t.traders {
		if trader.expireTimer != nil {
			trader.expireTimer.Stop()
		}
	}

	t.traders = make(map[uint32]*Trader, 16)
	t.byChar = make(map[int64]uint32, 16)
	t.count.Store(0)
}

func (t *Table) handleExpire(objectID uint32) {
	trader := t.Remove(objectID)
	if trader == nil {
		return
	}

	slog.Info("offline trader expired",
		"objectID", objectID,
		"characterID", trader.CharacterID,
		"duration", time.Since(trader.StartedAt))

	t.mu.RLock()
	fn := t.onExpire
	t.mu.RUnlock()

	if fn != nil {
		fn(objectID)
	}
}
