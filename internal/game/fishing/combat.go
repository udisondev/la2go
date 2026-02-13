// Package fishing implements the L2 Interlude fishing combat minigame.
//
// The fishing combat is a mode-based minigame where a player must use
// Pumping and Reeling skills to deplete the fish's HP before time runs out.
// The fish has a mode (resting/fighting) and optionally a deceptive mode
// (for hard fish) that affects which action is effective.
package fishing

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

// Combat mode constants.
const (
	ModeResting  = 0
	ModeFighting = 1
)

// Animation type constants.
const (
	AnimNone    = 0
	AnimPumping = 1
	AnimReeling = 2
)

// Action result constants (goodUse).
const (
	ResultNone    = 0
	ResultSuccess = 1
	ResultFailed  = 2
)

// Probabilities (out of 100).
const (
	probFightingStart  = 20 // 20% chance fighting mode at start
	probDeceptiveStart = 10 // 10% chance deceptive for hard fish
	probModeSwitch     = 30 // 30% chance mode switch per tick
	probDeceptiveFlip  = 10 // 10% chance deceptive flip per tick (hard only)
	probResist         = 10 // 10% chance action is resisted
)

// CombatState represents a snapshot of combat state sent to the client.
type CombatState struct {
	TimeLeft      int32
	FishHP        int32
	Mode          int32 // 0=resting, 1=fighting
	DeceptiveMode int32 // 0=normal, 1=deceptive
	GoodUse       int32 // 0=none, 1=success, 2=failed
	Anim          int32 // 0=none, 1=pumping, 2=reeling
	Penalty       int32
	HPBarColor    int32 // 0=normal, 1=deceptive indicator
}

// CombatListener receives combat events for packet dispatch.
type CombatListener interface {
	OnFishingTick(objectID int32, state CombatState)
	OnFishingEnd(objectID int32, win bool)
}

// Combat represents an active fishing combat session.
type Combat struct {
	mu sync.Mutex

	objectID int32 // player objectID
	fishID   int32 // fish item ID (reward)

	timeLeft      int32
	stop          int32 // mode switch cooldown (0 = ready, 1 = cooling)
	goodUse       int32
	anim          int32
	mode          int32
	deceptiveMode int32
	fishCurHP     int32
	fishMaxHP     int32
	regenHP       float64
	isUpperGrade  bool // true for fish_hard
	lureType      int32

	listener CombatListener
	done     atomic.Bool
	stopCh   chan struct{}
}

// NewCombat creates a new fishing combat session.
// combatDuration is in seconds.
func NewCombat(objectID, fishID, fishHP int32, hpRegen float64,
	combatDuration int32, fishGrade int32, lureType int32, listener CombatListener) *Combat {

	isHard := fishGrade == 2

	mode := int32(ModeResting)
	if rand.IntN(100) < probFightingStart {
		mode = ModeFighting
	}

	deceptive := int32(0)
	if isHard && rand.IntN(100) < probDeceptiveStart {
		deceptive = 1
	}

	return &Combat{
		objectID:      objectID,
		fishID:        fishID,
		timeLeft:      combatDuration,
		fishCurHP:     fishHP,
		fishMaxHP:     fishHP,
		regenHP:       hpRegen,
		isUpperGrade:  isHard,
		mode:          mode,
		deceptiveMode: deceptive,
		lureType:      lureType,
		listener:      listener,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the combat AI ticker (1Hz). The goroutine exits when combat
// ends (win/lose/timeout) or Stop is called.
func (c *Combat) Start() {
	go c.run()
}

// Stop terminates the combat prematurely (e.g. player disconnect).
func (c *Combat) Stop() {
	if c.done.CompareAndSwap(false, true) {
		close(c.stopCh)
	}
}

// ObjectID returns the player's object ID.
func (c *Combat) ObjectID() int32 { return c.objectID }

// FishID returns the reward fish item ID.
func (c *Combat) FishID() int32 { return c.fishID }

// MaxHP returns the fish's max HP.
func (c *Combat) MaxHP() int32 { return c.fishMaxHP }

// Mode returns the current fish mode. Safe for concurrent read.
func (c *Combat) Mode() int32 {
	c.mu.Lock()
	m := c.mode
	c.mu.Unlock()
	return m
}

// DeceptiveMode returns the current deceptive mode flag.
func (c *Combat) DeceptiveMode() int32 {
	c.mu.Lock()
	d := c.deceptiveMode
	c.mu.Unlock()
	return d
}

// LureType returns the lure type (0=newbie, 1=normal, 2=night).
func (c *Combat) LureType() int32 { return c.lureType }

// UsePumping handles the player's Pumping skill action.
// dmg is the calculated damage, pen is the penalty.
func (c *Combat) UsePumping(dmg, pen int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.done.Load() {
		return
	}

	c.anim = AnimPumping

	// 10% chance: resisted
	if rand.IntN(100) < probResist {
		c.goodUse = ResultFailed
		return
	}

	// Pumping is effective in resting mode (mode == 0).
	// Deceptive mode inverts the logic.
	effective := (c.mode == ModeResting && c.deceptiveMode == 0) ||
		(c.mode == ModeFighting && c.deceptiveMode == 1)

	if effective {
		c.fishCurHP -= dmg
		c.goodUse = ResultSuccess
	} else {
		c.fishCurHP += dmg
		c.goodUse = ResultFailed
	}

	c.fishCurHP -= pen
	c.clampHP()
}

// UseReeling handles the player's Reeling skill action.
// dmg is the calculated damage, pen is the penalty.
func (c *Combat) UseReeling(dmg, pen int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.done.Load() {
		return
	}

	c.anim = AnimReeling

	// 10% chance: resisted
	if rand.IntN(100) < probResist {
		c.goodUse = ResultFailed
		return
	}

	// Reeling is effective in fighting mode (mode == 1).
	// Deceptive mode inverts the logic.
	effective := (c.mode == ModeFighting && c.deceptiveMode == 0) ||
		(c.mode == ModeResting && c.deceptiveMode == 1)

	if effective {
		c.fishCurHP -= dmg
		c.goodUse = ResultSuccess
	} else {
		c.fishCurHP += dmg
		c.goodUse = ResultFailed
	}

	c.fishCurHP -= pen
	c.clampHP()
}

// run is the AI goroutine — ticks every second until combat ends.
func (c *Combat) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if end := c.tick(); end {
				return
			}
		}
	}
}

