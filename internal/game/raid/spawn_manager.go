package raid

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"
)

// RaidSpawnStore provides DB persistence for raid boss spawns.
type RaidSpawnStore interface {
	LoadAllRaidSpawns(ctx context.Context) ([]RaidSpawnRow, error)
	SaveRaidSpawn(ctx context.Context, row RaidSpawnRow) error
	DeleteRaidSpawn(ctx context.Context, bossID int32) error
}

// RaidSpawnRow mirrors db.RaidBossSpawnRow for decoupling.
type RaidSpawnRow struct {
	BossID      int32
	RespawnTime int64   // Unix seconds
	CurrentHP   float64
	CurrentMP   float64
}

// raidSpawnEntry tracks a single raid boss spawn state in memory.
type raidSpawnEntry struct {
	BossID      int32
	RespawnTime time.Time
	CurrentHP   float64
	CurrentMP   float64
	IsAlive     bool
}

// SpawnManager manages raid boss spawns with DB-backed respawn tracking.
//
// On startup:
//  1. Load raid boss spawn data from DB
//  2. Check if respawn time has passed — if yes, spawn immediately
//  3. If not, schedule respawn for remaining time
//
// On death:
//  1. Calculate respawn delay (12-24h by default)
//  2. Save respawn time to DB
//  3. Schedule respawn
//
// Phase 23: Raid Boss System.
// Java reference: RaidBossSpawnManager.java.
type SpawnManager struct {
	store RaidSpawnStore

	mu      sync.RWMutex
	entries map[int32]*raidSpawnEntry // bossID → entry

	// respawnFn is called when a raid boss should be spawned/respawned.
	// Injected by gameserver to avoid circular dependency.
	respawnFn func(bossID int32) error

	// configurable delays (seconds)
	respawnMinDelay int32
	respawnMaxDelay int32
}

// NewSpawnManager creates a new raid boss spawn manager.
func NewSpawnManager(store RaidSpawnStore, respawnFn func(bossID int32) error) *SpawnManager {
	return &SpawnManager{
		store:           store,
		entries:         make(map[int32]*raidSpawnEntry, 256),
		respawnFn:       respawnFn,
		respawnMinDelay: 43200, // 12h
		respawnMaxDelay: 86400, // 24h
	}
}

// SetRespawnDelays configures min/max respawn delays (in seconds).
func (m *SpawnManager) SetRespawnDelays(min, max int32) {
	m.respawnMinDelay = min
	m.respawnMaxDelay = max
}

// Init loads raid boss spawn data from DB and handles pending respawns.
// Bosses whose respawn time has passed are spawned immediately.
// Bosses still on cooldown are scheduled for later respawn.
func (m *SpawnManager) Init(ctx context.Context) error {
	rows, err := m.store.LoadAllRaidSpawns(ctx)
	if err != nil {
		return fmt.Errorf("load raid spawns: %w", err)
	}

	now := time.Now()
	spawnedCount := 0
	scheduledCount := 0

	for _, row := range rows {
		entry := &raidSpawnEntry{
			BossID:      row.BossID,
			RespawnTime: time.Unix(row.RespawnTime, 0),
			CurrentHP:   row.CurrentHP,
			CurrentMP:   row.CurrentMP,
		}

		m.mu.Lock()
		m.entries[row.BossID] = entry
		m.mu.Unlock()

		if row.RespawnTime <= 0 || now.After(entry.RespawnTime) {
			// Respawn time passed or never set — spawn now
			entry.IsAlive = true
			if m.respawnFn != nil {
				if err := m.respawnFn(row.BossID); err != nil {
					slog.Error("raid boss spawn on init",
						"bossID", row.BossID, "error", err)
					continue
				}
			}
			spawnedCount++
		} else {
			// Still on cooldown
			entry.IsAlive = false
			scheduledCount++
		}
	}

	slog.Info("raid boss spawn manager initialized",
		"loaded", len(rows),
		"spawned", spawnedCount,
		"scheduled", scheduledCount)

	return nil
}

