package offlinetrade

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// mockRepo implements Repository for testing.
type mockRepo struct {
	saved   []*Trader
	deleted []int64
	items   map[int64][]TradeEntry
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		items: make(map[int64][]TradeEntry),
	}
}

func (r *mockRepo) SaveTrader(_ context.Context, trader *Trader) error {
	r.saved = append(r.saved, trader)
	return nil
}

func (r *mockRepo) DeleteTrader(_ context.Context, characterID int64) error {
	r.deleted = append(r.deleted, characterID)
	return nil
}

func (r *mockRepo) LoadAll(_ context.Context) ([]*Trader, error) {
	return nil, nil
}

func (r *mockRepo) UpdateItems(_ context.Context, characterID int64, items []TradeEntry) error {
	r.items[characterID] = items
	return nil
}

func (r *mockRepo) DeleteAll(_ context.Context) error {
	return nil
}

func newTestPlayer(t *testing.T, objectID uint32, charID int64) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, charID, 1, "TestTrader", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	return p
}

func TestNewService(t *testing.T) {
	cfg := Config{Enabled: true, MaxDays: 7}
	svc := NewService(cfg, nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if !svc.Enabled() {
		t.Error("Enabled() = false, want true")
	}
	if svc.Count() != 0 {
		t.Errorf("Count() = %d, want 0", svc.Count())
	}
}

func TestService_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	svc := NewService(cfg, nil)
	if svc.Enabled() {
		t.Error("Enabled() = true, want false")
	}
}

func TestService_EnteredOfflineMode(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true, RealtimeSave: true}
	svc := NewService(cfg, repo)

	player := newTestPlayer(t, 1001, 42)
	// Настраиваем sell list
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ObjectID: 100, ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)
	player.SetPrivateStoreType(model.StoreSell)
	player.SetStoreMessage("Selling items!")

	ctx := context.Background()
	err := svc.EnteredOfflineMode(ctx, player, 1001, "testaccount")
	if err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	if svc.Count() != 1 {
		t.Errorf("Count() = %d, want 1", svc.Count())
	}
	if !svc.IsOfflineTrader(1001) {
		t.Error("IsOfflineTrader(1001) = false, want true")
	}
	if !svc.IsCharacterOffline(42) {
		t.Error("IsCharacterOffline(42) = false, want true")
	}

	// Проверяем что trader сохранён в repo (realtime save)
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved count = %d, want 1", len(repo.saved))
	}
	if repo.saved[0].CharacterID != 42 {
		t.Errorf("saved trader CharacterID = %d, want 42", repo.saved[0].CharacterID)
	}
}

func TestService_EnteredOfflineMode_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	svc := NewService(cfg, nil)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	err := svc.EnteredOfflineMode(context.Background(), player, 1001, "test")
	if err == nil {
		t.Error("EnteredOfflineMode should fail when disabled")
	}
}

func TestService_EnteredOfflineMode_NotInStoreMode(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	player := newTestPlayer(t, 1001, 42)
	// Player NOT in store mode (StoreNone)

	err := svc.EnteredOfflineMode(context.Background(), player, 1001, "test")
	if err == nil {
		t.Error("EnteredOfflineMode should fail when not in store mode")
	}
}

func TestService_EnteredOfflineMode_EmptyStore(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	// No sell list set → items empty

	err := svc.EnteredOfflineMode(context.Background(), player, 1001, "test")
	if err == nil {
		t.Error("EnteredOfflineMode should fail with empty store")
	}
}

func TestService_EnteredOfflineMode_BuyStore(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	player := newTestPlayer(t, 1001, 42)
	buyList := model.NewTradeList()
	buyList.AddItem(&model.TradeItem{ItemID: 57, Count: 5, Price: 500})
	player.SetBuyList(buyList)
	player.SetPrivateStoreType(model.StoreBuy)
	player.SetStoreMessage("Buying stuff")

	err := svc.EnteredOfflineMode(context.Background(), player, 1001, "test")
	if err != nil {
		t.Fatalf("EnteredOfflineMode for buy store: %v", err)
	}
	if svc.Count() != 1 {
		t.Errorf("Count() = %d, want 1", svc.Count())
	}
}

func TestService_RemoveTrader(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true}
	svc := NewService(cfg, repo)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	ctx := context.Background()
	if err := svc.EnteredOfflineMode(ctx, player, 1001, "test"); err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	removed := svc.RemoveTrader(ctx, 1001)
	if removed == nil {
		t.Fatal("RemoveTrader returned nil")
	}
	if svc.Count() != 0 {
		t.Errorf("Count() = %d, want 0", svc.Count())
	}

	// Проверяем delete в repo
	if len(repo.deleted) != 1 {
		t.Fatalf("repo.deleted count = %d, want 1", len(repo.deleted))
	}
	if repo.deleted[0] != 42 {
		t.Errorf("deleted charID = %d, want 42", repo.deleted[0])
	}
}

func TestService_RemoveByCharacter(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true}
	svc := NewService(cfg, repo)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	ctx := context.Background()
	if err := svc.EnteredOfflineMode(ctx, player, 1001, "test"); err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	removed := svc.RemoveByCharacter(ctx, 42)
	if removed == nil {
		t.Fatal("RemoveByCharacter returned nil")
	}
	if removed.ObjectID != 1001 {
		t.Errorf("removed.ObjectID = %d, want 1001", removed.ObjectID)
	}
	if svc.Count() != 0 {
		t.Errorf("Count() = %d, want 0", svc.Count())
	}
}

