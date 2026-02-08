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
