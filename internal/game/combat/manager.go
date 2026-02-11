package combat

import (
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// AIManagerInterface provides access to AI controllers without importing ai package.
// Phase 5.7: NPC Aggro & Basic AI.
type AIManagerInterface interface {
	GetController(objectID uint32) (AIController, error)
}

// AIController is a subset of ai.Controller used by CombatManager.
// Phase 5.7: NPC Aggro & Basic AI.
type AIController interface {
	NotifyDamage(attackerID uint32, damage int32)
}

// CombatManager координирует боевые действия (attacks, damage, death).
// Prevents import cycles: model → combat → gameserver.
//
// Phase 5.3: Basic Combat System.
// Phase 5.7: Added NPC attack support.
type CombatManager struct {
	// broadcastFunc is injected to avoid import cycle with gameserver
	broadcastFunc func(source *model.Player, data []byte, size int)

	// npcBroadcastFunc broadcasts from NPC position (no Player source needed).
	// Phase 5.7: NPC Aggro & Basic AI.
	npcBroadcastFunc func(x, y int32, data []byte, size int)

	// aiManager provides access to AI controllers for NotifyDamage callbacks.
	// Phase 5.7: NPC Aggro & Basic AI.
	aiManager AIManagerInterface

	// npcDeathFunc is called when NPC dies — triggers despawn + respawn.
	npcDeathFunc func(npc *model.Npc)

	// rewardFunc is called when NPC dies to give XP/SP to killer.
	// Phase 5.8: Experience & Leveling System.
	rewardFunc func(killer *model.Player, npc *model.Npc)
}

// SetNpcDeathFunc sets the callback for NPC death handling (despawn + respawn).
func (m *CombatManager) SetNpcDeathFunc(fn func(npc *model.Npc)) {
	m.npcDeathFunc = fn
}

// SetRewardFunc sets the callback for XP/SP reward on NPC kill.
// Phase 5.8: Experience & Leveling System.
func (m *CombatManager) SetRewardFunc(fn func(killer *model.Player, npc *model.Npc)) {
	m.rewardFunc = fn
}

// NewCombatManager creates new CombatManager.
// broadcastFunc должен вызывать ClientManager.BroadcastToVisibleNear().
//
// Phase 5.3: Basic Combat System.
// Phase 5.7: Added npcBroadcastFunc and aiManager.
func NewCombatManager(
	broadcastFunc func(*model.Player, []byte, int),
	npcBroadcastFunc func(int32, int32, []byte, int),
	aiManager AIManagerInterface,
) *CombatManager {
	return &CombatManager{
		broadcastFunc:    broadcastFunc,
		npcBroadcastFunc: npcBroadcastFunc,
		aiManager:        aiManager,
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
	} else if targetMonster, ok := target.Data.(*model.Monster); ok {
		// PvE: Monster target (Phase 5.7: Monster overrides Data)
		targetCharacter = targetMonster.Character
		targetPDef = targetMonster.PDef()
	} else if targetNpc, ok := target.Data.(*model.Npc); ok {
		// PvE: NPC target (non-aggressive NPC)
		targetCharacter = targetNpc.Character
		targetPDef = targetNpc.PDef()
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

		// Phase 5.7: Notify AI if target is NPC (for hate list / retaliation)
		if m.aiManager != nil {
			if ctrl, err := m.aiManager.GetController(target.ObjectID()); err == nil {
				ctrl.NotifyDamage(attacker.ObjectID(), damage)
			}
		}

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

			// Phase 5.7: Drop loot and trigger despawn/respawn if target is NPC/Monster
			// Phase 5.8: Reward XP/SP before loot/despawn
			if monster, ok := target.WorldObject.Data.(*model.Monster); ok {
				if m.rewardFunc != nil {
					m.rewardFunc(attacker, monster.Npc)
				}
				m.dropLoot(monster.Npc, attacker)
				if m.npcDeathFunc != nil {
					m.npcDeathFunc(monster.Npc)
				}
			} else if npc, ok := target.WorldObject.Data.(*model.Npc); ok {
				if m.rewardFunc != nil {
					m.rewardFunc(attacker, npc)
				}
				m.dropLoot(npc, attacker)
				if m.npcDeathFunc != nil {
					m.npcDeathFunc(npc)
				}
			}

			slog.Info("target died",
				"victim", target.Name(),
				"killer", attacker.Name())
		}
	}
}