func TestService_OnTransaction_RealtimeSave(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true, RealtimeSave: true}
	svc := NewService(cfg, repo)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	ctx := context.Background()
	if err := svc.EnteredOfflineMode(ctx, player, 1001, "test"); err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	// Обновляем items после транзакции
	newItems := []TradeEntry{{ItemIdentifier: 57, Count: 5, Price: 1000}}
	svc.OnTransaction(ctx, 1001, newItems)

	// Проверяем realtime save
	if items, ok := repo.items[42]; ok {
		if len(items) != 1 {
			t.Errorf("repo items count = %d, want 1", len(items))
		}
		if items[0].Count != 5 {
			t.Errorf("repo items[0].Count = %d, want 5", items[0].Count)
		}
	} else {
		t.Error("repo items not saved for charID 42")
	}
}

func TestService_OnTransaction_DisconnectFinished(t *testing.T) {
	cfg := Config{Enabled: true, DisconnectFinished: true}
	svc := NewService(cfg, nil)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	ctx := context.Background()
	if err := svc.EnteredOfflineMode(ctx, player, 1001, "test"); err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	// Магазин опустошается — items = nil
	svc.OnTransaction(ctx, 1001, nil)

	// Трейдер должен быть удалён
	if svc.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after empty transaction", svc.Count())
	}
}

func TestService_SaveAll(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true}
	svc := NewService(cfg, repo)

	for i := range 3 {
		player := newTestPlayer(t, uint32(1000+i), int64(i+1))
		player.SetPrivateStoreType(model.StoreSell)
		sellList := model.NewTradeList()
		sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
		player.SetSellList(sellList)

		if err := svc.EnteredOfflineMode(context.Background(), player, uint32(1000+i), "test"); err != nil {
			t.Fatalf("EnteredOfflineMode: %v", err)
		}
	}

	ctx := context.Background()
	if err := svc.SaveAll(ctx); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	if len(repo.saved) != 6 { // 3 from EnteredOfflineMode (realtime=false) + 3 from SaveAll... wait
		// realtime save is false, so EnteredOfflineMode doesn't save.
		// SaveAll saves all 3.
	}
	// Проверяем что SaveAll сохранил 3 трейдера
	// repo.saved has 3 entries from SaveAll
	if len(repo.saved) != 3 {
		t.Errorf("repo.saved count = %d, want 3", len(repo.saved))
	}
}

func TestService_Shutdown(t *testing.T) {
	repo := newMockRepo()
	cfg := Config{Enabled: true}
	svc := NewService(cfg, repo)

	player := newTestPlayer(t, 1001, 42)
	player.SetPrivateStoreType(model.StoreSell)
	sellList := model.NewTradeList()
	sellList.AddItem(&model.TradeItem{ItemID: 57, Count: 10, Price: 1000})
	player.SetSellList(sellList)

	if err := svc.EnteredOfflineMode(context.Background(), player, 1001, "test"); err != nil {
		t.Fatalf("EnteredOfflineMode: %v", err)
	}

	svc.Shutdown(context.Background())

	// После shutdown таблица должна быть пустой
	if svc.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after Shutdown", svc.Count())
	}

	// Проверяем что сохранение произошло
	if len(repo.saved) != 1 {
		t.Errorf("repo.saved count = %d, want 1", len(repo.saved))
	}
}

func TestService_MaxDays(t *testing.T) {
	cfg := Config{Enabled: true, MaxDays: 1}
	svc := NewService(cfg, nil)

	// MaxDays=1 → maxDur = 24h
	if svc.table.maxDur != 24*time.Hour {
		t.Errorf("table.maxDur = %v, want 24h", svc.table.maxDur)
	}
}

func TestService_MaxDaysZero(t *testing.T) {
	cfg := Config{Enabled: true, MaxDays: 0}
	svc := NewService(cfg, nil)

	// MaxDays=0 → unlimited
	if svc.table.maxDur != 0 {
		t.Errorf("table.maxDur = %v, want 0 (unlimited)", svc.table.maxDur)
	}
}

func TestService_IsOfflineTrader_NotFound(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	if svc.IsOfflineTrader(9999) {
		t.Error("IsOfflineTrader(9999) = true, want false")
	}
}

func TestService_IsCharacterOffline_NotFound(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	if svc.IsCharacterOffline(9999) {
		t.Error("IsCharacterOffline(9999) = true, want false")
	}
}

func TestService_RemoveTrader_NotFound(t *testing.T) {
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	removed := svc.RemoveTrader(context.Background(), 9999)
	if removed != nil {
		t.Error("RemoveTrader non-existent should return nil")
	}
}

func TestService_NoRepoSaveAll(t *testing.T) {
	// Без repo — SaveAll не паникует
	cfg := Config{Enabled: true}
	svc := NewService(cfg, nil)

	err := svc.SaveAll(context.Background())
	if err != nil {
		t.Fatalf("SaveAll without repo: %v", err)
	}
}
