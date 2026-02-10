package world

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// TestVisibilityManager_ReverseCache verifies reverse visibility index correctness.
// Phase 4.18 Optimization 1: Ensures GetObservers() returns correct observer list.
func TestVisibilityManager_ReverseCache(t *testing.T) {
	worldInstance := Instance()

	// Create visibility manager
	vm := NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)

	// Create 3 players in same region (all can see each other)
	regionX, regionY := int32(0), int32(0)
	baseX := regionX * RegionSize
	baseY := regionY * RegionSize

	players := make([]*model.Player, 3)
	for i := range 3 {
		player, err := model.NewPlayer(
			uint32(i+1),
			int64(i+1),
			int64(1),
			"TestPlayer"+string(rune('A'+i)),
			10,
			0,
			1,
		)
		if err != nil {
			t.Fatal(err)
		}

		// Place players in same region (1000 units apart)
		x := baseX + int32(i*1000)
		y := baseY + int32(i*1000)
		player.SetLocation(model.Location{X: x, Y: y, Z: 0, Heading: 0})

		// Add to world
		worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), player.Location())
		if err := worldInstance.AddObject(worldObj); err != nil {
			t.Fatal(err)
		}

		// Register with visibility manager
		vm.RegisterPlayer(player)
		players[i] = player
	}

	// Trigger batch update to build forward + reverse cache
	vm.UpdateAll()

	// Test 1: Verify reverse cache exists and is not nil
	cached := vm.reverseCache.Load()
	if cached == nil {
		t.Fatal("Reverse cache not initialized after UpdateAll()")
	}

	reverseCache, ok := cached.(map[uint32][]uint32)
	if !ok {
		t.Fatalf("Invalid reverse cache type: %T", cached)
	}

	// Test 2: Verify each player can be observed by others
	for i, player := range players {
		observers := vm.GetObservers(player.ObjectID())
		if observers == nil {
			t.Errorf("GetObservers returned nil for player %d", i)
			continue
		}

		// Expected: 2 other players can see this player (3 players total, excluding self)
		if len(observers) != 2 {
			t.Errorf("Player %d: expected 2 observers, got %d", i, len(observers))
			continue
		}

		// Verify observers are the OTHER two players (not self)
		foundOthers := 0
		for _, observerID := range observers {
			if observerID == player.ObjectID() {
				t.Errorf("Player %d: observer list includes self (objectID %d)", i, observerID)
			}
			for j, otherPlayer := range players {
				if j != i && observerID == otherPlayer.ObjectID() {
					foundOthers++
				}
			}
		}

		if foundOthers != 2 {
			t.Errorf("Player %d: expected to find 2 other players in observer list, found %d", i, foundOthers)
		}
	}

	// Test 3: Verify reverse cache contains at least our 3 players
	// Note: May contain more entries from previous tests (world singleton)
	// We only verify our players are present and have correct observers
	if len(reverseCache) < 3 {
		t.Errorf("Expected at least 3 entries in reverse cache, got %d", len(reverseCache))
	}

	// Verify all our players are in reverse cache
	for i, player := range players {
		if _, ok := reverseCache[player.ObjectID()]; !ok {
			t.Errorf("Player %d (objectID %d) not found in reverse cache", i, player.ObjectID())
		}
	}

	// Cleanup
	for _, player := range players {
		vm.UnregisterPlayer(player)
		worldInstance.RemoveObject(player.ObjectID())
	}
}

