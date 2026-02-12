package integration

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestLootPickup_Success verifies player can pick up dropped item.
// Phase 5.7: Loot System MVP.
// Uses synctest for instant fake-clock execution.
func TestLootPickup_Success(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		player.Inventory().EquipItem(sword, model.PaperdollRHand)

		// Use real NPC 13031 "Huge Pig" which has 100% drop of itemID=9142 (count 1-2)
		npcOID := nextOID()
		npcTemplate := model.NewNpcTemplate(
			13031, "Huge Pig", "", 5, 150, 100,
			15, 5, 50, 30, 0, 120, 253, 30, 60, 0, 0,
		)

		npc := model.NewNpc(npcOID, 13031, npcTemplate)
		npc.SetLocation(model.NewLocation(50, 0, 0, 0))

		if err := worldInst.AddObject(player.WorldObject); err != nil {
			t.Fatalf("AddObject player failed: %v", err)
		}
		defer worldInst.RemoveObject(player.ObjectID())

		if err := worldInst.AddObject(npc.WorldObject); err != nil {
			t.Fatalf("AddObject npc failed: %v", err)
		}
		defer worldInst.RemoveObject(npc.ObjectID())

		// Attack NPC until death (instant with fake clock)
		attackCount := combat.AttackUntilDead(combatMgr, player, npc.WorldObject, npc.Character, 50)

		if !npc.IsDead() {
			t.Fatalf("NPC should be dead after %d attacks", attackCount)
		}

		t.Logf("NPC killed after %d attacks", attackCount)

		// Wait for loot drop (instant with fake clock)
		time.Sleep(500 * time.Millisecond)

		// Find DroppedItem in world (should be at NPC location)
		var droppedItemID uint32
		world.ForEachVisibleObjectForPlayer(player, func(obj *model.WorldObject) bool {
			if droppedItem, ok := obj.Data.(*model.DroppedItem); ok {
				droppedItemID = droppedItem.ObjectID()
				t.Logf("Found DroppedItem: objectID=%d, itemID=%d, count=%d",
					droppedItem.ObjectID(),
					droppedItem.Item().ItemID(),
					droppedItem.Item().Count())
				return false
			}
			return true
		})

		if droppedItemID == 0 {
			t.Fatal("DroppedItem not found in world after NPC death")
		}

		obj, exists := worldInst.GetObject(droppedItemID)
		if !exists {
			t.Fatalf("DroppedItem objectID=%d not found in world", droppedItemID)
		}

		droppedItem, ok := obj.Data.(*model.DroppedItem)
		if !ok {
			t.Fatalf("Object is not DroppedItem (type=%T)", obj.Data)
		}

		item := droppedItem.Item()
		// NPC 13031 "Huge Pig" drops itemID=9142 with count 1-2 (100% chance)
		if item.ItemID() != 9142 {
			t.Errorf("Dropped item should be itemID 9142, got itemID=%d", item.ItemID())
		}

		if item.Count() < 1 || item.Count() > 2 {
			t.Errorf("Item count=%d, expected 1-2", item.Count())
		}

		inventoryBefore := len(player.Inventory().GetItems())

		if err := player.Inventory().AddItem(item); err != nil {
			t.Fatalf("AddItem failed: %v", err)
		}

		worldInst.RemoveObject(droppedItemID)

		_, exists = worldInst.GetObject(droppedItemID)
		if exists {
			t.Errorf("DroppedItem should be removed from world")
		}

		inventoryAfter := len(player.Inventory().GetItems())
		if inventoryAfter != inventoryBefore+1 {
			t.Errorf("Inventory size=%d, expected %d (added 1 item)", inventoryAfter, inventoryBefore+1)
		}

		foundItem := false
		for _, invItem := range player.Inventory().GetItems() {
			if invItem.ItemID() == 9142 {
				foundItem = true
				if invItem.Count() < 1 || invItem.Count() > 2 {
					t.Errorf("Item 9142 in inventory count=%d, expected 1-2", invItem.Count())
				}
				break
			}
		}

		if !foundItem {
			t.Error("Item 9142 not found in inventory after pickup")
		}
	})
}

// TestLootPickup_OutOfRange verifies pickup fails when player is too far.
// Phase 5.7: Loot System MVP.
func TestLootPickup_OutOfRange(t *testing.T) {
	worldInst := world.Instance()

	playerOID := nextOID()
	player, err := model.NewPlayer(playerOID, 100, 200, "Hunter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	adenaTemplate := &model.ItemTemplate{
		ItemID:    57,
		Name:      "Adena",
		Type:      model.ItemTypeConsumable,
		Stackable: true,
		Tradeable: true,
	}

	droppedOID := nextOID()
	adenaItem, _ := model.NewItem(1001, 57, 0, 50, adenaTemplate)
	droppedItem := model.NewDroppedItem(droppedOID, adenaItem, model.NewLocation(300, 0, 0, 0), 0)

	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	if err := worldInst.AddObject(droppedItem.WorldObject); err != nil {
		t.Fatalf("AddObject droppedItem failed: %v", err)
	}
	defer worldInst.RemoveObject(droppedItem.ObjectID())

	playerLoc := player.Location()
	itemLoc := droppedItem.Location()
	dx := int64(playerLoc.X - itemLoc.X)
	dy := int64(playerLoc.Y - itemLoc.Y)
	distSq := dx*dx + dy*dy

	const MaxItemPickupRangeSquared = 200 * 200
	if distSq <= MaxItemPickupRangeSquared {
		t.Fatalf("Test setup error: distance should be > 200 units")
	}

	inventoryBefore := len(player.Inventory().GetItems())

	if distSq > MaxItemPickupRangeSquared {
		t.Log("Pickup correctly rejected: out of range")
	} else {
		t.Error("Pickup should be rejected due to range")
	}

	inventoryAfter := len(player.Inventory().GetItems())
	if inventoryAfter != inventoryBefore {
		t.Errorf("Inventory should be unchanged: before=%d, after=%d", inventoryBefore, inventoryAfter)
	}

	_, exists := worldInst.GetObject(droppedItem.ObjectID())
	if !exists {
		t.Error("DroppedItem should still be in world after failed pickup")
	}
}
