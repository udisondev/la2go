package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/testutil"
)

// GameServerSuite tests GameServer with real TCP connections.
// Requires LoginServer running (for SessionManager integration).
type GameServerSuite struct {
	IntegrationSuite
	loginServer *login.Server
	gameServer  *gameserver.Server
	loginAddr   string // LoginServer address
	gameAddr    string // GameServer address
}

// SetupSuite initializes LoginServer and GameServer.
func (s *GameServerSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Create LoginServer config
	loginCfg := config.LoginServer{
		BindAddress:        "127.0.0.1",
		Port:               0, // random port
		AutoCreateAccounts: true,
		ShowLicence:        true, // Required for testutil.LoginClient.ReadLoginOk()
		GameServers: []config.GameServerEntry{
			{ID: 1, Name: "Bartz"},
		},
	}

	var err error
	s.loginServer, err = login.NewServer(loginCfg, s.db)
	if err != nil {
		s.T().Fatalf("failed to create login server: %v", err)
	}

	// Create GameServer config
	gameCfg := config.GameServer{
		BindAddress: "127.0.0.1",
		Port:        0, // random port
		ServerID:    1,
		HexID:       "c0a80001",
	}

	// Create CharacterRepository (Phase 4.6)
	charRepo := db.NewCharacterRepository(s.db.Pool())

	// Create GameServer with shared SessionManager
	s.gameServer, err = gameserver.NewServer(gameCfg, s.loginServer.SessionManager(), charRepo, &noopPersister{})
	if err != nil {
		s.T().Fatalf("failed to create game server: %v", err)
	}

	// Start LoginServer
	loginListener, loginAddr := testutil.ListenTCP(s.T())
	s.loginAddr = loginAddr

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	s.T().Cleanup(cancel)
	s.T().Cleanup(func() {
		_ = s.loginServer.Close()
	})

	t := s.T()
	go func() {
		if err := s.loginServer.Serve(ctx, loginListener); err != nil && err != context.Canceled {
			t.Logf("login server error: %v", err)
		}
	}()

	// Wait for LoginServer to start
	if err := testutil.WaitForTCPReady(s.loginAddr, 5*time.Second); err != nil {
		s.T().Fatalf("login server failed to start: %v", err)
	}

	// Start GameServer
	gameListener, gameAddr := testutil.ListenTCP(s.T())
	s.gameAddr = gameAddr

	s.T().Cleanup(func() {
		_ = s.gameServer.Close()
	})

	go func() {
		if err := s.gameServer.Serve(ctx, gameListener); err != nil && err != context.Canceled {
			t.Logf("game server error: %v", err)
		}
	}()

	// Wait for GameServer to start
	if err := testutil.WaitForTCPReady(s.gameAddr, 5*time.Second); err != nil {
		s.T().Fatalf("game server failed to start: %v", err)
	}
}

// TearDownSuite stops servers.
func (s *GameServerSuite) TearDownSuite() {
	s.IntegrationSuite.TearDownSuite()
}

// TestClientConnection tests that a client can connect and receive KeyPacket.
func (s *GameServerSuite) TestClientConnection() {
	client, err := testutil.NewGameClient(s.T(), s.gameAddr)
	s.Require().NoError(err, "failed to connect to game server")
	defer client.Close()

	// KeyPacket already read in NewGameClient
	s.NotNil(client.Encryption(), "encryption should be initialized")
}

// TestProtocolVersion tests that the server validates protocol version.
func (s *GameServerSuite) TestProtocolVersion() {
	client, err := testutil.NewGameClient(s.T(), s.gameAddr)
	s.Require().NoError(err)
	defer client.Close()

	// Send valid protocol version (0x0106)
	err = client.SendProtocolVersion(constants.ProtocolRevisionInterlude)
	s.NoError(err, "failed to send protocol version")

	// Server should not disconnect (no response expected)
	// Try to send another packet to verify connection is still alive
	err = client.SendProtocolVersion(constants.ProtocolRevisionInterlude)
	s.NoError(err, "connection should still be alive after valid protocol version")
}

// TestProtocolVersionInvalid tests that the server rejects invalid protocol version.
func (s *GameServerSuite) TestProtocolVersionInvalid() {
	client, err := testutil.NewGameClient(s.T(), s.gameAddr)
	s.Require().NoError(err)
	defer client.Close()

	// Send invalid protocol version
	err = client.SendProtocolVersion(0x9999)
	s.NoError(err, "send should succeed")

	// Server should close connection
	// Try to read a packet — should get EOF or error
	_, err = client.ReadPacket()
	s.Error(err, "server should close connection after invalid protocol version")
}

