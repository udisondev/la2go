package duel

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// Manager manages all active duels.
// Thread-safe for concurrent access.
type Manager struct {
	mu     sync.RWMutex
	duels  map[int32]*Duel       // duelID → Duel
	byPlayer map[uint32]int32    // objectID → duelID (quick lookup)
	nextID atomic.Int32
}

// NewManager creates a new duel manager.
func NewManager() *Manager {
	return &Manager{
		duels:    make(map[int32]*Duel, 16),
		byPlayer: make(map[uint32]int32, 32),
	}
}

// CanDuel checks if a player can participate in a duel.
// Returns empty string if OK, or reason string if not.
func CanDuel(p *model.Player) string {
	if p.IsDead() {
		return "hp_mp_below_50"
	}
	if p.CurrentHP() < p.MaxHP()/2 || p.CurrentMP() < p.MaxMP()/2 {
		return "hp_mp_below_50"
	}
	return ""
}

// CreateDuel creates a new duel.
// Does NOT start the countdown — caller must call StartDuel.
func (m *Manager) CreateDuel(playerA, playerB *model.Player, partyDuel bool) (*Duel, error) {
	// Проверяем что игроки не в дуэли
	m.mu.RLock()
	if _, ok := m.byPlayer[playerA.ObjectID()]; ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("player %s already in duel", playerA.Name())
	}
	if _, ok := m.byPlayer[playerB.ObjectID()]; ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("player %s already in duel", playerB.Name())
	}
	m.mu.RUnlock()

	id := m.nextID.Add(1)
	d := NewDuel(id, playerA, playerB, partyDuel)

	m.mu.Lock()
	m.duels[id] = d
	m.byPlayer[playerA.ObjectID()] = id
	m.byPlayer[playerB.ObjectID()] = id

	// Регистрируем всех участников party duel
	if partyDuel {
		if pa := playerA.GetParty(); pa != nil {
			for _, mem := range pa.Members() {
				m.byPlayer[mem.ObjectID()] = id
			}
		}
		if pb := playerB.GetParty(); pb != nil {
			for _, mem := range pb.Members() {
				m.byPlayer[mem.ObjectID()] = id
			}
		}
	}
	m.mu.Unlock()

	slog.Debug("duel created",
		"duelID", id,
		"playerA", playerA.Name(),
		"playerB", playerB.Name(),
		"party", partyDuel)

	return d, nil
}

// StartDuel begins the countdown and starts the duel lifecycle goroutine.
// onCountdown is called for each countdown tick (5,4,3,2,1).
// onStart is called when countdown reaches 0 (duel begins).
// onEnd is called when duel finishes with the result.
func (m *Manager) StartDuel(d *Duel, onCountdown func(d *Duel, count int32), onStart func(d *Duel), onEnd func(d *Duel, result Result)) {
	// Горутина завершается когда дуэль заканчивается.
	go m.runDuelLifecycle(d, onCountdown, onStart, onEnd)
}

// GetDuel returns a duel by ID.
func (m *Manager) GetDuel(duelID int32) *Duel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.duels[duelID]
}

// GetDuelByPlayer returns the active duel for a player.
func (m *Manager) GetDuelByPlayer(objectID uint32) *Duel {
	m.mu.RLock()
	duelID, ok := m.byPlayer[objectID]
	if !ok {
		m.mu.RUnlock()
		return nil
	}
	d := m.duels[duelID]
	m.mu.RUnlock()
	return d
}

// IsInDuel returns true if the player is in an active duel.
func (m *Manager) IsInDuel(objectID uint32) bool {
	m.mu.RLock()
	_, ok := m.byPlayer[objectID]
	m.mu.RUnlock()
	return ok
}

// OnPlayerDefeat handles a player being defeated (HP→1 in duel).
func (m *Manager) OnPlayerDefeat(objectID uint32) {
	d := m.GetDuelByPlayer(objectID)
	if d == nil || d.IsFinished() {
		return
	}
	d.OnPlayerDefeat(objectID)
}

// OnSurrender handles a player's surrender request.
func (m *Manager) OnSurrender(objectID uint32) {
	d := m.GetDuelByPlayer(objectID)
	if d == nil || d.IsFinished() {
		return
	}
	d.Surrender(objectID)
}

// RemoveDuel removes a duel from the manager.
func (m *Manager) RemoveDuel(duelID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.duels[duelID]
	if !ok {
		return
	}

	// Убираем всех участников
	for objID, did := range m.byPlayer {
		if did == duelID {
			delete(m.byPlayer, objID)
		}
	}
	delete(m.duels, duelID)

	slog.Debug("duel removed",
		"duelID", duelID,
		"playerA", d.playerA.Name(),
		"playerB", d.playerB.Name())
}

// DuelCount returns the number of active duels.
func (m *Manager) DuelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.duels)
}

// runDuelLifecycle runs the full duel lifecycle in a goroutine.
// Goroutine завершается когда дуэль заканчивается или cancelCh закрыт.
func (m *Manager) runDuelLifecycle(d *Duel, onCountdown func(*Duel, int32), onStart func(*Duel), onEnd func(*Duel, Result)) {
	defer m.RemoveDuel(d.id)

	// Phase 1: Countdown (5 → 0)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Сохраняем состояния при countdown=4
	saved := false

	for d.countdown.Load() > 0 {
		select {
		case <-d.cancelCh:
			return
		case <-ticker.C:
			count := d.countdown.Add(-1)

			if count == 4 && !saved {
				d.SaveAllConditions()
				saved = true
			}

			if count > 0 && onCountdown != nil {
				onCountdown(d, count)
			}

			if count == 0 {
				break
			}
		}
	}

	// Phase 2: Start fighting
	d.InitParticipants()

	if onStart != nil {
		onStart(d)
	}

	// Phase 3: Check conditions every second
	for {
		select {
		case <-d.cancelCh:
			return
		case <-ticker.C:
			result := d.CheckEndCondition()
			if result != ResultContinue {
				d.Finish()
				if onEnd != nil {
					onEnd(d, result)
				}
				return
			}
		}
	}
}
