package raid

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// GrandBossStore provides DB persistence for grand boss data.
type GrandBossStore interface {
	LoadAllGrandBosses(ctx context.Context) ([]GrandBossDataRow, error)
	SaveGrandBoss(ctx context.Context, row GrandBossDataRow) error
	GetGrandBoss(ctx context.Context, bossID int32) (*GrandBossDataRow, error)
}

// GrandBossDataRow mirrors db.GrandBossRow for decoupling.
type GrandBossDataRow struct {
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

// grandBossEntry tracks in-memory state for a single grand boss.
type grandBossEntry struct {
	BossID      int32
	Status      model.GrandBossStatus
	RespawnTime time.Time
	Boss        *model.GrandBoss // nil when dead/waiting
	LocX, LocY, LocZ int32
	Heading     int32
	CurrentHP   float64
	CurrentMP   float64
}

// GrandBossManager manages grand boss states, spawning, and DB persistence.
//
// Grand boss states:
//   - ALIVE (0): boss is spawned in the world
//   - DEAD (1): boss killed, waiting for respawn
//   - FIGHTING (2): boss engaged in combat (optional tracking)
//   - WAITING (3): waiting state (e.g., Antharas pre-fight wait)
//
// Periodic DB save every 5 minutes for status persistence.
//
// Phase 23: Raid Boss System.
// Java reference: GrandBossManager.java.
type GrandBossManager struct {
	store GrandBossStore

	mu      sync.RWMutex
	entries map[int32]*grandBossEntry // bossID → entry

	// spawnFn spawns a grand boss in the world. Injected by gameserver.
	spawnFn func(bossID int32) (*model.GrandBoss, error)
}

// NewGrandBossManager creates a new grand boss manager.
func NewGrandBossManager(store GrandBossStore, spawnFn func(int32) (*model.GrandBoss, error)) *GrandBossManager {
	return &GrandBossManager{
		store:   store,
		entries: make(map[int32]*grandBossEntry, 16),
		spawnFn: spawnFn,
	}
}

// Init loads grand boss data from DB and spawns bosses that are alive.
func (m *GrandBossManager) Init(ctx context.Context) error {
	rows, err := m.store.LoadAllGrandBosses(ctx)
	if err != nil {
		return fmt.Errorf("load grand bosses: %w", err)
	}

	now := time.Now()
	spawnedCount := 0

	for _, row := range rows {
		entry := &grandBossEntry{
			BossID:      row.BossID,
			Status:      model.GrandBossStatus(row.Status),
			RespawnTime: time.Unix(row.RespawnTime, 0),
			LocX:        row.LocX,
			LocY:        row.LocY,
			LocZ:        row.LocZ,
			Heading:     row.Heading,
			CurrentHP:   row.CurrentHP,
			CurrentMP:   row.CurrentMP,
		}

		m.mu.Lock()
		m.entries[row.BossID] = entry
		m.mu.Unlock()

		switch model.GrandBossStatus(row.Status) {
		case model.GrandBossAlive:
			// Spawn boss
			if m.spawnFn != nil {
				boss, spawnErr := m.spawnFn(row.BossID)
				if spawnErr != nil {
					slog.Error("grand boss spawn on init",
						"bossID", row.BossID, "error", spawnErr)
					continue
				}
				entry.Boss = boss
				spawnedCount++
			}

		case model.GrandBossDead:
			// Check if respawn time has passed
			if row.RespawnTime > 0 && now.After(entry.RespawnTime) {
				entry.Status = model.GrandBossAlive
				if m.spawnFn != nil {
					boss, spawnErr := m.spawnFn(row.BossID)
					if spawnErr != nil {
						slog.Error("grand boss respawn on init",
							"bossID", row.BossID, "error", spawnErr)
						continue
					}
					entry.Boss = boss
					spawnedCount++
				}
			}
			// Else: still dead, wait for respawn loop
		}
	}

	slog.Info("grand boss manager initialized",
		"loaded", len(rows),
		"spawned", spawnedCount)

	return nil
}

// GetStatus returns the current status of a grand boss.
func (m *GrandBossManager) GetStatus(bossID int32) model.GrandBossStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry, ok := m.entries[bossID]; ok {
		return entry.Status
	}
	return model.GrandBossAlive // default
}

// SetStatus updates grand boss status in memory and DB.
func (m *GrandBossManager) SetStatus(ctx context.Context, bossID int32, status model.GrandBossStatus) error {
	m.mu.Lock()
	entry, ok := m.entries[bossID]
	if !ok {
		entry = &grandBossEntry{BossID: bossID}
		m.entries[bossID] = entry
	}
	entry.Status = status
	m.mu.Unlock()

	return m.saveEntry(ctx, entry)
}

