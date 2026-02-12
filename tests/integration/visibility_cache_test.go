package integration

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestVisibilityCache_EndToEnd verifies full visibility cache workflow.
// 1. Create VisibilityManager and register player
// 2. Start manager (runs batch updates every 50ms)
// 3. Verify cache is populated after first update
// 4. Verify cache is used on subsequent queries (fast path)
// 5. Move player to different region, verify cache invalidated and recreated
func TestVisibilityCache_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		worldInstance := world.Instance()

		// Create player at known location
		playerOID := nextOID()
		player, err := model.NewPlayer(playerOID, int64(playerOID), 1, "TestPlayer", 10, 0, 1)
		if err != nil {
			t.Fatalf("NewPlayer failed: %v", err)
		}

		loc1 := model.NewLocation(150000, 150000, 0, 0)
		player.SetLocation(loc1)

		// Add test objects to world
		regionX1, regionY1 := world.CoordToRegionIndex(150000, 150000)
		region1 := worldInstance.GetRegionByIndex(regionX1, regionY1)
		if region1 == nil {
			t.Fatal("Region not found")
		}

		objectCount := 10
		for range objectCount {
			obj := model.NewWorldObject(nextOID(), "TestNPC", loc1)
			region1.AddVisibleObject(obj)
		}

		// Create and start VisibilityManager
		vm := world.NewVisibilityManager(worldInstance, 50*time.Millisecond, 100*time.Millisecond)
		vm.RegisterPlayer(player)

		// Start manager in background
		go vm.Start(ctx)

		// Wait for first batch update (instant with fake clock)
		time.Sleep(100 * time.Millisecond)

		// Verify cache was created
		cache := player.GetVisibilityCache()
		if cache == nil {
			t.Fatal("Cache not created after batch update")
		}

		// Verify cache contains objects (at least the ones we added)
		objects := cache.Objects()
		if len(objects) < objectCount {
			t.Errorf("Cache Objects() length = %d, want at least %d", len(objects), objectCount)
		}

		// Verify cache region matches player location
		if cache.RegionX() != regionX1 || cache.RegionY() != regionY1 {
			t.Errorf("Cache region (%d,%d) doesn't match player region (%d,%d)",
				cache.RegionX(), cache.RegionY(), regionX1, regionY1)
		}

		// Test fast path: use cached results
		cachedCount := 0
		world.ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
			cachedCount++
			return true
		})

		if cachedCount < objectCount {
			t.Errorf("Cached query returned %d objects, want at least %d", cachedCount, objectCount)
		}

		// Move player to different region
		loc2 := model.NewLocation(160000, 160000, 0, 0)
		player.SetLocation(loc2)

		regionX2, regionY2 := world.CoordToRegionIndex(160000, 160000)

		// Add objects to new region
		region2 := worldInstance.GetRegionByIndex(regionX2, regionY2)
		if region2 != nil {
			for range 5 {
				obj := model.NewWorldObject(nextOID(), "NPC2", loc2)
				region2.AddVisibleObject(obj)
			}
		}

		// Wait for cache update (instant with fake clock)
		time.Sleep(150 * time.Millisecond)

		// Verify cache was updated for new region
		newCache := player.GetVisibilityCache()
		if newCache == nil {
			t.Fatal("Cache should be updated for new region")
		}

		if newCache.RegionX() != regionX2 || newCache.RegionY() != regionY2 {
			t.Errorf("Cache not updated for new region: got (%d,%d), want (%d,%d)",
				newCache.RegionX(), newCache.RegionY(), regionX2, regionY2)
		}

		// Stop VM to avoid race
		cancel()
		time.Sleep(20 * time.Millisecond)

		// Cleanup
		vm.UnregisterPlayer(player)

		// Verify cache invalidated on unregister
		if player.GetVisibilityCache() != nil {
			t.Error("Cache should be nil after unregister")
		}
	})
}

// TestVisibilityCache_MultiplePlayersStressTest verifies batch update with multiple players.
// Simulates realistic scenario with 100 players moving around world.
func TestVisibilityCache_MultiplePlayersStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		worldInstance := world.Instance()

		// Create 100 players
		playerCount := 100
		players := make([]*model.Player, playerCount)

		for i := range playerCount {
			oid := nextOID()
			player, err := model.NewPlayer(oid, int64(oid), 1, "Player", 10, 0, 1)
			if err != nil {
				t.Fatalf("NewPlayer %d failed: %v", i, err)
			}

			// Distribute players across world
			loc := model.NewLocation(150000+int32(i*100), 150000+int32(i*100), 0, 0)
			player.SetLocation(loc)
			players[i] = player
		}

		// Populate world with 500 objects
		for i := range 500 {
			loc := model.NewLocation(150000+int32(i*50), 150000+int32(i*50), 0, 0)
			obj := model.NewWorldObject(nextOID(), "NPC", loc)

			regionX, regionY := world.CoordToRegionIndex(loc.X, loc.Y)
			region := worldInstance.GetRegionByIndex(regionX, regionY)
			if region != nil {
				region.AddVisibleObject(obj)
			}
		}

		// Create and start VisibilityManager
		vm := world.NewVisibilityManager(worldInstance, 50*time.Millisecond, 100*time.Millisecond)

		// Register all players
		for _, player := range players {
			vm.RegisterPlayer(player)
		}

		// Start manager
		go vm.Start(ctx)

		// Wait for several batch updates (instant with fake clock)
		time.Sleep(500 * time.Millisecond)

		// Verify all players have cache
		cachedCount := 0
		for _, player := range players {
			if player.GetVisibilityCache() != nil {
				cachedCount++
			}
		}

		if cachedCount < playerCount/2 {
			t.Errorf("Only %d/%d players have cache, expected at least 50%%",
				cachedCount, playerCount)
		}

		// Simulate player movement
		for i, player := range players {
			newLoc := model.NewLocation(150000+int32(i*150), 150000+int32(i*150), 0, 0)
			player.SetLocation(newLoc)
		}

		// Wait for cache updates after movement (instant with fake clock)
		time.Sleep(300 * time.Millisecond)

		// Verify caches were updated
		updatedCount := 0
		for _, player := range players {
			cache := player.GetVisibilityCache()
			if cache != nil && !cache.IsStale(500*time.Millisecond) {
				updatedCount++
			}
		}

		if updatedCount < playerCount/2 {
			t.Errorf("Only %d/%d players have fresh cache after movement",
				updatedCount, playerCount)
		}

		// Cleanup
		for _, player := range players {
			vm.UnregisterPlayer(player)
		}
	})
}
