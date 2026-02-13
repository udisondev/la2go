package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// --- helpers ---

func benchSeedItems(b *testing.B, itemRepo *ItemRepository, charID int64, count int) {
	b.Helper()
	items := make([]ItemRow, count)
	for i := range count {
		items[i] = ItemRow{
			ItemTypeID: int32(i + 100),
			OwnerID:    charID,
			Count:      1,
			Enchant:    0,
			Location:   0,
			SlotID:     -1,
		}
	}
	if err := itemRepo.SaveAll(context.Background(), charID, items); err != nil {
		b.Fatalf("seeding items: %v", err)
	}
}

func benchSeedSkills(b *testing.B, skillRepo *SkillRepository, charID int64, count int) {
	b.Helper()
	skills := make([]*model.SkillInfo, count)
	for i := range count {
		skills[i] = &model.SkillInfo{
			SkillID: int32(i + 1),
			Level:   1,
		}
	}
	if err := skillRepo.Save(context.Background(), charID, skills); err != nil {
		b.Fatalf("seeding skills: %v", err)
	}
}

func makeItemRows(count int, charID int64) []ItemRow {
	items := make([]ItemRow, count)
	for i := range count {
		items[i] = ItemRow{
			ItemTypeID: int32(i + 100),
			OwnerID:    charID,
			Count:      1,
			Enchant:    0,
			Location:   0,
			SlotID:     -1,
		}
	}
	return items
}

func makeSkills(count int) []*model.SkillInfo {
	skills := make([]*model.SkillInfo, count)
	for i := range count {
		skills[i] = &model.SkillInfo{
			SkillID: int32(i + 1),
			Level:   1,
		}
	}
	return skills
}

// --- ItemDefToTemplate benchmarks (no DB, pure computation) ---

func init() {
	// Ensure item templates are loaded for ItemDefToTemplate benchmarks
	if data.ItemTable == nil {
		if err := data.LoadItemTemplates(); err != nil {
			// Non-fatal, some benchmarks will skip
		}
	}
}

// BenchmarkItemDefToTemplate_Hit benchmarks template lookup for existing item.
// Expected: ~10-50ns (map lookup + struct allocation).
func BenchmarkItemDefToTemplate_Hit(b *testing.B) {
	b.ReportAllocs()
	// Find any valid item ID from loaded data
	var itemID int32
	for id := range data.ItemTable {
		itemID = id
		break
	}
	if itemID == 0 {
		b.Skip("no items loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = ItemDefToTemplate(itemID)
	}
}

// BenchmarkItemDefToTemplate_Miss benchmarks template lookup for non-existing item.
// Expected: ~5-10ns (map miss, return nil).
func BenchmarkItemDefToTemplate_Miss(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		_ = ItemDefToTemplate(999999)
	}
}

// --- ItemRepository benchmarks (requires DB) ---

// BenchmarkItemRepository_SaveAll_10Items benchmarks saving 10 items.
// Expected: ~1-5ms (DELETE + 10 INSERTs + COMMIT).
func BenchmarkItemRepository_SaveAll_10Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	items := makeItemRows(10, charID)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := itemRepo.SaveAll(ctx, charID, items); err != nil {
			b.Fatalf("SaveAll: %v", err)
		}
	}
}

// BenchmarkItemRepository_SaveAll_50Items benchmarks saving 50 items.
// Expected: ~2-10ms (DELETE + 50 INSERTs + COMMIT).
func BenchmarkItemRepository_SaveAll_50Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	items := makeItemRows(50, charID)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := itemRepo.SaveAll(ctx, charID, items); err != nil {
			b.Fatalf("SaveAll: %v", err)
		}
	}
}

// BenchmarkItemRepository_SaveAll_200Items benchmarks saving max inventory (200 items).
// Expected: ~5-20ms (DELETE + 200 INSERTs + COMMIT).
func BenchmarkItemRepository_SaveAll_200Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	items := makeItemRows(200, charID)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := itemRepo.SaveAll(ctx, charID, items); err != nil {
			b.Fatalf("SaveAll: %v", err)
		}
	}
}

// BenchmarkItemRepository_LoadByOwner_NoItems benchmarks loading with 0 items.
// Expected: ~500us-1ms (SELECT, 0 rows).
func BenchmarkItemRepository_LoadByOwner_NoItems(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		_, err := itemRepo.LoadByOwner(ctx, charID)
		if err != nil {
			b.Fatalf("LoadByOwner: %v", err)
		}
	}
}

