package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// ItemRepository управляет предметами в БД.
type ItemRepository struct {
	db *pgxpool.Pool
}

// NewItemRepository создаёт новый ItemRepository.
func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// LoadInventory загружает все предметы игрока из инвентаря.
func (r *ItemRepository) LoadInventory(ctx context.Context, ownerID int64) ([]*model.Item, error) {
	query := `
		SELECT item_id, owner_id, item_type, count, enchant, location, slot_id, created_at
		FROM items
		WHERE owner_id = $1 AND location = $2
		ORDER BY item_id
	`

	rows, err := r.db.Query(ctx, query, ownerID, int32(model.ItemLocationInventory))
	if err != nil {
		return nil, fmt.Errorf("querying inventory for owner %d: %w", ownerID, err)
	}
	defer rows.Close()

	// Pre-allocate для типичного инвентаря (20-100 items).
	// Capacity 50 покрывает 80% случаев без overallocation.
	items := make([]*model.Item, 0, 50)

	for rows.Next() {
		var itemID int64
		var ownerIDDB int64
		var itemType int32
		var count int32
		var enchant int32
		var location int32
		var slotID int32
		var createdAt time.Time

		err := rows.Scan(
			&itemID, &ownerIDDB, &itemType, &count, &enchant, &location, &slotID, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}

		// Создаём Item
		item, err := model.NewItem(ownerIDDB, itemType, count)
		if err != nil {
			return nil, fmt.Errorf("creating item model: %w", err)
		}

		// Устанавливаем остальные поля
		item.SetItemID(itemID)
		_ = item.SetEnchant(enchant)
		item.SetLocation(model.ItemLocation(location), slotID)
		item.SetCreatedAt(createdAt)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating item rows: %w", err)
	}

	return items, nil
}

// LoadPaperdoll загружает экипировку игрока (все equipped items).
func (r *ItemRepository) LoadPaperdoll(ctx context.Context, ownerID int64) ([]*model.Item, error) {
	query := `
		SELECT item_id, owner_id, item_type, count, enchant, location, slot_id, created_at
		FROM items
		WHERE owner_id = $1 AND location = $2
		ORDER BY slot_id
	`

	rows, err := r.db.Query(ctx, query, ownerID, int32(model.ItemLocationPaperdoll))
	if err != nil {
		return nil, fmt.Errorf("querying paperdoll for owner %d: %w", ownerID, err)
	}
	defer rows.Close()

	// Pre-allocate для paperdoll (14 equipment slots + weapons).
	// Capacity 20 покрывает все случаи.
	items := make([]*model.Item, 0, 20)

	for rows.Next() {
		var itemID int64
		var ownerIDDB int64
		var itemType int32
		var count int32
		var enchant int32
		var location int32
		var slotID int32
		var createdAt time.Time

		err := rows.Scan(
			&itemID, &ownerIDDB, &itemType, &count, &enchant, &location, &slotID, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}

		// Создаём Item
		item, err := model.NewItem(ownerIDDB, itemType, count)
		if err != nil {
			return nil, fmt.Errorf("creating item model: %w", err)
		}

		// Устанавливаем остальные поля
		item.SetItemID(itemID)
		_ = item.SetEnchant(enchant)
		item.SetLocation(model.ItemLocation(location), slotID)
		item.SetCreatedAt(createdAt)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating item rows: %w", err)
	}

	return items, nil
}

// Create создаёт новый предмет в БД.
func (r *ItemRepository) Create(ctx context.Context, item *model.Item) error {
	query := `
		INSERT INTO items (owner_id, item_type, count, enchant, location, slot_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING item_id, created_at
	`

	loc, slotID := item.Location()

	var itemID int64
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query,
		item.OwnerID(), item.ItemType(), item.Count(), item.Enchant(), int32(loc), slotID,
	).Scan(&itemID, &createdAt)

	if err != nil {
		return fmt.Errorf("creating item: %w", err)
	}

	// Устанавливаем ID и createdAt который вернула БД
	item.SetItemID(itemID)
	item.SetCreatedAt(createdAt)

	return nil
}

// Update обновляет предмет в БД.
func (r *ItemRepository) Update(ctx context.Context, item *model.Item) error {
	query := `
		UPDATE items
		SET count = $2, enchant = $3, location = $4, slot_id = $5
		WHERE item_id = $1
	`

	loc, slotID := item.Location()

	_, err := r.db.Exec(ctx, query,
		item.ItemID(), item.Count(), item.Enchant(), int32(loc), slotID,
	)

	if err != nil {
		return fmt.Errorf("updating item %d: %w", item.ItemID(), err)
	}

	return nil
}

// Delete удаляет предмет из БД.
func (r *ItemRepository) Delete(ctx context.Context, itemID int64) error {
	query := `DELETE FROM items WHERE item_id = $1`

	result, err := r.db.Exec(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("deleting item %d: %w", itemID, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("item %d not found", itemID)
	}

	return nil
}
