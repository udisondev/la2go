package world

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestForEachVisibleObject(t *testing.T) {
	w := Instance()

	// Clear any previous test objects
	// (Note: singleton world persists across tests)

	// Create objects in 3×3 grid of regions around Talking Island spawn
	baseX, baseY := int32(17000), int32(170000)

	objectIDs := make([]uint32, 0, 9)

	// Add objects to 9 regions (3×3 window)
	for dx := int32(-1); dx <= 1; dx++ {
		for dy := int32(-1); dy <= 1; dy++ {
			x := baseX + dx*RegionSize
			y := baseY + dy*RegionSize
			objectID := uint32(20000 + dx*10 + dy)

			loc := model.NewLocation(x, y, -3500, 0)
			obj := model.NewWorldObject(objectID, "VisibleObj", loc)

			if err := w.AddObject(obj); err != nil {
				t.Fatalf("AddObject() error = %v", err)
			}

			objectIDs = append(objectIDs, objectID)
		}
	}

	// Count visible objects from center
	count := CountVisibleObjects(w, baseX, baseY)

	if count != 9 {
		t.Errorf("CountVisibleObjects() = %d, want 9", count)
	}

	// Iterate and verify all objects are visible
	seen := make(map[uint32]bool)
	ForEachVisibleObject(w, baseX, baseY, func(obj *model.WorldObject) bool {
		seen[obj.ObjectID()] = true
		return true
	})

	if len(seen) != 9 {
		t.Errorf("ForEachVisibleObject() saw %d objects, want 9", len(seen))
	}

	// Cleanup
	for _, id := range objectIDs {
		w.RemoveObject(id)
	}
}

func TestForEachVisibleObject_EarlyStop(t *testing.T) {
	w := Instance()

	// Add 5 objects to same region
	loc := model.NewLocation(17000, 170000, -3500, 0)
	objectIDs := make([]uint32, 0, 5)

	for i := range 5 {
		objectID := uint32(30000 + i)
		obj := model.NewWorldObject(objectID, "StopTestObj", loc)
		if err := w.AddObject(obj); err != nil {
			t.Fatalf("AddObject() error = %v", err)
		}
		objectIDs = append(objectIDs, objectID)
	}

	// Iterate and stop after 3
	count := 0
	ForEachVisibleObject(w, 17000, 170000, func(obj *model.WorldObject) bool {
		count++
		return count < 3 // stop after 3
	})

	if count != 3 {
		t.Errorf("ForEachVisibleObject() with early stop count = %d, want 3", count)
	}

	// Cleanup
	for _, id := range objectIDs {
		w.RemoveObject(id)
	}
}

func TestForEachVisibleObject_InvalidCoordinates(t *testing.T) {
	w := Instance()

	// Should not panic or error with invalid coordinates
	count := CountVisibleObjects(w, WorldXMax+10000, WorldYMax+10000)

	if count != 0 {
		t.Errorf("CountVisibleObjects() with invalid coords = %d, want 0", count)
	}
}

func BenchmarkForEachVisibleObject(b *testing.B) {
	w := Instance()

	// Add objects to 3×3 grid
	baseX, baseY := int32(17000), int32(170000)
	objectIDs := make([]uint32, 0, 9)

	for dx := int32(-1); dx <= 1; dx++ {
		for dy := int32(-1); dy <= 1; dy++ {
			x := baseX + dx*RegionSize
			y := baseY + dy*RegionSize
			objectID := uint32(40000 + dx*10 + dy)

			loc := model.NewLocation(x, y, -3500, 0)
			obj := model.NewWorldObject(objectID, "BenchObj", loc)

			if err := w.AddObject(obj); err != nil {
				b.Fatal(err)
			}

			objectIDs = append(objectIDs, objectID)
		}
	}

	b.ResetTimer()
	for range b.N {
		ForEachVisibleObject(w, baseX, baseY, func(obj *model.WorldObject) bool {
			return true
		})
	}

	b.StopTimer()
	for _, id := range objectIDs {
		w.RemoveObject(id)
	}
}

