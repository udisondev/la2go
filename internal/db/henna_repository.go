package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HennaRow represents a single henna record from DB.
type HennaRow struct {
	Slot       int32
	DyeID      int32
	ClassIndex int32
}

// HennaRepository manages character_hennas table.
type HennaRepository struct {
	db *pgxpool.Pool
}

// NewHennaRepository creates a new HennaRepository.
func NewHennaRepository(db *pgxpool.Pool) *HennaRepository {
	return &HennaRepository{db: db}
}

// LoadByCharacterID loads all hennas for a character (class_index=0).
func (r *HennaRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]HennaRow, error) {
	query := `
		SELECT slot, dye_id, class_index
		FROM character_hennas
		WHERE character_id = $1 AND class_index = 0
		ORDER BY slot
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying hennas for character %d: %w", charID, err)
	}
	defer rows.Close()

	hennas := make([]HennaRow, 0, 3)
	for rows.Next() {
		var h HennaRow
		if err := rows.Scan(&h.Slot, &h.DyeID, &h.ClassIndex); err != nil {
			return nil, fmt.Errorf("scanning henna row: %w", err)
		}
		hennas = append(hennas, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating henna rows: %w", err)
	}

	return hennas, nil
}

// AddHenna inserts a henna into the given slot.
func (r *HennaRepository) AddHenna(ctx context.Context, charID int64, slot int32, dyeID int32) error {
	query := `
		INSERT INTO character_hennas (character_id, slot, dye_id, class_index)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (character_id, slot, class_index)
		DO UPDATE SET dye_id = $3
	`

	if _, err := r.db.Exec(ctx, query, charID, slot, dyeID); err != nil {
		return fmt.Errorf("inserting henna slot %d for character %d: %w", slot, charID, err)
	}

	return nil
}

// RemoveHenna removes a henna from the given slot.
func (r *HennaRepository) RemoveHenna(ctx context.Context, charID int64, slot int32) error {
	query := `DELETE FROM character_hennas WHERE character_id = $1 AND slot = $2 AND class_index = 0`

	if _, err := r.db.Exec(ctx, query, charID, slot); err != nil {
		return fmt.Errorf("deleting henna slot %d for character %d: %w", slot, charID, err)
	}

	return nil
}

// SaveAllTx saves all hennas within an existing transaction (full replace).
func (r *HennaRepository) SaveAllTx(ctx context.Context, tx pgx.Tx, charID int64, hennas []HennaRow) error {
	if _, err := tx.Exec(ctx, `DELETE FROM character_hennas WHERE character_id = $1 AND class_index = 0`, charID); err != nil {
		return fmt.Errorf("deleting existing hennas: %w", err)
	}

	for _, h := range hennas {
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_hennas (character_id, slot, dye_id, class_index) VALUES ($1, $2, $3, 0)`,
			charID, h.Slot, h.DyeID,
		); err != nil {
			return fmt.Errorf("inserting henna slot %d: %w", h.Slot, err)
		}
	}

	return nil
}

// Save saves all hennas using a standalone transaction.
func (r *HennaRepository) Save(ctx context.Context, charID int64, hennas []HennaRow) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("henna rollback failed", "characterID", charID, "error", err)
		}
	}()

	if err := r.SaveAllTx(ctx, tx, charID, hennas); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing hennas save: %w", err)
	}

	return nil
}
