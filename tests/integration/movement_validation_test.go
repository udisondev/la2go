package integration

import (
	"context"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

// TestMoveToLocation_ValidationReject tests that invalid movements are rejected.
// Scenario: Player tries to move to invalid Z coordinate (-25000).
// Expected: ValidateLocation response, position NOT updated.
// Note: StopMove broadcast not tested (requires World/VisibilityManager setup).
func TestMoveToLocation_ValidationReject(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup repositories
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()

	// Create handler (sessionManager = nil for this test)
	handler := gameserver.NewHandler(nil, clientMgr, charRepo)

	// Create test player at origin
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create GameClient
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16)) // 16-byte Blowfish key
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create MoveToLocation packet with invalid Z (-25000)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeMoveToLocation)
	w.WriteInt(10000)  // targetX
	w.WriteInt(10000)  // targetY
	w.WriteInt(-25000) // targetZ (INVALID — below -20000)
	w.WriteInt(0)      // originX
	w.WriteInt(0)      // originY
	w.WriteInt(0)      // originZ
	w.WriteInt(0)      // moveType
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

	// Verify: response packet is ValidateLocation
	if n == 0 {
		t.Fatal("Expected ValidateLocation response, got empty")
	}

	respOpcode := buf[0]
	if respOpcode != serverpackets.OpcodeValidateLocation {
		t.Errorf("Expected ValidateLocation opcode (0x%02X), got 0x%02X",
			serverpackets.OpcodeValidateLocation, respOpcode)
	}

	// Verify: player position NOT updated (still at origin)
	loc := player.Location()
	if loc.X != 0 || loc.Y != 0 || loc.Z != 0 {
		t.Errorf("Player position updated incorrectly: (%d,%d,%d), expected (0,0,0)",
			loc.X, loc.Y, loc.Z)
	}
}

// TestMoveToLocation_TeleportReject tests that teleportation attempts are rejected.
// Scenario: Player tries to move 10000 units away (max allowed: 9900).
// Expected: ValidateLocation response, position NOT updated.
func TestMoveToLocation_TeleportReject(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo)

	// Create player at origin
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create client
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16)) // 16-byte Blowfish key
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create MoveToLocation packet with teleportation (10000 units)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeMoveToLocation)
	w.WriteInt(10000) // targetX (10000 units away — TOO FAR)
	w.WriteInt(0)     // targetY
	w.WriteInt(0)     // targetZ
	w.WriteInt(0)     // originX
	w.WriteInt(0)     // originY
	w.WriteInt(0)     // originZ
	w.WriteInt(0)     // moveType
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}
	if n == 0 {
		t.Fatal("Expected ValidateLocation response, got empty")
	}

	// Verify ValidateLocation response
	respOpcode := buf[0]
	if respOpcode != serverpackets.OpcodeValidateLocation {
		t.Errorf("Expected ValidateLocation opcode (0x%02X), got 0x%02X",
			serverpackets.OpcodeValidateLocation, respOpcode)
	}

	// Verify position NOT changed
	loc := player.Location()
	if loc.X != 0 || loc.Y != 0 || loc.Z != 0 {
		t.Errorf("Player position changed: (%d,%d,%d), expected (0,0,0)",
			loc.X, loc.Y, loc.Z)
	}
}

// TestValidatePosition_DesyncCorrection tests desync detection and correction.
// Scenario: Client reports position 600 units away from server (above threshold).
// Expected: ValidateLocation response with server position.
func TestValidatePosition_DesyncCorrection(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo)

	// Create player at (0,0,0)
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create client
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16)) // 16-byte Blowfish key
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create ValidatePosition packet with desynced position (600 units away)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeValidatePosition)
	w.WriteInt(600) // clientX (600 units away — above 500 threshold)
	w.WriteInt(0)   // clientY
	w.WriteInt(0)   // clientZ
	w.WriteInt(0)   // heading
	w.WriteInt(0)   // vehicleID
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}
	if n == 0 {
		t.Fatal("Expected ValidateLocation response, got empty")
	}

	// Verify ValidateLocation response
	respOpcode := buf[0]
	if respOpcode != serverpackets.OpcodeValidateLocation {
		t.Errorf("Expected ValidateLocation opcode (0x%02X), got 0x%02X",
			serverpackets.OpcodeValidateLocation, respOpcode)
	}

	// Verify client position updated in Movement tracker
	clientX, clientY, clientZ, _ := player.Movement().ClientPosition()
	if clientX != 600 || clientY != 0 || clientZ != 0 {
		t.Errorf("Client position not updated: (%d,%d,%d), expected (600,0,0)",
			clientX, clientY, clientZ)
	}

	// Verify server position unchanged
	loc := player.Location()
	if loc.X != 0 || loc.Y != 0 || loc.Z != 0 {
		t.Errorf("Server position changed: (%d,%d,%d), expected (0,0,0)",
			loc.X, loc.Y, loc.Z)
	}
}

