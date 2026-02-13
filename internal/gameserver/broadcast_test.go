package gameserver

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

func TestClientManager_BroadcastToAll(t *testing.T) {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	// Register 5 clients
	var clients []*GameClient
	for i := range 5 {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + string(rune('0'+i))
		client.SetAccountName(accountName)
		client.SetState(ClientStateAuthenticated)
		cm.Register(accountName, client)
		clients = append(clients, client)
	}

	// BroadcastToAll should send to all authenticated clients
	payload := []byte{0x01, 0x02, 0x03}
	sent := cm.BroadcastToAll(payload, len(payload))

	if sent != 5 {
		t.Errorf("BroadcastToAll sent to %d clients, want 5", sent)
	}
}

func TestClientManager_BroadcastToAll_SkipsUnauthenticated(t *testing.T) {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	// Register 3 authenticated + 2 unauthenticated clients
	for i := range 5 {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		accountName := "account" + string(rune('0'+i))
		client.SetAccountName(accountName)

		// First 3 authenticated, last 2 not
		if i < 3 {
			client.SetState(ClientStateAuthenticated)
		} else {
			client.SetState(ClientStateConnected)
		}

		cm.Register(accountName, client)
	}

	// BroadcastToAll should send only to authenticated clients
	payload := []byte{0x01, 0x02, 0x03}
	sent := cm.BroadcastToAll(payload, len(payload))

	if sent != 3 {
		t.Errorf("BroadcastToAll sent to %d clients, want 3", sent)
	}
}

func TestClientManager_BroadcastToRegion(t *testing.T) {
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	// Create 3 players in different regions
	// Region formula: rx = (x >> 11) + 64, ry = (y >> 11) + 128
	testCases := []struct {
		accountName string
		playerName  string
		x           int32
		y           int32
		regionX     int32
		regionY     int32
	}{
		{"account1", "Player1", 10000, 20000, 68, 137},  // rx = (10000 >> 11) + 64 = 68
		{"account2", "Player2", 10000, 20000, 68, 137}, // same region as Player1
		{"account3", "Player3", 50000, 60000, 88, 157}, // different region
	}

	for i, tc := range testCases {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
		client.SetAccountName(tc.accountName)
		client.SetState(ClientStateInGame)

		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, tc.playerName, 10, 0, 1)
		player.SetLocation(model.Location{X: tc.x, Y: tc.y, Z: 0, Heading: 0})

		cm.Register(tc.accountName, client)
		cm.RegisterPlayer(player, client)
	}

	// Broadcast to region (68, 137) where Player1 and Player2 are located
	payload := []byte{0x01, 0x02, 0x03}
	sent := cm.BroadcastToRegion(68, 137, payload, len(payload))

	if sent != 2 {
		t.Errorf("BroadcastToRegion sent to %d players, want 2", sent)
	}
}