// tick processes one second of combat AI. Returns true if combat has ended.
func (c *Combat) tick() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.done.Load() {
		return true
	}

	c.timeLeft--

	// HP regen: fighting mode regenerates (unless deceptive inverts it).
	if c.mode == ModeFighting {
		if c.deceptiveMode == 0 {
			c.fishCurHP += int32(c.regenHP)
		}
	} else if c.deceptiveMode == 1 {
		c.fishCurHP += int32(c.regenHP)
	}

	// Mode switching logic (every other tick due to stop counter).
	if c.stop == 0 {
		c.stop = 1
		if rand.IntN(100) >= (100 - probModeSwitch) {
			if c.mode == ModeResting {
				c.mode = ModeFighting
			} else {
				c.mode = ModeResting
			}
		}
		if c.isUpperGrade && rand.IntN(100) >= (100-probDeceptiveFlip) {
			if c.deceptiveMode == 0 {
				c.deceptiveMode = 1
			} else {
				c.deceptiveMode = 0
			}
		}
	} else {
		c.stop--
	}

	// Check end conditions.
	// Fish escaped: HP regenerated to 200%.
	if c.fishCurHP >= c.fishMaxHP*2 {
		c.finish(false)
		return true
	}

	// Fish caught: HP depleted.
	if c.fishCurHP <= 0 {
		c.finish(true)
		return true
	}

	// Timeout.
	if c.timeLeft <= 0 {
		c.finish(false)
		return true
	}

	// Notify listener about tick state.
	if c.listener != nil {
		c.listener.OnFishingTick(c.objectID, CombatState{
			TimeLeft:      c.timeLeft,
			FishHP:        c.fishCurHP,
			Mode:          c.mode,
			DeceptiveMode: c.deceptiveMode,
			GoodUse:       c.goodUse,
			Anim:          c.anim,
			HPBarColor:    c.deceptiveMode,
		})
	}

	// Reset action state for next tick.
	c.goodUse = ResultNone
	c.anim = AnimNone

	return false
}

// finish ends the combat with the given result.
// Caller must hold c.mu.
func (c *Combat) finish(win bool) {
	if !c.done.CompareAndSwap(false, true) {
		return
	}
	close(c.stopCh)
	if c.listener != nil {
		c.listener.OnFishingEnd(c.objectID, win)
	}
}

// clampHP ensures fishCurHP stays in [0, fishMaxHP*2].
// Caller must hold c.mu.
func (c *Combat) clampHP() {
	if c.fishCurHP < 0 {
		c.fishCurHP = 0
	}
	maxAllowed := c.fishMaxHP * 2
	if c.fishCurHP > maxAllowed {
		c.fishCurHP = maxAllowed
	}
}