// BenchmarkItemRepository_LoadByOwner_50Items benchmarks loading 50 items.
// Expected: ~1-5ms (SELECT + 50 row scans).
func BenchmarkItemRepository_LoadByOwner_50Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	benchSeedItems(b, itemRepo, charID, 50)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		items, err := itemRepo.LoadByOwner(ctx, charID)
		if err != nil {
			b.Fatalf("LoadByOwner: %v", err)
		}
		if len(items) != 50 {
			b.Fatalf("expected 50 items, got %d", len(items))
		}
	}
}

// BenchmarkItemRepository_LoadByOwner_200Items benchmarks loading max inventory.
// Expected: ~2-10ms (SELECT + 200 row scans).
func BenchmarkItemRepository_LoadByOwner_200Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	benchSeedItems(b, itemRepo, charID, 200)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		items, err := itemRepo.LoadByOwner(ctx, charID)
		if err != nil {
			b.Fatalf("LoadByOwner: %v", err)
		}
		if len(items) != 200 {
			b.Fatalf("expected 200 items, got %d", len(items))
		}
	}
}

// --- SkillRepository benchmarks ---

// BenchmarkSkillRepository_Save_10Skills benchmarks saving 10 skills.
// Expected: ~1-5ms (DELETE + 10 INSERTs + COMMIT).
func BenchmarkSkillRepository_Save_10Skills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	skillRepo := NewSkillRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	skills := makeSkills(10)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := skillRepo.Save(ctx, charID, skills); err != nil {
			b.Fatalf("Save: %v", err)
		}
	}
}

// BenchmarkSkillRepository_Save_50Skills benchmarks saving 50 skills.
// Expected: ~2-10ms (DELETE + 50 INSERTs + COMMIT).
func BenchmarkSkillRepository_Save_50Skills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	skillRepo := NewSkillRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	skills := makeSkills(50)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := skillRepo.Save(ctx, charID, skills); err != nil {
			b.Fatalf("Save: %v", err)
		}
	}
}

// BenchmarkSkillRepository_Save_100Skills benchmarks saving 100 skills.
// Expected: ~5-20ms (DELETE + 100 INSERTs + COMMIT).
func BenchmarkSkillRepository_Save_100Skills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	skillRepo := NewSkillRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	skills := makeSkills(100)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := skillRepo.Save(ctx, charID, skills); err != nil {
			b.Fatalf("Save: %v", err)
		}
	}
}

// BenchmarkSkillRepository_LoadByCharacterID_NoSkills benchmarks loading with 0 skills.
// Expected: ~500us-1ms (SELECT, 0 rows).
func BenchmarkSkillRepository_LoadByCharacterID_NoSkills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	skillRepo := NewSkillRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		_, err := skillRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			b.Fatalf("LoadByCharacterID: %v", err)
		}
	}
}

// BenchmarkSkillRepository_LoadByCharacterID_50Skills benchmarks loading 50 skills.
// Expected: ~1-5ms (SELECT + 50 row scans).
func BenchmarkSkillRepository_LoadByCharacterID_50Skills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	skillRepo := NewSkillRepository(pool)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()
	benchSeedSkills(b, skillRepo, charID, 50)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		skills, err := skillRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			b.Fatalf("LoadByCharacterID: %v", err)
		}
		if len(skills) != 50 {
			b.Fatalf("expected 50 skills, got %d", len(skills))
		}
	}
}

// --- PlayerPersistenceService benchmarks ---

// BenchmarkPlayerPersistence_SavePlayer_NoItems benchmarks full save (character only, 0 items/skills).
// Expected: ~5-10ms (UPDATE character + DELETE items + DELETE skills).
func BenchmarkPlayerPersistence_SavePlayer_NoItems(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)
	skillRepo := NewSkillRepository(pool)
	svc := NewPlayerPersistenceService(pool, charRepo, itemRepo, skillRepo, nil, nil, nil, nil)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(17000, 170000, -3500, 0))
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := svc.SavePlayer(ctx, player); err != nil {
			b.Fatalf("SavePlayer: %v", err)
		}
	}
}

