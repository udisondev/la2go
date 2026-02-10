package world

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

func TestNewVisibilityManager(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	if vm == nil {
		t.Fatal("NewVisibilityManager returned nil")
	}

	if vm.Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", vm.Count())
	}
}

func TestVisibilityManager_RegisterUnregister(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	player1, _ := model.NewPlayer(1, 1, "Player1", 10, 0, 1)
	player2, _ := model.NewPlayer(2, 1, "Player2", 10, 0, 1)

	// Register players
	vm.RegisterPlayer(player1)
	if vm.Count() != 1 {
		t.Errorf("After register player1, Count() = %d, want 1", vm.Count())
	}

	vm.RegisterPlayer(player2)
	if vm.Count() != 2 {
		t.Errorf("After register player2, Count() = %d, want 2", vm.Count())
	}

	// Unregister player
	vm.UnregisterPlayer(player1)
	if vm.Count() != 1 {
		t.Errorf("After unregister player1, Count() = %d, want 1", vm.Count())
	}

	// Verify cache invalidated on unregister
	if player1.GetVisibilityCache() != nil {
		t.Error("Player1 cache should be nil after unregister")
	}
}

func TestVisibilityManager_UpdateAll(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	// Create test player at known location
	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0) // center of region
	player.SetLocation(loc)

	// Add some test objects to world
	regionX, regionY := CoordToRegionIndex(150000, 150000)
	region := world.GetRegion(regionX, regionY)
	if region != nil {
		for i := range 5 {
			obj := model.NewWorldObject(uint32(i+1), "TestNPC", loc)
			region.AddVisibleObject(obj)
		}
	}

	// Register player
	vm.RegisterPlayer(player)

	// Initially cache should be nil
	if player.GetVisibilityCache() != nil {
		t.Error("Initial cache should be nil")
	}

	// Run UpdateAll
	vm.UpdateAll()

	// Verify cache was created
	cache := player.GetVisibilityCache()
	if cache == nil {
		t.Fatal("Cache not created after UpdateAll")
	}

	// Verify cache contains objects
	objects := cache.Objects()
	if len(objects) < 5 {
		t.Errorf("Cache Objects() length = %d, want at least 5", len(objects))
	}

	// Verify cache region matches player location
	if cache.RegionX() != regionX || cache.RegionY() != regionY {
		t.Errorf("Cache region (%d,%d) doesn't match player region (%d,%d)",
			cache.RegionX(), cache.RegionY(), regionX, regionY)
	}
}

func TestVisibilityManager_UpdateAll_SkipFreshCache(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	regionX, regionY := CoordToRegionIndex(150000, 150000)

	// Phase 4.11 Tier 3: Compute correct fingerprint for cache
	fingerprint := vm.computeRegionFingerprint(regionX, regionY)

	// Manually set fresh cache with correct fingerprint
	freshCache := model.NewVisibilityCache([]*model.WorldObject{}, nil, nil, regionX, regionY, fingerprint)
	player.SetVisibilityCache(freshCache)

	vm.RegisterPlayer(player)

	// Run UpdateAll
	vm.UpdateAll()

	// Verify cache wasn't replaced (same instance)
	currentCache := player.GetVisibilityCache()
	if currentCache == nil {
		t.Fatal("Cache should still exist")
	}

	// Phase 4.11 Tier 3: Fresh cache with correct fingerprint should NOT be updated
	if currentCache.LastUpdate() != freshCache.LastUpdate() {
		t.Error("Fresh cache was updated when it shouldn't be")
	}
}

