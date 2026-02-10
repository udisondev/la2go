package model

import (
	"testing"
	"time"
)

func TestNewVisibilityCache(t *testing.T) {
	// Create test objects
	objects := []*WorldObject{
		NewWorldObject(1, "NPC1", Location{}),
		NewWorldObject(2, "NPC2", Location{}),
		NewWorldObject(3, "NPC3", Location{}),
	}

	cache := NewVisibilityCache(objects, 10, 20, 0)

	if cache == nil {
		t.Fatal("NewVisibilityCache returned nil")
	}

	if cache.RegionX() != 10 {
		t.Errorf("RegionX() = %d, want 10", cache.RegionX())
	}

	if cache.RegionY() != 20 {
		t.Errorf("RegionY() = %d, want 20", cache.RegionY())
	}

	cachedObjects := cache.Objects()
	if len(cachedObjects) != 3 {
		t.Errorf("Objects() length = %d, want 3", len(cachedObjects))
	}

	// Phase 4.11 Tier 1: Verify ownership transfer works correctly
	// Cache takes ownership of slice, so caller should not reuse it
	// Verify cache preserves original object references
	if cachedObjects[0].ObjectID() != 1 {
		t.Errorf("Cache object ID = %d, want 1", cachedObjects[0].ObjectID())
	}
	if cachedObjects[1].ObjectID() != 2 {
		t.Errorf("Cache object ID = %d, want 2", cachedObjects[1].ObjectID())
	}
	if cachedObjects[2].ObjectID() != 3 {
		t.Errorf("Cache object ID = %d, want 3", cachedObjects[2].ObjectID())
	}
}

// TestNewVisibilityCache_OwnershipTransfer verifies that cache takes ownership of slice.
// Phase 4.11 Tier 1: caller must not modify slice after passing to NewVisibilityCache.
func TestNewVisibilityCache_OwnershipTransfer(t *testing.T) {
	// Create test objects
	objects := []*WorldObject{
		NewWorldObject(1, "NPC1", Location{}),
		NewWorldObject(2, "NPC2", Location{}),
	}

	// Create cache (transfers ownership)
	cache := NewVisibilityCache(objects, 5, 10, 0)

	// Get cached objects
	cachedObjects := cache.Objects()
	if len(cachedObjects) != 2 {
		t.Fatalf("Objects() length = %d, want 2", len(cachedObjects))
	}

	// Verify cache has correct objects
	if cachedObjects[0].ObjectID() != 1 {
		t.Errorf("Cache object[0] ID = %d, want 1", cachedObjects[0].ObjectID())
	}
	if cachedObjects[1].ObjectID() != 2 {
		t.Errorf("Cache object[1] ID = %d, want 2", cachedObjects[1].ObjectID())
	}

	// IMPORTANT: Caller should NOT modify original slice after creating cache
	// This test documents the ownership transfer contract:
	// ❌ BAD: objects[0] = NewWorldObject(999, "Modified", Location{})
	// ✅ GOOD: caller discards objects reference immediately after NewVisibilityCache
}

func TestVisibilityCache_IsStale(t *testing.T) {
	cache := NewVisibilityCache([]*WorldObject{}, 0, 0, 0)

	// Fresh cache should not be stale
	if cache.IsStale(100 * time.Millisecond) {
		t.Error("Fresh cache reported as stale")
	}

	// Wait for cache to become stale
	time.Sleep(150 * time.Millisecond)

	if !cache.IsStale(100 * time.Millisecond) {
		t.Error("Old cache not reported as stale")
	}
}

func TestVisibilityCache_IsValidForRegion(t *testing.T) {
	cache := NewVisibilityCache([]*WorldObject{}, 10, 20, 0)

	tests := []struct {
		name     string
		regionX  int32
		regionY  int32
		expected bool
	}{
		{"same region", 10, 20, true},
		{"different X", 11, 20, false},
		{"different Y", 10, 21, false},
		{"both different", 11, 21, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cache.IsValidForRegion(tt.regionX, tt.regionY); got != tt.expected {
				t.Errorf("IsValidForRegion(%d, %d) = %v, want %v",
					tt.regionX, tt.regionY, got, tt.expected)
			}
		})
	}
}

func TestPlayer_VisibilityCache(t *testing.T) {
	player, err := NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Initially cache should be nil
	cache := player.GetVisibilityCache()
	if cache != nil {
		t.Errorf("Initial cache should be nil, got %v", cache)
	}

	// Set cache
	objects := []*WorldObject{
		NewWorldObject(1, "NPC1", Location{}),
		NewWorldObject(2, "NPC2", Location{}),
	}
	newCache := NewVisibilityCache(objects, 5, 5, 0)
	player.SetVisibilityCache(newCache)

	// Verify cache was set
	retrievedCache := player.GetVisibilityCache()
	if retrievedCache == nil {
		t.Fatal("Cache not set")
	}

	if retrievedCache.RegionX() != 5 {
		t.Errorf("Cache RegionX() = %d, want 5", retrievedCache.RegionX())
	}

	if len(retrievedCache.Objects()) != 2 {
		t.Errorf("Cache Objects() length = %d, want 2", len(retrievedCache.Objects()))
	}

	// Invalidate cache
	player.InvalidateVisibilityCache()

	invalidatedCache := player.GetVisibilityCache()
	if invalidatedCache != nil {
		t.Errorf("Cache should be nil after invalidation, got %v", invalidatedCache)
	}
}

func TestPlayer_VisibilityCache_Concurrent(t *testing.T) {
	player, err := NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Concurrent reads/writes
	done := make(chan bool)

	// Writers
	for range 10 {
		go func() {
			for range 100 {
				objects := []*WorldObject{
					NewWorldObject(1, "NPC", Location{}),
				}
				cache := NewVisibilityCache(objects, 0, 0, 0)
				player.SetVisibilityCache(cache)
			}
			done <- true
		}()
	}

	// Readers
	for range 20 {
		go func() {
			for range 1000 {
				_ = player.GetVisibilityCache()
			}
			done <- true
		}()
	}

	// Invalidators
	for range 5 {
		go func() {
			for range 50 {
				player.InvalidateVisibilityCache()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 35 {
		<-done
	}
}
