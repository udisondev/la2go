package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/game/manor"
)

// ManorRepository provides DB access for castle manor data.
type ManorRepository struct {
	pool *pgxpool.Pool
}

// NewManorRepository creates a new manor repository.
func NewManorRepository(pool *pgxpool.Pool) *ManorRepository {
	return &ManorRepository{pool: pool}
}

// LoadProduction loads seed production entries for a castle.
func (r *ManorRepository) LoadProduction(ctx context.Context, castleID int32) ([]manor.ProductionRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT castle_id, seed_id, amount, start_amount, price, next_period
		 FROM castle_manor_production WHERE castle_id = $1`, castleID)
	if err != nil {
		return nil, fmt.Errorf("query production castle %d: %w", castleID, err)
	}
	defer rows.Close()

	var result []manor.ProductionRow
	for rows.Next() {
		var row manor.ProductionRow
		if err := rows.Scan(&row.CastleID, &row.SeedID, &row.Amount, &row.StartAmount, &row.Price, &row.NextPeriod); err != nil {
			return nil, fmt.Errorf("scan production row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate production rows: %w", err)
	}
	return result, nil
}

// LoadProcure loads crop procurement entries for a castle.
func (r *ManorRepository) LoadProcure(ctx context.Context, castleID int32) ([]manor.ProcureRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT castle_id, crop_id, amount, start_amount, price, reward_type, next_period
		 FROM castle_manor_procure WHERE castle_id = $1`, castleID)
	if err != nil {
		return nil, fmt.Errorf("query procure castle %d: %w", castleID, err)
	}
	defer rows.Close()

	var result []manor.ProcureRow
	for rows.Next() {
		var row manor.ProcureRow
		if err := rows.Scan(&row.CastleID, &row.CropID, &row.Amount, &row.StartAmount, &row.Price, &row.RewardType, &row.NextPeriod); err != nil {
			return nil, fmt.Errorf("scan procure row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate procure rows: %w", err)
	}
	return result, nil
}

// SaveAll replaces all manor data for a castle (delete + insert).
func (r *ManorRepository) SaveAll(ctx context.Context, castleID int32, production []manor.ProductionRow, procure []manor.ProcureRow) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Удаление старых данных.
	if _, err := tx.Exec(ctx, `DELETE FROM castle_manor_production WHERE castle_id = $1`, castleID); err != nil {
		return fmt.Errorf("delete production castle %d: %w", castleID, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM castle_manor_procure WHERE castle_id = $1`, castleID); err != nil {
		return fmt.Errorf("delete procure castle %d: %w", castleID, err)
	}

	// Вставка production.
	for _, row := range production {
		if _, err := tx.Exec(ctx,
			`INSERT INTO castle_manor_production (castle_id, seed_id, amount, start_amount, price, next_period)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			row.CastleID, row.SeedID, row.Amount, row.StartAmount, row.Price, row.NextPeriod); err != nil {
			return fmt.Errorf("insert production seed %d: %w", row.SeedID, err)
		}
	}

	// Вставка procure.
	for _, row := range procure {
		if _, err := tx.Exec(ctx,
			`INSERT INTO castle_manor_procure (castle_id, crop_id, amount, start_amount, price, reward_type, next_period)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			row.CastleID, row.CropID, row.Amount, row.StartAmount, row.Price, row.RewardType, row.NextPeriod); err != nil {
			return fmt.Errorf("insert procure crop %d: %w", row.CropID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit manor tx: %w", err)
	}
	return nil
}

// DeleteAll removes all manor data for a castle.
func (r *ManorRepository) DeleteAll(ctx context.Context, castleID int32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM castle_manor_production WHERE castle_id = $1`, castleID); err != nil {
		return fmt.Errorf("delete production castle %d: %w", castleID, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM castle_manor_procure WHERE castle_id = $1`, castleID); err != nil {
		return fmt.Errorf("delete procure castle %d: %w", castleID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete tx: %w", err)
	}
	return nil
}
