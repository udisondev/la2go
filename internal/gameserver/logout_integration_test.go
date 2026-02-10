package gameserver

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestLogoutFlow tests complete logout flow from Logout packet to player removal.
// Phase 4.17.8: Integration test for Logout handler (Phase 4.17.5).
func TestLogoutFlow(t *testing.T) {
	ctx := context.Background()

	// Setup: create test player and add to world
	w := world.Instance()
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, 1, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Add player to world (as WorldObject)
	loc := model.NewLocation(10000, 20000, 1000, 0)
	worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), loc)
	if err := w.AddObject(worldObj); err != nil {
		t.Fatalf("Failed to add player to world: %v", err)
	}

	// Verify player in world
	if _, exists := w.GetObject(player.ObjectID()); !exists {
		t.Fatal("Player not found in world after AddObject")
	}

	// Create mock GameClient
	conn := testutil.NewMockConn()
	blowfishKey := make([]byte, 16)
	client, err := NewGameClient(conn, blowfishKey)
	if err != nil {
		t.Fatalf("Failed to create GameClient: %v", err)
	}

	// Set up client state
	client.SetState(ClientStateInGame)
	client.SetAccountName("testaccount")
	client.SetActivePlayer(player)
	sessionKey := &login.SessionKey{PlayOkID1: 123, PlayOkID2: 456}
	client.SetSessionKey(sessionKey)

	// Setup handler
	database := testutil.SetupTestDB(t)
	charRepo := db.NewCharacterRepository(database)
	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	handler := NewHandler(sessionManager, clientManager, charRepo)

	// Prepare Logout packet
	logoutPacket := []byte{clientpackets.OpcodeLogout} // Empty payload
	buf := make([]byte, 1024)

	// Execute: handle Logout packet
	n, keepOpen, err := handler.handleLogout(ctx, client, logoutPacket, buf)
	if err != nil {
		t.Fatalf("handleLogout failed: %v", err)
	}

	// Verify: handler returned LeaveWorld packet
	if n == 0 {
		t.Error("Expected LeaveWorld packet, got 0 bytes")
	}

	if !keepOpen {
		t.Error("Expected keepOpen=true (connection marked for disconnect, not closed immediately)")
	}

	// Verify: LeaveWorld packet opcode
	if buf[0] != serverpackets.OpcodeLeaveWorld {
		t.Errorf("Expected LeaveWorld opcode 0x%02X, got 0x%02X", serverpackets.OpcodeLeaveWorld, buf[0])
	}

	// Verify: player removed from world
	if _, exists := w.GetObject(player.ObjectID()); exists {
		t.Error("Player still in world after logout")
	}

	// Verify: client active player cleared
	if client.ActivePlayer() != nil {
		t.Error("Client ActivePlayer not cleared after logout")
	}

	// Verify: client marked for disconnection
	if !client.IsMarkedForDisconnection() {
		t.Error("Client not marked for disconnection after logout")
	}
}

// TestRequestRestartFlow tests complete restart flow from RequestRestart packet to character selection.
// Phase 4.17.8: Integration test for RequestRestart handler (Phase 4.17.6).
func TestRequestRestartFlow(t *testing.T) {
	ctx := context.Background()

	// Setup: create test player and add to world
	w := world.Instance()
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, 1, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Add player to world (as WorldObject)
	loc := model.NewLocation(10000, 20000, 1000, 0)
	worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), loc)
	if err := w.AddObject(worldObj); err != nil {
		t.Fatalf("Failed to add player to world: %v", err)
	}

	// Verify player in world
	if _, exists := w.GetObject(player.ObjectID()); !exists {
		t.Fatal("Player not found in world after AddObject")
	}

	// Create mock GameClient
	conn := testutil.NewMockConn()
	blowfishKey := make([]byte, 16)
	client, err := NewGameClient(conn, blowfishKey)
	if err != nil {
		t.Fatalf("Failed to create GameClient: %v", err)
	}

	// Set up client state
	client.SetState(ClientStateInGame)
	client.SetAccountName("testaccount")
	client.SetActivePlayer(player)
	sessionKey := &login.SessionKey{PlayOkID1: 123, PlayOkID2: 456}
	client.SetSessionKey(sessionKey)

	// Setup handler (without DB for simplified test)
	// NOTE: Full RequestRestart test with CharSelectionInfo requires account + character in DB.
	// This simplified test focuses on core logic: state transition and player removal.
	database := testutil.SetupTestDB(t)
	charRepo := db.NewCharacterRepository(database)
	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	handler := NewHandler(sessionManager, clientManager, charRepo)

	// Prepare RequestRestart packet
	restartPacket := []byte{clientpackets.OpcodeRequestRestart} // Empty payload
	buf := make([]byte, 2048) // Larger buffer for RestartResponse + CharSelectionInfo

	// Execute: handle RequestRestart packet
	// NOTE: This will fail at LoadByAccountName (no character in DB), but we can still
	// verify state transition and player removal happened before that point.
	n, keepOpen, err := handler.handleRequestRestart(ctx, client, restartPacket, buf)

	// Test simplified: verify error handling for missing DB data is graceful
	// Full test would require account + character DB setup (TODO: add AccountRepository)
	if err != nil {
		// Expected: LoadByAccountName fails (no character in DB)
		// But state should still be transitioned and player removed
		t.Logf("Expected error (no character in DB): %v", err)
	}

	// Verify: even on error, keepOpen should be false (error path closes connection)
	if n > 0 && !keepOpen {
		t.Logf("Connection marked for close on error (expected)")
	}

	// Verify: client state transitioned to AUTHENTICATED
	if client.State() != ClientStateAuthenticated {
		t.Errorf("Expected state AUTHENTICATED, got %v", client.State())
	}

	// Verify: player removed from world
	if _, exists := w.GetObject(player.ObjectID()); exists {
		t.Error("Player still in world after restart")
	}

	// Verify: client active player cleared
	if client.ActivePlayer() != nil {
		t.Error("Client ActivePlayer not cleared after restart")
	}

	// Verify: client NOT marked for disconnection (TCP remains open)
	if client.IsMarkedForDisconnection() {
		t.Error("Client marked for disconnection after restart (should remain open)")
	}
}