// TestVisibilityManager_ReverseCacheDistantPlayers verifies reverse cache handles distant players correctly.
// Players in different regions should NOT see each other → reverse cache should NOT include them.
func TestVisibilityManager_ReverseCacheDistantPlayers(t *testing.T) {
	worldInstance := Instance()

	vm := NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)

	// Create 2 players in DISTANT regions (too far to see each other)
	player1, err := model.NewPlayer(uint32(1), int64(1), int64(1), "Player1", 10, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	player1.SetLocation(model.Location{X: 0, Y: 0, Z: 0, Heading: 0})

	player2, err := model.NewPlayer(uint32(2), int64(2), int64(1), "Player2", 10, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Place player2 very far away (10 regions = 163,840 units away)
	// This is beyond visibility range (3×3 region grid = max 2 regions diagonal)
	player2.SetLocation(model.Location{X: 10 * RegionSize, Y: 10 * RegionSize, Z: 0, Heading: 0})

	// Add to world
	worldObj1 := model.NewWorldObject(player1.ObjectID(), player1.Name(), player1.Location())
	if err := worldInstance.AddObject(worldObj1); err != nil {
		t.Fatal(err)
	}

	worldObj2 := model.NewWorldObject(player2.ObjectID(), player2.Name(), player2.Location())
	if err := worldInstance.AddObject(worldObj2); err != nil {
		t.Fatal(err)
	}

	// Register with visibility manager
	vm.RegisterPlayer(player1)
	vm.RegisterPlayer(player2)

	// Trigger batch update
	vm.UpdateAll()

	// Test: Verify neither player can see the other
	// Note: GetObservers() returns nil (not empty slice) if object has no observers
	// This is correct behavior — no entry in reverse cache means no observers
	observers1 := vm.GetObservers(player1.ObjectID())
	// Expected: nil or empty slice (player2 is too far away)
	if observers1 != nil && len(observers1) != 0 {
		t.Errorf("Player1: expected 0 observers (player2 too far), got %d", len(observers1))
	}

	observers2 := vm.GetObservers(player2.ObjectID())
	if observers2 != nil && len(observers2) != 0 {
		t.Errorf("Player2: expected 0 observers (player1 too far), got %d", len(observers2))
	}

	// Cleanup
	vm.UnregisterPlayer(player1)
	vm.UnregisterPlayer(player2)
	worldInstance.RemoveObject(player1.ObjectID())
	worldInstance.RemoveObject(player2.ObjectID())
}

// TestVisibilityManager_ReverseCacheUpdate verifies reverse cache rebuilds correctly after player movement.
// Phase 4.18: Ensures reverse cache stays in sync with forward cache updates.
func TestVisibilityManager_ReverseCacheUpdate(t *testing.T) {
	worldInstance := Instance()

	vm := NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)

	// Create player1 at (0, 0)
	player1, err := model.NewPlayer(uint32(1), int64(1), int64(1), "Player1", 10, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	player1.SetLocation(model.Location{X: 0, Y: 0, Z: 0, Heading: 0})

	// Create player2 far away (initially invisible to player1)
	player2, err := model.NewPlayer(uint32(2), int64(2), int64(1), "Player2", 10, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Place player2 far away (10 regions = beyond visibility range)
	player2.SetLocation(model.Location{X: 10 * RegionSize, Y: 10 * RegionSize, Z: 0, Heading: 0})

	// Add to world
	worldObj1 := model.NewWorldObject(player1.ObjectID(), player1.Name(), player1.Location())
	if err := worldInstance.AddObject(worldObj1); err != nil {
		t.Fatal(err)
	}

	worldObj2 := model.NewWorldObject(player2.ObjectID(), player2.Name(), player2.Location())
	if err := worldInstance.AddObject(worldObj2); err != nil {
		t.Fatal(err)
	}

	vm.RegisterPlayer(player1)
	vm.RegisterPlayer(player2)

	// Initial update: players far apart, no visibility
	vm.UpdateAll()

	observers1Before := vm.GetObservers(player1.ObjectID())
	if len(observers1Before) != 0 {
		t.Errorf("Before move: expected 0 observers for player1, got %d", len(observers1Before))
	}

	// Move player2 close to player1 (same region)
	player2.SetLocation(model.Location{X: 1000, Y: 1000, Z: 0, Heading: 0})
	worldObj2.SetLocation(player2.Location())

	// Update again: players now close, should have visibility
	vm.UpdateAll()

	observers1After := vm.GetObservers(player1.ObjectID())
	if len(observers1After) != 1 {
		t.Errorf("After move: expected 1 observer for player1, got %d", len(observers1After))
	}

	// Verify observer is player2
	if len(observers1After) > 0 && observers1After[0] != player2.ObjectID() {
		t.Errorf("After move: expected observer to be player2 (objectID %d), got %d",
			player2.ObjectID(), observers1After[0])
	}

	// Cleanup
	vm.UnregisterPlayer(player1)
	vm.UnregisterPlayer(player2)
	worldInstance.RemoveObject(player1.ObjectID())
	worldInstance.RemoveObject(player2.ObjectID())
}
