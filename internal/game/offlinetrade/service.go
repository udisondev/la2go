package offlinetrade

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// Config holds offline trade configuration.
type Config struct {
	Enabled              bool
	MaxDays              int   // 0 = unlimited
	DisconnectFinished   bool  // remove trader when all items sold
	SetNameColor         bool  // change name color for offline traders
	NameColor            int32 // RGB color for offline trader name
	RealtimeSave         bool  // save to DB after each transaction
	RestoreOnStartup     bool  // restore offline traders on server start
}

// Service orchestrates offline trade operations.
// Bridges the in-memory Table with the DB Repository.
type Service struct {
	table  *Table
	repo   Repository
	config Config
}

// NewService creates a new offline trade service.
func NewService(cfg Config, repo Repository) *Service {
	var maxDur time.Duration
	if cfg.MaxDays > 0 {
		maxDur = time.Duration(cfg.MaxDays) * 24 * time.Hour
	}

	tbl := NewTable(maxDur)

	return &Service{
		table:  tbl,
		repo:   repo,
		config: cfg,
	}
}

// Table returns the in-memory trader table.
func (s *Service) Table() *Table {
	return s.table
}

// Enabled returns true if offline trade is enabled in config.
func (s *Service) Enabled() bool {
	return s.config.Enabled
}

// EnteredOfflineMode transitions a player into offline trade mode.
// Called from handleLogout / OnDisconnection when player has active store.
//
// Flow:
// 1. Build Trader from player state
// 2. Register in memory table
// 3. Save to DB (if realtime save enabled)
// 4. Set offline name color (if configured)
//
// Returns error if registration fails.
//
// Java reference: OfflineTradeUtil.enteredOfflineMode()
func (s *Service) EnteredOfflineMode(ctx context.Context, player *model.Player, objectID uint32, accountName string) error {
	if !s.config.Enabled {
		return fmt.Errorf("offline trade disabled")
	}

	storeType := player.PrivateStoreType()
	if !storeType.IsInStoreMode() {
		return fmt.Errorf("player not in store mode: %v", storeType)
	}

	// Собираем торговые предметы из списков игрока
	items := s.collectTradeEntries(player, storeType)
	if len(items) == 0 {
		return fmt.Errorf("no items in store")
	}

	trader := &Trader{
		CharacterID: player.CharacterID(),
		ObjectID:    objectID,
		AccountName: accountName,
		StoreType:   storeType,
		Title:       player.StoreMessage(),
		Items:       items,
		StartedAt:   time.Now(),
	}

	if err := s.table.Add(trader); err != nil {
		return fmt.Errorf("register offline trader: %w", err)
	}

	// Offline trader name color change deferred: requires Player.SetNameColor() + CharInfo update.
	// if s.config.SetNameColor {
	//     player.SetNameColor(s.config.NameColor)
	// }

	// Сохраняем в DB (если включен realtime save)
	if s.config.RealtimeSave && s.repo != nil {
		if err := s.repo.SaveTrader(ctx, trader); err != nil {
			slog.Error("save offline trader to DB",
				"characterID", trader.CharacterID,
				"error", err)
		}
	}

	slog.Info("player entered offline trade mode",
		"characterID", player.CharacterID(),
		"objectID", objectID,
		"storeType", storeType,
		"items", len(items))

	return nil
}

// OnTransaction updates trader items after a purchase.
// Called when another player buys/sells from the offline store.
// Removes the trader if store becomes empty (and config allows it).
func (s *Service) OnTransaction(ctx context.Context, objectID uint32, newItems []TradeEntry) {
	s.table.UpdateTraderItems(objectID, newItems)

	// Проверяем: если магазин опустел и DisconnectFinished — удаляем трейдера
	if s.config.DisconnectFinished && s.table.RemoveIfEmpty(objectID) {
		slog.Info("offline trader removed (store empty after transaction)",
			"objectID", objectID)

		// Удаляем из DB
		if s.repo != nil {
			trader := s.table.Get(objectID) // already removed, use charID from callback
			if trader != nil {
				if err := s.repo.DeleteTrader(ctx, trader.CharacterID); err != nil {
					slog.Error("delete empty offline trader from DB",
						"objectID", objectID,
						"error", err)
				}
			}
		}
		return
	}

	// Realtime save: обновляем items в DB
	if s.config.RealtimeSave && s.repo != nil {
		trader := s.table.Get(objectID)
		if trader != nil {
			if err := s.repo.UpdateItems(ctx, trader.CharacterID, newItems); err != nil {
				slog.Error("realtime save offline trader items",
					"objectID", objectID,
					"error", err)
			}
		}
	}
}

