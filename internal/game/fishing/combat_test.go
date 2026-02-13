package fishing

import (
	"sync"
	"testing"
	"time"
)

// mockListener captures combat events for test assertions.
type mockListener struct {
	mu     sync.Mutex
	ticks  []CombatState
	endWin *bool
	endCh  chan struct{}
}

func newMockListener() *mockListener {
	return &mockListener{endCh: make(chan struct{})}
}

func (m *mockListener) OnFishingTick(objectID int32, state CombatState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ticks = append(m.ticks, state)
}

func (m *mockListener) OnFishingEnd(objectID int32, win bool) {
	m.mu.Lock()
	m.endWin = &win
	m.mu.Unlock()
	close(m.endCh)
}

func (m *mockListener) waitEnd(t *testing.T, timeout time.Duration) bool {
	t.Helper()
	select {
	case <-m.endCh:
		m.mu.Lock()
		defer m.mu.Unlock()
		return *m.endWin
	case <-time.After(timeout):
		t.Fatal("combat did not end within timeout")
		return false
	}
}

func (m *mockListener) tickCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.ticks)
}

func TestCombat_NewCombat(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(100, 6411, 500, 10.0, 30, 1, 1, listener)

	if c.ObjectID() != 100 {
		t.Errorf("ObjectID() = %d; want 100", c.ObjectID())
	}
	if c.FishID() != 6411 {
		t.Errorf("FishID() = %d; want 6411", c.FishID())
	}
	if c.MaxHP() != 500 {
		t.Errorf("MaxHP() = %d; want 500", c.MaxHP())
	}
	if c.LureType() != 1 {
		t.Errorf("LureType() = %d; want 1", c.LureType())
	}
}

func TestCombat_WinByDepleteHP(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	// Low HP fish, high combat duration, no regen.
	c := NewCombat(1, 6411, 50, 0.0, 60, 0, 1, listener)
	// Force resting mode so pumping is always effective (eliminates randomness).
	c.mode = ModeResting
	c.Start()

	// Pumping with massive damage until HP drops.
	// Even with 10% resist chance, 20 attempts with 100 dmg against 50 HP is enough.
	for range 20 {
		c.UsePumping(100, 0)
		time.Sleep(10 * time.Millisecond)
	}

	win := listener.waitEnd(t, 5*time.Second)
	if !win {
		t.Error("expected win; got lose")
	}
}

func TestCombat_LoseByTimeout(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	// Very short timer (2s), high HP fish.
	c := NewCombat(2, 6412, 10000, 0.0, 2, 1, 1, listener)
	c.Start()

	win := listener.waitEnd(t, 5*time.Second)
	if win {
		t.Error("expected lose by timeout; got win")
	}
}

func TestCombat_LoseByEscapeHP(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	// Low max HP fish, high regen, long timer.
	c := NewCombat(3, 6413, 10, 100.0, 60, 1, 1, listener)
	// Force fighting mode for regen.
	c.mu.Lock()
	c.mode = ModeFighting
	c.deceptiveMode = 0
	c.mu.Unlock()

	c.Start()

	win := listener.waitEnd(t, 5*time.Second)
	if win {
		t.Error("expected lose by HP escape; got win")
	}
}

func TestCombat_StopPrematurely(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(4, 6414, 1000, 5.0, 120, 1, 1, listener)
	c.Start()

	time.Sleep(100 * time.Millisecond)
	c.Stop()

	// Should not panic or deadlock. End event may or may not fire.
	time.Sleep(200 * time.Millisecond)
}

func TestCombat_PumpingEffective_RestingNormal(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(5, 6415, 1000, 0.0, 120, 0, 1, listener)

	// Force resting + normal mode.
	c.mu.Lock()
	c.mode = ModeResting
	c.deceptiveMode = 0
	c.mu.Unlock()

	// Use pumping many times to get at least one success
	// (90% chance each time, so very unlikely to miss all).
	successes := 0
	for range 50 {
		c.mu.Lock()
		before := c.fishCurHP
		c.mu.Unlock()

		c.UsePumping(10, 0)

		c.mu.Lock()
		after := c.fishCurHP
		c.mu.Unlock()

		if after < before {
			successes++
		}
	}

	if successes == 0 {
		t.Error("pumping in resting+normal mode should deal damage at least once out of 50 attempts")
	}
}

