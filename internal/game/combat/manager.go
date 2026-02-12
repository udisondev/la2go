package combat

import (
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/data"
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

// HitResult содержит результат одной атаки для наблюдения в тестах.
type HitResult struct {
	AttackerID uint32
	TargetID   uint32
	Damage     int32
	Miss       bool
	Crit       bool
}

// CombatManager координирует боевые действия (attacks, damage, death).
// Prevents import cycles: model → combat → gameserver.
//
// Phase 5.3: Basic Combat System.
// Phase 5.7: Added NPC attack support.
// Phase 5.10: Added rates for drop system.
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

	// rates holds server drop rate multipliers.
	// Phase 5.10: DROP/LOOT System.
	rates *config.Rates

	// hitObserver — callback для наблюдения за результатами атак (nil в production).
	hitObserver func(HitResult)
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

// SetRates sets server drop rate multipliers.
// Phase 5.10: DROP/LOOT System.
func (m *CombatManager) SetRates(rates *config.Rates) {
	m.rates = rates
}

// SetHitObserver sets callback for observing attack results (for tests).
func (m *CombatManager) SetHitObserver(fn func(HitResult)) {
	m.hitObserver = fn
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

	// Notify observer (synchronously, before timer)
	if m.hitObserver != nil {
		m.hitObserver(HitResult{
			AttackerID: attacker.ObjectID(),
			TargetID:   target.ObjectID(),
			Damage:     damage,
			Miss:       miss,
			Crit:       crit,
		})
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

		// Check death — DoDie returns true only for the first caller (race-safe)
		if target.IsDead() && target.DoDie(attacker) {
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

// dropLoot handles item drops when NPC dies.
// Phase 5.10: Uses real NPC drop tables from data package.
//
// Workflow:
//  1. CalculateDrops() rolls chances per NPC template
//  2. For each dropped item: create Item + DroppedItem, add to world
//  3. Broadcast ItemOnGround to nearby players
//  4. Schedule auto-destroy timer
func (m *CombatManager) dropLoot(npc *model.Npc, killer *model.Player) {
	drops := CalculateDrops(npc.TemplateID(), m.rates)
	if len(drops) == 0 {
		return
	}

	npcLoc := npc.Location()
	worldInst := world.Instance()

	autoDestroyTime := 60 * time.Second
	if m.rates != nil && m.rates.ItemAutoDestroyTime > 0 {
		autoDestroyTime = time.Duration(m.rates.ItemAutoDestroyTime) * time.Second
	}

	for _, drop := range drops {
		m.spawnDroppedItem(drop, npc, killer, npcLoc, worldInst, autoDestroyTime)
	}
}

// spawnDroppedItem creates a single dropped item in world and broadcasts it.
// Phase 5.10: DROP/LOOT System.
func (m *CombatManager) spawnDroppedItem(
	drop DropResult,
	npc *model.Npc,
	killer *model.Player,
	npcLoc model.Location,
	worldInst *world.World,
	autoDestroyTime time.Duration,
) {
	// Look up item template from data
	itemDef := data.GetItemDef(drop.ItemID)
	var itemTemplate *model.ItemTemplate
	if itemDef != nil {
		itemTemplate = &model.ItemTemplate{
			ItemID:      itemDef.ID(),
			Name:        itemDef.Name(),
			Type:        itemTypeFromString(itemDef.Type()),
			PAtk:        itemDef.PAtk(),
			AttackRange: itemDef.AttackRange(),
			PDef:        itemDef.PDef(),
			Weight:      itemDef.Weight(),
			Stackable:   itemDef.IsStackable(),
			Tradeable:   itemDef.IsTradeable(),
		}
	} else {
		// Fallback template for unknown items
		itemTemplate = &model.ItemTemplate{
			ItemID:    drop.ItemID,
			Name:      "Unknown Item",
			Type:      model.ItemTypeEtcItem,
			Stackable: true,
		}
		slog.Warn("item template not found, using fallback",
			"itemID", drop.ItemID,
			"npc", npc.Name())
	}

	// Generate unique objectID in item range
	droppedObjectID := world.IDGenerator().NextItemID()

	item, err := model.NewItem(droppedObjectID, drop.ItemID, 0, drop.Count, itemTemplate)
	if err != nil {
		slog.Error("create loot item",
			"npc", npc.Name(),
			"itemID", drop.ItemID,
			"error", err)
		return
	}

	// Randomize position around NPC (±70 game units)
	const randomRange = 70
	offsetX := int32(rand.IntN(randomRange*2+1) - randomRange)
	offsetY := int32(rand.IntN(randomRange*2+1) - randomRange)
	dropLoc := model.NewLocation(npcLoc.X+offsetX, npcLoc.Y+offsetY, npcLoc.Z, 0)

	droppedItem := model.NewDroppedItem(droppedObjectID, item, dropLoc, npc.ObjectID())

	// Add to world
	if err := worldInst.AddItem(droppedItem); err != nil {
		slog.Error("add dropped item to world",
			"npc", npc.Name(),
			"itemID", drop.ItemID,
			"error", err)
		return
	}

	// Broadcast ItemOnGround packet
	itemOnGround := serverpackets.NewItemOnGround(droppedItem)
	pktData, err := itemOnGround.Write()
	if err != nil {
		slog.Error("write ItemOnGround packet",
			"itemID", drop.ItemID,
			"error", err)
		return
	}

	// Broadcast from NPC position (not killer's)
	if m.npcBroadcastFunc != nil {
		m.npcBroadcastFunc(npcLoc.X, npcLoc.Y, pktData, len(pktData))
	}

	// Auto-destroy timer
	time.AfterFunc(autoDestroyTime, func() {
		worldInst.RemoveObject(droppedObjectID)

		deleteObj := serverpackets.NewDeleteObject(int32(droppedObjectID))
		deleteData, err := deleteObj.Write()
		if err != nil {
			slog.Error("write DeleteObject for loot despawn",
				"objectID", droppedObjectID,
				"error", err)
			return
		}

		if m.npcBroadcastFunc != nil {
			m.npcBroadcastFunc(dropLoc.X, dropLoc.Y, deleteData, len(deleteData))
		}

		slog.Debug("loot despawned",
			"objectID", droppedObjectID,
			"itemID", drop.ItemID)
	})

	slog.Info("loot dropped",
		"npc", npc.Name(),
		"item", itemTemplate.Name,
		"count", drop.Count,
		"location", dropLoc,
		"objectID", droppedObjectID)
}

// itemTypeFromString converts item type string from data package to model.ItemType.
func itemTypeFromString(s string) model.ItemType {
	switch s {
	case "Weapon":
		return model.ItemTypeWeapon
	case "Armor":
		return model.ItemTypeArmor
	default:
		return model.ItemTypeEtcItem
	}
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

	// Notify observer (synchronously, before timer)
	if m.hitObserver != nil {
		m.hitObserver(HitResult{
			AttackerID: npc.ObjectID(),
			TargetID:   target.ObjectID(),
			Damage:     damage,
			Miss:       miss,
			Crit:       crit,
		})
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

		// Check death — DoDie returns true only for the first caller (race-safe)
		if target.IsDead() && target.DoDie(nil) {
			slog.Info("player killed by NPC",
				"victim", target.Name(),
				"killer", npc.Name())
		}
	}
}

// CombatMgr — global CombatManager instance.
// Initialized by cmd/gameserver/main.go with broadcast function.
// NOT safe for concurrent test assignment — tests that set this must NOT use t.Parallel().
//
// Phase 5.3: Basic Combat System.
var CombatMgr *CombatManager