// SaveAll saves all offline traders to DB (batch save).
// Called during graceful shutdown.
func (s *Service) SaveAll(ctx context.Context) error {
	if s.repo == nil {
		return nil
	}

	traders := s.table.ExportAll()
	if len(traders) == 0 {
		return nil
	}

	// Очищаем старые записи и сохраняем текущие
	if err := s.repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("delete old traders: %w", err)
	}

	var saved int
	for _, trader := range traders {
		if err := s.repo.SaveTrader(ctx, trader); err != nil {
			slog.Error("save offline trader on shutdown",
				"characterID", trader.CharacterID,
				"error", err)
			continue
		}
		saved++
	}

	slog.Info("offline traders saved on shutdown", "count", saved, "total", len(traders))
	return nil
}

// RestoreFromDB loads offline traders from DB into memory table.
// Called on server startup if RestoreOnStartup is enabled.
// Returns the number of restored traders.
func (s *Service) RestoreFromDB(ctx context.Context) (int, error) {
	if s.repo == nil || !s.config.RestoreOnStartup {
		return 0, nil
	}

	traders, err := s.repo.LoadAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("load offline traders: %w", err)
	}

	var restored int
	for _, trader := range traders {
		// ObjectID будет назначен при добавлении в мир (WorldObject)
		// На этапе загрузки — пропускаем трейдеров без ObjectID
		if trader.ObjectID == 0 {
			slog.Debug("skip offline trader without objectID",
				"characterID", trader.CharacterID)
			continue
		}

		if err := s.table.Add(trader); err != nil {
			slog.Warn("skip invalid offline trader",
				"characterID", trader.CharacterID,
				"error", err)
			continue
		}
		restored++
	}

	if restored > 0 {
		slog.Info("offline traders restored from DB", "count", restored)
	}

	// Очищаем DB после восстановления (Java pattern: delete after restore)
	if err := s.repo.DeleteAll(ctx); err != nil {
		slog.Error("cleanup DB after restore", "error", err)
	}

	return restored, nil
}

// Shutdown stops all timers and saves traders to DB.
func (s *Service) Shutdown(ctx context.Context) {
	if err := s.SaveAll(ctx); err != nil {
		slog.Error("offline trade shutdown save", "error", err)
	}
	s.table.StopAll()
}

// IsOfflineTrader returns true if the objectID belongs to an offline trader.
func (s *Service) IsOfflineTrader(objectID uint32) bool {
	return s.table.IsOfflineTrader(objectID)
}

// IsCharacterOffline returns true if the character is offline trading.
func (s *Service) IsCharacterOffline(characterID int64) bool {
	return s.table.IsCharacterOffline(characterID)
}

// RemoveTrader removes an offline trader (e.g., when player reconnects).
func (s *Service) RemoveTrader(ctx context.Context, objectID uint32) *Trader {
	trader := s.table.Remove(objectID)
	if trader == nil {
		return nil
	}

	if s.repo != nil {
		if err := s.repo.DeleteTrader(ctx, trader.CharacterID); err != nil {
			slog.Error("delete offline trader from DB on remove",
				"characterID", trader.CharacterID,
				"error", err)
		}
	}

	return trader
}

// RemoveByCharacter removes an offline trader by character ID.
// Used when player reconnects (lookup by character, not objectID).
func (s *Service) RemoveByCharacter(ctx context.Context, characterID int64) *Trader {
	trader := s.table.RemoveByCharacter(characterID)
	if trader == nil {
		return nil
	}

	if s.repo != nil {
		if err := s.repo.DeleteTrader(ctx, trader.CharacterID); err != nil {
			slog.Error("delete offline trader from DB on reconnect",
				"characterID", trader.CharacterID,
				"error", err)
		}
	}

	return trader
}

// Count returns the number of active offline traders.
func (s *Service) Count() int {
	return s.table.Count()
}

// collectTradeEntries extracts trade entries from player's active trade lists.
func (s *Service) collectTradeEntries(player *model.Player, storeType model.PrivateStoreType) []TradeEntry {
	switch storeType {
	case model.StoreSell, model.StorePackageSell:
		list := player.SellList()
		if list == nil {
			return nil
		}
		return tradeItemsToEntries(list.Items())

	case model.StoreBuy:
		list := player.BuyList()
		if list == nil {
			return nil
		}
		return tradeItemsToEntries(list.Items())

	default:
		return nil
	}
}

// tradeItemsToEntries converts model.TradeItem pointer slice to TradeEntry slice.
func tradeItemsToEntries(items []*model.TradeItem) []TradeEntry {
	if len(items) == 0 {
		return nil
	}
	entries := make([]TradeEntry, len(items))
	for i, item := range items {
		entries[i] = TradeEntry{
			ItemIdentifier: item.ItemID,
			Count:          int64(item.Count),
			Price:          item.Price,
		}
	}
	return entries
}
