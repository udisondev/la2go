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
// MVP implementation с mock target stats.
//
// Workflow:
//  1. Create mock target Character (TODO Phase 5.4: extract real Character)
//  2. Calculate miss/crit/damage
//  3. Broadcast Attack packet
//  4. Add to AttackStanceManager
//  5. Schedule damage application via onHitTimer
//
// Phase 5.3: Basic Combat System.
func (m *CombatManager) ExecuteAttack(attacker *model.Player, target *model.WorldObject) {
	// MVP: Create mock target character for damage calculation
	// TODO Phase 5.4: Extract real Character from WorldObject
	mockTarget := &model.Character{
		WorldObject: target,
	}
	// Set mock stats for damage calculation
	// These would normally come from Character template
	mockTargetLevel := int32(10)
	mockTargetPDef := int32(80 + mockTargetLevel*3) // Formula from GetBasePDef

	// Calculate miss/crit/damage
	miss := CalcHitMiss(attacker, mockTarget)
	crit := false
	damage := int32(0)

	if !miss {
		crit = CalcCrit(attacker, mockTarget)
		// For damage calculation, we need pDef
		// Use mock formula since we don't have real Character
		pAtk := float64(attacker.GetBasePAtk())
		pDef := float64(mockTargetPDef)

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
		m.onHitTimer(attacker, mockTarget, damage, crit, miss)
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
