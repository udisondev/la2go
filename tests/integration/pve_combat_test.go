package integration

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestPvECombat_PlayerVsNPC verifies Player can attack NPC.
// Phase 5.6: PvE Combat System.
func TestPvECombat_PlayerVsNPC(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	// Load templates
	data.InitStatBonuses()
	if err := data.LoadPlayerTemplates(); err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	// Setup combat managers
	clientMgr := gameserver.NewClientManager()

	attackStanceMgr := combat.NewAttackStanceManager()
	combat.AttackStanceMgr = attackStanceMgr
	attackStanceMgr.Start()
	defer attackStanceMgr.Stop()

	broadcastFunc := func(source *model.Player, data []byte, size int) {
		clientMgr.BroadcastToVisibleNear(source, data, size)
	}
	combatMgr := combat.NewCombatManager(broadcastFunc)
	combat.CombatMgr = combatMgr

	// Get world instance
	worldInst := world.Instance()

	// Create Player level 10 Human Fighter
	player, err := model.NewPlayer(1, 100, 200, "Hunter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Equip weapon (for faster kill)
	swordTemplate := &model.ItemTemplate{
		ItemID:      1,
		Name:        "Sword",
		Type:        model.ItemTypeWeapon,
		PAtk:        10,
		AttackRange: 40,
	}
	sword, _ := model.NewItem(1000, 1, 100, 1, swordTemplate)
	player.Inventory().AddItem(sword)
	player.Inventory().EquipItem(sword, model.PaperdollRHand)

	// Create NPC (Wolf level 5)
	npcTemplate := model.NewNpcTemplate(
		2000,    // templateID
		"Wolf",  // name
		"",      // title
		5,       // level
		150,     // maxHP (low for quick kill)
		100,     // maxMP
		15,      // pAtk
		5,       // mAtk
		50,      // pDef
		30,      // mDef
		0,       // race (0 = unknown)
		120,     // moveSpeed
		253,     // atkSpeed
		30,      // respawnMin
		60,      // respawnMax
	)

	npc := model.NewNpc(2, 2000, npcTemplate)
	npc.SetLocation(model.NewLocation(50, 0, 0, 0))

	// Add to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	if err := worldInst.AddObject(npc.WorldObject); err != nil {
		t.Fatalf("AddObject npc failed: %v", err)
	}
	defer worldInst.RemoveObject(npc.ObjectID())

	// Attack NPC
	npcHPBefore := npc.CurrentHP()
	combatMgr.ExecuteAttack(player, npc.WorldObject)
	time.Sleep(2 * time.Second) // Wait for attack delay

	npcHPAfter := npc.CurrentHP()
	damage := npcHPBefore - npcHPAfter

	t.Logf("Attack NPC: damage=%d (HP: %d → %d)", damage, npcHPBefore, npcHPAfter)

	// NPC should take damage
	if damage <= 0 {
		t.Errorf("NPC should take damage: damage=%d", damage)
	}

	// Attack until NPC dies
	attackCount := 0
	maxAttacks := 20 // Safety limit
	for !npc.IsDead() && attackCount < maxAttacks {
		combatMgr.ExecuteAttack(player, npc.WorldObject)
		time.Sleep(2 * time.Second)
		attackCount++
	}

	t.Logf("NPC killed after %d attacks", attackCount)

	// NPC should be dead
	if !npc.IsDead() {
		t.Errorf("NPC should be dead after %d attacks (HP: %d)", attackCount, npc.CurrentHP())
	}

	if npc.CurrentHP() != 0 {
		t.Errorf("NPC HP should be 0, got %d", npc.CurrentHP())
	}

	t.Log("PvE combat integration test passed!")
}
