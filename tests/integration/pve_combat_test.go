package integration

import (
	"testing"
	"testing/synctest"

	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestPvECombat_PlayerVsNPC verifies Player can attack NPC.
// Phase 5.6: PvE Combat System.
// Uses synctest for instant fake-clock execution (was 40s+ real time).
func TestPvECombat_PlayerVsNPC(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Combat managers inside bubble (goroutines use fake clock)
		clientMgr := gameserver.NewClientManager()

		attackStanceMgr := combat.NewAttackStanceManager(nil)
		combat.AttackStanceMgr = attackStanceMgr
		attackStanceMgr.Start()
		defer attackStanceMgr.Stop()

		broadcastFunc := func(source *model.Player, data []byte, size int) {
			clientMgr.BroadcastToVisibleNear(source, data, size)
		}
		combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
		combat.CombatMgr = combatMgr

		worldInst := world.Instance()

		playerOID := nextOID()
		player, err := model.NewPlayer(playerOID, 100, 200, "Hunter", 10, 0, 0)
		if err != nil {
			t.Fatalf("NewPlayer failed: %v", err)
		}
		player.SetLocation(model.NewLocation(0, 0, 0, 0))

		swordTemplate := &model.ItemTemplate{
			ItemID:      1,
			Name:        "Sword",
			Type:        model.ItemTypeWeapon,
			PAtk:        10,
			AttackRange: 40,
		}
		sword, _ := model.NewItem(1000, 1, 100, 1, swordTemplate)
		player.Inventory().AddItem(sword)
		_, _ = player.Inventory().EquipItem(sword, model.PaperdollRHand)

		npcOID := nextOID()
		npcTemplate := model.NewNpcTemplate(
			2000, "Wolf", "", 5, 150, 100,
			15, 5, 50, 30, 0, 120, 253, 30, 60, 0, 0,
		)

		npc := model.NewNpc(npcOID, 2000, npcTemplate)
		npc.SetLocation(model.NewLocation(50, 0, 0, 0))

		if err := worldInst.AddObject(player.WorldObject); err != nil {
			t.Fatalf("AddObject player failed: %v", err)
		}
		defer worldInst.RemoveObject(player.ObjectID())

		if err := worldInst.AddObject(npc.WorldObject); err != nil {
			t.Fatalf("AddObject npc failed: %v", err)
		}
		defer worldInst.RemoveObject(npc.ObjectID())

		// Attack NPC until first hit
		result, attempts := combat.AttackUntilHit(combatMgr, player, npc.WorldObject, 20)
		if result.Miss {
			t.Fatalf("could not land a hit in %d attempts", attempts)
		}

		t.Logf("First hit: damage=%d (attempts=%d)", result.Damage, attempts)

		if result.Damage <= 0 {
			t.Errorf("NPC should take damage: damage=%d", result.Damage)
		}

		// Attack until NPC dies
		attackCount := combat.AttackUntilDead(combatMgr, player, npc.WorldObject, npc.Character, 50)

		t.Logf("NPC killed after %d total attacks", attackCount)

		if !npc.IsDead() {
			t.Errorf("NPC should be dead after %d attacks (HP: %d)", attackCount, npc.CurrentHP())
		}

		if npc.CurrentHP() != 0 {
			t.Errorf("NPC HP should be 0, got %d", npc.CurrentHP())
		}
	})
}
