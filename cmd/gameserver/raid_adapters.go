package main

import (
	"context"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/raid"
)

// raidSpawnStoreAdapter adapts db.RaidRepository to raid.RaidSpawnStore.
type raidSpawnStoreAdapter struct {
	repo *db.RaidRepository
}

func (a *raidSpawnStoreAdapter) LoadAllRaidSpawns(ctx context.Context) ([]raid.RaidSpawnRow, error) {
	rows, err := a.repo.LoadAllRaidSpawns(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]raid.RaidSpawnRow, len(rows))
	for i, r := range rows {
		result[i] = raid.RaidSpawnRow{
			BossID:      r.BossID,
			RespawnTime: r.RespawnTime,
			CurrentHP:   r.CurrentHP,
			CurrentMP:   r.CurrentMP,
		}
	}
	return result, nil
}

func (a *raidSpawnStoreAdapter) SaveRaidSpawn(ctx context.Context, row raid.RaidSpawnRow) error {
	return a.repo.SaveRaidSpawn(ctx, db.RaidBossSpawnRow{
		BossID:      row.BossID,
		RespawnTime: row.RespawnTime,
		CurrentHP:   row.CurrentHP,
		CurrentMP:   row.CurrentMP,
	})
}

func (a *raidSpawnStoreAdapter) DeleteRaidSpawn(ctx context.Context, bossID int32) error {
	return a.repo.DeleteRaidSpawn(ctx, bossID)
}

// grandBossStoreAdapter adapts db.RaidRepository to raid.GrandBossStore.
type grandBossStoreAdapter struct {
	repo *db.RaidRepository
}

func (a *grandBossStoreAdapter) LoadAllGrandBosses(ctx context.Context) ([]raid.GrandBossDataRow, error) {
	rows, err := a.repo.LoadAllGrandBosses(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]raid.GrandBossDataRow, len(rows))
	for i, r := range rows {
		result[i] = raid.GrandBossDataRow{
			BossID:      r.BossID,
			LocX:        r.LocX,
			LocY:        r.LocY,
			LocZ:        r.LocZ,
			Heading:     r.Heading,
			CurrentHP:   r.CurrentHP,
			CurrentMP:   r.CurrentMP,
			RespawnTime: r.RespawnTime,
			Status:      r.Status,
		}
	}
	return result, nil
}

func (a *grandBossStoreAdapter) SaveGrandBoss(ctx context.Context, row raid.GrandBossDataRow) error {
	return a.repo.SaveGrandBoss(ctx, db.GrandBossRow{
		BossID:      row.BossID,
		LocX:        row.LocX,
		LocY:        row.LocY,
		LocZ:        row.LocZ,
		Heading:     row.Heading,
		CurrentHP:   row.CurrentHP,
		CurrentMP:   row.CurrentMP,
		RespawnTime: row.RespawnTime,
		Status:      row.Status,
	})
}

func (a *grandBossStoreAdapter) GetGrandBoss(ctx context.Context, bossID int32) (*raid.GrandBossDataRow, error) {
	row, err := a.repo.GetGrandBoss(ctx, bossID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return &raid.GrandBossDataRow{
		BossID:      row.BossID,
		LocX:        row.LocX,
		LocY:        row.LocY,
		LocZ:        row.LocZ,
		Heading:     row.Heading,
		CurrentHP:   row.CurrentHP,
		CurrentMP:   row.CurrentMP,
		RespawnTime: row.RespawnTime,
		Status:      row.Status,
	}, nil
}

// raidPointsStoreAdapter adapts db.RaidRepository to raid.RaidPointsStore.
type raidPointsStoreAdapter struct {
	repo *db.RaidRepository
}

func (a *raidPointsStoreAdapter) AddRaidPoints(ctx context.Context, characterID, bossID, points int32) error {
	return a.repo.AddRaidPoints(ctx, characterID, bossID, points)
}

func (a *raidPointsStoreAdapter) GetTotalRaidPoints(ctx context.Context, characterID int32) (int32, error) {
	return a.repo.GetTotalRaidPoints(ctx, characterID)
}

func (a *raidPointsStoreAdapter) GetTopRaidPointPlayers(ctx context.Context, limit int) ([]raid.RaidPointsEntry, error) {
	rows, err := a.repo.GetTopRaidPointPlayers(ctx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]raid.RaidPointsEntry, len(rows))
	for i, r := range rows {
		result[i] = raid.RaidPointsEntry{
			CharacterID: r.CharacterID,
			Points:      r.Points,
		}
	}
	return result, nil
}

func (a *raidPointsStoreAdapter) ResetAllRaidPoints(ctx context.Context) (int64, error) {
	return a.repo.ResetAllRaidPoints(ctx)
}
