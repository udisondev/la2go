package combat

import (
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// CombatManager координирует боевые действия (attacks, damage, death).
// Prevents import cycles: model → combat → gameserver.
//
// Phase 5.3: Basic Combat System.
type CombatManager struct {
	// broadcastFunc is injected to avoid import cycle with gameserver
	broadcastFunc func(source *model.Player, data []byte, size int)
}

// NewCombatManager creates new CombatManager.
// broadcastFunc должен вызывать ClientManager.BroadcastToVisibleNear().
//
// Phase 5.3: Basic Combat System.
func NewCombatManager(broadcastFunc func(*model.Player, []byte, int)) *CombatManager {
	return &CombatManager{
		broadcastFunc: broadcastFunc,
	}
}

// ExecuteAttack выполняет физическую атаку attacker → target.
//
// Phase 5.6: PvP + PvE implementation (Player vs Player/NPC).
//
// Workflow:
//  1. Extract target stats via type assertion (Player.GetPDef or Npc.PDef)
//  2. Calculate miss/crit/damage using real stats
//  3. Broadcast Attack packet
//  4. Add to AttackStanceManager
//  5. Schedule damage application via onHitTimer
//
// Phase 5.6: PvE Combat integration.
func (m *CombatManager) ExecuteAttack(attacker *model.Player, target *model.WorldObject) {
	// Phase 5.6: Extract target stats based on type (Player or Npc)
	var targetCharacter *model.Character
	var targetPDef int32

	// Try Player first (PvP)
	if targetPlayer, ok := target.Data.(*model.Player); ok {
		targetCharacter = targetPlayer.Character
		targetPDef = targetPlayer.GetPDef() // Phase 5.5: includes armor
	} else if targetNpc, ok := target.Data.(*model.Npc); ok {
		// PvE: NPC target
		targetCharacter = targetNpc.Character
		targetPDef = targetNpc.PDef() // From NPC template
	} else {
		// Unknown target type - skip
		slog.Warn("ExecuteAttack: unknown target type",
			"attacker", attacker.Name(),
			"targetID", target.ObjectID())
		return
	}

	// Calculate miss/crit/damage
	miss := CalcHitMiss(attacker, targetCharacter)
	crit := false
	damage := int32(0)

	if !miss {
		crit = CalcCrit(attacker, targetCharacter)

		// Phase 5.5: Use GetPAtk() (includes weapon bonus)
		pAtk := float64(attacker.GetPAtk())
		pDef := float64(targetPDef)

		// Simplified damage formula (76 × pAtk) / pDef × random × crit
		damageFloat := (76.0 * pAtk) / pDef
		damageFloat *= getRandomDamageMultiplier(attacker.Level())
		if crit {
			damageFloat *= 2.0
		}
		if damageFloat < 1 {
			damageFloat = 1
		}
		damage = int32(damageFloat)
	}

	// Create Attack packet
	attack := serverpackets.NewAttack(attacker, target)
	attack.AddHit(target.ObjectID(), damage, miss, crit)

	// Broadcast Attack packet immediately (LOD optimization)
	attackData, err := attack.Write()
	if err != nil {
		slog.Error("failed to write Attack packet",
			"attacker", attacker.Name(),
			"target", target.ObjectID(),
			"error", err)
		return
	}

	if m.broadcastFunc != nil {
		m.broadcastFunc(attacker, attackData, len(attackData))
	}

	// Add to combat stance (15-second window)
	if AttackStanceMgr != nil {
		AttackStanceMgr.AddAttackStance(attacker)
	}

	// Schedule damage application (delayed by attack speed)
	attackDelay := attacker.GetAttackDelay()
	time.AfterFunc(attackDelay, func() {
		m.onHitTimer(attacker, targetCharacter, damage, crit, miss)
	})

	slog.Debug("attack executed",
		"attacker", attacker.Name(),
		"target", target.Name(),
		"damage", damage,
		"miss", miss,
		"crit", crit)
}

// onHitTimer handles delayed damage application.
// Called after attack speed delay (time.AfterFunc).
//
// Phase 5.3: Basic Combat System.
func (m *CombatManager) onHitTimer(attacker *model.Player, target *model.Character, damage int32, crit bool, miss bool) {
	// Validate target still alive
	if target.IsDead() {
		return
	}

	// Apply damage
	if !miss && damage > 0 {
		// Reduce HP
		target.ReduceCurrentHP(damage)

		// Broadcast StatusUpdate (HP changed)
		statusUpdate := serverpackets.NewStatusUpdateForTarget(target)
		statusData, err := statusUpdate.Write()
		if err != nil {
			slog.Error("failed to write StatusUpdate packet",
				"target", target.Name(),
				"error", err)
			return
		}

		// Broadcast to visible players (LOD optimization)
		// Note: For broadcast, we need source Player position
		// For MVP, use attacker as broadcast source
		if m.broadcastFunc != nil {
			m.broadcastFunc(attacker, statusData, len(statusData))
		}

		// Check death
		if target.IsDead() {
			target.DoDie(attacker)
			slog.Info("target died",
				"victim", target.Name(),
				"killer", attacker.Name())
		}
	}
}

// Global CombatManager instance.
// Initialized by cmd/gameserver/main.go with broadcast function.
//
// Phase 5.3: Basic Combat System.
var CombatMgr *CombatManager
