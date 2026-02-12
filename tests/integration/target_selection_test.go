package integration

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestTargetSelection_Success tests successful target selection flow.
// Scenario: Player clicks on nearby object (within 2000 units).
// Expected: MyTargetSelected + StatusUpdate packets, player.Target() set.
func TestTargetSelection_Success(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, &noopCharRepo{}, &noopPersister{})

	// Get world instance
	worldInst := world.Instance()

	// Create player at origin
	playerOID := nextOID()
	player, err := model.NewPlayer(playerOID, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Add player to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	// Create target object nearby (500 units away)
	targetOID := nextOID()
	targetObj := model.NewWorldObject(targetOID, "TargetNPC", model.NewLocation(500, 0, 0, 0))
	if err := worldInst.AddObject(targetObj); err != nil {
		t.Fatalf("AddObject target failed: %v", err)
	}
	defer worldInst.RemoveObject(targetObj.ObjectID())

	// Create GameClient
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create RequestAction packet (simple click, not attack)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeRequestAction)
	w.WriteInt(int32(targetObj.ObjectID())) // objectID
	w.WriteInt(0)                            // originX
	w.WriteInt(0)                            // originY
	w.WriteInt(0)                            // originZ
	w.WriteByte(0)                           // actionType (simple click)
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: response packets sent
	if n == 0 {
		t.Fatal("Expected response packets (MyTargetSelected + StatusUpdate), got empty")
	}

	// Parse first packet (should be MyTargetSelected)
	if buf[0] != serverpackets.OpcodeMyTargetSelected {
		t.Errorf("Expected first packet opcode=0x%02X (MyTargetSelected), got 0x%02X",
			serverpackets.OpcodeMyTargetSelected, buf[0])
	}

	// Note: We don't parse StatusUpdate here because target is WorldObject (not Character)
	// so StatusUpdate won't be sent. This is expected behavior.

	// Verify: player target set
	if !player.HasTarget() {
		t.Error("Expected player.HasTarget()=true after target selection, got false")
	}
	if player.Target() == nil {
		t.Fatal("Expected player.Target() non-nil, got nil")
	}
	if player.Target().ObjectID() != targetObj.ObjectID() {
		t.Errorf("Expected player.Target().ObjectID()=%d, got %d",
			targetObj.ObjectID(), player.Target().ObjectID())
	}
}

// TestTargetSelection_TooFar tests target selection failure when target is too far.
// Scenario: Player clicks on object 3000 units away (max is 2000).
// Expected: Silent failure, no response packets, target NOT set.
func TestTargetSelection_TooFar(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, &noopCharRepo{}, &noopPersister{})

	// Get world instance
	worldInst := world.Instance()

	// Create player at origin
	playerOID := nextOID()
	player, err := model.NewPlayer(playerOID, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Add player to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	// Create target object far away (3000 units, beyond 2000 limit)
	targetOID := nextOID()
	targetObj := model.NewWorldObject(targetOID, "FarTarget", model.NewLocation(3000, 0, 0, 0))
	if err := worldInst.AddObject(targetObj); err != nil {
		t.Fatalf("AddObject target failed: %v", err)
	}
	defer worldInst.RemoveObject(targetObj.ObjectID())

	// Create GameClient
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create RequestAction packet
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeRequestAction)
	w.WriteInt(int32(targetObj.ObjectID())) // objectID
	w.WriteInt(0)                            // originX
	w.WriteInt(0)                            // originY
	w.WriteInt(0)                            // originZ
	w.WriteByte(0)                           // actionType
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: no response (silent failure)
	if n != 0 {
		t.Errorf("Expected no response for target too far, got %d bytes", n)
	}

	// Verify: player target NOT set
	if player.HasTarget() {
		t.Error("Expected player.HasTarget()=false after failed target selection, got true")
	}
	if player.Target() != nil {
		t.Error("Expected player.Target()=nil after failed target selection, got non-nil")
	}
}

// TestTargetSelection_NonExistent tests target selection failure when target doesn't exist.
// Scenario: Player clicks on non-existent objectID.
// Expected: Silent failure, no response, target NOT set.
func TestTargetSelection_NonExistent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, &noopCharRepo{}, &noopPersister{})

	// Create player
	playerOID := nextOID()
	player, err := model.NewPlayer(playerOID, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create GameClient
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create RequestAction packet with non-existent objectID
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeRequestAction)
	w.WriteInt(999999) // non-existent objectID
	w.WriteInt(0)      // originX
	w.WriteInt(0)      // originY
	w.WriteInt(0)      // originZ
	w.WriteByte(0)     // actionType
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: no response (silent failure)
	if n != 0 {
		t.Errorf("Expected no response for non-existent target, got %d bytes", n)
	}

	// Verify: player target NOT set
	if player.HasTarget() {
		t.Error("Expected player.HasTarget()=false, got true")
	}
}

// TestTargetSelection_AttackIntent tests attack intent detection (shift+click).
// Scenario: Player shift+clicks on target (actionType=1).
// Expected: Target set, MyTargetSelected sent, attack intent logged.
// Note: Auto-attack NOT implemented yet (TODO Phase 5.3).
func TestTargetSelection_AttackIntent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, &noopCharRepo{}, &noopPersister{})

	// Get world instance
	worldInst := world.Instance()

	// Create player
	playerOID := nextOID()
	player, err := model.NewPlayer(playerOID, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Add player to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	// Create target
	targetOID := nextOID()
	targetObj := model.NewWorldObject(targetOID, "Enemy", model.NewLocation(100, 0, 0, 0))
	if err := worldInst.AddObject(targetObj); err != nil {
		t.Fatalf("AddObject target failed: %v", err)
	}
	defer worldInst.RemoveObject(targetObj.ObjectID())

	// Create GameClient
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16), nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create RequestAction packet with attack intent (shift+click)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeRequestAction)
	w.WriteInt(int32(targetObj.ObjectID())) // objectID
	w.WriteInt(0)                            // originX
	w.WriteInt(0)                            // originY
	w.WriteInt(0)                            // originZ
	w.WriteByte(1)                           // actionType=1 (shift+click, attack intent)
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: response sent (MyTargetSelected)
	if n == 0 {
		t.Fatal("Expected MyTargetSelected response, got empty")
	}

	// Verify: target set
	if !player.HasTarget() {
		t.Error("Expected player.HasTarget()=true, got false")
	}
	if player.Target() == nil {
		t.Fatal("Expected player.Target() non-nil, got nil")
	}
	if player.Target().ObjectID() != targetObj.ObjectID() {
		t.Errorf("Expected target objectID=%d, got %d",
			targetObj.ObjectID(), player.Target().ObjectID())
	}

	// Note: Auto-attack NOT implemented yet (Phase 5.3)
	// This test verifies target selection works with attack intent flag
}
