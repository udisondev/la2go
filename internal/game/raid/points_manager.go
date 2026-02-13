package raid

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// RaidPointsStore provides DB persistence for raid boss points.
type RaidPointsStore interface {
	AddRaidPoints(ctx context.Context, characterID, bossID, points int32) error
	GetTotalRaidPoints(ctx context.Context, characterID int32) (int32, error)
	GetTopRaidPointPlayers(ctx context.Context, limit int) ([]RaidPointsEntry, error)
	ResetAllRaidPoints(ctx context.Context) (int64, error)
}

// RaidPointsEntry represents aggregated raid points for ranking.
type RaidPointsEntry struct {
	CharacterID int32
	Points      int32
}

// PointsManager manages raid boss kill points per player.
//
// Points are awarded when a raid boss dies based on boss level:
//   - points = max(1, bossLevel / 2)
//   - Split between party members if in party
//
// Weekly reset: every Monday points are cleared and
// top-ranked players receive clan reputation.
//
// Phase 23: Raid Boss System.
// Java reference: RaidBossPointsManager.java.
type PointsManager struct {
	store RaidPointsStore

	mu     sync.RWMutex
	cache  map[int32]int32 // characterID → total points (in-memory cache)
	loaded bool
}

// NewPointsManager creates a new raid points manager.
func NewPointsManager(store RaidPointsStore) *PointsManager {
	return &PointsManager{
		store: store,
		cache: make(map[int32]int32, 1024),
	}
}

// AddPoints awards raid points to a character for killing a raid boss.
// points = max(1, bossLevel / 2).
func (m *PointsManager) AddPoints(ctx context.Context, characterID, bossID, bossLevel int32) error {
	points := max(1, bossLevel/2)

	if err := m.store.AddRaidPoints(ctx, characterID, bossID, points); err != nil {
		return fmt.Errorf("add raid points char %d boss %d: %w", characterID, bossID, err)
	}

	// Update cache
	m.mu.Lock()
	m.cache[characterID] += points
	m.mu.Unlock()

	slog.Info("raid points awarded",
		"character", characterID,
		"bossID", bossID,
		"bossLevel", bossLevel,
		"points", points)

	return nil
}

// GetPoints returns total raid points for a character.
// Uses cache if available, otherwise queries DB.
func (m *PointsManager) GetPoints(ctx context.Context, characterID int32) (int32, error) {
	m.mu.RLock()
	if pts, ok := m.cache[characterID]; ok {
		m.mu.RUnlock()
		return pts, nil
	}
	m.mu.RUnlock()

	total, err := m.store.GetTotalRaidPoints(ctx, characterID)
	if err != nil {
		return 0, fmt.Errorf("get raid points char %d: %w", characterID, err)
	}

	m.mu.Lock()
	m.cache[characterID] = total
	m.mu.Unlock()

	return total, nil
}

// GetRanking returns top N players by raid points.
func (m *PointsManager) GetRanking(ctx context.Context, limit int) ([]RaidPointsEntry, error) {
	return m.store.GetTopRaidPointPlayers(ctx, limit)
}

// WeeklyReset clears all raid points (called every Monday).
// Returns number of deleted rows.
func (m *PointsManager) WeeklyReset(ctx context.Context) (int64, error) {
	deleted, err := m.store.ResetAllRaidPoints(ctx)
	if err != nil {
		return 0, fmt.Errorf("weekly raid points reset: %w", err)
	}

	m.mu.Lock()
	m.cache = make(map[int32]int32, 1024)
	m.mu.Unlock()

	slog.Info("raid points weekly reset completed", "deleted", deleted)
	return deleted, nil
}

// CalculatePoints returns points for killing a boss of given level.
func CalculatePoints(bossLevel int32) int32 {
	return max(1, bossLevel/2)
}