// TestAuthLoginWithValidSessionKey tests full authentication flow:
// 1. Authenticate on LoginServer to get SessionKey
// 2. Connect to GameServer
// 3. Send AuthLogin with SessionKey
// 4. Verify authentication succeeds
func (s *GameServerSuite) TestAuthLoginWithValidSessionKey() {
	// Step 1: Authenticate on LoginServer to get SessionKey
	loginClient, err := testutil.NewLoginClient(s.T(), s.loginAddr)
	s.Require().NoError(err, "failed to connect to login server")
	defer loginClient.Close()

	// Full login flow
	err = loginClient.SendAuthGameGuard()
	s.Require().NoError(err)
	err = loginClient.ReadGGAuth()
	s.Require().NoError(err)

	accountName := "testuser_gs"
	err = loginClient.SendRequestAuthLogin(accountName, "testpass")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := loginClient.ReadLoginOk()
	s.Require().NoError(err)
	s.NotZero(loginOkID1, "loginOkID1 should be set")
	s.NotZero(loginOkID2, "loginOkID2 should be set")

	// Get SessionKey from SessionManager (for testing)
	// In real scenario, client would get playOkID1/playOkID2 from PlayOk packet
	// For test, we construct SessionKey from loginOkID1/loginOkID2
	sessionKey := login.SessionKey{
		PlayOkID1:  0, // Not set in Phase 4.1 (no PlayOk packet yet)
		PlayOkID2:  0,
		LoginOkID1: loginOkID1,
		LoginOkID2: loginOkID2,
	}

	// Step 2: Connect to GameServer
	gameClient, err := testutil.NewGameClient(s.T(), s.gameAddr)
	s.Require().NoError(err, "failed to connect to game server")
	defer gameClient.Close()

	// Send ProtocolVersion
	err = gameClient.SendProtocolVersion(constants.ProtocolRevisionInterlude)
	s.Require().NoError(err)

	// Step 3: Send AuthLogin with SessionKey
	err = gameClient.SendAuthLogin(accountName, sessionKey)
	s.Require().NoError(err, "failed to send AuthLogin")

	// Step 4: Verify authentication succeeds
	// In Phase 4.1, GameServer doesn't send a response packet after AuthLogin
	// We verify success by checking that the connection stays alive
	// Try to send another packet — connection should still be open
	err = gameClient.SendProtocolVersion(constants.ProtocolRevisionInterlude)
	s.NoError(err, "connection should remain open after successful AuthLogin")
}

// TestAuthLoginWithInvalidSessionKey tests that invalid SessionKey is rejected.
func (s *GameServerSuite) TestAuthLoginWithInvalidSessionKey() {
	// Connect to GameServer (no LoginServer authentication)
	gameClient, err := testutil.NewGameClient(s.T(), s.gameAddr)
	s.Require().NoError(err)
	defer gameClient.Close()

	// Send ProtocolVersion
	err = gameClient.SendProtocolVersion(constants.ProtocolRevisionInterlude)
	s.Require().NoError(err)

	// Send AuthLogin with invalid SessionKey
	invalidSessionKey := login.SessionKey{
		PlayOkID1:  0x11111111,
		PlayOkID2:  0x22222222,
		LoginOkID1: 0x33333333,
		LoginOkID2: 0x44444444,
	}

	err = gameClient.SendAuthLogin("fakeuser", invalidSessionKey)
	s.Require().NoError(err, "send should succeed")

	// Server should close connection due to invalid SessionKey
	_, err = gameClient.ReadPacket()
	s.Error(err, "server should close connection after invalid SessionKey")
}

// TestMultipleClients tests that multiple clients can connect simultaneously.
func (s *GameServerSuite) TestMultipleClients() {
	const clientCount = 5

	clients := make([]*testutil.GameClient, clientCount)
	for i := range clientCount {
		client, err := testutil.NewGameClient(s.T(), s.gameAddr)
		s.Require().NoError(err, "client %d failed to connect", i)
		clients[i] = client

		// Send ProtocolVersion
		err = client.SendProtocolVersion(constants.ProtocolRevisionInterlude)
		s.NoError(err, "client %d failed to send protocol version", i)
	}

	// Clean up
	for i, client := range clients {
		err := client.Close()
		s.NoError(err, "client %d failed to close", i)
	}
}

// TestIntegrationGameServer runs the GameServerSuite.
func TestIntegrationGameServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	t.Parallel()

	suite.Run(t, new(GameServerSuite))
}
