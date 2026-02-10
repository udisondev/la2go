package model

import (
	"sync/atomic"
	"time"
)

// VisibilityCache caches visible objects for a player to avoid frequent region queries.
// Updated periodically by VisibilityManager (every 100ms).
// Immutable after creation — atomic.Value ensures safe concurrent reads.
// Phase 4.11 Tier 3: Added regionFingerprint for dirty region tracking.
type VisibilityCache struct {
	objects           []*WorldObject // visible objects (current + 8 surrounding regions)
	lastUpdate        time.Time      // when cache was last updated
	regionX           int32          // player's region X at cache time
	regionY           int32          // player's region Y at cache time
	regionFingerprint uint64         // XOR hash of 9 region versions (Tier 3)
}

// NewVisibilityCache creates a new visibility cache snapshot.
// IMPORTANT: Takes ownership of objects slice — caller MUST NOT modify it after calling.
// Phase 4.11 Tier 1: Eliminated copy to reduce allocations (-16.4% memory @ 10K players).
// Phase 4.11 Tier 3: Added regionFingerprint parameter for dirty tracking.
func NewVisibilityCache(objects []*WorldObject, regionX, regionY int32, regionFingerprint uint64) *VisibilityCache {
	// Transfer ownership: caller guarantees slice is not reused
	// No copy needed — slice is already isolated by getVisibleObjects()
	return &VisibilityCache{
		objects:           objects,
		lastUpdate:        time.Now(),
		regionX:           regionX,
		regionY:           regionY,
		regionFingerprint: regionFingerprint,
	}
}

// Objects returns cached visible objects (immutable slice).
// IMPORTANT: Do NOT modify returned slice — it's shared across goroutines.
func (c *VisibilityCache) Objects() []*WorldObject {
	return c.objects
}

// LastUpdate returns when cache was last updated.
func (c *VisibilityCache) LastUpdate() time.Time {
	return c.lastUpdate
}

// RegionX returns player's region X at cache time.
func (c *VisibilityCache) RegionX() int32 {
	return c.regionX
}

// RegionY returns player's region Y at cache time.
func (c *VisibilityCache) RegionY() int32 {
	return c.regionY
}

// IsStale returns true if cache is older than maxAge.
// Used by VisibilityManager to determine if update is needed.
func (c *VisibilityCache) IsStale(maxAge time.Duration) bool {
	return time.Since(c.lastUpdate) > maxAge
}

// IsValidForRegion returns true if cache was created for given region.
// If player moved to different region, cache must be invalidated.
func (c *VisibilityCache) IsValidForRegion(regionX, regionY int32) bool {
	return c.regionX == regionX && c.regionY == regionY
}

// RegionFingerprint returns XOR hash of 9 region versions at cache time.
// Phase 4.11 Tier 3: Used to skip cache update if regions unchanged.
func (c *VisibilityCache) RegionFingerprint() uint64 {
	return c.regionFingerprint
}

// PlayerVisibilityCache wraps atomic.Value for safe concurrent access to *VisibilityCache.
// Used by Player to store visibility cache.
type PlayerVisibilityCache struct {
	cache atomic.Value // stores *VisibilityCache
}

// NewPlayerVisibilityCache creates empty visibility cache for player.
func NewPlayerVisibilityCache() *PlayerVisibilityCache {
	pvc := &PlayerVisibilityCache{}
	pvc.cache.Store((*VisibilityCache)(nil)) // initialize with nil
	return pvc
}

// Get returns current visibility cache (may be nil if not initialized).
// Safe for concurrent reads.
func (pvc *PlayerVisibilityCache) Get() *VisibilityCache {
	v := pvc.cache.Load()
	if v == nil {
		return nil
	}
	return v.(*VisibilityCache)
}

// Set updates visibility cache atomically.
// Safe for concurrent writes (atomic.Value.Store is thread-safe).
func (pvc *PlayerVisibilityCache) Set(cache *VisibilityCache) {
	pvc.cache.Store(cache)
}

// Invalidate clears visibility cache (sets to nil).
// Called when player moves to different region or logs out.
func (pvc *PlayerVisibilityCache) Invalidate() {
	pvc.cache.Store((*VisibilityCache)(nil))
}