// TestForEachVisibleObjectByLOD tests LOD-aware visibility iteration.
// Phase 4.12: LOD-aware broadcast prioritization.
func TestForEachVisibleObjectByLOD(t *testing.T) {
	w := Instance()
	vm := NewVisibilityManager(w, 100, 200)

	// Create player at center
	player, err := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer() error = %v", err)
	}

	baseX, baseY := int32(17000), int32(170000)
	player.SetLocation(model.NewLocation(baseX, baseY, -3500, 0))

	// Add objects to 9 regions (3×3 window)
	// Center region: 1 object (NEAR)
	// Adjacent regions (4): 4 objects (MEDIUM)
	// Diagonal regions (4): 4 objects (FAR)
	objectIDs := make([]uint32, 0, 9)

	// Center (NEAR)
	centerObj := model.NewWorldObject(50000, "CenterNPC", model.NewLocation(baseX, baseY, -3500, 0))
	if err := w.AddObject(centerObj); err != nil {
		t.Fatalf("AddObject() error = %v", err)
	}
	objectIDs = append(objectIDs, 50000)

	// Adjacent (MEDIUM): [1], [3], [5], [7] in 3×3 grid
	adjacent := []struct{ dx, dy int32 }{
		{0, -1}, // top
		{-1, 0}, // left
		{1, 0},  // right
		{0, 1},  // bottom
	}
	for i, pos := range adjacent {
		x := baseX + pos.dx*RegionSize
		y := baseY + pos.dy*RegionSize
		objectID := uint32(50001 + i)

		obj := model.NewWorldObject(objectID, "AdjacentNPC", model.NewLocation(x, y, -3500, 0))
		if err := w.AddObject(obj); err != nil {
			t.Fatalf("AddObject() error = %v", err)
		}
		objectIDs = append(objectIDs, objectID)
	}

	// Diagonal (FAR): [0], [2], [6], [8] in 3×3 grid
	diagonal := []struct{ dx, dy int32 }{
		{-1, -1}, // top-left
		{1, -1},  // top-right
		{-1, 1},  // bottom-left
		{1, 1},   // bottom-right
	}
	for i, pos := range diagonal {
		x := baseX + pos.dx*RegionSize
		y := baseY + pos.dy*RegionSize
		objectID := uint32(50005 + i)

		obj := model.NewWorldObject(objectID, "DiagonalNPC", model.NewLocation(x, y, -3500, 0))
		if err := w.AddObject(obj); err != nil {
			t.Fatalf("AddObject() error = %v", err)
		}
		objectIDs = append(objectIDs, objectID)
	}

	// Register player and update visibility cache AFTER adding all objects
	vm.RegisterPlayer(player)
	player.InvalidateVisibilityCache()
	vm.UpdateAll() // Build cache with all objects

	// Test LODNear: should see 1 object (center)
	nearCount := 0
	ForEachVisibleObjectByLOD(player, LODNear, func(obj *model.WorldObject) bool {
		nearCount++
		return true
	})
	if nearCount != 1 {
		t.Errorf("LODNear count = %d, want 1", nearCount)
	}

	// Test LODMedium: should see 4 objects (adjacent)
	mediumCount := 0
	ForEachVisibleObjectByLOD(player, LODMedium, func(obj *model.WorldObject) bool {
		mediumCount++
		return true
	})
	if mediumCount != 4 {
		t.Errorf("LODMedium count = %d, want 4", mediumCount)
	}

	// Test LODFar: should see 4 objects (diagonal)
	farCount := 0
	ForEachVisibleObjectByLOD(player, LODFar, func(obj *model.WorldObject) bool {
		farCount++
		return true
	})
	if farCount != 4 {
		t.Errorf("LODFar count = %d, want 4", farCount)
	}

	// Test LODAll: should see 9 objects (all)
	allCount := 0
	ForEachVisibleObjectByLOD(player, LODAll, func(obj *model.WorldObject) bool {
		allCount++
		return true
	})
	if allCount != 9 {
		t.Errorf("LODAll count = %d, want 9", allCount)
	}

	// Cleanup
	vm.UnregisterPlayer(player)
	for _, id := range objectIDs {
		w.RemoveObject(id)
	}
}

// TestForEachNearObject tests convenience wrapper for near objects.
// Phase 4.12: Most critical events should use near-only broadcast.
func TestForEachNearObject(t *testing.T) {
	w := Instance()
	vm := NewVisibilityManager(w, 100, 200)

	player, _ := model.NewPlayer(1, 1, "TestPlayer", 10, 0, 1)
	baseX, baseY := int32(17000), int32(170000)
	player.SetLocation(model.NewLocation(baseX, baseY, -3500, 0))

	// Add 1 near object + 1 far object
	nearObj := model.NewWorldObject(60000, "NearNPC", model.NewLocation(baseX, baseY, -3500, 0))
	farObj := model.NewWorldObject(60001, "FarNPC", model.NewLocation(baseX+RegionSize, baseY+RegionSize, -3500, 0))

	w.AddObject(nearObj)
	w.AddObject(farObj)

	vm.RegisterPlayer(player)
	player.InvalidateVisibilityCache()
	vm.UpdateAll() // Build cache with objects

	// ForEachNearObject should only see near object
	count := 0
	ForEachNearObject(player, func(obj *model.WorldObject) bool {
		count++
		if obj.ObjectID() != 60000 {
			t.Errorf("ForEachNearObject() saw object %d, want only 60000", obj.ObjectID())
		}
		return true
	})

	if count != 1 {
		t.Errorf("ForEachNearObject() count = %d, want 1", count)
	}

	// Cleanup
	vm.UnregisterPlayer(player)
	w.RemoveObject(60000)
	w.RemoveObject(60001)
}
