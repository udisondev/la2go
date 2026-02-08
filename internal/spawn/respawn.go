package spawn

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// RespawnTask represents a scheduled respawn task
type RespawnTask struct {
	Spawn       *model.Spawn
	RespawnTime time.Time
}

// RespawnTaskManager manages scheduled respawns
type RespawnTaskManager struct {
	spawnManager *Manager
	ticker       *time.Ticker
	stopCh       chan struct{}

	mu    sync.RWMutex
	tasks map[int64]*RespawnTask // spawnID → task
}

// NewRespawnTaskManager creates new respawn task manager
func NewRespawnTaskManager(spawnManager *Manager) *RespawnTaskManager {
	return &RespawnTaskManager{
		spawnManager: spawnManager,
		stopCh:       make(chan struct{}),
		tasks:        make(map[int64]*RespawnTask),
	}
}

// Start starts respawn task manager (blocks until context is canceled)
func (m *RespawnTaskManager) Start(ctx context.Context) error {
	m.ticker = time.NewTicker(1 * time.Second)
	defer m.ticker.Stop()

	slog.Info("respawn task manager started", "interval", "1s")

	for {
		select {
		case <-ctx.Done():
			slog.Info("respawn task manager stopping")
			return ctx.Err()

		case <-m.stopCh:
			slog.Info("respawn task manager stopped")
			return nil

		case now := <-m.ticker.C:
			m.processTasks(ctx, now)
		}
	}
}

// Stop stops respawn task manager
func (m *RespawnTaskManager) Stop() {
	close(m.stopCh)
}

// ScheduleRespawn schedules respawn after delay (in seconds)
func (m *RespawnTaskManager) ScheduleRespawn(spawn *model.Spawn, delaySeconds int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	respawnTime := time.Now().Add(time.Duration(delaySeconds) * time.Second)

	task := &RespawnTask{
		Spawn:       spawn,
		RespawnTime: respawnTime,
	}

	m.tasks[spawn.SpawnID()] = task

	slog.Debug("respawn scheduled",
		"spawnID", spawn.SpawnID(),
		"templateID", spawn.TemplateID(),
		"delaySeconds", delaySeconds,
		"respawnTime", respawnTime.Format(time.RFC3339))
}

// CancelRespawn cancels scheduled respawn
func (m *RespawnTaskManager) CancelRespawn(spawnID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tasks, spawnID)

	slog.Debug("respawn cancelled", "spawnID", spawnID)
}

// processTasks processes respawn tasks that are due
func (m *RespawnTaskManager) processTasks(ctx context.Context, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find tasks that are due
	dueTasks := make([]*RespawnTask, 0)

	for spawnID, task := range m.tasks {
		if now.After(task.RespawnTime) || now.Equal(task.RespawnTime) {
			dueTasks = append(dueTasks, task)
			delete(m.tasks, spawnID)
		}
	}

	// Process due tasks (outside lock to avoid blocking)
	if len(dueTasks) > 0 {
		go m.executeTasks(ctx, dueTasks)
	}
}

// executeTasks executes respawn tasks
func (m *RespawnTaskManager) executeTasks(ctx context.Context, tasks []*RespawnTask) {
	for _, task := range tasks {
		spawn := task.Spawn

		// Check if spawn is still full
		if spawn.CurrentCount() >= spawn.MaximumCount() {
			slog.Debug("respawn skipped (spawn full)",
				"spawnID", spawn.SpawnID(),
				"currentCount", spawn.CurrentCount(),
				"maximumCount", spawn.MaximumCount())
			continue
		}

		// Respawn NPC
		npc, err := m.spawnManager.ScheduleRespawn(ctx, spawn)
		if err != nil {
			slog.Error("respawn failed",
				"spawnID", spawn.SpawnID(),
				"templateID", spawn.TemplateID(),
				"error", err)
			continue
		}

		slog.Info("NPC respawned",
			"objectID", npc.ObjectID(),
			"name", npc.Name(),
			"spawnID", spawn.SpawnID())
	}
}

// TaskCount returns number of scheduled respawn tasks
func (m *RespawnTaskManager) TaskCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks)
}

// GetTask returns respawn task for spawn (for testing)
func (m *RespawnTaskManager) GetTask(spawnID int64) (*RespawnTask, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[spawnID]
	return task, ok
}