func TestVisibilityManager_UpdateAll_InvalidateOnRegionChange(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	loc1 := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc1)

	regionX1, regionY1 := CoordToRegionIndex(150000, 150000)

	// Set cache for region 1
	oldCache := model.NewVisibilityCache([]*model.WorldObject{}, nil, nil, regionX1, regionY1, 0)
	player.SetVisibilityCache(oldCache)

	vm.RegisterPlayer(player)

	// Move player to different region
	loc2 := model.NewLocation(160000, 160000, 0, 0)
	player.SetLocation(loc2)

	regionX2, regionY2 := CoordToRegionIndex(160000, 160000)

	// Verify player actually moved to different region
	if regionX1 == regionX2 && regionY1 == regionY2 {
		t.Skip("Test locations are in same region, skipping")
	}

	// Run UpdateAll
	vm.UpdateAll()

	// Verify cache was updated (different region)
	newCache := player.GetVisibilityCache()
	if newCache == nil {
		t.Fatal("Cache should be updated")
	}

	if newCache.RegionX() == regionX1 || newCache.RegionY() == regionY1 {
		t.Error("Cache should be updated for new region")
	}
}

func TestVisibilityManager_Start_Stop(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 50*time.Millisecond, 100*time.Millisecond)

	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	vm.RegisterPlayer(player)

	// Start manager with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- vm.Start(ctx)
	}()

	// Wait for at least 2 ticks (100ms)
	time.Sleep(120 * time.Millisecond)

	// Cancel context to stop manager
	cancel()

	// Wait for manager to stop
	err := <-done
	if err != context.Canceled && err != context.DeadlineExceeded {
		t.Errorf("Start() error = %v, want context.Canceled or DeadlineExceeded", err)
	}

	// Verify cache was updated at least once
	cache := player.GetVisibilityCache()
	if cache == nil {
		t.Error("Cache should be created during periodic updates")
	}
}

// TestVisibilityManager_UpdateAll_SkipUnchangedRegions tests fingerprint-based skip logic.
// Phase 4.11 Tier 3: Cache should be skipped if regions unchanged (80% skip rate expected).
func TestVisibilityManager_UpdateAll_SkipUnchangedRegions(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	regionX, regionY := CoordToRegionIndex(150000, 150000)

	// Compute initial fingerprint
	initialFP := vm.computeRegionFingerprint(regionX, regionY)

	// Set fresh cache with initial fingerprint
	freshCache := model.NewVisibilityCache([]*model.WorldObject{}, nil, nil, regionX, regionY, initialFP)
	player.SetVisibilityCache(freshCache)

	vm.RegisterPlayer(player)

	// Run UpdateAll (should skip because regions unchanged)
	vm.UpdateAll()

	currentCache := player.GetVisibilityCache()
	if currentCache.LastUpdate() != freshCache.LastUpdate() {
		t.Error("Cache was updated when regions unchanged")
	}

	// Now add object to region (changes fingerprint)
	region := world.GetRegion(regionX, regionY)
	obj := model.NewWorldObject(100, "TestNPC", model.Location{})
	region.AddVisibleObject(obj)

	// Run UpdateAll (should update because fingerprint changed)
	vm.UpdateAll()

	currentCache = player.GetVisibilityCache()
	if currentCache.LastUpdate() == freshCache.LastUpdate() {
		t.Error("Cache was NOT updated when region changed")
	}

	// Verify new cache has updated fingerprint
	newFP := vm.computeRegionFingerprint(regionX, regionY)
	if currentCache.RegionFingerprint() != newFP {
		t.Errorf("Cache fingerprint = %d, want %d", currentCache.RegionFingerprint(), newFP)
	}
}

func TestVisibilityManager_Concurrent(t *testing.T) {
	world := Instance()
	vm := NewVisibilityManager(world, 10*time.Millisecond, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start manager
	go vm.Start(ctx)

	// Concurrent register/unregister
	done := make(chan bool)

	for i := range 10 {
		go func(id int) {
			for j := range 10 {
				player, _ := model.NewPlayer(int64(id*100+j), 1, "Player", 10, 0, 1)
				loc := model.NewLocation(150000, 150000, 0, 0)
				player.SetLocation(loc)

				vm.RegisterPlayer(player)
				time.Sleep(5 * time.Millisecond)
				vm.UnregisterPlayer(player)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	cancel()
}
