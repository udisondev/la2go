package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

func TestNewClientManager(t *testing.T) {
	cm := NewClientManager()
	if cm == nil {
		t.Fatal("NewClientManager returned nil")
	}

	if cm.Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", cm.Count())
	}

	if cm.PlayerCount() != 0 {
		t.Errorf("Initial PlayerCount() = %d, want 0", cm.PlayerCount())
	}
}

func TestClientManager_Register_Unregister(t *testing.T) {
	cm := NewClientManager()

	conn1 := testutil.NewMockConn()
	client1, _ := NewGameClient(conn1, make([]byte, 16), nil, 0, 0)
	client1.SetAccountName("account1")

	conn2 := testutil.NewMockConn()
	client2, _ := NewGameClient(conn2, make([]byte, 16), nil, 0, 0)
	client2.SetAccountName("account2")

	// Register clients
	cm.Register("account1", client1)
	if cm.Count() != 1 {
		t.Errorf("After register client1, Count() = %d, want 1", cm.Count())
	}

	cm.Register("account2", client2)
	if cm.Count() != 2 {
		t.Errorf("After register client2, Count() = %d, want 2", cm.Count())
	}

	// Get client
	got := cm.GetClient("account1")
	if got != client1 {
		t.Error("GetClient returned wrong client")
	}

	// Unregister client
	cm.Unregister("account1")
	if cm.Count() != 1 {
		t.Errorf("After unregister client1, Count() = %d, want 1", cm.Count())
	}

	// Verify client removed
	got = cm.GetClient("account1")
	if got != nil {
		t.Error("GetClient should return nil after unregister")
	}
}

func TestClientManager_RegisterPlayer(t *testing.T) {
	cm := NewClientManager()

	conn := testutil.NewMockConn()
	client, _ := NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	client.SetAccountName("testaccount")

	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)

	// Register client
	cm.Register("testaccount", client)

	// Register player
	cm.RegisterPlayer(player, client)
	if cm.PlayerCount() != 1 {
		t.Errorf("After RegisterPlayer, PlayerCount() = %d, want 1", cm.PlayerCount())
	}

	// Get client by player
	got := cm.GetClientByPlayer(player)
	if got != client {
		t.Error("GetClientByPlayer returned wrong client")
	}

	// Unregister player
	cm.UnregisterPlayer(player)
	if cm.PlayerCount() != 0 {
		t.Errorf("After UnregisterPlayer, PlayerCount() = %d, want 0", cm.PlayerCount())
	}
}

func TestClientManager_ForEachClient(t *testing.T) {
	cm := NewClientManager()

	// Register 5 clients
	for i := range 5 {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), nil, 0, 0)
		accountName := "account" + string(rune('0'+i))
		client.SetAccountName(accountName)
		cm.Register(accountName, client)
	}

	// Iterate all
	count := 0
	cm.ForEachClient(func(c *GameClient) bool {
		count++
		return true
	})

	if count != 5 {
		t.Errorf("ForEachClient iterated %d clients, want 5", count)
	}

	// Early stop
	count = 0
	cm.ForEachClient(func(c *GameClient) bool {
		count++
		return count < 3 // stop after 3
	})

	if count != 3 {
		t.Errorf("ForEachClient with early stop iterated %d clients, want 3", count)
	}
}

func TestClientManager_ForEachPlayer(t *testing.T) {
	cm := NewClientManager()

	// Register 3 players
	for i := range 3 {
		conn := testutil.NewMockConn()
		client, _ := NewGameClient(conn, make([]byte, 16), nil, 0, 0)
		accountName := "account" + string(rune('0'+i))
		client.SetAccountName(accountName)

		player, _ := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player", 10, 0, 1)

		cm.Register(accountName, client)
		cm.RegisterPlayer(player, client)
	}

	// Iterate all
	count := 0
	cm.ForEachPlayer(func(p *model.Player, c *GameClient) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("ForEachPlayer iterated %d players, want 3", count)
	}
}

func TestClientManager_Unregister_RemovesPlayerMapping(t *testing.T) {
	cm := NewClientManager()

	conn := testutil.NewMockConn()
	client, _ := NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	client.SetAccountName("testaccount")

	player, _ := model.NewPlayer(1, 1, 1, "TestPlayer", 10, 0, 1)

	cm.Register("testaccount", client)
	cm.RegisterPlayer(player, client)

	// Verify player registered
	if cm.PlayerCount() != 1 {
		t.Fatal("Player should be registered")
	}

	// Unregister client (should also remove player mapping)
	cm.Unregister("testaccount")

	// Verify player mapping removed
	if cm.PlayerCount() != 0 {
		t.Error("Player mapping should be removed when client unregistered")
	}
}