// TestValidatePosition_AbnormalZ tests Z-coordinate boundary enforcement.
// Scenario: Client reports abnormal Z (-25000, below -20000 limit).
// Expected: ValidateLocation response, player teleported to lastServerPosition.
func TestValidatePosition_AbnormalZ(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo)

	// Create player at (1000,1000,0)
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(1000, 1000, 0, 0))
	player.Movement().SetLastServerPosition(1000, 1000, 0)

	// Create client
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16)) // 16-byte Blowfish key
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Create ValidatePosition packet with abnormal Z
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeValidatePosition)
	w.WriteInt(1000)   // clientX
	w.WriteInt(1000)   // clientY
	w.WriteInt(-25000) // clientZ (ABNORMAL — below -20000)
	w.WriteInt(0)      // heading
	w.WriteInt(0)      // vehicleID
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}
	if n == 0 {
		t.Fatal("Expected ValidateLocation response, got empty")
	}

	// Verify ValidateLocation response
	respOpcode := buf[0]
	if respOpcode != serverpackets.OpcodeValidateLocation {
		t.Errorf("Expected ValidateLocation opcode (0x%02X), got 0x%02X",
			serverpackets.OpcodeValidateLocation, respOpcode)
	}

	// Verify player teleported to lastServerPosition (1000,1000,0)
	loc := player.Location()
	if loc.X != 1000 || loc.Y != 1000 || loc.Z != 0 {
		t.Errorf("Player position not reset: (%d,%d,%d), expected (1000,1000,0)",
			loc.X, loc.Y, loc.Z)
	}
}

// TestMoveToLocation_NormalFlow tests valid movement flow.
// Scenario: Player moves 1000 units (valid distance).
// Expected: CharMoveToLocation broadcast, position updated, lastServerPosition updated.
func TestMoveToLocation_NormalFlow(t *testing.T) {
	dbConn := testutil.SetupTestDB(t)
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup
	charRepo := db.NewCharacterRepository(dbConn)
	clientMgr := gameserver.NewClientManager()
	handler := gameserver.NewHandler(nil, clientMgr, charRepo)

	// Create player at origin
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create client
	conn := testutil.NewMockConn()
	client, err := gameserver.NewGameClient(conn, make([]byte, 16)) // 16-byte Blowfish key
	if err != nil {
		t.Fatalf("NewGameClient failed: %v", err)
	}
	client.SetState(gameserver.ClientStateInGame)
	client.SetActivePlayer(player)
	client.SetAccountName("testaccount")

	// Register client in ClientManager (for broadcast)
	clientMgr.RegisterPlayer(player, client)

	// Create MoveToLocation packet (valid move: 1000 units)
	w := packet.NewWriter(64)
	w.WriteByte(clientpackets.OpcodeMoveToLocation)
	w.WriteInt(1000) // targetX
	w.WriteInt(0)    // targetY
	w.WriteInt(0)    // targetZ
	w.WriteInt(0)    // originX
	w.WriteInt(0)    // originY
	w.WriteInt(0)    // originZ
	w.WriteInt(0)    // moveType
	pktData := w.Bytes()

	// Process packet
	buf := make([]byte, 1024)
	n, keepAlive, err := handler.HandlePacket(ctx, client, pktData, buf)

	// Verify
	if err != nil {
		t.Errorf("HandlePacket returned error: %v", err)
	}
	if !keepAlive {
		t.Errorf("Expected keepAlive=true, got false")
	}

	// MoveToLocation does NOT send response to client (client-predicted)
	if n != 0 {
		t.Errorf("Expected no response, got %d bytes", n)
	}

	// Verify player position updated
	loc := player.Location()
	if loc.X != 1000 || loc.Y != 0 || loc.Z != 0 {
		t.Errorf("Player position not updated: (%d,%d,%d), expected (1000,0,0)",
			loc.X, loc.Y, loc.Z)
	}

	// Verify lastServerPosition updated
	lastX, lastY, lastZ := player.Movement().LastServerPosition()
	if lastX != 1000 || lastY != 0 || lastZ != 0 {
		t.Errorf("LastServerPosition not updated: (%d,%d,%d), expected (1000,0,0)",
			lastX, lastY, lastZ)
	}
}