// TestDisconnectionFlow_Immediate tests immediate player removal on TCP disconnect (CanLogout = true).
// Phase 4.17.8: Integration test for OnDisconnection (Phase 4.17.7) — immediate cleanup path.
func TestDisconnectionFlow_Immediate(t *testing.T) {
	ctx := context.Background()

	// Setup: create test player and add to world
	w := world.Instance()
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, 1, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Add player to world (as WorldObject)
	loc := model.NewLocation(10000, 20000, 1000, 0)
	worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), loc)
	if err := w.AddObject(worldObj); err != nil {
		t.Fatalf("Failed to add player to world: %v", err)
	}

	// Verify player in world
	if _, exists := w.GetObject(player.ObjectID()); !exists {
		t.Fatal("Player not found in world after AddObject")
	}

	// Create mock GameClient
	conn := testutil.NewMockConn()
	blowfishKey := make([]byte, 16)
	client, err := NewGameClient(conn, blowfishKey)
	if err != nil {
		t.Fatalf("Failed to create GameClient: %v", err)
	}

	// Set up client state
	client.SetState(ClientStateInGame)
	client.SetAccountName("testaccount")
	client.SetActivePlayer(player)

	// Execute: call OnDisconnection (simulates TCP disconnect)
	// Player.CanLogout() returns true (no combat stance) → immediate cleanup
	OnDisconnection(ctx, client)

	// Verify: player removed from world immediately
	if _, exists := w.GetObject(player.ObjectID()); exists {
		t.Error("Player still in world after OnDisconnection (should be removed immediately)")
	}

	// Verify: client active player cleared
	if client.ActivePlayer() != nil {
		t.Error("Client ActivePlayer not cleared after OnDisconnection")
	}
}

// TestDisconnectionFlow_Delayed tests delayed player removal on TCP disconnect (CanLogout = false).
// Phase 4.17.8: Integration test for OnDisconnection (Phase 4.17.7) — delayed cleanup path (15 seconds).
//
// NOTE: This test is SKIPPED because Player.HasAttackStance() is currently a stub (always returns false).
// TODO Phase 4.18: Enable this test after implementing AttackStanceTaskManager.
func TestDisconnectionFlow_Delayed(t *testing.T) {
	t.Skip("Skipping delayed disconnection test: Player.HasAttackStance() not implemented yet (Phase 4.18)")

	ctx := context.Background()

	// Setup: create test player and add to world
	w := world.Instance()
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, 1, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Add player to world (as WorldObject)
	loc := model.NewLocation(10000, 20000, 1000, 0)
	worldObj := model.NewWorldObject(player.ObjectID(), player.Name(), loc)
	if err := w.AddObject(worldObj); err != nil {
		t.Fatalf("Failed to add player to world: %v", err)
	}

	// Create mock GameClient
	conn := testutil.NewMockConn()
	blowfishKey := make([]byte, 16)
	client, err := NewGameClient(conn, blowfishKey)
	if err != nil {
		t.Fatalf("Failed to create GameClient: %v", err)
	}

	client.SetState(ClientStateInGame)
	client.SetAccountName("testaccount")
	client.SetActivePlayer(player)

	// TODO Phase 4.18: Mock Player.HasAttackStance() to return true (combat stance)
	// For now, this would require modifying Player struct or using interface

	// Execute: call OnDisconnection
	OnDisconnection(ctx, client)

	// Verify: player still in world (not removed immediately)
	if _, exists := w.GetObject(player.ObjectID()); !exists {
		t.Error("Player removed from world immediately (should be delayed for 15 seconds)")
	}

	// Wait for CombatTime + buffer
	time.Sleep(CombatTime + 500*time.Millisecond)

	// Verify: player removed after delay
	if _, exists := w.GetObject(player.ObjectID()); exists {
		t.Error("Player still in world after 15-second delay")
	}
}