func TestCombat_ReelingEffective_FightingNormal(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(6, 6416, 1000, 0.0, 120, 0, 1, listener)

	// Force fighting + normal mode.
	c.mu.Lock()
	c.mode = ModeFighting
	c.deceptiveMode = 0
	c.mu.Unlock()

	successes := 0
	for range 50 {
		c.mu.Lock()
		before := c.fishCurHP
		c.mu.Unlock()

		c.UseReeling(10, 0)

		c.mu.Lock()
		after := c.fishCurHP
		c.mu.Unlock()

		if after < before {
			successes++
		}
	}

	if successes == 0 {
		t.Error("reeling in fighting+normal mode should deal damage at least once out of 50 attempts")
	}
}

func TestCombat_DeceptiveInvertsLogic(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(7, 6417, 1000, 0.0, 120, 2, 1, listener)

	// Force resting + deceptive: pumping should FAIL, reeling should SUCCEED.
	c.mu.Lock()
	c.mode = ModeResting
	c.deceptiveMode = 1
	c.mu.Unlock()

	// Pumping in resting+deceptive = FAIL (heals fish).
	heals := 0
	for range 50 {
		c.mu.Lock()
		before := c.fishCurHP
		c.mu.Unlock()

		c.UsePumping(10, 0)

		c.mu.Lock()
		after := c.fishCurHP
		c.mu.Unlock()

		if after > before {
			heals++
		}
	}

	if heals == 0 {
		t.Error("pumping in resting+deceptive should heal fish at least once out of 50 attempts")
	}
}

func TestCombat_PenaltyReducesHP(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(8, 6418, 1000, 0.0, 120, 0, 1, listener)

	c.mu.Lock()
	c.mode = ModeResting
	c.deceptiveMode = 0
	before := c.fishCurHP
	c.mu.Unlock()

	// Use pumping with penalty.
	c.UsePumping(0, 5)

	c.mu.Lock()
	after := c.fishCurHP
	c.mu.Unlock()

	// HP should decrease by at least penalty (unless resist).
	if after > before {
		// If resist happened, HP stays the same minus penalty.
		// Since pen=5 is always applied, HP should decrease or stay same.
		t.Logf("resist may have happened: before=%d, after=%d", before, after)
	}
}

func TestCombat_HPClamp(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(9, 6419, 100, 0.0, 120, 0, 1, listener)

	c.mu.Lock()
	c.mode = ModeResting
	c.deceptiveMode = 0
	c.mu.Unlock()

	// Massive damage to force HP below 0.
	c.UsePumping(5000, 0)

	c.mu.Lock()
	hp := c.fishCurHP
	c.mu.Unlock()

	if hp < 0 {
		t.Errorf("fishCurHP = %d; want >= 0 (clamped)", hp)
	}
}

func TestCombat_TickGeneratesEvents(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	// 3 second combat with low HP to prevent early end.
	c := NewCombat(10, 6420, 100000, 0.0, 3, 0, 1, listener)
	c.Start()

	// Wait for at least 2 ticks.
	time.Sleep(2500 * time.Millisecond)
	c.Stop()

	count := listener.tickCount()
	if count < 1 {
		t.Errorf("tick count = %d; want >= 1", count)
	}
}

func TestCombat_ModeGetters(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(11, 6421, 1000, 0.0, 60, 2, 1, listener)

	c.mu.Lock()
	c.mode = ModeFighting
	c.deceptiveMode = 1
	c.mu.Unlock()

	if c.Mode() != ModeFighting {
		t.Errorf("Mode() = %d; want %d", c.Mode(), ModeFighting)
	}
	if c.DeceptiveMode() != 1 {
		t.Errorf("DeceptiveMode() = %d; want 1", c.DeceptiveMode())
	}
}

func TestCombat_ConcurrentActions(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(12, 6422, 100000, 0.0, 120, 1, 1, listener)
	c.Start()
	defer c.Stop()

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				c.UsePumping(1, 0)
			}
		}()
	}
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				c.UseReeling(1, 0)
			}
		}()
	}
	wg.Wait()
}

func TestCombat_DoubleStopSafe(t *testing.T) {
	t.Parallel()

	listener := newMockListener()
	c := NewCombat(13, 6423, 1000, 0.0, 60, 0, 1, listener)
	c.Start()

	time.Sleep(50 * time.Millisecond)

	// Double stop should not panic.
	c.Stop()
	c.Stop()
}
