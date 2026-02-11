package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestVisibilityCache_LODBucketing verifies LOD visibility cache correctness.
// Phase 4.14: Integration test for LOD bucketing (near/medium/far split).
//
// NOTE: Broadcast packet reduction tests are SKIPPED due to architectural limitation:
// Player.ObjectID() returns 0 by default (Player and WorldObject are separate entities).
// Broadcast methods search by Player.ObjectID(), but visibility cache contains WorldObject IDs.
// This requires architectural refactor (link Player ↔ WorldObject) — TODO Phase 4.15.
//
// THIS TEST ONLY VERIFIES: Visibility cache LOD bucketing works correctly.
//
// Test scenario:
// - sourceObj at center region (17000, 170000)
// - 1 nearPlayer in same region → should see sourceObj in LODNear bucket
// - 4 mediumPlayers in adjacent regions → should see sourceObj in LODMedium bucket
// - 4 farPlayers in diagonal regions → should see sourceObj in LODFar bucket
// - Total: 9 target players
//
// Verified:
// - LOD bucketing works correctly (near/medium/far split)
// - Visibility cache populated for all players
// - sourceObj correctly distributed across LOD buckets based on distance
func TestVisibilityCache_LODBucketing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	worldInstance := world.Instance()

	// Test coordinates: Talking Island spawn area
	baseX, baseY := int32(17000), int32(170000)
	regionSize := world.RegionSize // 4096 game units

	// Create sourcePlayer at center
	sourcePlayer, err := model.NewPlayer(1, 1, 1, "SourcePlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	sourcePlayer.SetLocation(model.NewLocation(baseX, baseY, -3500, 0))

	// Register sourcePlayer as WorldObject (required for visibility)
	// IMPORTANT: Use unique ObjectID for each WorldObject (Player.ObjectID() returns 0 by default)
	sourceObjID := uint32(10000)
	sourceObj := model.NewWorldObject(sourceObjID, sourcePlayer.Name(), sourcePlayer.Location())
	if err := worldInstance.AddObject(sourceObj); err != nil {
		t.Fatalf("AddObject(sourcePlayer) failed: %v", err)
	}

	// DEBUG: Verify sourcePlayer added to correct region
	sourceRegionX, sourceRegionY := world.CoordToRegionIndex(baseX, baseY)
	sourceRegion := worldInstance.GetRegionByIndex(sourceRegionX, sourceRegionY)
	if sourceRegion != nil {
		t.Logf("sourcePlayer added to region(%d,%d), objectCount=%d",
			sourceRegionX, sourceRegionY, len(sourceRegion.GetVisibleObjectsSnapshot()))
	}

	// Create target players in 3×3 grid around sourcePlayer
	targetPlayers := make([]*model.Player, 0, 9)
	expectedNear := 0
	expectedMedium := 0
	expectedAll := 0

	playerID := int64(2)

	// NEAR: 1 player in center region (same as source)
	nearPlayer, _ := model.NewPlayer(uint32(playerID), playerID, 1, "NearPlayer", 10, 0, 1)
	nearPlayer.SetLocation(model.NewLocation(baseX+100, baseY+100, -3500, 0))
	nearObjID := uint32(10000 + playerID)
	nearObj := model.NewWorldObject(nearObjID, nearPlayer.Name(), nearPlayer.Location())
	worldInstance.AddObject(nearObj)
	targetPlayers = append(targetPlayers, nearPlayer)
	expectedNear++
	expectedMedium++
	expectedAll++
	playerID++

	// MEDIUM: 4 players in adjacent regions (share edge with center)
	adjacentOffsets := []struct{ dx, dy int32 }{
		{0, -1}, // top
		{-1, 0}, // left
		{1, 0},  // right
		{0, 1},  // bottom
	}

	for _, offset := range adjacentOffsets {
		player, _ := model.NewPlayer(uint32(playerID), playerID, 1, "MediumPlayer", 10, 0, 1)
		x := baseX + offset.dx*int32(regionSize)
		y := baseY + offset.dy*int32(regionSize)
		player.SetLocation(model.NewLocation(x, y, -3500, 0))
		objID := uint32(10000 + playerID)
		obj := model.NewWorldObject(objID, player.Name(), player.Location())
		worldInstance.AddObject(obj)
		targetPlayers = append(targetPlayers, player)
		expectedMedium++
		expectedAll++
		playerID++
	}

	// FAR: 4 players in diagonal regions (share corner with center)
	diagonalOffsets := []struct{ dx, dy int32 }{
		{-1, -1}, // top-left
		{1, -1},  // top-right
		{-1, 1},  // bottom-left
		{1, 1},   // bottom-right
	}

	for _, offset := range diagonalOffsets {
		player, _ := model.NewPlayer(uint32(playerID), playerID, 1, "FarPlayer", 10, 0, 1)
		x := baseX + offset.dx*int32(regionSize)
		y := baseY + offset.dy*int32(regionSize)
		player.SetLocation(model.NewLocation(x, y, -3500, 0))
		objID := uint32(10000 + playerID)
		obj := model.NewWorldObject(objID, player.Name(), player.Location())
		worldInstance.AddObject(obj)
		targetPlayers = append(targetPlayers, player)
		expectedAll++
		playerID++
	}

	// Start VisibilityManager to populate caches
	vm := world.NewVisibilityManager(worldInstance, 50*time.Millisecond, 100*time.Millisecond)
	vm.RegisterPlayer(sourcePlayer)
	for _, player := range targetPlayers {
		vm.RegisterPlayer(player)
	}

	go vm.Start(ctx)

	// Wait for visibility caches to populate (need 2-3 batch updates)
	time.Sleep(300 * time.Millisecond)

	// Verify caches are populated
	if sourcePlayer.GetVisibilityCache() == nil {
		t.Fatal("sourcePlayer cache not created")
	}
	for i, player := range targetPlayers {
		cache := player.GetVisibilityCache()
		if cache == nil {
			t.Fatalf("targetPlayer[%d] cache not created", i)
		}

		// DEBUG: Check if sourcePlayer is in target's cache and which LOD bucket
		near := cache.NearObjects()
		medium := cache.MediumObjects()
		far := cache.FarObjects()

		foundInNear := false
		foundInMedium := false
		foundInFar := false

		for _, obj := range near {
			if obj.ObjectID() == sourceObjID {
				foundInNear = true
				break
			}
		}
		for _, obj := range medium {
			if obj.ObjectID() == sourceObjID {
				foundInMedium = true
				break
			}
		}
		for _, obj := range far {
			if obj.ObjectID() == sourceObjID {
				foundInFar = true
				break
			}
		}

		// Get player's region
		playerLoc := player.Location()
		playerRegionX, playerRegionY := world.CoordToRegionIndex(playerLoc.X, playerLoc.Y)
		sourceRegionX, sourceRegionY := world.CoordToRegionIndex(baseX, baseY)

		t.Logf("targetPlayer[%d] (%s) at region(%d,%d), source at region(%d,%d): near=%v, medium=%v, far=%v ✅",
			i, player.Name(), playerRegionX, playerRegionY, sourceRegionX, sourceRegionY,
			foundInNear, foundInMedium, foundInFar)

		// Verify LOD bucketing correctness
		if i == 0 {
			// NearPlayer: should see sourceObj in NEAR bucket only
			if !foundInNear || foundInMedium || foundInFar {
				t.Errorf("NearPlayer LOD bucketing incorrect: near=%v, medium=%v, far=%v (want: true/false/false)",
					foundInNear, foundInMedium, foundInFar)
			}
		} else if i >= 1 && i <= 4 {
			// MediumPlayers: should see sourceObj in MEDIUM bucket only
			if foundInNear || !foundInMedium || foundInFar {
				t.Errorf("MediumPlayer[%d] LOD bucketing incorrect: near=%v, medium=%v, far=%v (want: false/true/false)",
					i, foundInNear, foundInMedium, foundInFar)
			}
		} else {
			// FarPlayers: should see sourceObj in FAR bucket only
			if foundInNear || foundInMedium || !foundInFar {
				t.Errorf("FarPlayer[%d] LOD bucketing incorrect: near=%v, medium=%v, far=%v (want: false/false/true)",
					i, foundInNear, foundInMedium, foundInFar)
			}
		}
	}

	t.Logf("✅ Visibility cache LOD bucketing verified: 1 near / 4 medium / 4 far players")
	t.Logf("Expected broadcast reduction: Near=-89%% (1 vs 9), Medium=-56%% (5 vs 9)")
	t.Logf("NOTE: Actual broadcast tests skipped — requires Player↔WorldObject linking (Phase 4.15)")

	// Cleanup
	vm.UnregisterPlayer(sourcePlayer)
	for _, player := range targetPlayers {
		vm.UnregisterPlayer(player)
	}
	worldInstance.RemoveObject(sourceObjID)
	// Remove target player objects
	for i := range len(targetPlayers) {
		worldInstance.RemoveObject(uint32(10000 + int64(i+2)))
	}
}

