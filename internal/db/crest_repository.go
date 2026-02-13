package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/game/crest"
)

// CrestRepository implements crest.Store backed by PostgreSQL.
type CrestRepository struct {
	pool *pgxpool.Pool
}

// Compile-time check.
var _ crest.Store = (*CrestRepository)(nil)

// NewCrestRepository creates a new crest repository.
func NewCrestRepository(pool *pgxpool.Pool) *CrestRepository {
	return &CrestRepository{pool: pool}
}

// LoadAll fetches every crest row from the database.
func (r *CrestRepository) LoadAll(ctx context.Context) ([]crest.CrestRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT crest_id, data, type FROM crests`)
	if err != nil {
		return nil, fmt.Errorf("query crests: %w", err)
	}
	defer rows.Close()

	var result []crest.CrestRow
	for rows.Next() {
		var row crest.CrestRow
		if err := rows.Scan(&row.CrestID, &row.Data, &row.Type); err != nil {
			return nil, fmt.Errorf("scan crest row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate crest rows: %w", err)
	}
	return result, nil
}

// Insert adds a new crest to the database.
func (r *CrestRepository) Insert(ctx context.Context, row crest.CrestRow) error {
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO crests (crest_id, data, type) VALUES ($1, $2, $3)`,
		row.CrestID, row.Data, row.Type); err != nil {
		return fmt.Errorf("insert crest %d: %w", row.CrestID, err)
	}
	return nil
}

// Delete removes a crest from the database.
func (r *CrestRepository) Delete(ctx context.Context, crestID int32) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM crests WHERE crest_id = $1`, crestID); err != nil {
		return fmt.Errorf("delete crest %d: %w", crestID, err)
	}
	return nil
}
