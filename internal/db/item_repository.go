package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ItemRow представляет строку таблицы items для загрузки/сохранения.
type ItemRow struct {
	ItemTypeID int32 // item_type в БД (= model.Item.ItemID() = template ID)
	OwnerID    int64 // character_id
	Count      int32
	Enchant    int32
	Location   int32 // 0=inventory, 1=paperdoll
	SlotID     int32 // paperdoll slot (-1 если не equipped)
}

// ItemRepository управляет предметами в БД.
type ItemRepository struct {
	db *pgxpool.Pool
}

// NewItemRepository создаёт новый ItemRepository.
func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// LoadByOwner загружает все предметы персонажа.
func (r *ItemRepository) LoadByOwner(ctx context.Context, ownerID int64) ([]ItemRow, error) {
	query := `
		SELECT item_type, count, enchant, location, slot_id
		FROM items
		WHERE owner_id = $1
		ORDER BY item_id
	`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("querying items for owner %d: %w", ownerID, err)
	}
	defer rows.Close()

	items := make([]ItemRow, 0, 32)
	for rows.Next() {
		var row ItemRow
		row.OwnerID = ownerID
		if err := rows.Scan(&row.ItemTypeID, &row.Count, &row.Enchant, &row.Location, &row.SlotID); err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}
		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating item rows: %w", err)
	}

	return items, nil
}

// SaveAll сохраняет все предметы персонажа (полная перезапись).
// Удаляет старые, вставляет новые в одной транзакции.
func (r *ItemRepository) SaveAll(ctx context.Context, ownerID int64, items []ItemRow) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `DELETE FROM items WHERE owner_id = $1`, ownerID); err != nil {
		return fmt.Errorf("deleting existing items: %w", err)
	}

	for _, item := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO items (owner_id, item_type, count, enchant, location, slot_id) VALUES ($1, $2, $3, $4, $5, $6)`,
			ownerID, item.ItemTypeID, item.Count, item.Enchant, item.Location, item.SlotID,
		); err != nil {
			return fmt.Errorf("inserting item type %d: %w", item.ItemTypeID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing items save: %w", err)
	}

	return nil
}
