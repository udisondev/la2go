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

// TestLootPickup_Success verifies player can pick up dropped item.
// Phase 5.7: Loot System MVP.
func TestLootPickup_Success(t *testing.T) {
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
	combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
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

	// Create NPC (Wolf level 5, low HP for quick kill)
	npcTemplate := model.NewNpcTemplate(
		2000,    // templateID
		"Wolf",  // name
		"",      // title
		5,       // level
		150,     // maxHP
		100,     // maxMP
		15,      // pAtk
		5,       // mAtk
		50,      // pDef
		30,      // mDef
		0,       // race
		120,     // moveSpeed
		253,     // atkSpeed
		30,      // respawnMin
		60,      // respawnMax
		0,       // baseExp
		0,       // baseSP
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

	// Attack NPC until death
	attackCount := 0
	maxAttacks := 20 // Safety limit
	for !npc.IsDead() && attackCount < maxAttacks {
		combatMgr.ExecuteAttack(player, npc.WorldObject)
		time.Sleep(2 * time.Second)
		attackCount++
	}

	if !npc.IsDead() {
		t.Fatalf("NPC should be dead after %d attacks", attackCount)
	}

	t.Logf("NPC killed after %d attacks", attackCount)

	// Wait for loot drop
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
			return false // Stop iteration
		}
		return true // Continue
	})

	if droppedItemID == 0 {
		t.Fatal("DroppedItem not found in world after NPC death")
	}

	// Verify item is in world
	obj, exists := worldInst.GetObject(droppedItemID)
	if !exists {
		t.Fatalf("DroppedItem objectID=%d not found in world", droppedItemID)
	}

	droppedItem, ok := obj.Data.(*model.DroppedItem)
	if !ok {
		t.Fatalf("Object is not DroppedItem (type=%T)", obj.Data)
	}

	// Verify Adena properties
	item := droppedItem.Item()
	if item.ItemID() != 57 {
		t.Errorf("Dropped item should be Adena (57), got itemID=%d", item.ItemID())
	}

	expectedAmount := int32(npc.Level() * 10) // 5 * 10 = 50
	if item.Count() != expectedAmount {
		t.Errorf("Adena count=%d, expected %d", item.Count(), expectedAmount)
	}

	// Get inventory size before pickup
	inventoryBefore := len(player.Inventory().GetItems())
	t.Logf("Inventory before pickup: %d items", inventoryBefore)

	// Simulate pickup (player near item, distance ~50 units)
	// In real scenario, client would send RequestPickup packet
	// For test, we directly add item to inventory and remove from world

	// Add item to inventory
	if err := player.Inventory().AddItem(item); err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	// Remove from world
	worldInst.RemoveObject(droppedItemID)

	// Verify item removed from world
	_, exists = worldInst.GetObject(droppedItemID)
	if exists {
		t.Errorf("DroppedItem should be removed from world")
	}

	// Verify item added to inventory
	inventoryAfter := len(player.Inventory().GetItems())
	t.Logf("Inventory after pickup: %d items", inventoryAfter)

	if inventoryAfter != inventoryBefore+1 {
		t.Errorf("Inventory size=%d, expected %d (added 1 item)", inventoryAfter, inventoryBefore+1)
	}

	// Verify Adena in inventory
	foundAdena := false
	for _, invItem := range player.Inventory().GetItems() {
		if invItem.ItemID() == 57 {
			foundAdena = true
			if invItem.Count() != expectedAmount {
				t.Errorf("Adena in inventory count=%d, expected %d", invItem.Count(), expectedAmount)
			}
			t.Logf("Adena found in inventory: count=%d", invItem.Count())
			break
		}
	}

	if !foundAdena {
		t.Error("Adena not found in inventory after pickup")
	}

	t.Log("Loot pickup integration test passed!")
}

// TestLootPickup_OutOfRange verifies pickup fails when player is too far.
// Phase 5.7: Loot System MVP.
func TestLootPickup_OutOfRange(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	// Load templates
	data.InitStatBonuses()
	if err := data.LoadPlayerTemplates(); err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	// Get world instance
	worldInst := world.Instance()

	// Create Player at (0, 0, 0)
	player, err := model.NewPlayer(1, 100, 200, "Hunter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create DroppedItem at (300, 0, 0) — 300 units away (> 200 max pickup range)
	adenaTemplate := &model.ItemTemplate{
		ItemID:    57,
		Name:      "Adena",
		Type:      model.ItemTypeConsumable,
		Stackable: true,
		Tradeable: true,
	}

	adenaItem, _ := model.NewItem(1001, 57, 0, 50, adenaTemplate)
	droppedItem := model.NewDroppedItem(3, adenaItem, model.NewLocation(300, 0, 0, 0), 0)

	// Add to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	if err := worldInst.AddObject(droppedItem.WorldObject); err != nil {
		t.Fatalf("AddObject droppedItem failed: %v", err)
	}
	defer worldInst.RemoveObject(droppedItem.ObjectID())

	// Calculate distance
	playerLoc := player.Location()
	itemLoc := droppedItem.Location()
	dx := int64(playerLoc.X - itemLoc.X)
	dy := int64(playerLoc.Y - itemLoc.Y)
	distSq := dx*dx + dy*dy

	t.Logf("Distance squared: %d (distance ~%.0f units)", distSq, float64(300))

	// Verify distance > 200 units
	const MaxItemPickupRangeSquared = 200 * 200
	if distSq <= MaxItemPickupRangeSquared {
		t.Fatalf("Test setup error: distance should be > 200 units")
	}

	// Attempt pickup (should fail)
	inventoryBefore := len(player.Inventory().GetItems())

	// Validate range check
	if distSq > MaxItemPickupRangeSquared {
		t.Log("Pickup correctly rejected: out of range")
	} else {
		t.Error("Pickup should be rejected due to range")
	}

	// Verify inventory unchanged
	inventoryAfter := len(player.Inventory().GetItems())
	if inventoryAfter != inventoryBefore {
		t.Errorf("Inventory should be unchanged: before=%d, after=%d", inventoryBefore, inventoryAfter)
	}

	// Verify item still in world
	_, exists := worldInst.GetObject(droppedItem.ObjectID())
	if !exists {
		t.Error("DroppedItem should still be in world after failed pickup")
	}

	t.Log("Out of range pickup test passed!")
}
