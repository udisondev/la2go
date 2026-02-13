package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RaidBossSpawnRow represents a row from raidboss_spawnlist.
type RaidBossSpawnRow struct {
	BossID      int32
	RespawnTime int64   // Unix seconds
	CurrentHP   float64
	CurrentMP   float64
}

// GrandBossRow represents a row from grandboss_data.
type GrandBossRow struct {
	BossID      int32
	LocX        int32
	LocY        int32
	LocZ        int32
	Heading     int32
	CurrentHP   float64
	CurrentMP   float64
	RespawnTime int64 // Unix seconds
	Status      int16
}

// RaidPointsRow represents a row from character_raid_points.
type RaidPointsRow struct {
	CharacterID int32
	BossID      int32
	Points      int32
}

// RaidRepository provides CRUD for raid boss DB tables.
type RaidRepository struct {
	pool *pgxpool.Pool
}

// NewRaidRepository creates a new RaidRepository.
func NewRaidRepository(pool *pgxpool.Pool) *RaidRepository {
	return &RaidRepository{pool: pool}
}

// --- raidboss_spawnlist ---

// LoadAllRaidSpawns loads all raid boss spawn records.
func (r *RaidRepository) LoadAllRaidSpawns(ctx context.Context) ([]RaidBossSpawnRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT boss_id, respawn_time, current_hp, current_mp FROM raidboss_spawnlist`)
	if err != nil {
		return nil, fmt.Errorf("query raidboss_spawnlist: %w", err)
	}
	defer rows.Close()

	var result []RaidBossSpawnRow
	for rows.Next() {
		var row RaidBossSpawnRow
		if err := rows.Scan(&row.BossID, &row.RespawnTime, &row.CurrentHP, &row.CurrentMP); err != nil {
			return nil, fmt.Errorf("scan raidboss_spawnlist: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// SaveRaidSpawn inserts or updates a raid boss spawn record.
func (r *RaidRepository) SaveRaidSpawn(ctx context.Context, row RaidBossSpawnRow) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO raidboss_spawnlist (boss_id, respawn_time, current_hp, current_mp)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (boss_id) DO UPDATE SET
		   respawn_time = EXCLUDED.respawn_time,
		   current_hp   = EXCLUDED.current_hp,
		   current_mp   = EXCLUDED.current_mp`,
		row.BossID, row.RespawnTime, row.CurrentHP, row.CurrentMP)
	if err != nil {
		return fmt.Errorf("upsert raidboss_spawnlist boss %d: %w", row.BossID, err)
	}
	return nil
}

