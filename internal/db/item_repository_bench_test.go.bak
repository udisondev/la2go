package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// Benchmark LoadInventory — с разными sizes (10, 50, 100 items)
// Проверяет эффективность pre-allocation для типичных случаев
func BenchmarkItemRepository_LoadInventory(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			// ВАЖНО: setupTestDB() очищает таблицы, поэтому вызываем его ДО создания персонажа
			pool := setupTestDB(b)
			repo := NewItemRepository(pool)
			charRepo := NewCharacterRepository(pool)

			// Setup test character (required for FK constraint)
			// Создаём ПОСЛЕ setupTestDB(), чтобы TRUNCATE не удалил его
			player, err := model.NewPlayer(0, 0, 1, "ItemTestOwner", 75, 0, 0)
			if err != nil {
				b.Fatalf("creating player: %v", err)
			}
			if err := charRepo.Create(context.Background(), player); err != nil {
				b.Fatalf("creating test player: %v", err)
			}

			// Используем реальный characterID из БД
			ownerID := player.CharacterID()

			// Создаём N предметов в инвентаре
			for i := 0; i < size; i++ {
				item, err := model.NewItem(ownerID, 1000+int32(i), 1)
				if err != nil {
					b.Fatalf("creating item %d: %v", i, err)
				}
				item.SetLocation(model.ItemLocationInventory, 0)

				if err := repo.Create(context.Background(), item); err != nil {
					b.Fatalf("creating test item %d: %v", i, err)
				}
			}

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				items, err := repo.LoadInventory(ctx, ownerID)
				if err != nil {
					b.Errorf("LoadInventory failed: %v", err)
				}
				if len(items) != size {
					b.Errorf("expected %d items, got %d", size, len(items))
				}
			}
		})
	}
}

// Benchmark LoadPaperdoll — проверяет эффективность pre-allocation для equipment
func BenchmarkItemRepository_LoadPaperdoll(b *testing.B) {
	pool := setupTestDB(b)
	repo := NewItemRepository(pool)
	charRepo := NewCharacterRepository(pool)

	// Setup test character (required for FK constraint)
	player, err := model.NewPlayer(0, 0, 1, "PaperdollTestOwner", 75, 0, 0)
	if err != nil {
		b.Fatalf("creating player: %v", err)
	}
	if err := charRepo.Create(context.Background(), player); err != nil {
		b.Fatalf("creating test player: %v", err)
	}

	// Используем реальный characterID из БД
	ownerID := player.CharacterID()

	// Setup test character с полной экипировкой (14 slots)
	const numSlots = 14

	for i := 0; i < numSlots; i++ {
		item, err := model.NewItem(ownerID, 2000+int32(i), 1)
		if err != nil {
			b.Fatalf("creating item %d: %v", i, err)
		}
		item.SetLocation(model.ItemLocationPaperdoll, int32(i))

		if err := repo.Create(context.Background(), item); err != nil {
			b.Fatalf("creating test item %d: %v", i, err)
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items, err := repo.LoadPaperdoll(ctx, ownerID)
		if err != nil {
			b.Errorf("LoadPaperdoll failed: %v", err)
		}
		if len(items) != numSlots {
			b.Errorf("expected %d items, got %d", numSlots, len(items))
		}
	}
}
