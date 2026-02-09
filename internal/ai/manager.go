package ai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// TickManager manages AI ticks for all registered NPCs
type TickManager struct {
	controllers     sync.Map // map[uint32]Controller — objectID → controller
	ticker          *time.Ticker
	stopCh          chan struct{}
	controllerCount atomic.Int32 // cached count of controllers (O(1) access)
}

// NewTickManager creates new AI tick manager
func NewTickManager() *TickManager {
	return &TickManager{
		stopCh: make(chan struct{}),
	}
}

// Register registers AI controller for NPC
func (m *TickManager) Register(objectID uint32, controller Controller) {
	m.controllers.Store(objectID, controller)
	m.controllerCount.Add(1) // Update cached count
	controller.Start()

	slog.Debug("AI controller registered",
		"objectID", objectID,
		"intention", controller.CurrentIntention())
}

// Unregister unregisters AI controller
func (m *TickManager) Unregister(objectID uint32) {
	value, ok := m.controllers.LoadAndDelete(objectID)
	if !ok {
		return
	}

	m.controllerCount.Add(-1) // Update cached count

	controller := value.(Controller)
	controller.Stop()

	slog.Debug("AI controller unregistered", "objectID", objectID)
}

// Start starts AI tick loop (blocks until context is canceled)
func (m *TickManager) Start(ctx context.Context) error {
	m.ticker = time.NewTicker(1 * time.Second)
	defer m.ticker.Stop()

	slog.Info("AI tick manager started", "interval", "1s")

	for {
		select {
		case <-ctx.Done():
			slog.Info("AI tick manager stopping")
			return ctx.Err()

		case <-m.stopCh:
			slog.Info("AI tick manager stopped")
			return nil

		case <-m.ticker.C:
			m.tickAll()
		}
	}
}

// Stop stops AI tick loop
func (m *TickManager) Stop() {
	close(m.stopCh)
}

// tickAll ticks all registered controllers
func (m *TickManager) tickAll() {
	count := 0

	m.controllers.Range(func(key, value any) bool {
		controller := value.(Controller)
		controller.Tick()
		count++
		return true
	})

	if count > 0 && IsDebugEnabled() {
		slog.Debug("AI tick completed", "controllers", count)
	}
}

// Count returns number of registered controllers (O(1) cached count)
// IMPORTANT: Count is cached atomically and updated when controllers are registered/unregistered.
// This is a performance optimization to avoid O(N) Range() on sync.Map.
func (m *TickManager) Count() int {
	return int(m.controllerCount.Load())
}

// GetController returns controller for NPC
func (m *TickManager) GetController(objectID uint32) (Controller, error) {
	value, ok := m.controllers.Load(objectID)
	if !ok {
		return nil, fmt.Errorf("controller not found for objectID %d", objectID)
	}
	return value.(Controller), nil
}
