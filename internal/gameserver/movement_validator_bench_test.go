package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// newBenchPlayer creates a player at given position for benchmarks.
// Uses panic instead of t.Fatal since b.Fatal is not available in setup.
func newBenchPlayer(x, y, z int32) *model.Player {
	p, err := model.NewPlayer(1, 100, 200, "BenchPlayer", 10, 0, 0)
	if err != nil {
		panic(err)
	}
	p.SetLocation(model.NewLocation(x, y, z, 0))
	return p
}

// BenchmarkValidateMoveToLocation_Valid benchmarks a valid movement (typical hot path).
// Expected: ~10-20ns (3x int64 comparisons + arithmetic).
func BenchmarkValidateMoveToLocation_Valid(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 1000, 1000, 500)
	}
}

// BenchmarkValidateMoveToLocation_InvalidZ benchmarks Z-bounds failure (early return).
// Expected: ~5ns (first comparison fails).
func BenchmarkValidateMoveToLocation_InvalidZ(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 1000, 1000, -50000)
	}
}

// BenchmarkValidateMoveToLocation_TooFar benchmarks distance > max (anti-teleport).
// Expected: ~10ns (Z check pass, distance calc, max distance fail).
func BenchmarkValidateMoveToLocation_TooFar(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 20000, 20000, 0)
	}
}

// BenchmarkValidateMoveToLocation_TooClose benchmarks distance < min (anti-spam).
// Expected: ~10ns (Z pass, distance calc, min distance fail).
func BenchmarkValidateMoveToLocation_TooClose(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 1, 0, 0)
	}
}

// BenchmarkValidateMoveToLocation_ZeroDistance benchmarks click-same-position (allowed).
// Expected: ~10ns (Z pass, distance=0 special case pass).
func BenchmarkValidateMoveToLocation_ZeroDistance(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(100, 200, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 100, 200, 0)
	}
}

// BenchmarkValidateMoveToLocation_MaxBoundary benchmarks exactly at max distance boundary.
// Expected: ~10ns (all checks pass, close to boundary).
func BenchmarkValidateMoveToLocation_MaxBoundary(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 9900, 0, 0)
	}
}

// BenchmarkValidateMoveToLocation_Diagonal benchmarks diagonal move (realistic scenario).
// Expected: ~10ns (both dx and dy contribute to distance).
func BenchmarkValidateMoveToLocation_Diagonal(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(1000, 2000, -100)

	b.ResetTimer()
	for range b.N {
		_ = ValidateMoveToLocation(player, 2500, 3500, 200)
	}
}

// BenchmarkValidatePositionDesync_NoDesync benchmarks no desync (most common path).
// Expected: ~5-10ns (1x distance^2 calc + comparison).
func BenchmarkValidatePositionDesync_NoDesync(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(1000, 2000, 0)

	b.ResetTimer()
	for range b.N {
		_, _ = ValidatePositionDesync(player, 1010, 2010, 0)
	}
}

// BenchmarkValidatePositionDesync_Warning benchmarks warning-level desync.
// Expected: ~5-10ns (same calc, different branch outcome).
func BenchmarkValidatePositionDesync_Warning(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_, _ = ValidatePositionDesync(player, 501, 0, 0)
	}
}

// BenchmarkValidatePositionDesync_Critical benchmarks critical desync (potential hack).
// Expected: ~5-10ns (same calc).
func BenchmarkValidatePositionDesync_Critical(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(0, 0, 0)

	b.ResetTimer()
	for range b.N {
		_, _ = ValidatePositionDesync(player, 1000, 1000, 0)
	}
}

// BenchmarkValidatePositionDesync_Identical benchmarks identical positions (zero diff).
// Expected: ~5ns (fastest path: diff=0, no correction).
func BenchmarkValidatePositionDesync_Identical(b *testing.B) {
	b.ReportAllocs()
	player := newBenchPlayer(5000, 6000, 0)

	b.ResetTimer()
	for range b.N {
		_, _ = ValidatePositionDesync(player, 5000, 6000, 0)
	}
}
