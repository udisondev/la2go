package combat

import (
	"testing"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestDropLoot_RealNpcDropTable tests the full flow:
// Monster death → CalculateDrops → DroppedItem created → added to world.
func TestDropLoot_RealNpcDropTable(t *testing.T) {
	// NPC 13031 "Huge Pig": 100% group, 100% item, itemID=9142, min=1, max=2
	template := model.NewNpcTemplate(
		13031, "Huge Pig", "",
		70, 680, 2000,
		8, 60000, 5, 200000,
		300, 160, 253,
		0, 0, 0, 0,
	)

	npc := model.NewNpc(world.IDGenerator().NextNpcID(), 13031, template)
	npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	killer, err := model.NewPlayer(world.IDGenerator().NextPlayerID(), 1, 1, "TestKiller", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	killer.SetLocation(model.NewLocation(17050, 170050, -3500, 0))
	killer.WorldObject.Data = killer

	// Track broadcast calls
	var broadcastCalls int
	broadcastFn := func(source *model.Player, pktData []byte, size int) {
		broadcastCalls++
	}
	npcBroadcastFn := func(x, y int32, pktData []byte, size int) {}

	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
		ItemAutoDestroyTime:       300, // Long timeout for test
	}

	mgr := NewCombatManager(broadcastFn, npcBroadcastFn, nil)
	mgr.SetRates(rates)

	// Act: trigger dropLoot
	mgr.dropLoot(npc, killer)

	// Assert: item should be in world
	worldInst := world.Instance()

	// Check that broadcast was called (ItemOnGround packet)
	if broadcastCalls != 1 {
		t.Errorf("expected 1 broadcast call (ItemOnGround), got %d", broadcastCalls)
	}

	// Verify dropped item is in the world by checking items map
	foundItems := 0
	worldInst.ForEachItem(func(di *model.DroppedItem) bool {
		item := di.Item()
		if item.ItemID() == 9142 {
			foundItems++
			if item.Count() < 1 || item.Count() > 2 {
				t.Errorf("Adena count = %d, want 1-2", item.Count())
			}
			// Check position is near NPC (±70)
			loc := di.Location()
			dx := loc.X - 17000
			dy := loc.Y - 170000
			if dx < -70 || dx > 70 || dy < -70 || dy > 70 {
				t.Errorf("item position too far from NPC: (%d, %d), expected ±70 from (17000, 170000)", loc.X, loc.Y)
			}
		}
		return true
	})

	if foundItems != 1 {
		t.Errorf("expected 1 dropped item in world, found %d", foundItems)
	}
}

// TestDropLoot_NpcWithoutDropTable verifies no items appear for NPC without drops.
func TestDropLoot_NpcWithoutDropTable(t *testing.T) {
	// Use template ID that has no drops
	template := model.NewNpcTemplate(
		99999, "NoDrop NPC", "",
		5, 100, 50,
		10, 50, 5, 25,
		0, 100, 253,
		0, 0, 0, 0,
	)

	npc := model.NewNpc(world.IDGenerator().NextNpcID(), 99999, template)
	npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	killer, err := model.NewPlayer(world.IDGenerator().NextPlayerID(), 2, 1, "TestKiller2", 5, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	killer.SetLocation(model.NewLocation(17050, 170050, -3500, 0))
	killer.WorldObject.Data = killer

	var broadcastCalls int
	broadcastFn := func(source *model.Player, pktData []byte, size int) {
		broadcastCalls++
	}

	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
		ItemAutoDestroyTime:       300,
	}

	mgr := NewCombatManager(broadcastFn, nil, nil)
	mgr.SetRates(rates)

	mgr.dropLoot(npc, killer)

	if broadcastCalls != 0 {
		t.Errorf("expected 0 broadcasts for NPC without drops, got %d", broadcastCalls)
	}
}

// TestDropLoot_MultipleItemDrop tests NPC with multiple drop groups.
func TestDropLoot_MultipleItemDrop(t *testing.T) {
	// NPC 18003 "Bearded Keltir": multiple groups, Adena 70%, arrows, etc.
	template := model.NewNpcTemplate(
		18003, "Bearded Keltir", "",
		1, 39, 40,
		4, 44, 5, 26,
		1000, 110, 253,
		0, 0, 29, 2,
	)

	npc := model.NewNpc(world.IDGenerator().NextNpcID(), 18003, template)
	npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	killer, err := model.NewPlayer(world.IDGenerator().NextPlayerID(), 3, 1, "TestKiller3", 5, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	killer.SetLocation(model.NewLocation(17050, 170050, -3500, 0))
	killer.WorldObject.Data = killer

	// Use 100x rate to ensure most drops happen
	rates := &config.Rates{
		DeathDropChanceMultiplier: 100.0,
		DeathDropAmountMultiplier: 1.0,
		ItemAutoDestroyTime:       300,
	}

	var broadcastCalls int
	broadcastFn := func(source *model.Player, pktData []byte, size int) {
		broadcastCalls++
	}

	mgr := NewCombatManager(broadcastFn, nil, nil)
	mgr.SetRates(rates)

	mgr.dropLoot(npc, killer)

	// With 100x multiplier, should get several items
	if broadcastCalls < 2 {
		t.Errorf("expected multiple broadcast calls with 100x rates, got %d", broadcastCalls)
	}
}
