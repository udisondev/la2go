package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubClassRow represents a single subclass record from DB.
type SubClassRow struct {
	ClassID    int32
	ClassIndex int32
	Exp        int64
	SP         int64
	Level      int32
}

// SubClassRepository manages character_subclasses table.
type SubClassRepository struct {
	db *pgxpool.Pool
}

// NewSubClassRepository creates a new SubClassRepository.
func NewSubClassRepository(db *pgxpool.Pool) *SubClassRepository {
	return &SubClassRepository{db: db}
}

// LoadByCharacterID loads all subclasses for a character.
func (r *SubClassRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]SubClassRow, error) {
	query := `
		SELECT class_id, class_index, exp, sp, level
		FROM character_subclasses
		WHERE character_id = $1
		ORDER BY class_index ASC
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying subclasses for character %d: %w", charID, err)
	}
	defer rows.Close()

	result := make([]SubClassRow, 0, 3)
	for rows.Next() {
		var s SubClassRow
		if err := rows.Scan(&s.ClassID, &s.ClassIndex, &s.Exp, &s.SP, &s.Level); err != nil {
			return nil, fmt.Errorf("scanning subclass row: %w", err)
		}
		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating subclass rows: %w", err)
	}

	return result, nil
}

// Insert adds a new subclass record.
func (r *SubClassRepository) Insert(ctx context.Context, charID int64, row SubClassRow) error {
	query := `
		INSERT INTO character_subclasses (character_id, class_id, class_index, exp, sp, level)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.db.Exec(ctx, query, charID, row.ClassID, row.ClassIndex, row.Exp, row.SP, row.Level); err != nil {
		return fmt.Errorf("inserting subclass index %d for character %d: %w", row.ClassIndex, charID, err)
	}

	return nil
}

// Update updates an existing subclass record.
func (r *SubClassRepository) Update(ctx context.Context, charID int64, row SubClassRow) error {
	query := `
		UPDATE character_subclasses
		SET exp = $1, sp = $2, level = $3, class_id = $4
		WHERE character_id = $5 AND class_index = $6
	`

	if _, err := r.db.Exec(ctx, query, row.Exp, row.SP, row.Level, row.ClassID, charID, row.ClassIndex); err != nil {
		return fmt.Errorf("updating subclass index %d for character %d: %w", row.ClassIndex, charID, err)
	}

	return nil
}

// Delete removes a subclass by index.
func (r *SubClassRepository) Delete(ctx context.Context, charID int64, classIndex int32) error {
	query := `DELETE FROM character_subclasses WHERE character_id = $1 AND class_index = $2`

	if _, err := r.db.Exec(ctx, query, charID, classIndex); err != nil {
		return fmt.Errorf("deleting subclass index %d for character %d: %w", classIndex, charID, err)
	}

	return nil
}

// SaveAllTx saves all subclasses within an existing transaction (full replace).
func (r *SubClassRepository) SaveAllTx(ctx context.Context, tx pgx.Tx, charID int64, rows []SubClassRow) error {
	if _, err := tx.Exec(ctx, `DELETE FROM character_subclasses WHERE character_id = $1`, charID); err != nil {
		return fmt.Errorf("deleting existing subclasses: %w", err)
	}

	for _, s := range rows {
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_subclasses (character_id, class_id, class_index, exp, sp, level)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			charID, s.ClassID, s.ClassIndex, s.Exp, s.SP, s.Level,
		); err != nil {
			return fmt.Errorf("inserting subclass index %d: %w", s.ClassIndex, err)
		}
	}

	return nil
}

// Save saves all subclasses using a standalone transaction.
func (r *SubClassRepository) Save(ctx context.Context, charID int64, rows []SubClassRow) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("subclass rollback failed", "characterID", charID, "error", err)
		}
	}()

	if err := r.SaveAllTx(ctx, tx, charID, rows); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing subclasses save: %w", err)
	}

	return nil
}
