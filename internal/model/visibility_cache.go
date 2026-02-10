package model

import (
	"sync/atomic"
	"time"
)

// VisibilityCache caches visible objects for a player to avoid frequent region queries.
// Updated periodically by VisibilityManager (every 100ms).
// Immutable after creation — atomic.Value ensures safe concurrent reads.
// Phase 4.11 Tier 3: Added regionFingerprint for dirty region tracking.
// Phase 4.11 Tier 4: Split into LOD buckets (near/medium/far) for broadcast optimization.
type VisibilityCache struct {
	// Phase 4.11 Tier 4: LOD buckets (Level of Detail)
	nearObjects   []*WorldObject // same region (highest priority, ~50 objects)
	mediumObjects []*WorldObject // adjacent regions (medium priority, ~200 objects)
	farObjects    []*WorldObject // distant regions (low priority, ~200 objects)

	lastUpdate        time.Time // when cache was last updated
	regionX           int32     // player's region X at cache time
	regionY           int32     // player's region Y at cache time
	regionFingerprint uint64    // XOR hash of 9 region versions (Tier 3)
}

// NewVisibilityCache creates a new visibility cache snapshot with LOD buckets.
// IMPORTANT: Takes ownership of all slices — caller MUST NOT modify them after calling.
// Phase 4.11 Tier 1: Eliminated copy to reduce allocations (-16.4% memory @ 10K players).
// Phase 4.11 Tier 3: Added regionFingerprint parameter for dirty tracking.
// Phase 4.11 Tier 4: Split into near/medium/far LOD buckets for broadcast optimization.
func NewVisibilityCache(nearObjects, mediumObjects, farObjects []*WorldObject, regionX, regionY int32, regionFingerprint uint64) *VisibilityCache {
	// Transfer ownership: caller guarantees slices are not reused
	// No copy needed — slices are already isolated by getVisibleObjects()
	return &VisibilityCache{
		nearObjects:       nearObjects,
		mediumObjects:     mediumObjects,
		farObjects:        farObjects,
		lastUpdate:        time.Now(),
		regionX:           regionX,
		regionY:           regionY,
		regionFingerprint: regionFingerprint,
	}
}

// Objects returns all cached visible objects (near + medium + far combined).
// IMPORTANT: Do NOT modify returned slice — it's shared across goroutines.
// Phase 4.11 Tier 4: Combines all LOD buckets for backward compatibility.
func (c *VisibilityCache) Objects() []*WorldObject {
	// Combine all buckets (lazy allocation, done once per query)
	total := len(c.nearObjects) + len(c.mediumObjects) + len(c.farObjects)
	result := make([]*WorldObject, 0, total)
	result = append(result, c.nearObjects...)
	result = append(result, c.mediumObjects...)
	result = append(result, c.farObjects...)
	return result
}

// NearObjects returns objects in same region (highest priority).
// Phase 4.11 Tier 4: LOD near bucket (~50 objects).
func (c *VisibilityCache) NearObjects() []*WorldObject {
	return c.nearObjects
}

// MediumObjects returns objects in adjacent regions (medium priority).
// Phase 4.11 Tier 4: LOD medium bucket (~200 objects).
func (c *VisibilityCache) MediumObjects() []*WorldObject {
	return c.mediumObjects
}

// FarObjects returns objects in distant regions (low priority).
// Phase 4.11 Tier 4: LOD far bucket (~200 objects).
func (c *VisibilityCache) FarObjects() []*WorldObject {
	return c.farObjects
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