// BenchmarkPlayerPersistence_SavePlayer_50Items benchmarks full save with 50 items.
// Expected: ~10-20ms (UPDATE + DELETE items + 50 INSERTs + DELETE skills).
func BenchmarkPlayerPersistence_SavePlayer_50Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)
	skillRepo := NewSkillRepository(pool)
	svc := NewPlayerPersistenceService(pool, charRepo, itemRepo, skillRepo, nil, nil, nil, nil)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(17000, 170000, -3500, 0))
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}

	// Add items to player inventory
	for i := range 50 {
		tmpl := &model.ItemTemplate{
			ItemID:    int32(i + 100),
			Name:      fmt.Sprintf("Item%d", i),
			Type:      model.ItemTypeEtcItem,
			Stackable: true,
		}
		item, err := model.NewItem(uint32(i+1), int32(i+100), player.CharacterID(), 1, tmpl)
		if err != nil {
			b.Fatalf("NewItem: %v", err)
		}
		if err := player.Inventory().AddItem(item); err != nil {
			b.Fatalf("AddItem: %v", err)
		}
	}

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := svc.SavePlayer(ctx, player); err != nil {
			b.Fatalf("SavePlayer: %v", err)
		}
	}
}

// BenchmarkPlayerPersistence_SavePlayer_WithSkills benchmarks full save with items + skills.
// Expected: ~15-30ms (UPDATE + 50 items + 50 skills).
func BenchmarkPlayerPersistence_SavePlayer_WithSkills(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)
	skillRepo := NewSkillRepository(pool)
	svc := NewPlayerPersistenceService(pool, charRepo, itemRepo, skillRepo, nil, nil, nil, nil)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(17000, 170000, -3500, 0))
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}

	// Add items
	for i := range 50 {
		tmpl := &model.ItemTemplate{
			ItemID:    int32(i + 100),
			Name:      fmt.Sprintf("Item%d", i),
			Type:      model.ItemTypeEtcItem,
			Stackable: true,
		}
		item, err := model.NewItem(uint32(i+1), int32(i+100), player.CharacterID(), 1, tmpl)
		if err != nil {
			b.Fatalf("NewItem: %v", err)
		}
		if err := player.Inventory().AddItem(item); err != nil {
			b.Fatalf("AddItem: %v", err)
		}
	}

	// Add skills
	for i := range 50 {
		player.AddSkill(int32(i+1), 1, false)
	}

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if err := svc.SavePlayer(ctx, player); err != nil {
			b.Fatalf("SavePlayer: %v", err)
		}
	}
}

// BenchmarkPlayerPersistence_LoadPlayerData_NoItems benchmarks loading 0 items/skills.
// Expected: ~1-3ms (2 SELECTs, 0 rows).
func BenchmarkPlayerPersistence_LoadPlayerData_NoItems(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)
	skillRepo := NewSkillRepository(pool)
	svc := NewPlayerPersistenceService(pool, charRepo, itemRepo, skillRepo, nil, nil, nil, nil)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		_, err := svc.LoadPlayerData(ctx, charID)
		if err != nil {
			b.Fatalf("LoadPlayerData: %v", err)
		}
	}
}

// BenchmarkPlayerPersistence_LoadPlayerData_50Items benchmarks loading 50 items + 50 skills.
// Expected: ~2-10ms (2 SELECTs + 100 row scans).
func BenchmarkPlayerPersistence_LoadPlayerData_50Items(b *testing.B) {
	pool := setupTestDB(b)
	charRepo := NewCharacterRepository(pool)
	itemRepo := NewItemRepository(pool)
	skillRepo := NewSkillRepository(pool)
	svc := NewPlayerPersistenceService(pool, charRepo, itemRepo, skillRepo, nil, nil, nil, nil)

	player, err := model.NewPlayer(0, 0, 1, "BenchHero", 40, 0, 0)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	if err := charRepo.Create(context.Background(), "benchaccount", player); err != nil {
		b.Fatalf("creating player: %v", err)
	}
	charID := player.CharacterID()

	benchSeedItems(b, itemRepo, charID, 50)
	benchSeedSkills(b, skillRepo, charID, 50)

	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		pd, err := svc.LoadPlayerData(ctx, charID)
		if err != nil {
			b.Fatalf("LoadPlayerData: %v", err)
		}
		if len(pd.Items) != 50 {
			b.Fatalf("expected 50 items, got %d", len(pd.Items))
		}
		if len(pd.Skills) != 50 {
			b.Fatalf("expected 50 skills, got %d", len(pd.Skills))
		}
	}
}
