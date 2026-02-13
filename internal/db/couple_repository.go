package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// CoupleRepository provides database access for the couples table.
// Phase 33: Marriage System.
type CoupleRepository struct {
	pool *pgxpool.Pool
}

// NewCoupleRepository creates a new CoupleRepository.
func NewCoupleRepository(pool *pgxpool.Pool) *CoupleRepository {
	return &CoupleRepository{pool: pool}
}

// CoupleRow represents a database row from the couples table.
type CoupleRow struct {
	ID          int32
	Player1ID   int32
	Player2ID   int32
	Married     bool
	AffiancedAt time.Time
	MarriedAt   *time.Time
}

// LoadAll loads all couple records from the database.
func (r *CoupleRepository) LoadAll(ctx context.Context) ([]CoupleRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, player1_id, player2_id, married, affianced_at, married_at
		 FROM couples ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query couples: %w", err)
	}
	defer rows.Close()

	var result []CoupleRow
	for rows.Next() {
		var row CoupleRow
		if err := rows.Scan(&row.ID, &row.Player1ID, &row.Player2ID,
			&row.Married, &row.AffiancedAt, &row.MarriedAt); err != nil {
			return nil, fmt.Errorf("scan couple: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// Create inserts a new couple record and returns the generated ID.
func (r *CoupleRepository) Create(ctx context.Context, p1ID, p2ID int32) (int32, error) {
	// Ensure player1_id < player2_id (CHECK constraint).
	if p1ID > p2ID {
		p1ID, p2ID = p2ID, p1ID
	}

	var id int32
	err := r.pool.QueryRow(ctx,
		`INSERT INTO couples (player1_id, player2_id, married, affianced_at)
		 VALUES ($1, $2, false, NOW())
		 RETURNING id`,
		p1ID, p2ID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert couple: %w", err)
	}
	return id, nil
}

// UpdateMarried marks a couple as married and sets the wedding date.
func (r *CoupleRepository) UpdateMarried(ctx context.Context, coupleID int32) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE couples SET married = true, married_at = NOW() WHERE id = $1`,
		coupleID)
	if err != nil {
		return fmt.Errorf("update married: %w", err)
	}
	return nil
}

// Delete removes a couple record by ID.
func (r *CoupleRepository) Delete(ctx context.Context, coupleID int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM couples WHERE id = $1`, coupleID)
	if err != nil {
		return fmt.Errorf("delete couple: %w", err)
	}
	return nil
}

// FindByPlayer looks up a couple by either player ID.
func (r *CoupleRepository) FindByPlayer(ctx context.Context, playerID int32) (*CoupleRow, error) {
	var row CoupleRow
	err := r.pool.QueryRow(ctx,
		`SELECT id, player1_id, player2_id, married, affianced_at, married_at
		 FROM couples
		 WHERE player1_id = $1 OR player2_id = $1`,
		playerID).Scan(&row.ID, &row.Player1ID, &row.Player2ID,
		&row.Married, &row.AffiancedAt, &row.MarriedAt)
	if err != nil {
		return nil, fmt.Errorf("find couple by player %d: %w", playerID, err)
	}
	return &row, nil
}

// RowToCouple converts a CoupleRow to a model.Couple.
func RowToCouple(row CoupleRow) *model.Couple {
	return &model.Couple{
		ID:          row.ID,
		Player1ID:   row.Player1ID,
		Player2ID:   row.Player2ID,
		Married:     row.Married,
		AffiancedAt: row.AffiancedAt,
		MarriedAt:   row.MarriedAt,
	}
}
