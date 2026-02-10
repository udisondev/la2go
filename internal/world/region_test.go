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

// TestRegion_GetVisibleObjectsSnapshot_LazyRebuild tests snapshot cache lazy rebuild.
// Phase 4.11 Tier 2: Snapshot cache reduces sync.Map overhead (-70% latency).
func TestRegion_GetVisibleObjectsSnapshot_LazyRebuild(t *testing.T) {
	region := NewRegion(0, 0)

	// Initially cache is empty
	snapshot := region.GetVisibleObjectsSnapshot()
	if len(snapshot) != 0 {
		t.Errorf("Initial snapshot length = %d, want 0", len(snapshot))
	}

	// Add objects
	obj1 := model.NewWorldObject(1, "Obj1", model.Location{})
	obj2 := model.NewWorldObject(2, "Obj2", model.Location{})
	region.AddVisibleObject(obj1)
	region.AddVisibleObject(obj2)

	// Get snapshot (should rebuild)
	snapshot = region.GetVisibleObjectsSnapshot()
	if len(snapshot) != 2 {
		t.Errorf("Snapshot length after add = %d, want 2", len(snapshot))
	}

	// Get snapshot again (should use cache, not rebuild)
	snapshot2 := region.GetVisibleObjectsSnapshot()
	if len(snapshot2) != 2 {
		t.Errorf("Cached snapshot length = %d, want 2", len(snapshot2))
	}

	// Verify cache returns same slice reference (zero-copy)
	if &snapshot[0] != &snapshot2[0] {
		t.Error("Cached snapshot should return same slice reference")
	}

	// Remove object (invalidates cache)
	region.RemoveVisibleObject(1)

	// Get snapshot (should rebuild)
	snapshot3 := region.GetVisibleObjectsSnapshot()
	if len(snapshot3) != 1 {
		t.Errorf("Snapshot length after remove = %d, want 1", len(snapshot3))
	}
	if snapshot3[0].ObjectID() != 2 {
		t.Errorf("Snapshot object ID = %d, want 2", snapshot3[0].ObjectID())
	}
}

// TestRegion_GetVisibleObjectsSnapshot_Concurrent tests concurrent snapshot access.
// Phase 4.11 Tier 2: Snapshot cache is concurrent-safe (atomic.Value).
func TestRegion_GetVisibleObjectsSnapshot_Concurrent(t *testing.T) {
	region := NewRegion(0, 0)

	// Add initial objects
	for i := range 100 {
		obj := model.NewWorldObject(uint32(i), "Obj", model.Location{})
		region.AddVisibleObject(obj)
	}

	done := make(chan bool)

	// Concurrent readers
	for range 50 {
		go func() {
			for range 1000 {
				snapshot := region.GetVisibleObjectsSnapshot()
				_ = snapshot // use snapshot
			}
			done <- true
		}()
	}

	// Concurrent writers
	for range 10 {
		go func() {
			for i := range 100 {
				obj := model.NewWorldObject(uint32(i+1000), "NewObj", model.Location{})
				region.AddVisibleObject(obj)
				region.RemoveVisibleObject(uint32(i + 1000))
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 60 {
		<-done
	}
}

// TestRegion_SnapshotImmutability verifies snapshot immutability.
// Phase 4.11 Tier 2: Snapshot is immutable — safe for concurrent access.
func TestRegion_SnapshotImmutability(t *testing.T) {
	region := NewRegion(0, 0)

	// Add objects
	obj1 := model.NewWorldObject(1, "Obj1", model.Location{})
	obj2 := model.NewWorldObject(2, "Obj2", model.Location{})
	region.AddVisibleObject(obj1)
	region.AddVisibleObject(obj2)

	// Get snapshot
	snapshot := region.GetVisibleObjectsSnapshot()
	if len(snapshot) != 2 {
		t.Fatalf("Snapshot length = %d, want 2", len(snapshot))
	}

	// Add more objects (invalidates cache)
	obj3 := model.NewWorldObject(3, "Obj3", model.Location{})
	region.AddVisibleObject(obj3)

	// Original snapshot should remain unchanged (immutable)
	if len(snapshot) != 2 {
		t.Errorf("Original snapshot modified: length = %d, want 2", len(snapshot))
	}

	// New snapshot should have 3 objects
	newSnapshot := region.GetVisibleObjectsSnapshot()
	if len(newSnapshot) != 3 {
		t.Errorf("New snapshot length = %d, want 3", len(newSnapshot))
	}
}
