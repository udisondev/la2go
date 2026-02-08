package world

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestInstance_Singleton(t *testing.T) {
	w1 := Instance()
	w2 := Instance()

	if w1 != w2 {
		t.Error("Instance() should return same instance (singleton)")
	}
}

func TestWorld_RegionCount(t *testing.T) {
	w := Instance()

	expectedCount := RegionsX * RegionsY
	if w.RegionCount() != expectedCount {
		t.Errorf("RegionCount() = %d, want %d", w.RegionCount(), expectedCount)
	}
}

func TestWorld_GetRegion(t *testing.T) {
	w := Instance()

	// Valid coordinates
	region := w.GetRegion(0, 0)
	if region == nil {
		t.Error("GetRegion(0, 0) returned nil, want valid region")
	}

	// Invalid coordinates (out of bounds)
	region = w.GetRegion(WorldXMax+10000, WorldYMax+10000)
	if region != nil {
		t.Error("GetRegion(out of bounds) should return nil")
	}
}

func TestWorld_GetRegionByIndex(t *testing.T) {
	w := Instance()

	// Valid index
	region := w.GetRegionByIndex(0, 0)
	if region == nil {
		t.Error("GetRegionByIndex(0, 0) returned nil")
	}
	if region.RX() != 0 || region.RY() != 0 {
		t.Errorf("GetRegionByIndex(0, 0) region = (%d, %d), want (0, 0)", region.RX(), region.RY())
	}

	// Invalid index
	region = w.GetRegionByIndex(-1, 0)
	if region != nil {
		t.Error("GetRegionByIndex(-1, 0) should return nil")
	}

	region = w.GetRegionByIndex(RegionsX, 0)
	if region != nil {
		t.Error("GetRegionByIndex(RegionsX, 0) should return nil")
	}
}

func TestWorld_AddRemoveObject(t *testing.T) {
	w := Instance()

	// Create test object at valid location
	loc := model.NewLocation(17000, 170000, -3500, 0)
	obj := model.NewWorldObject(9999, "TestNPC", loc)

	// Add object
	if err := w.AddObject(obj); err != nil {
		t.Fatalf("AddObject() error = %v", err)
	}

	// Verify object is in world
	got, ok := w.GetObject(9999)
	if !ok {
		t.Error("GetObject() after AddObject() returned false")
	}
	if got.ObjectID() != 9999 {
		t.Errorf("GetObject() objectID = %d, want 9999", got.ObjectID())
	}

	// Verify object is in region
	region := w.GetRegion(17000, 170000)
	if region == nil {
		t.Fatal("GetRegion() returned nil")
	}

	found := false
	region.ForEachVisibleObject(func(o *model.WorldObject) bool {
		if o.ObjectID() == 9999 {
			found = true
			return false
		}
		return true
	})

	if !found {
		t.Error("object not found in region after AddObject()")
	}

	// Remove object
	w.RemoveObject(9999)

	// Verify object is removed from world
	_, ok = w.GetObject(9999)
	if ok {
		t.Error("GetObject() after RemoveObject() returned true, want false")
	}

	// Verify object is removed from region
	found = false
	region.ForEachVisibleObject(func(o *model.WorldObject) bool {
		if o.ObjectID() == 9999 {
			found = true
			return false
		}
		return true
	})

	if found {
		t.Error("object still in region after RemoveObject()")
	}
}

func TestWorld_AddObject_InvalidCoordinates(t *testing.T) {
	w := Instance()

	// Create object at invalid location (way out of bounds)
	loc := model.NewLocation(WorldXMax+100000, WorldYMax+100000, 0, 0)
	obj := model.NewWorldObject(8888, "InvalidObj", loc)

	// AddObject should return error
	if err := w.AddObject(obj); err == nil {
		t.Error("AddObject() with invalid coordinates should return error")
	}
}

func TestWorld_SurroundingRegions_3x3Window(t *testing.T) {
	w := Instance()

	// Get region at center of world
	region := w.GetRegionByIndex(OffsetX, OffsetY)
	if region == nil {
		t.Fatal("GetRegionByIndex() returned nil")
	}

	surrounding := region.SurroundingRegions()

	// Should have 9 regions (3×3 including center)
	if len(surrounding) != 9 {
		t.Errorf("SurroundingRegions() length = %d, want 9", len(surrounding))
	}

	// Verify center region is included
	foundCenter := false
	for _, r := range surrounding {
		if r.RX() == OffsetX && r.RY() == OffsetY {
			foundCenter = true
			break
		}
	}

	if !foundCenter {
		t.Error("SurroundingRegions() should include center region")
	}
}

func TestWorld_SurroundingRegions_EdgeRegion(t *testing.T) {
	w := Instance()

	// Get corner region (0, 0) — should have fewer than 9 neighbors
	region := w.GetRegionByIndex(0, 0)
	if region == nil {
		t.Fatal("GetRegionByIndex(0, 0) returned nil")
	}

	surrounding := region.SurroundingRegions()

	// Corner region should have 4 surrounding regions (including itself)
	// (0,0), (1,0), (0,1), (1,1)
	if len(surrounding) != 4 {
		t.Errorf("SurroundingRegions() for corner region length = %d, want 4", len(surrounding))
	}
}

func BenchmarkWorld_GetRegion(b *testing.B) {
	w := Instance()

	for range b.N {
		w.GetRegion(17000, 170000)
	}
}

func BenchmarkWorld_AddObject(b *testing.B) {
	w := Instance()
	loc := model.NewLocation(17000, 170000, -3500, 0)

	b.ResetTimer()
	for i := range b.N {
		obj := model.NewWorldObject(uint32(i+100000), "BenchObj", loc)
		if err := w.AddObject(obj); err != nil {
			b.Fatal(err)
		}
	}
}
