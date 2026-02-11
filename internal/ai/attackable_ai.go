package ai

import (
	"log/slog"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// AttackFunc is a callback to execute NPC attack on a target WorldObject.
// Injected by SpawnManager to avoid import cycle with CombatManager.
// Phase 5.7: NPC Aggro & Basic AI.
type AttackFunc func(monster *model.Monster, target *model.WorldObject)

// ScanFunc scans visible objects around position (x, y).
// Injected by SpawnManager to avoid import cycle with world package.
// Phase 5.7: NPC Aggro & Basic AI.
type ScanFunc func(x, y int32, fn func(*model.WorldObject) bool)

// GetObjectFunc looks up a WorldObject by objectID.
// Injected by SpawnManager to avoid import cycle with world package.
// Phase 5.7: NPC Aggro & Basic AI.
type GetObjectFunc func(objectID uint32) (*model.WorldObject, bool)

// AttackableAI implements AI for aggressive monsters.
// State machine: IDLE → ACTIVE (scan for targets) → ATTACK (execute attacks).
// Phase 5.7: NPC Aggro & Basic AI.
// Java reference: AttackableAI.thinkActive(), Attackable.addDamageHate()
type AttackableAI struct {
	monster   *model.Monster
	isRunning atomic.Bool

	// globalAggro starts at -10 (10-tick spawn immunity).
	// Incremented each tick. When >= 0, NPC can detect players.
	// Receiving damage sets it to 0 (cancels immunity).
	globalAggro atomic.Int32

	attackFunc    AttackFunc
	scanFunc      ScanFunc
	getObjectFunc GetObjectFunc
}

// NewAttackableAI creates a new AttackableAI controller for an aggressive monster.
// Phase 5.7: NPC Aggro & Basic AI.
func NewAttackableAI(
	monster *model.Monster,
	attackFunc AttackFunc,
	scanFunc ScanFunc,
	getObjectFunc GetObjectFunc,
) *AttackableAI {
	return &AttackableAI{
		monster:       monster,
		attackFunc:    attackFunc,
		scanFunc:      scanFunc,
		getObjectFunc: getObjectFunc,
	}
}

// Start starts the AI controller.
// Sets globalAggro to -10 for 10-tick spawn immunity.
func (ai *AttackableAI) Start() {
	ai.isRunning.Store(true)
	ai.globalAggro.Store(-10) // 10-tick spawn immunity
	ai.SetIntention(model.IntentionActive)

	if IsDebugEnabled() {
		slog.Debug("attackable AI started",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID(),
			"aggroRange", ai.monster.AggroRange())
	}
}

// Stop stops the AI controller.
func (ai *AttackableAI) Stop() {
	ai.isRunning.Store(false)
	ai.SetIntention(model.IntentionIdle)
	ai.monster.AggroList().Clear()
	ai.monster.ClearTarget()

	if IsDebugEnabled() {
		slog.Debug("attackable AI stopped",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID())
	}
}

// SetIntention sets AI intention on the underlying NPC.
func (ai *AttackableAI) SetIntention(intention model.Intention) {
	oldIntention := ai.monster.Intention()
	ai.monster.SetIntention(intention)

	if oldIntention != intention && IsDebugEnabled() {
		slog.Debug("attackable AI intention changed",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID(),
			"from", oldIntention,
			"to", intention)
	}
}

// CurrentIntention returns current AI intention.
func (ai *AttackableAI) CurrentIntention() model.Intention {
	return ai.monster.Intention()
}

// Npc returns the underlying NPC.
func (ai *AttackableAI) Npc() *model.Npc {
	return ai.monster.Npc
}

// NotifyDamage handles NPC receiving damage.
// Cancels spawn immunity and adds attacker to hate list.
// If idle, switches to attack mode.
// Phase 5.7: NPC Aggro & Basic AI.
func (ai *AttackableAI) NotifyDamage(attackerID uint32, damage int32) {
	if !ai.isRunning.Load() {
		return
	}

	if ai.monster.IsDead() {
		return
	}

	// Cancel spawn immunity
	if ai.globalAggro.Load() < 0 {
		ai.globalAggro.Store(0)
	}

	// Calculate hate and add to hate list
	hate := model.CalcHateValue(damage, ai.monster.Level())
	ai.monster.AggroList().AddHate(attackerID, hate)
	ai.monster.AggroList().AddDamage(attackerID, int64(damage))

	// If idle/active, switch to attack
	currentIntention := ai.CurrentIntention()
	if currentIntention != model.IntentionAttack {
		mostHated := ai.monster.AggroList().GetMostHated()
		if mostHated != 0 {
			ai.monster.SetTarget(mostHated)
			ai.SetIntention(model.IntentionAttack)
		}
	}

	if IsDebugEnabled() {
		slog.Debug("attackable AI notified of damage",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID(),
			"attackerID", attackerID,
			"damage", damage,
			"hate", hate)
	}
}

// Tick performs AI tick (called every second by TickManager).
// Phase 5.7: NPC Aggro & Basic AI.
func (ai *AttackableAI) Tick() {
	if !ai.isRunning.Load() {
		return
	}

	if ai.monster.IsDead() {
		return
	}

	// Increment globalAggro (spawn immunity countdown)
	if ai.globalAggro.Load() < 0 {
		ai.globalAggro.Add(1)
	}

	currentIntention := ai.CurrentIntention()

	switch currentIntention {
	case model.IntentionAttack:
		ai.thinkAttack()
	case model.IntentionActive, model.IntentionIdle:
		ai.thinkActive()
	}
}

// thinkActive scans for players in aggro range.
// If player found, adds hate and switches to ATTACK.
// Phase 5.7: NPC Aggro & Basic AI.
// Java reference: AttackableAI.thinkActive()
func (ai *AttackableAI) thinkActive() {
	// Spawn immunity still active
	if ai.globalAggro.Load() < 0 {
		return
	}

	if ai.scanFunc == nil {
		return
	}

	npcLoc := ai.monster.Location()
	aggroRange := ai.monster.AggroRange()
	aggroRangeSq := int64(aggroRange) * int64(aggroRange)

	ai.scanFunc(npcLoc.X, npcLoc.Y, func(obj *model.WorldObject) bool {
		// Skip self
		if obj.ObjectID() == ai.monster.ObjectID() {
			return true
		}

		// Only aggro on players
		player, ok := obj.Data.(*model.Player)
		if !ok {
			return true
		}

		// Skip dead players
		if player.IsDead() {
			return true
		}

		// Check distance
		playerLoc := obj.Location()
		distSq := npcLoc.DistanceSquared(playerLoc)
		if distSq > aggroRangeSq {
			return true
		}

		// Player in aggro range — add initial hate
		ai.monster.AggroList().AddHate(obj.ObjectID(), 1)
		return true
	})

	// If we have targets, switch to attack
	if !ai.monster.AggroList().IsEmpty() {
		mostHated := ai.monster.AggroList().GetMostHated()
		if mostHated != 0 {
			ai.monster.SetTarget(mostHated)
			ai.SetIntention(model.IntentionAttack)

			if IsDebugEnabled() {
				slog.Debug("attackable AI acquired target",
					"npc", ai.monster.Name(),
					"objectID", ai.monster.ObjectID(),
					"targetID", mostHated)
			}
		}
	}
}

// thinkAttack validates current target and executes attack.
// If target is invalid (dead, out of range, gone), returns to ACTIVE.
// Phase 5.7: NPC Aggro & Basic AI.
// Java reference: AttackableAI.thinkAttack()
func (ai *AttackableAI) thinkAttack() {
	// Get most hated target
	targetID := ai.monster.AggroList().GetMostHated()
	if targetID == 0 {
		// No targets — return to active
		ai.monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		return
	}

	// Update target to most hated
	ai.monster.SetTarget(targetID)

	// Look up target in world
	if ai.getObjectFunc == nil {
		return
	}

	targetObj, found := ai.getObjectFunc(targetID)
	if !found {
		// Target disconnected — remove from hate list
		ai.monster.AggroList().Remove(targetID)
		ai.monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		return
	}

	// Validate target is alive
	if player, ok := targetObj.Data.(*model.Player); ok {
		if player.IsDead() {
			ai.monster.AggroList().Remove(targetID)
			ai.monster.ClearTarget()
			ai.SetIntention(model.IntentionActive)
			return
		}
	}

	// Check attack range (NPC is stationary in Phase 5.7, movement in Phase 5.10)
	npcLoc := ai.monster.Location()
	targetLoc := targetObj.Location()
	distSq := npcLoc.DistanceSquared(targetLoc)

	// Attack range: use max physical attack range (100 units)
	const attackRangeSq = 100 * 100
	if distSq > attackRangeSq {
		// Target out of attack range — skip attack (no movement in Phase 5.7)
		// Don't clear target, player might come back
		return
	}

	// Execute attack via callback
	if ai.attackFunc != nil {
		ai.attackFunc(ai.monster, targetObj)
	}
}
