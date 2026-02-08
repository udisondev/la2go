package world

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestNewRegion(t *testing.T) {
	region := NewRegion(10, 20)

	if region.RX() != 10 {
		t.Errorf("RX() = %d, want 10", region.RX())
	}
	if region.RY() != 20 {
		t.Errorf("RY() = %d, want 20", region.RY())
	}
}

func TestRegion_AddRemoveVisibleObject(t *testing.T) {
	region := NewRegion(0, 0)
	obj := model.NewWorldObject(100, "TestObj", model.Location{})

	// Add object
	region.AddVisibleObject(obj)

	// Verify object is in region
	count := 0
	region.ForEachVisibleObject(func(o *model.WorldObject) bool {
		count++
		if o.ObjectID() != 100 {
			t.Errorf("ForEachVisibleObject() objectID = %d, want 100", o.ObjectID())
		}
		return true
	})

	if count != 1 {
		t.Errorf("ForEachVisibleObject() count = %d, want 1", count)
	}

	// Remove object
	region.RemoveVisibleObject(100)

	// Verify object is removed
	count = 0
	region.ForEachVisibleObject(func(o *model.WorldObject) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("after RemoveVisibleObject() ForEachVisibleObject() count = %d, want 0", count)
	}
}

func TestRegion_ForEachVisibleObject_EarlyStop(t *testing.T) {
	region := NewRegion(0, 0)

	// Add multiple objects
	for i := range 10 {
		obj := model.NewWorldObject(uint32(i), "Obj", model.Location{})
		region.AddVisibleObject(obj)
	}

	// Iterate and stop after 5
	count := 0
	region.ForEachVisibleObject(func(o *model.WorldObject) bool {
		count++
		return count < 5 // stop after 5
	})

	if count != 5 {
		t.Errorf("ForEachVisibleObject() with early stop count = %d, want 5", count)
	}
}

func TestRegion_SurroundingRegions(t *testing.T) {
	region := NewRegion(5, 5)

	surrounding := []*Region{
		NewRegion(4, 4), NewRegion(5, 4), NewRegion(6, 4),
		NewRegion(4, 5), region, NewRegion(6, 5),
		NewRegion(4, 6), NewRegion(5, 6), NewRegion(6, 6),
	}

	region.SetSurroundingRegions(surrounding)

	got := region.SurroundingRegions()
	if len(got) != len(surrounding) {
		t.Errorf("SurroundingRegions() length = %d, want %d", len(got), len(surrounding))
	}

	// Verify immutability: SurroundingRegions() returns same slice reference (zero-copy)
	// IMPORTANT: surroundingRegions is immutable after World initialization
	// This is a performance optimization (100K calls/sec, -7.2 MB/sec allocation rate)
	gotAgain := region.SurroundingRegions()
	if &got[0] != &gotAgain[0] {
		t.Error("SurroundingRegions() should return same slice reference (zero-copy)")
	}

	// Verify content
	for i, r := range got {
		if r != surrounding[i] {
			t.Errorf("SurroundingRegions()[%d] = %v, want %v", i, r, surrounding[i])
		}
	}
}
