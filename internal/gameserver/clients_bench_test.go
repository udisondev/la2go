package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkClientManager_GetClientByObjectID measures O(N) linear scan overhead.
// Phase 4.11 Priority 0: Critical hot path — called for EACH visible player.
// Baseline expectation:
//   - 10 players: ~50ns (best case: player at index 1)
//   - 100 players: ~500ns (average case: player at index 50)
//   - 1000 players: ~5µs (worst case: player at index 999)
//   - 1000 players (miss): ~5µs (miss case: full scan, not found)
func BenchmarkClientManager_GetClientByObjectID(b *testing.B) {
	scenarios := []struct {
		name        string
		playerCount int
		targetIndex int // -1 for miss case
	}{
		{"10players_best", 10, 1},
		{"100players_average", 100, 50},
		{"1000players_worst", 1000, 999},
		{"1000players_miss", 1000, -1},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			cm := NewClientManager()

			// Populate clientManager with fake players
			var targetObjectID uint32
			for i := range scenario.playerCount {
				characterID := int64(0x10000000 + i)
				player, err := model.NewPlayer(
					characterID,
					int64(1000+i), // accountID
					"player_"+string(rune('A'+i%26)),
					int32(1+(i%80)),  // level (1-80)
					0,                 // raceID
					0,                 // classID
				)
				if err != nil {
					b.Fatalf("failed to create player: %v", err)
				}

				client := &GameClient{} // Minimal client for benchmark
				cm.RegisterPlayer(player, client)

				// Mark target objectID
				if i == scenario.targetIndex {
					targetObjectID = uint32(characterID)
				}
			}

			// Miss case: use non-existent objectID
			if scenario.targetIndex == -1 {
				targetObjectID = 0xFFFFFFFF
			}

			// Benchmark GetClientByObjectID
			b.ResetTimer()
			for range b.N {
				_ = cm.GetClientByObjectID(targetObjectID)
			}
		})
	}
}

// BenchmarkClientManager_GetClientByObjectID_Parallel measures concurrent read contention.
// Expected baseline: ~50ns @ 10 players (no RWMutex contention for reads).
func BenchmarkClientManager_GetClientByObjectID_Parallel(b *testing.B) {
	cm := NewClientManager()

	// Populate with 100 players
	var targetObjectID uint32
	for i := range 100 {
		characterID := int64(0x10000000 + i)
		player, err := model.NewPlayer(
			characterID,
			int64(1000+i),
			"player_"+string(rune('A'+i%26)),
			int32(1+(i%80)),
			0,
			0,
		)
		if err != nil {
			b.Fatalf("failed to create player: %v", err)
		}

		client := &GameClient{}
		cm.RegisterPlayer(player, client)

		if i == 50 {
			targetObjectID = uint32(characterID)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cm.GetClientByObjectID(targetObjectID)
		}
	})
}
