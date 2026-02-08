package world

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkRegion_SurroundingRegions — zero-copy vs copy (before optimization)
func BenchmarkRegion_SurroundingRegions(b *testing.B) {
	// Create region with 9 surrounding regions (3×3 window)
	region := NewRegion(10, 10)
	surrounding := make([]*Region, 9)
	for i := range surrounding {
		surrounding[i] = NewRegion(int32(i/3+9), int32(i%3+9))
	}
	region.SetSurroundingRegions(surrounding)

	b.Run("ZeroCopy_current", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			_ = region.SurroundingRegions()
		}
	})
}

// BenchmarkVisibility_ForEachVisibleObject — realistic visibility query workload
func BenchmarkVisibility_ForEachVisibleObject(b *testing.B) {
	// Setup: 3×3 region grid with 50 objects per region
	world := Instance()

	centerX, centerY := int32(150000), int32(150000)
	rx, ry := CoordToRegionIndex(centerX, centerY)

	// Add 50 objects per region (9 regions × 50 = 450 objects total)
	objectID := uint32(1)
	for dx := int32(-1); dx <= 1; dx++ {
		for dy := int32(-1); dy <= 1; dy++ {
			region := world.GetRegion(rx+dx, ry+dy)
			if region == nil {
				continue
			}

			for range 50 {
				loc := model.NewLocation(centerX+dx*1000, centerY+dy*1000, 0, 0)
				obj := model.NewWorldObject(objectID, "TestNPC", loc)
				region.AddVisibleObject(obj)
				objectID++
			}
		}
	}

	b.Run("ForEachVisibleObject_current_region", func(b *testing.B) {
		b.ReportAllocs()

		region := world.GetRegion(rx, ry)

		b.ResetTimer()
		for range b.N {
			count := 0
			region.ForEachVisibleObject(func(obj *model.WorldObject) bool {
				count++
				return true
			})
		}
	})

	b.Run("ForEachVisibleObject_9_regions", func(b *testing.B) {
		b.ReportAllocs()

		region := world.GetRegion(rx, ry)
		surrounding := region.SurroundingRegions()

		b.ResetTimer()
		for range b.N {
			count := 0

			// Current region
			region.ForEachVisibleObject(func(obj *model.WorldObject) bool {
				count++
				return true
			})

			// Surrounding regions
			for _, r := range surrounding {
				if r == nil {
					continue
				}
				r.ForEachVisibleObject(func(obj *model.WorldObject) bool {
					count++
					return true
				})
			}
		}
	})
}

// BenchmarkRegion_AddRemoveObject — concurrent add/remove operations
func BenchmarkRegion_AddRemoveObject(b *testing.B) {
	region := NewRegion(10, 10)

	b.Run("Add", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for i := range b.N {
			loc := model.NewLocation(0, 0, 0, 0)
			obj := model.NewWorldObject(uint32(i), "TestNPC", loc)
			region.AddVisibleObject(obj)
		}
	})

	b.Run("Remove", func(b *testing.B) {
		// Pre-fill region
		for i := range 1000 {
			loc := model.NewLocation(0, 0, 0, 0)
			obj := model.NewWorldObject(uint32(i), "TestNPC", loc)
			region.AddVisibleObject(obj)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			region.RemoveVisibleObject(uint32(i % 1000))
		}
	})
}