// OnRaidBossDeath handles raid boss death: calculates respawn time, saves to DB.
func (m *SpawnManager) OnRaidBossDeath(ctx context.Context, bossID int32, hp, mp float64) error {
	delay := m.calculateRespawnDelay()
	respawnTime := time.Now().Add(time.Duration(delay) * time.Second)

	m.mu.Lock()
	entry, ok := m.entries[bossID]
	if !ok {
		entry = &raidSpawnEntry{BossID: bossID}
		m.entries[bossID] = entry
	}
	entry.RespawnTime = respawnTime
	entry.CurrentHP = hp
	entry.CurrentMP = mp
	entry.IsAlive = false
	m.mu.Unlock()

	// Persist to DB
	row := RaidSpawnRow{
		BossID:      bossID,
		RespawnTime: respawnTime.Unix(),
		CurrentHP:   hp,
		CurrentMP:   mp,
	}
	if err := m.store.SaveRaidSpawn(ctx, row); err != nil {
		return fmt.Errorf("save raid spawn on death boss %d: %w", bossID, err)
	}

	slog.Info("raid boss death recorded",
		"bossID", bossID,
		"respawnDelay", delay,
		"respawnTime", respawnTime.Format(time.RFC3339))

	return nil
}

// OnRaidBossSpawned marks boss as alive after successful spawn.
func (m *SpawnManager) OnRaidBossSpawned(bossID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[bossID]
	if !ok {
		entry = &raidSpawnEntry{BossID: bossID}
		m.entries[bossID] = entry
	}
	entry.IsAlive = true
}

// CheckPendingRespawns checks all entries for bosses ready to respawn.
// Returns list of bossIDs that should be spawned now.
// Called periodically by the respawn tick loop.
func (m *SpawnManager) CheckPendingRespawns() []int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var ready []int32

	for bossID, entry := range m.entries {
		if !entry.IsAlive && !entry.RespawnTime.IsZero() && now.After(entry.RespawnTime) {
			ready = append(ready, bossID)
		}
	}

	return ready
}

// RunRespawnLoop periodically checks for pending respawns.
// Blocks until context is canceled.
func (m *SpawnManager) RunRespawnLoop(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	slog.Info("raid boss respawn loop started", "interval", "30s")

	for {
		select {
		case <-ctx.Done():
			slog.Info("raid boss respawn loop stopping")
			return ctx.Err()
		case <-ticker.C:
			ready := m.CheckPendingRespawns()
			for _, bossID := range ready {
				if m.respawnFn != nil {
					if err := m.respawnFn(bossID); err != nil {
						slog.Error("raid boss respawn",
							"bossID", bossID, "error", err)
						continue
					}
				}
				m.OnRaidBossSpawned(bossID)
				slog.Info("raid boss respawned", "bossID", bossID)
			}
		}
	}
}

// IsAlive returns whether raid boss is currently alive.
func (m *SpawnManager) IsAlive(bossID int32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry, ok := m.entries[bossID]; ok {
		return entry.IsAlive
	}
	return false
}

// GetRespawnTime returns scheduled respawn time for a dead raid boss.
// Returns zero time if boss is alive or not tracked.
func (m *SpawnManager) GetRespawnTime(bossID int32) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry, ok := m.entries[bossID]; ok && !entry.IsAlive {
		return entry.RespawnTime
	}
	return time.Time{}
}

// EntryCount returns number of tracked raid bosses.
func (m *SpawnManager) EntryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// calculateRespawnDelay returns respawn delay in seconds [min, max].
func (m *SpawnManager) calculateRespawnDelay() int32 {
	min := m.respawnMinDelay
	max := m.respawnMaxDelay
	if max <= min {
		return min
	}
	return min + rand.Int32N(max-min+1)
}
