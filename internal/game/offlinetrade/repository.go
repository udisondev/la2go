package offlinetrade

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// Repository defines the persistence interface for offline traders.
type Repository interface {
	// SaveTrader saves an offline trader to the database (upsert).
	SaveTrader(ctx context.Context, trader *Trader) error

	// DeleteTrader removes an offline trader from the database.
	DeleteTrader(ctx context.Context, characterID int64) error

	// LoadAll loads all offline traders from the database.
	LoadAll(ctx context.Context) ([]*Trader, error)

	// UpdateItems replaces the item list for a specific offline trader.
	UpdateItems(ctx context.Context, characterID int64, items []TradeEntry) error

	// DeleteAll removes all offline traders from the database.
	DeleteAll(ctx context.Context) error
}

// PgRepository implements Repository using PostgreSQL.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PostgreSQL-backed offline trade repository.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

// SaveTrader saves an offline trader to the database.
// Uses a transaction to atomically update both trade header and items.
func (r *PgRepository) SaveTrader(ctx context.Context, trader *Trader) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert trade header
	_, err = tx.Exec(ctx, `
		INSERT INTO character_offline_trade (char_id, created_at, store_type, title)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (char_id) DO UPDATE SET
			created_at = EXCLUDED.created_at,
			store_type = EXCLUDED.store_type,
			title      = EXCLUDED.title`,
		trader.CharacterID,
		trader.StartedAt.Unix(),
		int16(trader.StoreType),
		trader.Title,
	)
	if err != nil {
		return fmt.Errorf("upsert trade header: %w", err)
	}

	// Replace items: delete old, insert new
	_, err = tx.Exec(ctx, `DELETE FROM character_offline_trade_items WHERE char_id = $1`, trader.CharacterID)
	if err != nil {
		return fmt.Errorf("delete old items: %w", err)
	}

	if len(trader.Items) > 0 {
		batch := &pgx.Batch{}
		for _, item := range trader.Items {
			batch.Queue(
				`INSERT INTO character_offline_trade_items (char_id, item, count, price) VALUES ($1, $2, $3, $4)`,
				trader.CharacterID, item.ItemIdentifier, item.Count, item.Price,
			)
		}
		br := tx.SendBatch(ctx, batch)
		for range trader.Items {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("insert item: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// DeleteTrader removes an offline trader from the database.
func (r *PgRepository) DeleteTrader(ctx context.Context, characterID int64) error {
	// Items are cascade-deleted via FK.
	_, err := r.pool.Exec(ctx, `DELETE FROM character_offline_trade WHERE char_id = $1`, characterID)
	if err != nil {
		return fmt.Errorf("delete trader %d: %w", characterID, err)
	}
	return nil
}

// LoadAll loads all offline traders from the database.
// Returns traders with their items populated.
func (r *PgRepository) LoadAll(ctx context.Context) ([]*Trader, error) {
	// Load headers
	rows, err := r.pool.Query(ctx, `
		SELECT char_id, created_at, store_type, title
		FROM character_offline_trade
		ORDER BY char_id`)
	if err != nil {
		return nil, fmt.Errorf("query traders: %w", err)
	}
	defer rows.Close()

	traderMap := make(map[int64]*Trader)
	var traders []*Trader

	for rows.Next() {
		var (
			charID    int64
			createdAt int64
			storeType int16
			title     string
		)
		if err := rows.Scan(&charID, &createdAt, &storeType, &title); err != nil {
			return nil, fmt.Errorf("scan trader: %w", err)
		}
		t := &Trader{
			CharacterID: charID,
			StoreType:   model.PrivateStoreType(storeType),
			Title:       title,
			StartedAt:   time.Unix(createdAt, 0),
		}
		traderMap[charID] = t
		traders = append(traders, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate traders: %w", err)
	}

	if len(traders) == 0 {
		return traders, nil
	}

	// Load items for all traders
	itemRows, err := r.pool.Query(ctx, `
		SELECT char_id, item, count, price
		FROM character_offline_trade_items
		ORDER BY char_id`)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var (
			charID int64
			itemID int32
			count  int64
			price  int64
		)
		if err := itemRows.Scan(&charID, &itemID, &count, &price); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if t, ok := traderMap[charID]; ok {
			t.Items = append(t.Items, TradeEntry{
				ItemIdentifier: itemID,
				Count:          count,
				Price:          price,
			})
		}
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return traders, nil
}

// UpdateItems replaces the item list for a specific offline trader.
// Used for realtime save after each transaction.
func (r *PgRepository) UpdateItems(ctx context.Context, characterID int64, items []TradeEntry) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM character_offline_trade_items WHERE char_id = $1`, characterID)
	if err != nil {
		return fmt.Errorf("delete items: %w", err)
	}

	if len(items) > 0 {
		batch := &pgx.Batch{}
		for _, item := range items {
			batch.Queue(
				`INSERT INTO character_offline_trade_items (char_id, item, count, price) VALUES ($1, $2, $3, $4)`,
				characterID, item.ItemIdentifier, item.Count, item.Price,
			)
		}
		br := tx.SendBatch(ctx, batch)
		for range items {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("insert item: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// DeleteAll removes all offline traders from the database.
// Called on graceful shutdown after batch-saving current state.
func (r *PgRepository) DeleteAll(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM character_offline_trade`)
	if err != nil {
		return fmt.Errorf("delete all: %w", err)
	}
	return nil
}
