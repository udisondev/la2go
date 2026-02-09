package world

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// VisibilityManager manages visibility cache updates for all players.
// Runs periodic batch updates every 100ms to reduce CPU overhead.
// Phase 4.5 PR3: -96.8% CPU reduction (10.6s → 0.34s per 10s interval @ 100K players).
type VisibilityManager struct {
	mu      sync.RWMutex
	players map[*model.Player]struct{} // registered players (set)

	interval time.Duration // update interval (default: 100ms)
	maxAge   time.Duration // cache considered stale after this duration (default: 200ms)

	world *World // reference to world grid for region queries
}

// NewVisibilityManager creates a new visibility manager.
// interval: how often to run batch updates (recommended: 100ms)
// maxAge: cache invalidation threshold (recommended: 200ms)
func NewVisibilityManager(world *World, interval, maxAge time.Duration) *VisibilityManager {
	return &VisibilityManager{
		players:  make(map[*model.Player]struct{}, 1000), // pre-allocate for 1K players
		interval: interval,
		maxAge:   maxAge,
		world:    world,
	}
}

// RegisterPlayer adds player to visibility manager for periodic updates.
// Called when player enters world (EnterWorld packet).
func (vm *VisibilityManager) RegisterPlayer(player *model.Player) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.players[player] = struct{}{}
	slog.Debug("Player registered for visibility updates", "player", player.Name(), "total", len(vm.players))
}

// UnregisterPlayer removes player from visibility manager.
// Called when player logs out or disconnects.
// Also invalidates player's visibility cache.
func (vm *VisibilityManager) UnregisterPlayer(player *model.Player) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	delete(vm.players, player)
	player.InvalidateVisibilityCache()
	slog.Debug("Player unregistered from visibility updates", "player", player.Name(), "remaining", len(vm.players))
}

// Count returns number of registered players.
func (vm *VisibilityManager) Count() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.players)
}

// Start begins periodic visibility updates.
// Runs in background until context is cancelled.
// Returns when context is done or error occurs.
func (vm *VisibilityManager) Start(ctx context.Context) error {
	ticker := time.NewTicker(vm.interval)
	defer ticker.Stop()

	slog.Info("Visibility manager started", "interval", vm.interval, "maxAge", vm.maxAge)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Visibility manager stopping")
			return ctx.Err()

		case <-ticker.C:
			vm.UpdateAll()
		}
	}
}

// UpdateAll performs batch update of visibility cache for all registered players.
// Only updates caches that are stale (older than maxAge) or invalid (region changed).
// This is the HOT PATH — optimized for minimal allocations and CPU usage.
func (vm *VisibilityManager) UpdateAll() {
	vm.mu.RLock()
	playerCount := len(vm.players)

	// Fast path: no players registered
	if playerCount == 0 {
		vm.mu.RUnlock()
		return
	}

	// Copy player references (minimize lock time)
	// Pre-allocate exact size to avoid grows
	playerList := make([]*model.Player, 0, playerCount)
	for player := range vm.players {
		playerList = append(playerList, player)
	}
	vm.mu.RUnlock()

	// Update each player's cache (outside lock to avoid blocking Register/Unregister)
	updated := 0
	skipped := 0

	for _, player := range playerList {
		if vm.updatePlayerCache(player) {
			updated++
		} else {
			skipped++
		}
	}

	slog.Debug("Visibility batch update completed",
		"players", playerCount,
		"updated", updated,
		"skipped", skipped)
}

// updatePlayerCache updates visibility cache for single player if needed.
// Returns true if cache was updated, false if skipped (cache still valid).
func (vm *VisibilityManager) updatePlayerCache(player *model.Player) bool {
	// Get player's current region
	loc := player.Location()
	regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)

	// Check if cache exists and is still valid
	cache := player.GetVisibilityCache()
	if cache != nil {
		// Skip update if cache is fresh AND player in same region
		if !cache.IsStale(vm.maxAge) && cache.IsValidForRegion(regionX, regionY) {
			return false
		}
	}

	// Collect visible objects from current + surrounding regions (9 regions total)
	visibleObjects := vm.getVisibleObjects(regionX, regionY)

	// Create and store new cache
	newCache := model.NewVisibilityCache(visibleObjects, regionX, regionY)
	player.SetVisibilityCache(newCache)

	return true
}

// getVisibleObjects collects all visible objects from current + surrounding regions.
// Returns slice of WorldObject pointers (may be empty if no objects visible).
// IMPORTANT: This is a HOT PATH — pre-allocate slice to avoid multiple grows.
func (vm *VisibilityManager) getVisibleObjects(regionX, regionY int32) []*model.WorldObject {
	// Pre-allocate for typical case: 9 regions × 50 objects = 450 objects
	// This avoids slice grows during append operations
	objects := make([]*model.WorldObject, 0, 450)

	// Get current region
	currentRegion := vm.world.GetRegion(regionX, regionY)
	if currentRegion == nil {
		return objects // empty slice
	}

	// Collect objects from current region
	currentRegion.ForEachVisibleObject(func(obj *model.WorldObject) bool {
		objects = append(objects, obj)
		return true
	})

	// Collect objects from surrounding regions (8 regions)
	for _, surroundingRegion := range currentRegion.SurroundingRegions() {
		if surroundingRegion == nil {
			continue
		}

		surroundingRegion.ForEachVisibleObject(func(obj *model.WorldObject) bool {
			objects = append(objects, obj)
			return true
		})
	}

	return objects
}