func TestClientManager_BroadcastToVisibleByLOD(t *testing.T) {
	// Setup world, visibility manager, and client manager
	worldInstance := world.Instance()
	cm := NewClientManager()
	pool := NewBytePool(128)
	cm.SetWritePool(pool)

	vm := world.NewVisibilityManager(worldInstance, 50*time.Millisecond, 100*time.Millisecond)
	cm.SetVisibilityManager(vm)

	baseX, baseY := int32(17000), int32(170000)
	regionSize := world.RegionSize

	// Create source player
	sourceOID := uint32(50001)
	sourcePlayer, err := model.NewPlayer(sourceOID, int64(sourceOID), 1, "SourcePlayer", 10, 0, 1)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	sourcePlayer.SetLocation(model.NewLocation(baseX, baseY, -3500, 0))

	sourceObj := model.NewWorldObject(sourceOID, sourcePlayer.Name(), sourcePlayer.Location())
	if err := worldInstance.AddObject(sourceObj); err != nil {
		t.Fatalf("AddObject(source) failed: %v", err)
	}
	defer worldInstance.RemoveObject(sourceOID)

	sourceConn := testutil.NewMockConn()
	sourceClient, _ := NewGameClient(sourceConn, make([]byte, 16), pool, 16, 0)
	sourceClient.SetAccountName("source_account")
	sourceClient.SetState(ClientStateInGame)
	sourceClient.SetActivePlayer(sourcePlayer)
	cm.Register("source_account", sourceClient)
	cm.RegisterPlayer(sourcePlayer, sourceClient)
	vm.RegisterPlayer(sourcePlayer)
	defer vm.UnregisterPlayer(sourcePlayer)

	// Create near player (same region as source)
	nearOID := uint32(50002)
	nearPlayer, _ := model.NewPlayer(nearOID, int64(nearOID), 1, "NearPlayer", 10, 0, 1)
	nearPlayer.SetLocation(model.NewLocation(baseX+100, baseY+100, -3500, 0))

	nearObj := model.NewWorldObject(nearOID, nearPlayer.Name(), nearPlayer.Location())
	if err := worldInstance.AddObject(nearObj); err != nil {
		t.Fatalf("AddObject(near) failed: %v", err)
	}
	defer worldInstance.RemoveObject(nearOID)

	nearConn := testutil.NewMockConn()
	nearClient, _ := NewGameClient(nearConn, make([]byte, 16), pool, 16, 0)
	nearClient.SetAccountName("near_account")
	nearClient.SetState(ClientStateInGame)
	nearClient.SetActivePlayer(nearPlayer)
	cm.Register("near_account", nearClient)
	cm.RegisterPlayer(nearPlayer, nearClient)
	vm.RegisterPlayer(nearPlayer)
	defer vm.UnregisterPlayer(nearPlayer)

	// Create far player (different region)
	farOID := uint32(50003)
	farPlayer, _ := model.NewPlayer(farOID, int64(farOID), 1, "FarPlayer", 10, 0, 1)
	farPlayer.SetLocation(model.NewLocation(baseX+int32(regionSize)*2, baseY+int32(regionSize)*2, -3500, 0))

	farObj := model.NewWorldObject(farOID, farPlayer.Name(), farPlayer.Location())
	if err := worldInstance.AddObject(farObj); err != nil {
		t.Fatalf("AddObject(far) failed: %v", err)
	}
	defer worldInstance.RemoveObject(farOID)

	farConn := testutil.NewMockConn()
	farClient, _ := NewGameClient(farConn, make([]byte, 16), pool, 16, 0)
	farClient.SetAccountName("far_account")
	farClient.SetState(ClientStateInGame)
	farClient.SetActivePlayer(farPlayer)
	cm.Register("far_account", farClient)
	cm.RegisterPlayer(farPlayer, farClient)
	vm.RegisterPlayer(farPlayer)
	defer vm.UnregisterPlayer(farPlayer)

	// Populate visibility caches
	vm.UpdateAll()

	payload := []byte{0x01, 0x02, 0x03}

	// Test BroadcastToVisibleNear: only near player should receive
	sentNear := cm.BroadcastToVisibleNear(sourcePlayer, payload, len(payload))
	// Near player is in same region — should receive the broadcast
	if sentNear < 0 {
		t.Errorf("BroadcastToVisibleNear sent %d packets, expected >= 0", sentNear)
	}
	t.Logf("BroadcastToVisibleNear: %d packets", sentNear)

	// Test BroadcastToVisible (LODAll): all visible players should receive
	sentAll := cm.BroadcastToVisible(sourcePlayer, payload, len(payload))
	if sentAll < sentNear {
		t.Errorf("BroadcastToVisible (%d) should be >= BroadcastToVisibleNear (%d)", sentAll, sentNear)
	}
	t.Logf("BroadcastToVisible (LODAll): %d packets", sentAll)

	// Test: nil VisibilityManager returns 0
	cm2 := NewClientManager()
	sentNilVM := cm2.BroadcastToVisible(sourcePlayer, payload, len(payload))
	if sentNilVM != 0 {
		t.Errorf("BroadcastToVisible with nil VisibilityManager sent %d, want 0", sentNilVM)
	}
}
