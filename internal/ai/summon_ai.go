package ai

import (
	"log/slog"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// SummonAttackFunc executes summon's attack on a target.
// Injected to avoid import cycle with CombatManager.
// Phase 19: Pets/Summons System.
type SummonAttackFunc func(summon *model.Summon, target *model.WorldObject)

// SummonAI implements AI controller for summoned creatures (pets and servitors).
// State machine: IDLE → FOLLOW (track owner) → ATTACK (fight target).
// Phase 19: Pets/Summons System.
type SummonAI struct {
	summon    *model.Summon
	isRunning atomic.Bool

	// Callbacks (injected to avoid import cycles)
	getObjectFunc GetObjectFunc
	attackFunc    SummonAttackFunc
}

// NewSummonAI creates a new SummonAI controller.
func NewSummonAI(
	summon *model.Summon,
	getObjectFunc GetObjectFunc,
	attackFunc SummonAttackFunc,
) *SummonAI {
	return &SummonAI{
		summon:        summon,
		getObjectFunc: getObjectFunc,
		attackFunc:    attackFunc,
	}
}

// Start starts the summon AI. Sets initial intention to follow owner.
func (ai *SummonAI) Start() {
	ai.isRunning.Store(true)
	ai.summon.SetFollow(true)
	ai.SetIntention(model.IntentionFollow)

	if IsDebugEnabled() {
		slog.Debug("summon AI started",
			"objectID", ai.summon.ObjectID(),
			"name", ai.summon.Name(),
			"ownerID", ai.summon.OwnerID(),
			"type", ai.summon.Type())
	}
}

// Stop stops the summon AI.
func (ai *SummonAI) Stop() {
	ai.isRunning.Store(false)
	ai.summon.ClearTarget()
	ai.SetIntention(model.IntentionIdle)

	if IsDebugEnabled() {
		slog.Debug("summon AI stopped",
			"objectID", ai.summon.ObjectID(),
			"name", ai.summon.Name())
	}
}

// SetIntention sets AI intention on the underlying summon.
func (ai *SummonAI) SetIntention(intention model.Intention) {
	ai.summon.SetIntention(intention)
}

// CurrentIntention returns current AI intention.
func (ai *SummonAI) CurrentIntention() model.Intention {
	return ai.summon.Intention()
}

// Npc returns nil — summon is not an NPC.
// Implements Controller interface for compatibility with TickManager.
func (ai *SummonAI) Npc() *model.Npc {
	return nil
}

// NotifyDamage handles summon receiving damage.
// If following, switches to attack the attacker.
func (ai *SummonAI) NotifyDamage(attackerID uint32, damage int32) {
	if !ai.isRunning.Load() {
		return
	}
	if ai.summon.IsDead() {
		return
	}

	// If no current target, retaliate against attacker
	if ai.summon.Target() == 0 {
		ai.summon.SetTarget(attackerID)
		ai.summon.SetFollow(false)
		ai.SetIntention(model.IntentionAttack)
	}
}

// Tick performs AI tick (called every second by TickManager).
func (ai *SummonAI) Tick() {
	if !ai.isRunning.Load() {
		return
	}
	if ai.summon.IsDead() {
		return
	}

	switch ai.CurrentIntention() {
	case model.IntentionFollow:
		ai.thinkFollow()
	case model.IntentionAttack:
		ai.thinkAttack()
	case model.IntentionIdle:
		if ai.summon.IsFollowing() {
			ai.SetIntention(model.IntentionFollow)
		}
	}
}

// thinkFollow moves summon towards owner if too far.
func (ai *SummonAI) thinkFollow() {
	if ai.getObjectFunc == nil {
		return
	}

	ownerObj, found := ai.getObjectFunc(ai.summon.OwnerID())
	if !found {
		return
	}

	summonLoc := ai.summon.Location()
	ownerLoc := ownerObj.Location()
	distSq := summonLoc.DistanceSquared(ownerLoc)

	// Follow distance: teleport if > 2000 units, move if > 100 units
	const teleportDistSq = 2000 * 2000
	const followDistSq = 100 * 100

	if distSq > teleportDistSq {
		// Teleport near owner
		ai.summon.SetLocation(model.NewLocation(
			ownerLoc.X+50, ownerLoc.Y+50, ownerLoc.Z, ownerLoc.Heading,
		))
	} else if distSq > followDistSq {
		// Move towards owner (simplified: snap to nearby position)
		ai.summon.SetLocation(model.NewLocation(
			ownerLoc.X+30, ownerLoc.Y+30, ownerLoc.Z, ownerLoc.Heading,
		))
	}
}

// thinkAttack validates target and executes attack.
func (ai *SummonAI) thinkAttack() {
	targetID := ai.summon.Target()
	if targetID == 0 {
		ai.returnToFollow()
		return
	}

	if ai.getObjectFunc == nil {
		return
	}

	targetObj, found := ai.getObjectFunc(targetID)
	if !found {
		ai.summon.ClearTarget()
		ai.returnToFollow()
		return
	}

	// Check target alive
	if player, ok := targetObj.Data.(*model.Player); ok {
		if player.IsDead() {
			ai.summon.ClearTarget()
			ai.returnToFollow()
			return
		}
	}

	// Check attack range
	summonLoc := ai.summon.Location()
	targetLoc := targetObj.Location()
	distSq := summonLoc.DistanceSquared(targetLoc)

	const attackRangeSq = 100 * 100
	if distSq > attackRangeSq {
		// Too far to attack — move closer (simplified: snap)
		return
	}

	// Execute attack via callback
	if ai.attackFunc != nil {
		ai.attackFunc(ai.summon, targetObj)
	}
}

// returnToFollow sets summon back to follow mode.
func (ai *SummonAI) returnToFollow() {
	if ai.summon.IsFollowing() {
		ai.SetIntention(model.IntentionFollow)
	} else {
		ai.SetIntention(model.IntentionIdle)
	}
}

// OrderAttack commands summon to attack a specific target.
func (ai *SummonAI) OrderAttack(targetID uint32) {
	if !ai.isRunning.Load() || ai.summon.IsDead() {
		return
	}
	ai.summon.SetTarget(targetID)
	ai.summon.SetFollow(false)
	ai.SetIntention(model.IntentionAttack)
}

// OrderFollow commands summon to follow owner.
func (ai *SummonAI) OrderFollow() {
	if !ai.isRunning.Load() {
		return
	}
	ai.summon.ClearTarget()
	ai.summon.SetFollow(true)
	ai.SetIntention(model.IntentionFollow)
}

// OrderStop commands summon to stop and stay in place.
func (ai *SummonAI) OrderStop() {
	if !ai.isRunning.Load() {
		return
	}
	ai.summon.ClearTarget()
	ai.summon.SetFollow(false)
	ai.SetIntention(model.IntentionIdle)
}

// Summon returns the underlying Summon model.
func (ai *SummonAI) Summon() *model.Summon {
	return ai.summon
}
