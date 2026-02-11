package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
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
	// TODO: This test requires:
	// 1. World instance with populated regions
	// 2. VisibilityManager running to populate caches
	// 3. Players registered as WorldObjects
	// Current limitation: visibility cache depends on full world setup
	// which is not available in unit tests without integration test infrastructure.
	//
	// For now, skip this test. Will implement in Phase 4.14 when we have
	// proper test fixtures for world + visibility system.
	t.Skip("Requires world integration test infrastructure (Phase 4.14)")
}
