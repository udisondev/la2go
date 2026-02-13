package ai

import (
	"log/slog"
	"math"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// AttackFunc is a callback to execute NPC attack on a target WorldObject.
// Injected by SpawnManager to avoid import cycle with CombatManager.
type AttackFunc func(monster *model.Monster, target *model.WorldObject)

// ScanFunc scans visible objects around position (x, y).
// Injected by SpawnManager to avoid import cycle with world package.
type ScanFunc func(x, y int32, fn func(*model.WorldObject) bool)

// GetObjectFunc looks up a WorldObject by objectID.
// Injected by SpawnManager to avoid import cycle with world package.
type GetObjectFunc func(objectID uint32) (*model.WorldObject, bool)

// NpcCastFunc is a callback to execute NPC skill cast on a target.
// Injected by SpawnManager. If nil, NPC skill casting is disabled.
type NpcCastFunc func(monster *model.Monster, target *model.WorldObject, skillID, skillLevel int32)

// MoveNpcFunc is a callback to move NPC toward a location.
// Injected by SpawnManager. If nil, NPC movement (chase/walk) is disabled.
type MoveNpcFunc func(npc *model.Npc, x, y, z int32)

// AI constants matching Java L2J Mobius AttackableAI / NpcConfig.
const (
	randomWalkRate    = 30   // 1/30 chance of random walk per tick (~3.3%)
	maxDriftRange     = 300  // max distance NPC can drift from spawn
	maxDriftRangeSq   = int64(maxDriftRange) * int64(maxDriftRange)
	chaseRangeNormal  = 1500 // max chase distance for regular monsters
	chaseRangeRaid    = 3000 // max chase distance for raids
	hateForgetChance  = 500  // 1/500 per tick chance to forget aggro at full HP
	attackRangeBase   = 100  // default physical attack range (atkRange + collision radius)
	factionZTolerance = 600  // max Z-difference for faction call
)

// AttackableAI implements AI for aggressive monsters.
// State machine: IDLE → ACTIVE (scan/random walk) → ATTACK (chase/cast/attack).
// Phase 15: NPC skill casting, chase, faction call, random walk, hate decay, return to spawn.
type AttackableAI struct {
	monster   *model.Monster
	isRunning atomic.Bool

	// globalAggro starts at -10 (10-tick spawn immunity).
	// Incremented each tick. When >= 0, NPC can detect players.
	// Receiving damage sets it to 0 (cancels immunity).
	globalAggro atomic.Int32

	// attackTimeout tracks when to abandon attack (Java: MAX_ATTACK_TIMEOUT = 2 min).
	attackTimeout atomic.Int64 // UnixMilli deadline

	// Skill cooldowns: map[skillID]readyAtUnixMilli.
	// Simple approach — no sync needed, only accessed from Tick() goroutine.
	skillCooldowns map[int32]int64

	// Callbacks (injected to avoid import cycles)
	attackFunc    AttackFunc
	scanFunc      ScanFunc
	getObjectFunc GetObjectFunc
	castFunc      NpcCastFunc
	moveFunc      MoveNpcFunc
}

// NewAttackableAI creates a new AttackableAI controller for an aggressive monster.
func NewAttackableAI(
	monster *model.Monster,
	attackFunc AttackFunc,
	scanFunc ScanFunc,
	getObjectFunc GetObjectFunc,
) *AttackableAI {
	return &AttackableAI{
		monster:        monster,
		attackFunc:     attackFunc,
		scanFunc:       scanFunc,
		getObjectFunc:  getObjectFunc,
		skillCooldowns: make(map[int32]int64),
	}
}

// SetCastFunc sets the NPC skill cast callback.
func (ai *AttackableAI) SetCastFunc(fn NpcCastFunc) {
	ai.castFunc = fn
}

// SetMoveFunc sets the NPC movement callback.
func (ai *AttackableAI) SetMoveFunc(fn MoveNpcFunc) {
	ai.moveFunc = fn
}

// Start starts the AI controller.
// Sets globalAggro to -10 for 10-tick spawn immunity.
func (ai *AttackableAI) Start() {
	ai.isRunning.Store(true)
	ai.globalAggro.Store(-10)
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
// Cancels spawn immunity, adds attacker to hate list,
// triggers faction call, and switches to attack mode.
func (ai *AttackableAI) NotifyDamage(attackerID uint32, damage int32) {
	if !ai.isRunning.Load() || ai.monster.IsDead() {
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
			ai.attackTimeout.Store(time.Now().Add(2 * time.Minute).UnixMilli())
		}
	}

	// Faction call: nearby NPCs of same clan should help
	ai.callFaction(attackerID)

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
func (ai *AttackableAI) Tick() {
	if !ai.isRunning.Load() || ai.monster.IsDead() {
		return
	}

	// Increment globalAggro (spawn immunity countdown)
	if g := ai.globalAggro.Load(); g < 0 {
		ai.globalAggro.Add(1)
	} else if g > 0 {
		ai.globalAggro.Add(-1)
	}

	switch ai.CurrentIntention() {
	case model.IntentionAttack:
		ai.thinkAttack()
	case model.IntentionActive, model.IntentionIdle:
		ai.thinkActive()
	}
}

// thinkActive scans for players, performs random walk, checks hate decay.
// Java reference: AttackableAI.thinkActive()
func (ai *AttackableAI) thinkActive() {
	// Spawn immunity still active
	if ai.globalAggro.Load() < 0 {
		return
	}

	// Feature 5: Hate decay — forget aggro when at full HP/MP (1/500 chance per tick)
	ai.checkHateDecay()

	// Feature 4: Random walk — 1/30 chance per tick to walk near spawn
	ai.tryRandomWalk()

	// Feature 8 (partial): Return to spawn if drifted too far
	ai.checkReturnToSpawn()

	// Scan for players in aggro range
	if ai.scanFunc == nil {
		return
	}

	npcLoc := ai.monster.Location()
	aggroRange := ai.monster.AggroRange()
	aggroRangeSq := int64(aggroRange) * int64(aggroRange)

	ai.scanFunc(npcLoc.X, npcLoc.Y, func(obj *model.WorldObject) bool {
		if obj.ObjectID() == ai.monster.ObjectID() {
			return true
		}

		player, ok := obj.Data.(*model.Player)
		if !ok || player.IsDead() {
			return true
		}

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
			ai.attackTimeout.Store(time.Now().Add(2 * time.Minute).UnixMilli())

			if IsDebugEnabled() {
				slog.Debug("attackable AI acquired target",
					"npc", ai.monster.Name(),
					"objectID", ai.monster.ObjectID(),
					"targetID", mostHated)
			}
		}
	}
}

// thinkAttack validates target, tries skill casting, chases, and executes attacks.
// Java reference: AttackableAI.thinkAttack()
func (ai *AttackableAI) thinkAttack() {
	// Check attack timeout (2 min without landing a hit)
	if time.Now().UnixMilli() > ai.attackTimeout.Load() {
		ai.returnHome()
		return
	}

	// Get most hated target
	targetID := ai.monster.AggroList().GetMostHated()
	if targetID == 0 {
		ai.monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		return
	}

	ai.monster.SetTarget(targetID)

	if ai.getObjectFunc == nil {
		return
	}

	targetObj, found := ai.getObjectFunc(targetID)
	if !found {
		ai.monster.AggroList().Remove(targetID)
		ai.monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		return
	}

	// Validate target is alive
	if isTargetDead(targetObj) {
		ai.monster.AggroList().Remove(targetID)
		ai.monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		return
	}

	// Feature 8: Check if too far from spawn — return home
	if ai.isTooFarFromSpawn() {
		ai.returnHome()
		return
	}

	// Calculate distance to target
	npcLoc := ai.monster.Location()
	targetLoc := targetObj.Location()
	distSq := npcLoc.DistanceSquared(targetLoc)
	dist := math.Sqrt(float64(distSq))

	// Determine attack range
	atkRange := ai.getAttackRange()

	// Feature 1: Try NPC skill casting first
	if ai.trySkillCast(targetObj, dist) {
		return
	}

	// Feature 2: Chase — move toward target if out of attack range
	if dist > float64(atkRange) {
		ai.chaseTarget(targetObj)
		return
	}

	// Execute physical attack via callback
	if ai.attackFunc != nil {
		ai.attackFunc(ai.monster, targetObj)
		// Reset attack timeout on successful attack attempt
		ai.attackTimeout.Store(time.Now().Add(2 * time.Minute).UnixMilli())
	}
}

// Feature 1: trySkillCast selects and casts an NPC skill if available.
// Returns true if a skill was cast (caller should skip physical attack).
// Java reference: AttackableAI skill selection priority.
func (ai *AttackableAI) trySkillCast(target *model.WorldObject, dist float64) bool {
	if ai.castFunc == nil {
		return false
	}

	npcDef := data.GetNpcDef(ai.monster.TemplateID())
	if npcDef == nil {
		return false
	}

	skills := npcDef.Skills()
	if len(skills) == 0 {
		return false
	}

	now := time.Now().UnixMilli()

	// Collect usable skills (off cooldown, in range, enough MP)
	type usableSkill struct {
		id, level int32
		tmpl      *data.SkillTemplate
	}
	var candidates []usableSkill

	for _, sk := range skills {
		// Check cooldown
		if readyAt, ok := ai.skillCooldowns[sk.SkillID()]; ok && now < readyAt {
			continue
		}

		tmpl := data.GetSkillTemplate(sk.SkillID(), sk.SkillLevel())
		if tmpl == nil {
			continue
		}

		// Check cast range (0 means melee / self)
		if tmpl.CastRange > 0 && dist > float64(tmpl.CastRange) {
			continue
		}

		// Check MP
		if tmpl.MpConsume > 0 && ai.monster.CurrentMP() < tmpl.MpConsume {
			continue
		}

		candidates = append(candidates, usableSkill{
			id:    sk.SkillID(),
			level: sk.SkillLevel(),
			tmpl:  tmpl,
		})
	}

	if len(candidates) == 0 {
		return false
	}

	// Pick a random usable skill (Java: Rnd.get(size))
	chosen := candidates[rand.IntN(len(candidates))]

	// Cast the skill
	ai.castFunc(ai.monster, target, chosen.id, chosen.level)

	// Set cooldown (reuseDelay from template, minimum 1s)
	cooldown := max(int64(chosen.tmpl.ReuseDelay), 1000)
	ai.skillCooldowns[chosen.id] = now + cooldown

	if IsDebugEnabled() {
		slog.Debug("NPC cast skill",
			"npc", ai.monster.Name(),
			"skill", chosen.tmpl.Name,
			"skillID", chosen.id,
			"level", chosen.level,
			"target", target.ObjectID())
	}

	return true
}

// Feature 2: chaseTarget moves NPC toward the target.
func (ai *AttackableAI) chaseTarget(target *model.WorldObject) {
	if ai.moveFunc == nil {
		return
	}

	targetLoc := target.Location()
	ai.moveFunc(ai.monster.Npc, targetLoc.X, targetLoc.Y, targetLoc.Z)

	if IsDebugEnabled() {
		slog.Debug("NPC chasing target",
			"npc", ai.monster.Name(),
			"targetID", target.ObjectID(),
			"targetX", targetLoc.X,
			"targetY", targetLoc.Y)
	}
}

// Feature 3: callFaction calls nearby NPCs of same clan for help.
// Java reference: AttackableAI lines 2422-2504.
func (ai *AttackableAI) callFaction(attackerID uint32) {
	if ai.scanFunc == nil || ai.getObjectFunc == nil {
		return
	}

	npcDef := data.GetNpcDef(ai.monster.TemplateID())
	if npcDef == nil || len(npcDef.Clans()) == 0 {
		return
	}

	factionRange := npcDef.ClanHelpRange()
	if factionRange <= 0 {
		factionRange = 300 // default faction range
	}
	factionRangeSq := int64(factionRange) * int64(factionRange)

	npcLoc := ai.monster.Location()
	callerClans := npcDef.Clans()
	callerID := npcDef.ID()

	ai.scanFunc(npcLoc.X, npcLoc.Y, func(obj *model.WorldObject) bool {
		if obj.ObjectID() == ai.monster.ObjectID() {
			return true
		}

		// Check if it's a monster with AI
		nearbyMonster, ok := obj.Data.(*model.Monster)
		if !ok || nearbyMonster.IsDead() {
			return true
		}

		// Check distance
		objLoc := obj.Location()
		distSq := npcLoc.DistanceSquared(objLoc)
		if distSq > factionRangeSq {
			return true
		}

		// Check Z tolerance
		dz := npcLoc.Z - objLoc.Z
		if dz < 0 {
			dz = -dz
		}
		if dz > factionZTolerance {
			return true
		}

		// Check intention — only call idle/active NPCs
		intent := nearbyMonster.Intention()
		if intent != model.IntentionIdle && intent != model.IntentionActive {
			return true
		}

		// Check clan match
		nearbyDef := data.GetNpcDef(nearbyMonster.TemplateID())
		if nearbyDef == nil {
			return true
		}
		if !nearbyDef.IsClan(callerClans) {
			return true
		}

		// Check ignore list
		if nearbyDef.IgnoresNpcID(callerID) {
			return true
		}

		// Add minimal hate to the nearby NPC so it attacks the aggressor
		nearbyMonster.AggroList().AddHate(attackerID, 1)

		if IsDebugEnabled() {
			slog.Debug("faction call",
				"caller", ai.monster.Name(),
				"helper", nearbyMonster.Name(),
				"attacker", attackerID)
		}

		return true
	})
}

// Feature 4: tryRandomWalk makes NPC walk randomly near spawn point.
// Java: 1/RANDOM_WALK_RATE chance per tick, within MAX_DRIFT_RANGE.
func (ai *AttackableAI) tryRandomWalk() {
	if ai.moveFunc == nil {
		return
	}

	spawn := ai.monster.Spawn()
	if spawn == nil {
		return
	}

	// 1/30 chance per tick (~3.3%)
	if rand.IntN(randomWalkRate) != 0 {
		return
	}

	spawnLoc := spawn.Location()

	// Random offset within maxDriftRange
	dx := rand.Int32N(int32(maxDriftRange)*2+1) - int32(maxDriftRange)
	dy := rand.Int32N(int32(maxDriftRange)*2+1) - int32(maxDriftRange)

	newX := spawnLoc.X + dx
	newY := spawnLoc.Y + dy

	ai.moveFunc(ai.monster.Npc, newX, newY, spawnLoc.Z)

	if IsDebugEnabled() {
		slog.Debug("NPC random walk",
			"npc", ai.monster.Name(),
			"toX", newX,
			"toY", newY)
	}
}

// Feature 5: checkHateDecay clears aggro list if NPC is at full HP/MP.
// Java: 1/500 chance per tick when at full health.
func (ai *AttackableAI) checkHateDecay() {
	if ai.monster.AggroList().IsEmpty() {
		return
	}

	// Only decay when at full HP and MP
	if ai.monster.CurrentHP() < ai.monster.MaxHP() ||
		ai.monster.CurrentMP() < ai.monster.MaxMP() {
		return
	}

	// 1/500 chance per tick (0.2%)
	if rand.IntN(hateForgetChance) != 0 {
		return
	}

	ai.monster.AggroList().Clear()
	ai.monster.ClearTarget()

	if IsDebugEnabled() {
		slog.Debug("NPC hate decayed — cleared aggro list",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID())
	}
}

// checkReturnToSpawn moves NPC toward spawn if it drifted too far (thinkActive).
func (ai *AttackableAI) checkReturnToSpawn() {
	if ai.moveFunc == nil {
		return
	}

	spawn := ai.monster.Spawn()
	if spawn == nil {
		return
	}

	npcLoc := ai.monster.Location()
	spawnLoc := spawn.Location()
	distSq := npcLoc.DistanceSquared(spawnLoc)

	if distSq > maxDriftRangeSq {
		ai.moveFunc(ai.monster.Npc, spawnLoc.X, spawnLoc.Y, spawnLoc.Z)

		if IsDebugEnabled() {
			slog.Debug("NPC returning to spawn (idle drift)",
				"npc", ai.monster.Name())
		}
	}
}

// isTooFarFromSpawn checks if NPC is beyond max chase distance from spawn.
func (ai *AttackableAI) isTooFarFromSpawn() bool {
	spawn := ai.monster.Spawn()
	if spawn == nil {
		return false
	}

	npcLoc := ai.monster.Location()
	spawnLoc := spawn.Location()
	distSq := npcLoc.DistanceSquared(spawnLoc)

	maxDist := int64(chaseRangeNormal)
	if data.IsRaidBoss(ai.monster.TemplateID()) || data.IsGrandBoss(ai.monster.TemplateID()) {
		maxDist = int64(chaseRangeRaid)
	}
	maxDistSq := maxDist * maxDist

	return distSq > maxDistSq
}

// returnHome clears aggro, resets HP/MP, and moves NPC back to spawn.
// Java reference: Attackable.returnHome()
func (ai *AttackableAI) returnHome() {
	ai.monster.AggroList().Clear()
	ai.monster.ClearTarget()
	ai.SetIntention(model.IntentionActive)

	// Restore full HP/MP
	ai.monster.SetCurrentHP(ai.monster.MaxHP())
	ai.monster.SetCurrentMP(ai.monster.MaxMP())

	// Move to spawn point
	spawn := ai.monster.Spawn()
	if spawn != nil && ai.moveFunc != nil {
		spawnLoc := spawn.Location()
		ai.moveFunc(ai.monster.Npc, spawnLoc.X, spawnLoc.Y, spawnLoc.Z)
	}

	if IsDebugEnabled() {
		slog.Debug("NPC returning home",
			"npc", ai.monster.Name(),
			"objectID", ai.monster.ObjectID())
	}
}

// getAttackRange returns the physical attack range for this NPC.
func (ai *AttackableAI) getAttackRange() int32 {
	npcDef := data.GetNpcDef(ai.monster.TemplateID())
	if npcDef != nil && npcDef.AtkRange() > 0 {
		return npcDef.AtkRange()
	}
	return attackRangeBase
}

// isTargetDead checks if the target WorldObject's creature is dead.
func isTargetDead(obj *model.WorldObject) bool {
	if player, ok := obj.Data.(*model.Player); ok {
		return player.IsDead()
	}
	return false
}
