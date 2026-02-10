package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// Benchmark UpdateLocation — HOT PATH (5-10M calls/sec на пике с 100K игроков)
func BenchmarkCharacterRepository_UpdateLocation(b *testing.B) {
	pool := setupTestDB(b)
	repo := NewCharacterRepository(pool)

	// Setup test character
	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 75, 0, 0)
	if err != nil {
		b.Fatalf("creating player: %v", err)
	}

	if err := repo.Create(context.Background(), player); err != nil {
		b.Fatalf("creating test player: %v", err)
	}

	ctx := context.Background()
	loc := model.NewLocation(10000, 20000, 3000, 1500)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := repo.UpdateLocation(ctx, player.CharacterID(), loc); err != nil {
				b.Errorf("UpdateLocation failed: %v", err)
			}
		}
	})
}

// Benchmark UpdateStats — HOT PATH (5-10M calls/sec на пике)
func BenchmarkCharacterRepository_UpdateStats(b *testing.B) {
	pool := setupTestDB(b)
	repo := NewCharacterRepository(pool)

	// Setup test character
	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 75, 0, 0)
	if err != nil {
		b.Fatalf("creating player: %v", err)
	}

	if err := repo.Create(context.Background(), player); err != nil {
		b.Fatalf("creating test player: %v", err)
	}

	ctx := context.Background()
	hp, mp, cp := int32(1000), int32(500), int32(100)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := repo.UpdateStats(ctx, player.CharacterID(), hp, mp, cp); err != nil {
				b.Errorf("UpdateStats failed: %v", err)
			}
		}
	})
}

// Benchmark LoadByID — WARM PATH
func BenchmarkCharacterRepository_LoadByID(b *testing.B) {
	pool := setupTestDB(b)
	repo := NewCharacterRepository(pool)

	// Setup test character
	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 75, 0, 0)
	if err != nil {
		b.Fatalf("creating player: %v", err)
	}

	if err := repo.Create(context.Background(), player); err != nil {
		b.Fatalf("creating test player: %v", err)
	}

	ctx := context.Background()
	characterID := player.CharacterID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.LoadByID(ctx, characterID)
		if err != nil {
			b.Errorf("LoadByID failed: %v", err)
		}
	}
}

// Benchmark LoadByAccountID — с pre-allocation
// Тестирует memory allocations для разных sizes
func BenchmarkCharacterRepository_LoadByAccountID(b *testing.B) {
	pool := setupTestDB(b)
	repo := NewCharacterRepository(pool)

	// Создаём 5 персонажей для одного аккаунта
	const accountID = int64(1)
	const numChars = 5

	for i := 0; i < numChars; i++ {
		name := fmt.Sprintf("BenchHero%d", i)
		player, err := model.NewPlayer(uint32(i), 0, accountID, name, 75, 0, 0)
		if err != nil {
			b.Fatalf("creating player %d: %v", i, err)
		}

		if err := repo.Create(context.Background(), player); err != nil {
			b.Fatalf("creating test player %d: %v", i, err)
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		players, err := repo.LoadByAccountID(ctx, accountID)
		if err != nil {
			b.Errorf("LoadByAccountID failed: %v", err)
		}
		if len(players) != numChars {
			b.Errorf("expected %d characters, got %d", numChars, len(players))
		}
	}
}