// dropLoot handles item drop when NPC dies.
// Phase 5.7: Loot System MVP.
//
// MVP: Drops fixed amount of Adena (game currency).
// TODO Phase 5.8: Loot tables, random drops, equipment drops.
func (m *CombatManager) dropLoot(npc *model.Npc, killer *model.Player) {
	// MVP: Drop Adena (itemID=57, game currency)
	// Amount based on NPC level
	adenaAmount := int32(npc.Level() * 10) // Level 5 → 50 Adena

	// Create Adena item
	adenaTemplate := &model.ItemTemplate{
		ItemID:    57,
		Name:      "Adena",
		Type:      model.ItemTypeConsumable,
		Stackable: true,
		Tradeable: true,
	}

	// Generate unique objectID for dropped item
	// TODO Phase 5.8: Use proper ObjectIDGenerator
	droppedObjectID := uint32(0x00000001 + npc.ObjectID()%0x0FFFFFFF)

	adenaItem, err := model.NewItem(droppedObjectID, 57, 0, adenaAmount, adenaTemplate)
	if err != nil {
		slog.Error("failed to create loot item",
			"npc", npc.Name(),
			"error", err)
		return
	}

	// Create DroppedItem at NPC location
	npcLoc := npc.Location()
	droppedItem := model.NewDroppedItem(droppedObjectID, adenaItem, npcLoc, 0)

	// Add to world
	worldInst := world.Instance()
	if err := worldInst.AddObject(droppedItem.WorldObject); err != nil {
		slog.Error("failed to add dropped item to world",
			"npc", npc.Name(),
			"error", err)
		return
	}

	// Broadcast ItemOnGround packet
	itemOnGround := serverpackets.NewItemOnGround(droppedItem)
	itemData, err := itemOnGround.Write()
	if err != nil {
		slog.Error("failed to write ItemOnGround packet",
			"npc", npc.Name(),
			"error", err)
		return
	}

	// Broadcast to visible players (LOD optimization)
	// Use killer as broadcast source for visibility
	if m.broadcastFunc != nil {
		m.broadcastFunc(killer, itemData, len(itemData))
	}

	// Loot despawn timer: remove dropped item after 60 seconds
	time.AfterFunc(60*time.Second, func() {
		worldInst.RemoveObject(droppedObjectID)

		// Broadcast DeleteObject to nearby players
		deleteObj := serverpackets.NewDeleteObject(int32(droppedObjectID))
		deleteData, err := deleteObj.Write()
		if err != nil {
			slog.Error("failed to write DeleteObject for loot despawn",
				"objectID", droppedObjectID,
				"error", err)
			return
		}

		if m.npcBroadcastFunc != nil {
			m.npcBroadcastFunc(npcLoc.X, npcLoc.Y, deleteData, len(deleteData))
		}

		slog.Debug("loot despawned",
			"objectID", droppedObjectID,
			"item", adenaTemplate.Name)
	})

	slog.Info("loot dropped",
		"npc", npc.Name(),
		"item", adenaTemplate.Name,
		"amount", adenaAmount,
		"location", npcLoc,
		"objectID", droppedObjectID)
}

// ExecuteNpcAttack executes physical attack from NPC to player target.
// Phase 5.7: NPC Aggro & Basic AI.
func (m *CombatManager) ExecuteNpcAttack(npc *model.Npc, target *model.WorldObject) {
	// Extract target Player
	targetPlayer, ok := target.Data.(*model.Player)
	if !ok {
		slog.Warn("ExecuteNpcAttack: target is not a Player",
			"npc", npc.Name(),
			"targetID", target.ObjectID())
		return
	}

	// Validate target alive
	if targetPlayer.IsDead() {
		return
	}

	// Calculate miss/crit/damage
	miss := CalcHitMissGeneric()
	crit := false
	damage := int32(0)

	if !miss {
		crit = CalcCritGeneric()

		pAtk := float64(npc.PAtk())
		pDef := float64(targetPlayer.GetPDef())
		if pDef < 1 {
			pDef = 1
		}

		damageFloat := (76.0 * pAtk) / pDef
		damageFloat *= getRandomDamageMultiplier(npc.Level())
		if crit {
			damageFloat *= 2.0
		}
		if damageFloat < 1 {
			damageFloat = 1
		}
		damage = int32(damageFloat)
	}

	// Create Attack packet (NPC attacker)
	npcLoc := npc.Location()
	attack := serverpackets.NewNpcAttack(npc.ObjectID(), npcLoc, target)
	attack.AddHit(target.ObjectID(), damage, miss, crit)

	attackData, err := attack.Write()
	if err != nil {
		slog.Error("failed to write NPC Attack packet",
			"npc", npc.Name(),
			"target", target.Name(),
			"error", err)
		return
	}

	// Broadcast from NPC position
	if m.npcBroadcastFunc != nil {
		m.npcBroadcastFunc(npcLoc.X, npcLoc.Y, attackData, len(attackData))
	}

	// Schedule damage application with NPC attack speed delay
	atkSpeed := npc.AtkSpeed()
	if atkSpeed < 1 {
		atkSpeed = 253 // default NPC attack speed
	}
	attackDelay := time.Duration(500000/atkSpeed) * time.Millisecond

	time.AfterFunc(attackDelay, func() {
		m.onNpcHitTimer(npc, targetPlayer, damage, miss)
	})

	slog.Debug("NPC attack executed",
		"npc", npc.Name(),
		"target", targetPlayer.Name(),
		"damage", damage,
		"miss", miss,
		"crit", crit)
}

// onNpcHitTimer handles delayed damage from NPC attack.
// Phase 5.7: NPC Aggro & Basic AI.
func (m *CombatManager) onNpcHitTimer(npc *model.Npc, target *model.Player, damage int32, miss bool) {
	if target.IsDead() {
		return
	}

	if !miss && damage > 0 {
		target.ReduceCurrentHP(damage)

		// Broadcast StatusUpdate
		statusUpdate := serverpackets.NewStatusUpdateForTarget(target.Character)
		statusData, err := statusUpdate.Write()
		if err != nil {
			slog.Error("failed to write StatusUpdate for NPC hit",
				"npc", npc.Name(),
				"target", target.Name(),
				"error", err)
			return
		}

		// Broadcast from NPC position
		npcLoc := npc.Location()
		if m.npcBroadcastFunc != nil {
			m.npcBroadcastFunc(npcLoc.X, npcLoc.Y, statusData, len(statusData))
		}

		// Check death
		if target.IsDead() {
			target.DoDie(nil) // NPC kill — no Player killer

			slog.Info("player killed by NPC",
				"victim", target.Name(),
				"killer", npc.Name())
		}
	}
}

// Global CombatManager instance.
// Initialized by cmd/gameserver/main.go with broadcast function.
//
// Phase 5.3: Basic Combat System.
var CombatMgr *CombatManager