// DeleteRaidSpawn removes a raid boss spawn record.
func (r *RaidRepository) DeleteRaidSpawn(ctx context.Context, bossID int32) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM raidboss_spawnlist WHERE boss_id = $1`, bossID)
	if err != nil {
		return fmt.Errorf("delete raidboss_spawnlist boss %d: %w", bossID, err)
	}
	return nil
}

// --- grandboss_data ---

// LoadAllGrandBosses loads all grand boss state records.
func (r *RaidRepository) LoadAllGrandBosses(ctx context.Context) ([]GrandBossRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT boss_id, loc_x, loc_y, loc_z, heading,
		        current_hp, current_mp, respawn_time, status
		 FROM grandboss_data`)
	if err != nil {
		return nil, fmt.Errorf("query grandboss_data: %w", err)
	}
	defer rows.Close()

	var result []GrandBossRow
	for rows.Next() {
		var row GrandBossRow
		if err := rows.Scan(
			&row.BossID, &row.LocX, &row.LocY, &row.LocZ, &row.Heading,
			&row.CurrentHP, &row.CurrentMP, &row.RespawnTime, &row.Status,
		); err != nil {
			return nil, fmt.Errorf("scan grandboss_data: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// SaveGrandBoss inserts or updates a grand boss state record.
func (r *RaidRepository) SaveGrandBoss(ctx context.Context, row GrandBossRow) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO grandboss_data
		   (boss_id, loc_x, loc_y, loc_z, heading, current_hp, current_mp, respawn_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (boss_id) DO UPDATE SET
		   loc_x = EXCLUDED.loc_x, loc_y = EXCLUDED.loc_y, loc_z = EXCLUDED.loc_z,
		   heading = EXCLUDED.heading,
		   current_hp = EXCLUDED.current_hp, current_mp = EXCLUDED.current_mp,
		   respawn_time = EXCLUDED.respawn_time, status = EXCLUDED.status`,
		row.BossID, row.LocX, row.LocY, row.LocZ, row.Heading,
		row.CurrentHP, row.CurrentMP, row.RespawnTime, row.Status)
	if err != nil {
		return fmt.Errorf("upsert grandboss_data boss %d: %w", row.BossID, err)
	}
	return nil
}

// GetGrandBoss loads a single grand boss record by ID.
func (r *RaidRepository) GetGrandBoss(ctx context.Context, bossID int32) (*GrandBossRow, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT boss_id, loc_x, loc_y, loc_z, heading,
		        current_hp, current_mp, respawn_time, status
		 FROM grandboss_data WHERE boss_id = $1`, bossID)

	var gb GrandBossRow
	if err := row.Scan(
		&gb.BossID, &gb.LocX, &gb.LocY, &gb.LocZ, &gb.Heading,
		&gb.CurrentHP, &gb.CurrentMP, &gb.RespawnTime, &gb.Status,
	); err != nil {
		return nil, fmt.Errorf("get grandboss_data boss %d: %w", bossID, err)
	}
	return &gb, nil
}

// --- character_raid_points ---

// LoadRaidPoints loads all raid points for a character.
func (r *RaidRepository) LoadRaidPoints(ctx context.Context, characterID int32) ([]RaidPointsRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT character_id, boss_id, points
		 FROM character_raid_points WHERE character_id = $1`, characterID)
	if err != nil {
		return nil, fmt.Errorf("query raid points for character %d: %w", characterID, err)
	}
	defer rows.Close()

	var result []RaidPointsRow
	for rows.Next() {
		var row RaidPointsRow
		if err := rows.Scan(&row.CharacterID, &row.BossID, &row.Points); err != nil {
			return nil, fmt.Errorf("scan raid points: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// AddRaidPoints adds points for a character's raid boss kill.
func (r *RaidRepository) AddRaidPoints(ctx context.Context, characterID, bossID, points int32) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO character_raid_points (character_id, boss_id, points)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (character_id, boss_id) DO UPDATE SET
		   points = character_raid_points.points + EXCLUDED.points`,
		characterID, bossID, points)
	if err != nil {
		return fmt.Errorf("add raid points char %d boss %d: %w", characterID, bossID, err)
	}
	return nil
}

// GetTotalRaidPoints returns total raid points for a character.
func (r *RaidRepository) GetTotalRaidPoints(ctx context.Context, characterID int32) (int32, error) {
	var total int32
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(points), 0) FROM character_raid_points
		 WHERE character_id = $1`, characterID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum raid points for character %d: %w", characterID, err)
	}
	return total, nil
}

// GetTopRaidPointPlayers returns top N players by total raid points.
func (r *RaidRepository) GetTopRaidPointPlayers(ctx context.Context, limit int) ([]RaidPointsRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT character_id, 0 AS boss_id, SUM(points) AS points
		 FROM character_raid_points
		 GROUP BY character_id
		 ORDER BY points DESC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("top raid point players: %w", err)
	}
	defer rows.Close()

	var result []RaidPointsRow
	for rows.Next() {
		var row RaidPointsRow
		if err := rows.Scan(&row.CharacterID, &row.BossID, &row.Points); err != nil {
			return nil, fmt.Errorf("scan top raid points: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ResetAllRaidPoints deletes all raid points (weekly reset).
func (r *RaidRepository) ResetAllRaidPoints(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM character_raid_points`)
	if err != nil {
		return 0, fmt.Errorf("reset raid points: %w", err)
	}
	return result.RowsAffected(), nil
}
