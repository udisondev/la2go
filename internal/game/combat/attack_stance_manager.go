package combat

import (
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// CombatTime is the duration of combat stance (15 seconds).
// After this time, player exits combat and auto-attack stops.
//
// Phase 5.3: Basic Combat System.
// Java reference: AttackStanceTaskManager.COMBAT_TIME = 15000 ms.
const CombatTime = 15 * time.Second

// AttackStanceManager tracks combat state for all players.
// Player enters combat when attacking, exits after 15 seconds of inactivity.
//
// Thread-safety: sync.Map for concurrent access (multiple players attacking simultaneously).
//
// Phase 5.3: Basic Combat System.
// Java reference: AttackStanceTaskManager.java (ConcurrentHashMap + scheduled task).
type AttackStanceManager struct {
	stances       sync.Map      // key: *model.Player, value: time.Time (last attack timestamp)
	stopCh        chan struct{} // Close this channel to stop cleanup goroutine
	wg            sync.WaitGroup
	broadcastFunc func(source *model.Player, data []byte, size int)
}

// NewAttackStanceManager creates new AttackStanceManager instance.
// Must call Start() to begin cleanup goroutine.
//
// Phase 5.3: Basic Combat System.
func NewAttackStanceManager(broadcastFunc func(*model.Player, []byte, int)) *AttackStanceManager {
	return &AttackStanceManager{
		stopCh:        make(chan struct{}),
		broadcastFunc: broadcastFunc,
	}
}

// Start launches cleanup goroutine (ticker 1 second).
// Goroutine runs until Stop() is called.
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) Start() {
	m.wg.Add(1)
	go m.run()
}

// Stop terminates cleanup goroutine.
// Blocks until goroutine exits (safe shutdown).
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// AddAttackStance adds player to combat state.
// Updates timestamp to current time (extends combat duration).
//
// Called when player performs physical attack.
//
// Thread-safe: sync.Map.Store is concurrent-safe.
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) AddAttackStance(player *model.Player) {
	player.MarkAttackStance()
	m.stances.Store(player, time.Now())
}

// RemoveAttackStance removes player from combat state.
// Player exits combat immediately (no AutoAttackStop broadcast).
//
// Thread-safe: sync.Map.Delete is concurrent-safe.
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) RemoveAttackStance(player *model.Player) {
	m.stances.Delete(player)
}

// HasAttackStance returns true if player is in combat state.
//
// Thread-safe: sync.Map.Load is concurrent-safe.
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) HasAttackStance(player *model.Player) bool {
	_, exists := m.stances.Load(player)
	return exists
}

// run is cleanup goroutine (ticker 1 second).
// Removes expired entries (> 15 seconds since last attack).
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) run() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired combat entries (> 15 seconds).
// Sends AutoAttackStop packet to all visible players.
//
// Phase 5.3: Basic Combat System.
func (m *AttackStanceManager) cleanup() {
	now := time.Now()

	m.stances.Range(func(key, value any) bool {
		player := key.(*model.Player)
		timestamp := value.(time.Time)

		// Check if combat expired
		if now.Sub(timestamp) > CombatTime {
			// Remove from combat
			m.stances.Delete(key)

			// Send AutoAttackStop packet
			autoAttackStop := serverpackets.NewAutoAttackStop(player.ObjectID())
			data, err := autoAttackStop.Write()
			if err != nil {
				slog.Error("failed to write AutoAttackStop packet",
					"character", player.Name(),
					"error", err)
				return true
			}

			// Broadcast to visible players
			if m.broadcastFunc != nil {
				m.broadcastFunc(player, data, len(data))
			}

			slog.Debug("combat stance expired",
				"character", player.Name(),
				"duration", now.Sub(timestamp))

			// TODO Phase 5.4: Stop AI auto-attack
			// player.GetAI().SetAutoAttacking(false)
		}

		return true // Continue iteration
	})
}

// AttackStanceMgr — global AttackStanceManager instance.
// Initialized by cmd/gameserver/main.go.
// NOT safe for concurrent test assignment — tests that set this must NOT use t.Parallel().
//
// Phase 5.3: Basic Combat System.
var AttackStanceMgr *AttackStanceManager
