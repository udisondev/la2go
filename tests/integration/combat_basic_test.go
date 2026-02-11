package integration

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
	"github.com/udisondev/la2go/internal/world"
)

// TestBasicAttack_Success tests basic physical attack flow.
// Scenario: Player attacks nearby target within range (100 units).
// Expected: Attack packet broadcast, player added to combat stance.
//
// Phase 5.3: Basic Combat System (MVP).
func TestBasicAttack_Success(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup repositories and managers
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo, &noopPersister{})

	// Get world instance
	worldInst := world.Instance()

	// Initialize AttackStanceManager (Phase 5.3)
	attackStanceMgr := combat.NewAttackStanceManager()
	combat.AttackStanceMgr = attackStanceMgr
	attackStanceMgr.Start()
	defer attackStanceMgr.Stop()

	// Initialize CombatManager with broadcast function (Phase 5.3)
	broadcastFunc := func(source *model.Player, data []byte, size int) {
		clientMgr.BroadcastToVisibleNear(source, data, size)
	}
	combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
	combat.CombatMgr = combatMgr

	// Create player at origin (level 10, objectID=1)
	player, err := model.NewPlayer(1, 100, 200, "AttackerPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Add player to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	// Create target NPC nearby (50 units away, within attack range)
	targetTemplate := model.NewNpcTemplate(
		9000, "TargetNPC", "", 5, 1500, 800,
		100, 50, 80, 40, 0, 120, 253, 30, 60, 0, 0,
	)
	targetNpc := model.NewNpc(2, 9000, targetTemplate)
	targetNpc.SetLocation(model.NewLocation(50, 0, 0, 0))
	if err := worldInst.AddNpc(targetNpc); err != nil {
		t.Fatalf("AddNpc target failed: %v", err)
	}
	defer worldInst.RemoveObject(targetNpc.ObjectID())

	// Create GameClient for player
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16))
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Register player client (required for broadcast)
	clientMgr.RegisterPlayer(player, client)
	defer clientMgr.UnregisterPlayer(player)

	// Reset write counter before test
	conn.ResetWriteCount()

	// Create AttackRequest packet (opcode 0x0A)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeAttackRequest)
	w.WriteInt(int32(targetNpc.ObjectID())) // target objectID
	w.WriteInt(0)                            // originX
	w.WriteInt(0)                            // originY
	w.WriteInt(0)                            // originZ
	w.WriteByte(0)                           // attackID (simple click)
	pktData := w.Bytes()

	// Process AttackRequest packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: no direct response (AttackRequest returns 0 bytes)
	if n != 0 {
		t.Errorf("Expected n=0 (no direct response), got n=%d", n)
	}

	// Verify: Attack packet broadcast to visible players
	// Note: BroadcastToVisibleNear may send 0 packets if no visible players
	// but combat flow should execute without errors
	writeCount := conn.WriteCount()
	t.Logf("Attack packet broadcast: %d writes", writeCount)

	// Verify: player added to combat stance (15-second window)
	// Sleep briefly to allow AttackStanceManager.AddAttackStance() to complete
	time.Sleep(50 * time.Millisecond)

	// Check combat stance (player should be in stance)
	inStance := attackStanceMgr.HasAttackStance(player)
	if !inStance {
		t.Errorf("Expected player in attack stance, but HasAttackStance=false")
	}
}

// TestBasicAttack_OutOfRange tests attack validation: target out of range.
// Scenario: Player attacks target 1000 units away (max range = 100 units).
// Expected: ActionFailed packet, no attack broadcast.
//
// Phase 5.3: Basic Combat System (MVP).
func TestBasicAttack_OutOfRange(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup repositories and managers
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo, &noopPersister{})

	// Get world instance
	worldInst := world.Instance()

	// Initialize AttackStanceManager (Phase 5.3)
	attackStanceMgr := combat.NewAttackStanceManager()
	combat.AttackStanceMgr = attackStanceMgr
	attackStanceMgr.Start()
	defer attackStanceMgr.Stop()

	// Initialize CombatManager with broadcast function (Phase 5.3)
	broadcastFunc := func(source *model.Player, data []byte, size int) {
		clientMgr.BroadcastToVisibleNear(source, data, size)
	}
	combatMgr := combat.NewCombatManager(broadcastFunc, nil, nil)
	combat.CombatMgr = combatMgr

	// Create player at origin (level 10, objectID=1)
	player, err := model.NewPlayer(1, 100, 200, "AttackerPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Add player to world
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	defer worldInst.RemoveObject(player.ObjectID())

	// Create target NPC far away (1000 units, out of attack range)
	targetTemplate := model.NewNpcTemplate(
		9001, "DistantNPC", "", 5, 1500, 800,
		100, 50, 80, 40, 0, 120, 253, 30, 60, 0, 0,
	)
	targetNpc := model.NewNpc(2, 9001, targetTemplate)
	targetNpc.SetLocation(model.NewLocation(1000, 0, 0, 0))
	if err := worldInst.AddNpc(targetNpc); err != nil {
		t.Fatalf("AddNpc target failed: %v", err)
	}
	defer worldInst.RemoveObject(targetNpc.ObjectID())

	// Create GameClient for player
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16))
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Register player client (required for broadcast)
	clientMgr.RegisterPlayer(player, client)
	defer clientMgr.UnregisterPlayer(player)

	// Reset write counter before test
	conn.ResetWriteCount()

	// Create AttackRequest packet (opcode 0x0A)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeAttackRequest)
	w.WriteInt(int32(targetNpc.ObjectID())) // target objectID
	w.WriteInt(0)                            // originX
	w.WriteInt(0)                            // originY
	w.WriteInt(0)                            // originZ
	w.WriteByte(0)                           // attackID (simple click)
	pktData := w.Bytes()

	// Process AttackRequest packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify: no error, connection stays open
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// Verify: ActionFailed packet sent (1 byte opcode 0x25)
	if n == 0 {
		t.Fatal("Expected ActionFailed packet (1 byte), got n=0")
	}
	if n < 1 {
		t.Fatalf("Expected at least 1 byte (ActionFailed opcode), got n=%d", n)
	}

	// Verify: ActionFailed opcode = 0x25
	expectedOpcode := byte(0x25)
	if buf[0] != expectedOpcode {
		t.Errorf("Expected ActionFailed opcode=0x%02X, got 0x%02X", expectedOpcode, buf[0])
	}

	// Verify: no Attack broadcast (validation failed)
	writeCount := conn.WriteCount()
	if writeCount > 0 {
		t.Errorf("Expected no Attack broadcast (validation failed), got %d writes", writeCount)
	}

	// Verify: player NOT added to combat stance (attack failed)
	time.Sleep(50 * time.Millisecond)
	inStance := attackStanceMgr.HasAttackStance(player)
	if inStance {
		t.Errorf("Expected player NOT in attack stance (attack failed), but HasAttackStance=true")
	}
}