// TestBroadcastPacketReduction_LOD verifies broadcast packet reduction with LOD API.
// Phase 4.16: Integration test for broadcast optimization using LOD-aware methods.
//
// Prerequisites: Phase 4.15 completed (Player↔WorldObject linking via objectID).
//
// Test scenario:
// - sourcePlayer with GameClient at center region
// - 9 target players with GameClients (1 near, 4 medium, 4 far)
// - Populate visibility caches via VisibilityManager
// - Measure packet counts for different broadcast methods
//
// Expected results (LOD levels are EXCLUSIVE, not cumulative):
// - BroadcastToAll: 10 packets (all players receive, including source)
// - BroadcastToVisibleNear: 1 packet (only near player receives)
// - BroadcastToVisibleMedium: 4 packets (only medium players receive, near excluded)
// - BroadcastToVisible: 9 packets (all visible players receive, backward compat)
//
// Packet reduction verification:
// - Near vs All: -90% (1 vs 10 packets) — optimal for critical events
// - Medium vs All: -60% (4 vs 10 packets) — optimal for zone events
func TestBroadcastPacketReduction_LOD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	worldInstance := world.Instance()

	// Test coordinates: Talking Island spawn area
	baseX, baseY := int32(17000), int32(170000)
	regionSize := world.RegionSize

	// Create sourcePlayer with proper objectID (Phase 4.15)
	sourceObjectID := uint32(0x10000001)
	sourcePlayer, err := model.NewPlayer(sourceObjectID, 1, 1, "SourcePlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	sourcePlayer.SetLocation(model.NewLocation(baseX, baseY, -3500, 0))

	// Register sourcePlayer as WorldObject
	sourceObj := model.NewWorldObject(sourceObjectID, sourcePlayer.Name(), sourcePlayer.Location())
	if err := worldInstance.AddObject(sourceObj); err != nil {
		t.Fatalf("AddObject(sourcePlayer) failed: %v", err)
	}

	// Setup ClientManager and register sourcePlayer
	cm := gameserver.NewClientManager()
	pool := gameserver.NewBytePool(128)
	cm.SetWritePool(pool)
	sourceConn := testutil.NewMockConn()
	sourceClient, _ := gameserver.NewGameClient(sourceConn, make([]byte, 16), pool, 16, 0)
	sourceClient.SetAccountName("source_account")
	sourceClient.SetState(gameserver.ClientStateInGame)
	sourceClient.SetActivePlayer(sourcePlayer) // Phase 4.18 Fix: Set ActivePlayer for broadcast
	cm.Register("source_account", sourceClient)
	cm.RegisterPlayer(sourcePlayer, sourceClient)

	// Create 9 target players with GameClients (1 near, 4 medium, 4 far)
	targetPlayers := make([]*model.Player, 0, 9)
	targetClients := make([]*gameserver.GameClient, 0, 9)

	objectID := uint32(0x10000002)

	// NEAR: 1 player in center region (same as source)
	nearPlayer, _ := model.NewPlayer(objectID, int64(objectID), 1, "NearPlayer", 10, 0, 1)
	nearPlayer.SetLocation(model.NewLocation(baseX+100, baseY+100, -3500, 0))
	nearObj := model.NewWorldObject(objectID, nearPlayer.Name(), nearPlayer.Location())
	worldInstance.AddObject(nearObj)

	nearConn := testutil.NewMockConn()
	nearClient, _ := gameserver.NewGameClient(nearConn, make([]byte, 16), pool, 16, 0)
	nearClient.SetAccountName("near_account")
	nearClient.SetState(gameserver.ClientStateInGame)
	nearClient.SetActivePlayer(nearPlayer) // Phase 4.18 Fix: Set ActivePlayer for broadcast
	cm.Register("near_account", nearClient)
	cm.RegisterPlayer(nearPlayer, nearClient)

	targetPlayers = append(targetPlayers, nearPlayer)
	targetClients = append(targetClients, nearClient)
	objectID++

	// MEDIUM: 4 players in adjacent regions (share edge with center)
	adjacentOffsets := []struct{ dx, dy int32 }{
		{0, -1}, // top
		{-1, 0}, // left
		{1, 0},  // right
		{0, 1},  // bottom
	}

	for _, offset := range adjacentOffsets {
		player, _ := model.NewPlayer(objectID, int64(objectID), 1, "MediumPlayer", 10, 0, 1)
		x := baseX + offset.dx*int32(regionSize)
		y := baseY + offset.dy*int32(regionSize)
		player.SetLocation(model.NewLocation(x, y, -3500, 0))

		obj := model.NewWorldObject(objectID, player.Name(), player.Location())
		worldInstance.AddObject(obj)

		conn := testutil.NewMockConn()
		client, _ := gameserver.NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		client.SetAccountName(fmt.Sprintf("medium_account_%d", objectID))
		client.SetState(gameserver.ClientStateInGame)
		client.SetActivePlayer(player) // Phase 4.18 Fix: Set ActivePlayer for broadcast
		cm.Register(fmt.Sprintf("medium_account_%d", objectID), client)
		cm.RegisterPlayer(player, client)

		targetPlayers = append(targetPlayers, player)
		targetClients = append(targetClients, client)
		objectID++
	}

	// FAR: 4 players in diagonal regions (share corner with center)
	diagonalOffsets := []struct{ dx, dy int32 }{
		{-1, -1}, // top-left
		{1, -1},  // top-right
		{-1, 1},  // bottom-left
		{1, 1},   // bottom-right
	}

	for _, offset := range diagonalOffsets {
		player, _ := model.NewPlayer(objectID, int64(objectID), 1, "FarPlayer", 10, 0, 1)
		x := baseX + offset.dx*int32(regionSize)
		y := baseY + offset.dy*int32(regionSize)
		player.SetLocation(model.NewLocation(x, y, -3500, 0))

		obj := model.NewWorldObject(objectID, player.Name(), player.Location())
		worldInstance.AddObject(obj)

		conn := testutil.NewMockConn()
		client, _ := gameserver.NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		client.SetAccountName(fmt.Sprintf("far_account_%d", objectID))
		client.SetState(gameserver.ClientStateInGame)
		client.SetActivePlayer(player) // Phase 4.18 Fix: Set ActivePlayer for broadcast
		cm.Register(fmt.Sprintf("far_account_%d", objectID), client)
		cm.RegisterPlayer(player, client)

		targetPlayers = append(targetPlayers, player)
		targetClients = append(targetClients, client)
		objectID++
	}

	// Start VisibilityManager to populate caches
	vm := world.NewVisibilityManager(worldInstance, 50*time.Millisecond, 100*time.Millisecond)

	// Link VisibilityManager to ClientManager (required for reverse cache lookups)
	cm.SetVisibilityManager(vm)

	vm.RegisterPlayer(sourcePlayer)
	for _, player := range targetPlayers {
		vm.RegisterPlayer(player)
	}

	go vm.Start(ctx)

	// Wait for visibility caches to populate (need 2-3 batch updates)
	time.Sleep(300 * time.Millisecond)

	// Force immediate update to ensure reverse cache is built
	vm.UpdateAll()

	// Verify caches are populated
	if sourcePlayer.GetVisibilityCache() == nil {
		t.Fatal("sourcePlayer cache not created")
	}
	for i, player := range targetPlayers {
		if player.GetVisibilityCache() == nil {
			t.Fatalf("targetPlayer[%d] cache not created", i)
		}
	}

	// Verify reverse cache exists
	observers := vm.GetObservers(sourcePlayer.ObjectID())
	if observers == nil {
		t.Fatal("Reverse cache not built after UpdateAll()")
	}
	t.Logf("✅ Reverse cache built: sourcePlayer has %d observers", len(observers))

	// Create test packet (dummy data)
	testPacket := make([]byte, 64)
	testPacket[0] = 0x03 // dummy opcode
	copy(testPacket[1:], []byte("test broadcast packet"))
	payloadLen := 23

	// Test 1: BroadcastToAll (baseline — should queue to all 10 players)
	// NOTE: sent count = packets queued to sendCh (async write architecture).
	// Actual delivery to conn happens in writePump (not started in test).
	sentAll := cm.BroadcastToAll(testPacket, payloadLen)
	if sentAll != 10 { // 9 targets + 1 source
		t.Errorf("BroadcastToAll sent %d packets, want 10 (9 targets + 1 source)", sentAll)
	}
	t.Logf("BroadcastToAll: %d packets queued (baseline)", sentAll)

	// Test 2: BroadcastToVisibleNear (should queue only to near player)
	sentNear := cm.BroadcastToVisibleNear(sourcePlayer, testPacket, payloadLen)
	if sentNear != 1 {
		t.Errorf("BroadcastToVisibleNear sent %d packets, want 1 (only near player)", sentNear)
	}

	reductionNear := float64(sentAll-sentNear) / float64(sentAll) * 100
	t.Logf("BroadcastToVisibleNear: %d packets queued (-%.1f%% vs BroadcastToAll)", sentNear, reductionNear)

	// Test 3: BroadcastToVisibleMedium (should queue to medium only = 4 players)
	// Note: LOD levels are EXCLUSIVE, not cumulative. LODMedium = medium only, not near+medium.
	sentMedium := cm.BroadcastToVisibleMedium(sourcePlayer, testPacket, payloadLen)
	if sentMedium != 4 {
		t.Errorf("BroadcastToVisibleMedium sent %d packets, want 4 (medium only)", sentMedium)
	}

	reductionMedium := float64(sentAll-sentMedium) / float64(sentAll) * 100
	t.Logf("BroadcastToVisibleMedium: %d packets queued (-%.1f%% vs BroadcastToAll)", sentMedium, reductionMedium)

	// Test 4: BroadcastToVisible (backward compat — should queue to all visible = 9 players)
	sentVisible := cm.BroadcastToVisible(sourcePlayer, testPacket, payloadLen)
	if sentVisible != 9 {
		t.Errorf("BroadcastToVisible sent %d packets, want 9 (all visible players)", sentVisible)
	}
	t.Logf("BroadcastToVisible: %d packets queued (backward compat, LODAll)", sentVisible)

	// Final verification
	t.Logf("📊 Packet Reduction Summary:")
	t.Logf("  - Near:   1 packet  (-90%% vs All) — critical events (movement, combat)")
	t.Logf("  - Medium: 4 packets (-60%% vs All) — zone events (NPC spawn)")
	t.Logf("  - All:    9 packets (baseline)    — global events (announcement)")
	t.Logf("✅ Phase 4.16: Broadcast packet reduction verified!")

	// Cleanup
	vm.UnregisterPlayer(sourcePlayer)
	for _, player := range targetPlayers {
		vm.UnregisterPlayer(player)
	}
	worldInstance.RemoveObject(sourceObjectID)
	for i := range targetPlayers {
		worldInstance.RemoveObject(uint32(0x10000002 + i))
	}
}

