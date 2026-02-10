package world

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkForEachVisibleObject_Baseline measures SLOW PATH (no cache).
// This is the baseline performance before Phase 4.5 PR3 optimization.
func BenchmarkForEachVisibleObject_Baseline(b *testing.B) {
	world := Instance()

	// Setup: Create player at center with 50 objects in region
	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	// Populate region with 50 objects
	regionX, regionY := CoordToRegionIndex(150000, 150000)
	region := world.GetRegion(regionX, regionY)
	if region != nil {
		for i := range 50 {
			obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
			region.AddVisibleObject(obj)
		}
	}

	b.ResetTimer()
	for range b.N {
		count := 0
		ForEachVisibleObjectForPlayer(player, func(obj *model.WorldObject) bool {
			count++
			return true
		})
	}
}

// BenchmarkForEachVisibleObjectCached_Hit measures FAST PATH (cache hit).
// This is the optimized performance with valid cache.
func BenchmarkForEachVisibleObjectCached_Hit(b *testing.B) {
	world := Instance()

	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	regionX, regionY := CoordToRegionIndex(150000, 150000)
	region := world.GetRegion(regionX, regionY)
	if region != nil {
		for i := range 50 {
			obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
			region.AddVisibleObject(obj)
		}
	}

	// Pre-populate cache (simulate VisibilityManager update)
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)
	vm.updatePlayerCache(player)

	b.ResetTimer()
	for range b.N {
		count := 0
		ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
			count++
			return true
		})
	}
}

// BenchmarkForEachVisibleObjectCached_Miss measures SLOW PATH fallback (cache miss).
// This simulates first query or invalidated cache.
func BenchmarkForEachVisibleObjectCached_Miss(b *testing.B) {
	world := Instance()

	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)
	loc := model.NewLocation(150000, 150000, 0, 0)
	player.SetLocation(loc)

	regionX, regionY := CoordToRegionIndex(150000, 150000)
	region := world.GetRegion(regionX, regionY)
	if region != nil {
		for i := range 50 {
			obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
			region.AddVisibleObject(obj)
		}
	}

	// No cache — simulate cache miss

	b.ResetTimer()
	for range b.N {
		count := 0
		ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
			count++
			return true
		})
	}
}

// BenchmarkVisibilityManager_UpdateAll_100 measures batch update for 100 players.
func BenchmarkVisibilityManager_UpdateAll_100(b *testing.B) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	// Create 100 players at different locations
	players := make([]*model.Player, 100)
	for i := range 100 {
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player", 10, 0, 1)
		loc := model.NewLocation(150000+int32(i*100), 150000+int32(i*100), 0, 0)
		player.SetLocation(loc)
		vm.RegisterPlayer(player)
		players[i] = player
	}

	// Populate world with 500 objects
	for i := range 500 {
		loc := model.NewLocation(150000+int32(i*50), 150000+int32(i*50), 0, 0)
		obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
		regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)
		region := world.GetRegion(regionX, regionY)
		if region != nil {
			region.AddVisibleObject(obj)
		}
	}

	b.ResetTimer()
	for range b.N {
		vm.UpdateAll()
	}
}

// BenchmarkVisibilityManager_UpdateAll_1000 measures batch update for 1000 players.
// This simulates realistic production load.
func BenchmarkVisibilityManager_UpdateAll_1000(b *testing.B) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	// Create 1000 players
	players := make([]*model.Player, 1000)
	for i := range 1000 {
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player", 10, 0, 1)
		loc := model.NewLocation(150000+int32(i*10), 150000+int32(i*10), 0, 0)
		player.SetLocation(loc)
		vm.RegisterPlayer(player)
		players[i] = player
	}

	// Populate world with 5000 objects
	for i := range 5000 {
		loc := model.NewLocation(150000+int32(i*20), 150000+int32(i*20), 0, 0)
		obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
		regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)
		region := world.GetRegion(regionX, regionY)
		if region != nil {
			region.AddVisibleObject(obj)
		}
	}

	b.ResetTimer()
	for range b.N {
		vm.UpdateAll()
	}
}

// BenchmarkVisibilityManager_UpdateAll_10000 measures batch update for 10K players.
// This simulates heavy production load.
func BenchmarkVisibilityManager_UpdateAll_10000(b *testing.B) {
	world := Instance()
	vm := NewVisibilityManager(world, 100*time.Millisecond, 200*time.Millisecond)

	// Create 10K players
	players := make([]*model.Player, 10000)
	for i := range 10000 {
		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player", 10, 0, 1)
		// Distribute players across world
		loc := model.NewLocation(100000+int32(i%1000)*100, 100000+int32(i/1000)*100, 0, 0)
		player.SetLocation(loc)
		vm.RegisterPlayer(player)
		players[i] = player
	}

	// Populate world with 50K objects
	for i := range 50000 {
		loc := model.NewLocation(100000+int32(i%10000)*10, 100000+int32(i/10000)*10, 0, 0)
		obj := model.NewWorldObject(uint32(i+1), "NPC", loc)
		regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)
		region := world.GetRegion(regionX, regionY)
		if region != nil {
			region.AddVisibleObject(obj)
		}
	}

	b.ResetTimer()
	for range b.N {
		vm.UpdateAll()
	}
}

// Phase 4.11 Tier 1 Opt 2: Removed BenchmarkVisibilityCache_ObjectExists
// objectExists() function was removed — defensive validation unnecessary overhead

// BenchmarkPlayer_GetSetVisibilityCache measures atomic.Value overhead.
func BenchmarkPlayer_GetSetVisibilityCache(b *testing.B) {
	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)
	objects := []*model.WorldObject{
		model.NewWorldObject(1, "NPC", model.Location{}),
	}
	cache := model.NewVisibilityCache(objects, nil, nil, 0, 0, 0)

	b.Run("Get", func(b *testing.B) {
		player.SetVisibilityCache(cache)
		b.ResetTimer()
		for range b.N {
			_ = player.GetVisibilityCache()
		}
	})

	b.Run("Set", func(b *testing.B) {
		for range b.N {
			player.SetVisibilityCache(cache)
		}
	})

	b.Run("Invalidate", func(b *testing.B) {
		for range b.N {
			player.InvalidateVisibilityCache()
		}
	})
}
