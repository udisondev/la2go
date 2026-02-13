package quest

import (
	"context"
	"sync"
	"time"
)

// TimerFunc is the callback for quest timers.
type TimerFunc func(timerName string, player PlayerRef, npcObjectID uint32)

// PlayerRef is a minimal player reference for timers.
// Avoids importing full model.Player to keep quest package decoupled.
type PlayerRef interface {
	ObjectID() uint32
	Name() string
}

// Timer represents a single quest timer that fires after a delay.
// Thread-safe: cancel can be called from any goroutine.
type Timer struct {
	name        string
	questName   string
	playerObjID uint32
	npcObjID    uint32
	cancel      context.CancelFunc
	done        chan struct{}
}

// Name returns the timer name.
func (t *Timer) Name() string { return t.name }

// Cancel stops the timer before it fires.
func (t *Timer) Cancel() {
	t.cancel()
	<-t.done // дождаться завершения горутины
}

// TimerManager manages active quest timers.
// Thread-safe for concurrent timer creation/cancellation.
type TimerManager struct {
	mu     sync.Mutex
	timers map[string]*Timer // key: "questName:timerName:playerObjectID"
}

// NewTimerManager creates a new timer manager.
func NewTimerManager() *TimerManager {
	return &TimerManager{
		timers: make(map[string]*Timer, 32),
	}
}

// timerKey generates a unique key for a timer.
func timerKey(questName, timerName string, playerObjID uint32) string {
	// Формат: questName:timerName:playerObjID
	return questName + ":" + timerName + ":" + intToString(int(playerObjID))
}

// StartTimer creates and starts a new timer.
// If a timer with the same key already exists, it is cancelled first.
func (tm *TimerManager) StartTimer(
	questName, timerName string,
	delay time.Duration,
	player PlayerRef,
	npcObjID uint32,
	callback TimerFunc,
) *Timer {
	key := timerKey(questName, timerName, player.ObjectID())

	tm.mu.Lock()
	// Отменяем существующий таймер с тем же ключом
	if old, ok := tm.timers[key]; ok {
		old.cancel()
		delete(tm.timers, key)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	t := &Timer{
		name:        timerName,
		questName:   questName,
		playerObjID: player.ObjectID(),
		npcObjID:    npcObjID,
		cancel:      cancel,
		done:        done,
	}

	tm.timers[key] = t
	tm.mu.Unlock()

	go func() {
		defer close(done)

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			callback(timerName, player, npcObjID)
		}

		// Удаляем таймер из менеджера после срабатывания
		tm.mu.Lock()
		if current, ok := tm.timers[key]; ok && current == t {
			delete(tm.timers, key)
		}
		tm.mu.Unlock()
	}()

	return t
}

// CancelTimer cancels a timer by key components.
// Returns true if timer was found and cancelled.
func (tm *TimerManager) CancelTimer(questName, timerName string, playerObjID uint32) bool {
	key := timerKey(questName, timerName, playerObjID)

	tm.mu.Lock()
	t, ok := tm.timers[key]
	if ok {
		delete(tm.timers, key)
	}
	tm.mu.Unlock()

	if ok {
		t.cancel()
		<-t.done
		return true
	}
	return false
}

// CancelAllForPlayer cancels all timers for a specific player.
func (tm *TimerManager) CancelAllForPlayer(playerObjID uint32) {
	suffix := ":" + intToString(int(playerObjID))

	tm.mu.Lock()
	var toCancel []*Timer
	for key, t := range tm.timers {
		if len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix {
			toCancel = append(toCancel, t)
			delete(tm.timers, key)
		}
	}
	tm.mu.Unlock()

	for _, t := range toCancel {
		t.cancel()
		<-t.done
	}
}

// CancelAllForQuest cancels all timers for a specific quest.
func (tm *TimerManager) CancelAllForQuest(questName string) {
	prefix := questName + ":"

	tm.mu.Lock()
	var toCancel []*Timer
	for key, t := range tm.timers {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			toCancel = append(toCancel, t)
			delete(tm.timers, key)
		}
	}
	tm.mu.Unlock()

	for _, t := range toCancel {
		t.cancel()
		<-t.done
	}
}

// ActiveCount returns the number of active timers.
func (tm *TimerManager) ActiveCount() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return len(tm.timers)
}

// Shutdown cancels all active timers.
func (tm *TimerManager) Shutdown() {
	tm.mu.Lock()
	all := make([]*Timer, 0, len(tm.timers))
	for _, t := range tm.timers {
		all = append(all, t)
	}
	tm.timers = make(map[string]*Timer)
	tm.mu.Unlock()

	for _, t := range all {
		t.cancel()
		<-t.done
	}
}