// OnGrandBossDeath handles grand boss death: sets status to DEAD, calculates respawn.
func (m *GrandBossManager) OnGrandBossDeath(ctx context.Context, bossID int32, respawnDelay int64) error {
	respawnTime := time.Now().Add(time.Duration(respawnDelay) * time.Second)

	m.mu.Lock()
	entry, ok := m.entries[bossID]
	if !ok {
		entry = &grandBossEntry{BossID: bossID}
		m.entries[bossID] = entry
	}
	entry.Status = model.GrandBossDead
	entry.RespawnTime = respawnTime
	entry.Boss = nil
	entry.CurrentHP = 0
	entry.CurrentMP = 0
	m.mu.Unlock()

	if err := m.saveEntry(ctx, entry); err != nil {
		return fmt.Errorf("save grand boss death %d: %w", bossID, err)
	}

	slog.Info("grand boss death recorded",
		"bossID", bossID,
		"respawnDelay", respawnDelay,
		"respawnTime", respawnTime.Format(time.RFC3339))

	return nil
}

// OnGrandBossSpawned registers a live grand boss.
func (m *GrandBossManager) OnGrandBossSpawned(bossID int32, boss *model.GrandBoss) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[bossID]
	if !ok {
		entry = &grandBossEntry{BossID: bossID}
		m.entries[bossID] = entry
	}
	entry.Status = model.GrandBossAlive
	entry.Boss = boss
}

// GetBoss returns the live GrandBoss instance (nil if dead).
func (m *GrandBossManager) GetBoss(bossID int32) *model.GrandBoss {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry, ok := m.entries[bossID]; ok {
		return entry.Boss
	}
	return nil
}

// CheckPendingRespawns returns grand boss IDs ready to respawn.
func (m *GrandBossManager) CheckPendingRespawns() []int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var ready []int32

	for bossID, entry := range m.entries {
		if entry.Status == model.GrandBossDead &&
			!entry.RespawnTime.IsZero() &&
			now.After(entry.RespawnTime) {
			ready = append(ready, bossID)
		}
	}

	return ready
}

// RunSaveLoop periodically saves grand boss states to DB.
// Blocks until context is canceled.
func (m *GrandBossManager) RunSaveLoop(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	slog.Info("grand boss save loop started", "interval", "5m")

	for {
		select {
		case <-ctx.Done():
			// Final save before exit
			m.saveAll(ctx)
			slog.Info("grand boss save loop stopping")
			return ctx.Err()
		case <-ticker.C:
			m.saveAll(ctx)
		}
	}
}

// RunRespawnLoop periodically checks for pending grand boss respawns.
// Blocks until context is canceled.
func (m *GrandBossManager) RunRespawnLoop(ctx context.Context) error {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	slog.Info("grand boss respawn loop started", "interval", "60s")

	for {
		select {
		case <-ctx.Done():
			slog.Info("grand boss respawn loop stopping")
			return ctx.Err()
		case <-ticker.C:
			ready := m.CheckPendingRespawns()
			for _, bossID := range ready {
				if m.spawnFn != nil {
					boss, err := m.spawnFn(bossID)
					if err != nil {
						slog.Error("grand boss respawn",
							"bossID", bossID, "error", err)
						continue
					}
					m.OnGrandBossSpawned(bossID, boss)
					slog.Info("grand boss respawned", "bossID", bossID)
				}
			}
		}
	}
}

// EntryCount returns number of tracked grand bosses.
func (m *GrandBossManager) EntryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// saveEntry persists a single grand boss entry to DB.
func (m *GrandBossManager) saveEntry(ctx context.Context, entry *grandBossEntry) error {
	m.mu.RLock()
	row := GrandBossDataRow{
		BossID:      entry.BossID,
		LocX:        entry.LocX,
		LocY:        entry.LocY,
		LocZ:        entry.LocZ,
		Heading:     entry.Heading,
		CurrentHP:   entry.CurrentHP,
		CurrentMP:   entry.CurrentMP,
		RespawnTime: entry.RespawnTime.Unix(),
		Status:      int16(entry.Status),
	}

	// Update location from live boss if present
	if entry.Boss != nil {
		loc := entry.Boss.Location()
		row.LocX = loc.X
		row.LocY = loc.Y
		row.LocZ = loc.Z
		row.Heading = int32(loc.Heading)
		row.CurrentHP = float64(entry.Boss.CurrentHP())
		row.CurrentMP = float64(entry.Boss.CurrentMP())
	}
	m.mu.RUnlock()

	return m.store.SaveGrandBoss(ctx, row)
}

// saveAll persists all entries to DB.
func (m *GrandBossManager) saveAll(ctx context.Context) {
	m.mu.RLock()
	entries := make([]*grandBossEntry, 0, len(m.entries))
	for _, e := range m.entries {
		entries = append(entries, e)
	}
	m.mu.RUnlock()

	saved := 0
	for _, entry := range entries {
		if err := m.saveEntry(ctx, entry); err != nil {
			slog.Error("save grand boss state", "bossID", entry.BossID, "error", err)
			continue
		}
		saved++
	}

	if saved > 0 {
		slog.Debug("grand boss states saved", "count", saved)
	}
}
