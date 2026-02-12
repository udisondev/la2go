package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// PlayerPersistenceService атомарно сохраняет/загружает данные игрока.
// MVP: последовательное сохранение (character → items → skills), без единой транзакции.
type PlayerPersistenceService struct {
	pool      *pgxpool.Pool
	charRepo  *CharacterRepository
	itemRepo  *ItemRepository
	skillRepo *SkillRepository
}

// NewPlayerPersistenceService создаёт новый сервис.
func NewPlayerPersistenceService(
	pool *pgxpool.Pool,
	charRepo *CharacterRepository,
	itemRepo *ItemRepository,
	skillRepo *SkillRepository,
) *PlayerPersistenceService {
	return &PlayerPersistenceService{
		pool:      pool,
		charRepo:  charRepo,
		itemRepo:  itemRepo,
		skillRepo: skillRepo,
	}
}

// SavePlayer saves all player data (character, items, skills) in a single transaction.
// Ensures consistency: either all data is saved or none.
func (s *PlayerPersistenceService) SavePlayer(ctx context.Context, player *model.Player) error {
	charID := player.CharacterID()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction for character %d: %w", charID, err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("rollback failed", "characterID", charID, "error", err)
		}
	}()

	// 1. Save character (location, stats, exp, sp, level)
	if err := s.charRepo.UpdateTx(ctx, tx, player); err != nil {
		return fmt.Errorf("saving character %d: %w", charID, err)
	}

	// 2. Save items
	items := player.Inventory().GetItems()
	itemRows := make([]ItemRow, 0, len(items))
	for _, item := range items {
		itemRows = append(itemRows, ItemRow{
			ItemTypeID: item.ItemID(),
			OwnerID:    charID,
			Count:      item.Count(),
			Enchant:    item.Enchant(),
			Location:   int32(item.Location()),
			SlotID:     item.Slot(),
		})
	}

	if err := s.itemRepo.SaveAllTx(ctx, tx, charID, itemRows); err != nil {
		return fmt.Errorf("saving items for character %d: %w", charID, err)
	}

	// 3. Save skills
	skills := player.Skills()
	if err := s.skillRepo.SaveTx(ctx, tx, charID, skills); err != nil {
		return fmt.Errorf("saving skills for character %d: %w", charID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction for character %d: %w", charID, err)
	}

	slog.Info("player data saved",
		"characterID", charID,
		"character", player.Name(),
		"items", len(itemRows),
		"skills", len(skills))

	return nil
}

// LoadPlayerData загружает items и skills для существующего player.
func (s *PlayerPersistenceService) LoadPlayerData(ctx context.Context, charID int64) ([]ItemRow, []*model.SkillInfo, error) {
	itemRows, err := s.itemRepo.LoadByOwner(ctx, charID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading items for character %d: %w", charID, err)
	}

	skills, err := s.skillRepo.LoadByCharacterID(ctx, charID)
	if err != nil {
		return nil, nil, fmt.Errorf("loading skills for character %d: %w", charID, err)
	}

	return itemRows, skills, nil
}

// ItemDefToTemplate ищет item definition по ID и конвертирует в *model.ItemTemplate.
// Возвращает nil если item definition не найден.
func ItemDefToTemplate(itemID int32) *model.ItemTemplate {
	def := data.GetItemDef(itemID)
	if def == nil {
		return nil
	}
	return &model.ItemTemplate{
		ItemID:      def.ID(),
		Name:        def.Name(),
		Type:        itemTypeFromString(def.Type()),
		PAtk:        def.PAtk(),
		AttackRange: def.AttackRange(),
		PDef:        def.PDef(),
		Weight:      def.Weight(),
		Stackable:   def.IsStackable(),
		Tradeable:   def.IsTradeable(),
	}
}

// itemTypeFromString конвертирует строку типа в model.ItemType.
func itemTypeFromString(s string) model.ItemType {
	switch s {
	case "Weapon":
		return model.ItemTypeWeapon
	case "Armor":
		return model.ItemTypeArmor
	default:
		return model.ItemTypeEtcItem
	}
}
