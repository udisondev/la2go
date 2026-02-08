package model

import (
	"testing"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		name    string
		x       int32
		y       int32
		z       int32
		heading uint16
		want    Location
	}{
		{
			name:    "zero values",
			x:       0,
			y:       0,
			z:       0,
			heading: 0,
			want:    Location{X: 0, Y: 0, Z: 0, Heading: 0},
		},
		{
			name:    "positive coordinates",
			x:       100,
			y:       200,
			z:       300,
			heading: 1000,
			want:    Location{X: 100, Y: 200, Z: 300, Heading: 1000},
		},
		{
			name:    "negative coordinates",
			x:       -100,
			y:       -200,
			z:       -300,
			heading: 32768,
			want:    Location{X: -100, Y: -200, Z: -300, Heading: 32768},
		},
		{
			name:    "max heading",
			x:       0,
			y:       0,
			z:       0,
			heading: 65535,
			want:    Location{X: 0, Y: 0, Z: 0, Heading: 65535},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLocation(tt.x, tt.y, tt.z, tt.heading)
			if got != tt.want {
				t.Errorf("NewLocation() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLocation_WithHeading(t *testing.T) {
	original := NewLocation(100, 200, 300, 1000)

	tests := []struct {
		name       string
		newHeading uint16
		want       Location
	}{
		{
			name:       "change heading",
			newHeading: 2000,
			want:       Location{X: 100, Y: 200, Z: 300, Heading: 2000},
		},
		{
			name:       "zero heading",
			newHeading: 0,
			want:       Location{X: 100, Y: 200, Z: 300, Heading: 0},
		},
		{
			name:       "max heading",
			newHeading: 65535,
			want:       Location{X: 100, Y: 200, Z: 300, Heading: 65535},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := original.WithHeading(tt.newHeading)

			// Проверяем что новый Location имеет правильные значения
			if got != tt.want {
				t.Errorf("WithHeading() = %+v, want %+v", got, tt.want)
			}

			// ВАЖНО: проверяем immutability — оригинал не должен измениться
			if original.Heading != 1000 {
				t.Errorf("WithHeading() mutated original location: got heading %d, want 1000", original.Heading)
			}
			if original.X != 100 || original.Y != 200 || original.Z != 300 {
				t.Errorf("WithHeading() mutated original coordinates: %+v", original)
			}
		})
	}
}

func TestLocation_WithCoordinates(t *testing.T) {
	original := NewLocation(100, 200, 300, 1000)

	tests := []struct {
		name string
		x    int32
		y    int32
		z    int32
		want Location
	}{
		{
			name: "change coordinates",
			x:    400,
			y:    500,
			z:    600,
			want: Location{X: 400, Y: 500, Z: 600, Heading: 1000},
		},
		{
			name: "zero coordinates",
			x:    0,
			y:    0,
			z:    0,
			want: Location{X: 0, Y: 0, Z: 0, Heading: 1000},
		},
		{
			name: "negative coordinates",
			x:    -100,
			y:    -200,
			z:    -300,
			want: Location{X: -100, Y: -200, Z: -300, Heading: 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := original.WithCoordinates(tt.x, tt.y, tt.z)

			// Проверяем что новый Location имеет правильные значения
			if got != tt.want {
				t.Errorf("WithCoordinates() = %+v, want %+v", got, tt.want)
			}

			// ВАЖНО: проверяем immutability — оригинал не должен измениться
			if original.X != 100 || original.Y != 200 || original.Z != 300 {
				t.Errorf("WithCoordinates() mutated original: %+v", original)
			}
			if original.Heading != 1000 {
				t.Errorf("WithCoordinates() mutated heading: got %d, want 1000", original.Heading)
			}
		})
	}
}

func TestLocation_DistanceSquared(t *testing.T) {
	tests := []struct {
		name  string
		loc1  Location
		loc2  Location
		want  int64
	}{
		{
			name:  "same location",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(0, 0, 0, 0),
			want:  0,
		},
		{
			name:  "distance on X axis",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(10, 0, 0, 0),
			want:  100, // 10^2
		},
		{
			name:  "distance on Y axis",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(0, 10, 0, 0),
			want:  100, // 10^2
		},
		{
			name:  "distance on Z axis",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(0, 0, 10, 0),
			want:  100, // 10^2
		},
		{
			name:  "3-4-5 triangle",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(3, 4, 0, 0),
			want:  25, // 3^2 + 4^2 = 25 (distance 5)
		},
		{
			name:  "3D distance",
			loc1:  NewLocation(0, 0, 0, 0),
			loc2:  NewLocation(1, 2, 2, 0),
			want:  9, // 1^2 + 2^2 + 2^2 = 9 (distance 3)
		},
		{
			name:  "negative coordinates",
			loc1:  NewLocation(-10, -10, -10, 0),
			loc2:  NewLocation(10, 10, 10, 0),
			want:  1200, // 20^2 + 20^2 + 20^2 = 1200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.loc1.DistanceSquared(tt.loc2)
			if got != tt.want {
				t.Errorf("DistanceSquared() = %d, want %d", got, tt.want)
			}

			// Distance должна быть симметричной
			gotReverse := tt.loc2.DistanceSquared(tt.loc1)
			if gotReverse != tt.want {
				t.Errorf("DistanceSquared() reverse = %d, want %d", gotReverse, tt.want)
			}
		})
	}
}

func TestLocation_ZeroValue(t *testing.T) {
	// Проверяем что zero value Location — valid struct
	var loc Location

	if loc.X != 0 || loc.Y != 0 || loc.Z != 0 || loc.Heading != 0 {
		t.Errorf("zero value Location not initialized correctly: %+v", loc)
	}

	// Zero value должен работать с методами
	newLoc := loc.WithHeading(100)
	if newLoc.Heading != 100 {
		t.Errorf("WithHeading() on zero value failed: %+v", newLoc)
	}

	dist := loc.DistanceSquared(NewLocation(10, 10, 10, 0))
	if dist != 300 {
		t.Errorf("DistanceSquared() on zero value failed: got %d, want 300", dist)
	}
}

// Benchmark для DistanceSquared (hot path в movement calculations)
func BenchmarkLocation_DistanceSquared(b *testing.B) {
	loc1 := NewLocation(1000, 2000, 3000, 0)
	loc2 := NewLocation(1100, 2200, 3300, 0)

	b.ResetTimer()
	for b.Loop() {
		_ = loc1.DistanceSquared(loc2)
	}
}
