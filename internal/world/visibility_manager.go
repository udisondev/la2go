package world

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// VisibilityManager manages visibility cache updates for all players.
// Runs periodic batch updates every 100ms to reduce CPU overhead.
// Phase 4.5 PR3: -96.8% CPU reduction (10.6s → 0.34s per 10s interval @ 100K players).
// Phase 4.11 Tier 2 Opt 5: Parallel updates with worker pool for 10K+ players scalability.
type VisibilityManager struct {
	mu      sync.RWMutex
	players map[*model.Player]struct{} // registered players (set)

	interval time.Duration // update interval (default: 100ms)
	maxAge   time.Duration // cache considered stale after this duration (default: 200ms)

	world *World // reference to world grid for region queries

	// Phase 4.11 Tier 2 Opt 5: Worker pool configuration
	numWorkers int // number of parallel workers (default: runtime.NumCPU())
}

// NewVisibilityManager creates a new visibility manager.
// interval: how often to run batch updates (recommended: 100ms)
// maxAge: cache invalidation threshold (recommended: 200ms)
// Phase 4.11 Tier 2 Opt 5: Defaults to parallel updates with NumCPU() workers.
func NewVisibilityManager(world *World, interval, maxAge time.Duration) *VisibilityManager {
	return &VisibilityManager{
		players:    make(map[*model.Player]struct{}, 1000), // pre-allocate for 1K players
		interval:   interval,
		maxAge:     maxAge,
		world:      world,
		numWorkers: runtime.NumCPU(), // default: use all CPU cores
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

// SetNumWorkers sets number of parallel workers for batch updates.
// Phase 4.11 Tier 2 Opt 5: Configure worker pool size.
// Recommended: runtime.NumCPU() for balanced workload.
// Useful for tuning: fewer workers = less contention, more workers = better parallelism.
func (vm *VisibilityManager) SetNumWorkers(n int) {
	if n < 1 {
		n = 1
	}
	vm.numWorkers = n
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
// Phase 4.11 Tier 2 Opt 5: Parallel updates with worker pool for 100+ players.
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

	// Phase 4.11 Tier 2 Opt 5: Use parallel updates for 1000+ players
	// For small player counts (<1000), sequential is faster (no goroutine overhead)
	// Threshold tuned via benchmarks: 100 players = +154% regression, 1000 = +29% regression
	if playerCount < 1000 {
		vm.updateAllSequential(playerList, playerCount)
	} else {
		vm.updateAllParallel(playerList, playerCount)
	}
}

// updateAllSequential performs sequential update for small player counts (<100).
// Avoids goroutine overhead for small workloads.
func (vm *VisibilityManager) updateAllSequential(playerList []*model.Player, playerCount int) {
	updated := 0
	skipped := 0

	for _, player := range playerList {
		if vm.updatePlayerCache(player) {
			updated++
		} else {
			skipped++
		}
	}

	slog.Debug("Visibility batch update completed (sequential)",
		"players", playerCount,
		"updated", updated,
		"skipped", skipped)
}

// updateAllParallel performs parallel update using worker pool for large player counts (100+).
// Phase 4.11 Tier 2 Opt 5: -75% latency expected (27.7s → 7s @ 10K players).
func (vm *VisibilityManager) updateAllParallel(playerList []*model.Player, playerCount int) {
	var updated, skipped atomic.Int32

	// Calculate chunk size (divide players among workers)
	numWorkers := vm.numWorkers
	if numWorkers > playerCount {
		numWorkers = playerCount // don't create more workers than players
	}
	chunkSize := playerCount / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Spawn workers
	for i := range numWorkers {
		// Calculate chunk boundaries for this worker
		start := i * chunkSize
		end := start + chunkSize

		// Last worker takes remaining players (handles division remainder)
		if i == numWorkers-1 {
			end = playerCount
		}

		// Worker goroutine
		go func(chunk []*model.Player) {
			defer wg.Done()

			for _, player := range chunk {
				if vm.updatePlayerCache(player) {
					updated.Add(1)
				} else {
					skipped.Add(1)
				}
			}
		}(playerList[start:end])
	}

	// Wait for all workers to complete
	wg.Wait()

	slog.Debug("Visibility batch update completed (parallel)",
		"players", playerCount,
		"workers", numWorkers,
		"updated", updated.Load(),
		"skipped", skipped.Load())
}

// updatePlayerCache updates visibility cache for single player if needed.
// Returns true if cache was updated, false if skipped (cache still valid).
// Phase 4.11 Tier 3: Added fingerprint check to skip update if regions unchanged.
func (vm *VisibilityManager) updatePlayerCache(player *model.Player) bool {
	// Get player's current region
	loc := player.Location()
	regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)

	// Check if cache exists and is still valid
	cache := player.GetVisibilityCache()
	if cache != nil {
		// Phase 4.11 Tier 3: Skip if cache is fresh AND player in same region AND regions unchanged
		if !cache.IsStale(vm.maxAge) && cache.IsValidForRegion(regionX, regionY) {
			// Compute current fingerprint
			currentFP := vm.computeRegionFingerprint(regionX, regionY)
			if cache.RegionFingerprint() == currentFP {
				return false // SKIP: regions unchanged
			}
		}
	}

	// Phase 4.11 Tier 4: Collect visible objects split by LOD buckets
	// Ownership transferred to cache (no copy needed)
	near, medium, far := vm.getVisibleObjectsLOD(regionX, regionY)

	// Phase 4.11 Tier 3: Compute fingerprint for new cache
	fingerprint := vm.computeRegionFingerprint(regionX, regionY)

	// Create and store new cache (takes ownership of all slices)
	newCache := model.NewVisibilityCache(near, medium, far, regionX, regionY, fingerprint)
	player.SetVisibilityCache(newCache)

	return true
}

// computeRegionFingerprint computes XOR hash of 9 region versions.
// Phase 4.11 Tier 3: Used to detect if regions changed since last cache update.
// Returns 0 if region is nil (edge of map).
func (vm *VisibilityManager) computeRegionFingerprint(regionX, regionY int32) uint64 {
	fp := uint64(0)

	// Get current region
	currentRegion := vm.world.GetRegion(regionX, regionY)
	if currentRegion == nil {
		return 0
	}

	// XOR current region version
	fp = currentRegion.Version()

	// XOR surrounding region versions (with bit shift to avoid collision)
	for i, region := range currentRegion.SurroundingRegions() {
		if region == nil {
			continue
		}
		// Shift each region version by different amount to avoid XOR cancellation
		fp ^= region.Version() << uint(i*8%64)
	}

	return fp
}

// getVisibleObjectsLOD collects visible objects split by LOD (Level of Detail) buckets.
// Returns three slices: near (same region), medium (adjacent), far (distant).
// Phase 4.11 Tier 2: Use snapshot cache instead of sync.Map.Range (-70% latency).
// Phase 4.11 Tier 4: Split into LOD buckets for broadcast optimization (-30% expected).
func (vm *VisibilityManager) getVisibleObjectsLOD(regionX, regionY int32) (near, medium, far []*model.WorldObject) {
	// Pre-allocate for typical distribution: near=50, medium=200, far=200
	near = make([]*model.WorldObject, 0, 50)
	medium = make([]*model.WorldObject, 0, 200)
	far = make([]*model.WorldObject, 0, 200)

	// Get current region
	currentRegion := vm.world.GetRegion(regionX, regionY)
	if currentRegion == nil {
		return near, medium, far // empty slices
	}

	// Collect objects from current region (NEAR bucket — highest priority)
	snapshot := currentRegion.GetVisibleObjectsSnapshot()
	near = append(near, snapshot...)

	// Collect objects from surrounding regions (split into MEDIUM and FAR)
	// SurroundingRegions returns 3×3 grid (9 regions including center):
	// [0][1][2]
	// [3][4][5]  where [4] is center (already in near)
	// [6][7][8]
	//
	// Adjacent (MEDIUM): [1], [3], [5], [7] — share edge with center
	// Diagonal (FAR): [0], [2], [6], [8] — share only corner with center
	surroundingRegions := currentRegion.SurroundingRegions()
	for i, region := range surroundingRegions {
		if region == nil || region == currentRegion {
			continue // skip nil or center region
		}

		snapshot := region.GetVisibleObjectsSnapshot()

		// Determine if region is adjacent or diagonal
		// Adjacent indices: 1, 3, 5, 7 (share edge)
		// Diagonal indices: 0, 2, 6, 8 (share corner)
		isAdjacent := (i == 1 || i == 3 || i == 5 || i == 7)

		if isAdjacent {
			medium = append(medium, snapshot...)
		} else {
			far = append(far, snapshot...)
		}
	}

	return near, medium, far
}
