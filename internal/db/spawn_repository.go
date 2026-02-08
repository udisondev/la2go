package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// SpawnRepository handles spawn CRUD operations
type SpawnRepository struct {
	pool *pgxpool.Pool
}

// NewSpawnRepository creates a new spawn repository
func NewSpawnRepository(pool *pgxpool.Pool) *SpawnRepository {
	return &SpawnRepository{pool: pool}
}

// LoadAll loads all spawns from database
func (r *SpawnRepository) LoadAll(ctx context.Context) ([]*model.Spawn, error) {
	query := `
		SELECT spawn_id, template_id, x, y, z, heading, maximum_count, do_respawn
		FROM spawns
		ORDER BY spawn_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("loading all spawns: %w", err)
	}
	defer rows.Close()

	spawns := make([]*model.Spawn, 0, 50) // pre-allocate for typical count

	for rows.Next() {
		var (
			spawnID      int64
			templateID   int32
			x, y, z      int32
			heading      uint16
			maximumCount int32
			doRespawn    bool
		)

		if err := rows.Scan(&spawnID, &templateID, &x, &y, &z, &heading, &maximumCount, &doRespawn); err != nil {
			return nil, fmt.Errorf("scanning spawn row: %w", err)
		}

		spawn := model.NewSpawn(spawnID, templateID, x, y, z, heading, maximumCount, doRespawn)
		spawns = append(spawns, spawn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating spawn rows: %w", err)
	}

	return spawns, nil
}

// LoadByID loads spawn by ID
func (r *SpawnRepository) LoadByID(ctx context.Context, spawnID int64) (*model.Spawn, error) {
	query := `
		SELECT spawn_id, template_id, x, y, z, heading, maximum_count, do_respawn
		FROM spawns
		WHERE spawn_id = $1
	`

	var (
		id           int64
		templateID   int32
		x, y, z      int32
		heading      uint16
		maximumCount int32
		doRespawn    bool
	)

	err := r.pool.QueryRow(ctx, query, spawnID).Scan(
		&id, &templateID, &x, &y, &z, &heading, &maximumCount, &doRespawn,
	)
	if err != nil {
		return nil, fmt.Errorf("loading spawn %d: %w", spawnID, err)
	}

	return model.NewSpawn(id, templateID, x, y, z, heading, maximumCount, doRespawn), nil
}

// Create creates new spawn
func (r *SpawnRepository) Create(ctx context.Context, spawn *model.Spawn) (int64, error) {
	query := `
		INSERT INTO spawns (template_id, x, y, z, heading, maximum_count, do_respawn)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING spawn_id
	`

	loc := spawn.Location()
	var spawnID int64

	err := r.pool.QueryRow(ctx, query,
		spawn.TemplateID(),
		loc.X,
		loc.Y,
		loc.Z,
		loc.Heading,
		spawn.MaximumCount(),
		spawn.DoRespawn(),
	).Scan(&spawnID)

	if err != nil {
		return 0, fmt.Errorf("creating spawn for template %d: %w", spawn.TemplateID(), err)
	}

	return spawnID, nil
}
