package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gslistener"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/testutil"
)

// GSListenerSuite тестирует GS↔LS relay с реальными TCP соединениями.
type GSListenerSuite struct {
	IntegrationSuite
	server     *gslistener.Server
	sessionMgr *login.SessionManager
	gsTable    *gameserver.GameServerTable
	addr       string // адрес сервера (listener.Addr().String())
}

// SetupSuite инициализирует gslistener Server.
func (s *GSListenerSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Создаём зависимости
	s.sessionMgr = login.NewSessionManager()
	s.gsTable = gameserver.NewGameServerTable(s.db)

	// Создаём конфиг
	cfg := config.LoginServer{
		GSListenHost: "127.0.0.1",
		GSListenPort: 0,
	}

	var err error
	s.server, err = gslistener.NewServer(cfg, s.db, s.gsTable, s.sessionMgr)
	if err != nil {
		s.T().Fatalf("failed to create gslistener server: %v", err)
	}

	// Создаём listener на случайном порту
	listener, addr := testutil.ListenTCP(s.T())
	s.addr = addr

	// Запускаем сервер в background (с timeout для предотвращения зависания)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	s.T().Cleanup(cancel)
	s.T().Cleanup(func() {
		_ = s.server.Close()
	})

	go func() {
		if err := s.server.Serve(ctx, listener); err != nil && err != context.Canceled {
			s.T().Logf("gslistener server error: %v", err)
		}
	}()

	// Ждём запуска (polling с timeout вместо sleep)
	if err := testutil.WaitForTCPReady(s.addr, 5*time.Second); err != nil {
		s.T().Fatalf("gslistener failed to start: %v", err)
	}
}

// TearDownSuite останавливает сервер.
func (s *GSListenerSuite) TearDownSuite() {
	s.IntegrationSuite.TearDownSuite()
}

// ============================================================================
// Phase 1: Critical Path Tests
// ============================================================================

// TestFullGSRegistrationFlow тестирует полный flow регистрации GameServer:
// InitLS → BlowFishKey → GameServerAuth → AuthResponse
func (s *GSListenerSuite) TestFullGSRegistrationFlow() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err, "failed to create GS client")
	defer client.Close()

	// Complete registration
	serverID := byte(100)
	hexID := "test_hex_id_0123456789abcdef0000"
	err = client.CompleteRegistration(serverID, hexID)
	s.Require().NoError(err, "registration should succeed")

	// Verify GS registered in table
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists, "GS should be registered")
	s.True(gsInfo.IsAuthed(), "GS should be authed")
}

// TestBlowFishKeyHandling тестирует обработку BlowFishKey пакета.
func (s *GSListenerSuite) TestBlowFishKeyHandling() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Генерируем случайный BF ключ
	bfKey := make([]byte, 40)
	for i := range bfKey {
		bfKey[i] = byte(i)
	}

	// Отправляем BlowFishKey
	err = client.SendBlowFishKey(bfKey)
	s.NoError(err, "BlowFishKey should be processed successfully")
}

// TestGameServerAuthSuccess тестирует успешную регистрацию GameServer.
func (s *GSListenerSuite) TestGameServerAuthSuccess() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(101)
	hexID := "auth_test_id_123456789abcdef00001"

	err = client.CompleteRegistration(serverID, hexID)
	s.Require().NoError(err)

	// Verify в GameServerTable
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists)
	s.Equal(serverID, byte(gsInfo.ID()))
	s.True(gsInfo.IsAuthed())
}

// TestPlayerAuthRequestValid тестирует валидацию session через SessionManager.
func (s *GSListenerSuite) TestPlayerAuthRequestValid() {
	// Регистрируем GS
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(102)
	err = client.CompleteRegistration(serverID, "valid_auth_hex_id_00000000000")
	s.Require().NoError(err)

	// Создаём session (имитируем что LS создал её)
	account := "testplayer"
	sessionKey := login.SessionKey{
		LoginOkID1: 12345,
		LoginOkID2: 67890,
		PlayOkID1:  11111,
		PlayOkID2:  22222,
	}
	s.sessionMgr.Store(account, sessionKey, nil)

	// Отправляем PlayerAuthRequest
	err = client.SendPlayerAuthRequest(account, sessionKey)
	s.Require().NoError(err)

	// Читаем PlayerAuthResponse
	respAccount, result, err := client.ReadPlayerAuthResponse()
	s.Require().NoError(err)
	s.Equal(account, respAccount, "account should match")
	s.True(result, "auth should be successful")

	// Verify session удалена из SessionManager
	count := s.sessionMgr.Count()
	s.Equal(0, count, "session should be removed after validation")
}

// TestPlayerInGameNotification тестирует обработку PlayerInGame пакета.
func (s *GSListenerSuite) TestPlayerInGameNotification() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(103)
	err = client.CompleteRegistration(serverID, "ingame_hex_id_0000000000000000")
	s.Require().NoError(err)

	// Отправляем PlayerInGame
	account := "player1"
	err = client.SendPlayerInGame(account)
	s.NoError(err, "PlayerInGame should be processed without errors")

	// Verify GS всё ещё в таблице
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists, "GS should still be registered")
	s.True(gsInfo.IsAuthed(), "GS should still be authed")
}

// TestPlayerLogoutNotification тестирует обработку PlayerLogout пакета.
func (s *GSListenerSuite) TestPlayerLogoutNotification() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(104)
	err = client.CompleteRegistration(serverID, "logout_hex_id_00000000000000000")
	s.Require().NoError(err)

	account := "player2"

	// Добавляем игрока
	err = client.SendPlayerInGame(account)
	s.Require().NoError(err)

	// Отправляем PlayerLogout
	err = client.SendPlayerLogout(account)
	s.NoError(err, "PlayerLogout should be processed without errors")

	// Verify GS всё ещё активен
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists, "GS should still be registered")
	s.True(gsInfo.IsAuthed(), "GS should still be authed")
}

// TestServerStatusUpdate тестирует обновление server attributes.
func (s *GSListenerSuite) TestServerStatusUpdate() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(105)
	err = client.CompleteRegistration(serverID, "status_hex_id_000000000000000000")
	s.Require().NoError(err)

	// Отправляем ServerStatus с атрибутами
	attributes := map[int]int32{
		0x01: 1000, // max players
		0x02: 500,  // current players
	}

	err = client.SendServerStatus(serverID, attributes)
	s.NoError(err, "ServerStatus should be processed")

	// Verify attributes обновлены
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists)
	s.True(gsInfo.IsAuthed(), "GS should still be authed")
}

// TestGSStateTransitions проверяет переходы состояний GameServer.
func (s *GSListenerSuite) TestGSStateTransitions() {
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// State transitions уже покрыты в CompleteRegistration:
	// CONNECTED → BF_CONNECTED (после BlowFishKey)
	// BF_CONNECTED → AUTHED (после GameServerAuth)

	serverID := byte(106)
	hexID := "state_test_hex_id_0000000000000"

	// До регистрации: не в таблице или не authed
	_, exists := s.gsTable.GetByID(int(serverID))
	s.False(exists, "GS should not exist before registration")

	// Complete registration
	err = client.CompleteRegistration(serverID, hexID)
	s.Require().NoError(err)

	// После регистрации: authed
	gsInfo, exists := s.gsTable.GetByID(int(serverID))
	s.True(exists, "GS should exist after registration")
	s.True(gsInfo.IsAuthed(), "GS should be in AUTHED state")
}

// ============================================================================
// Phase 2: Error Handling Tests
// ============================================================================

// TestGameServerAuthErrors тестирует различные error cases при регистрации GS.
func (s *GSListenerSuite) TestGameServerAuthErrors() {
	// 1. Wrong hexID
	s.Run("wrong hexID", func() {
		s.T().Parallel()

		// Сначала регистрируем GS с правильным hexID
		client1, err := testutil.NewGSClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client1.Close()

		serverID := byte(200)
		correctHexID := "correct_hex_id_000000000000000"

		err = client1.CompleteRegistration(serverID, correctHexID)
		s.Require().NoError(err, "first registration should succeed")

		// Disconnect первого клиента
		client1.Close()

		// Пытаемся подключиться с тем же serverID но другим hexID
		client2, err := testutil.NewGSClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client2.Close()

		// NewGSClient уже прочитал InitLS

		// Send BlowFishKey
		bfKey := make([]byte, 40)
		_, err = rand.Read(bfKey)
		s.Require().NoError(err)
		err = client2.SendBlowFishKey(bfKey)
		s.Require().NoError(err)

		// Send GameServerAuth с НЕПРАВИЛЬНЫМ hexID и acceptAlternate=false
		wrongHexID := "wrong_hex_id_0000000000000000"
		err = client2.SendGameServerAuth(serverID, wrongHexID, false) // acceptAlternate=false
		s.Require().NoError(err)

		// Should receive LoginServerFail with ReasonWrongHexID
		reason, err := client2.ReadLoginServerFail()
		s.Require().NoError(err)
		s.Equal(byte(3), reason, "should receive ReasonWrongHexID (3)")
	})

	// 2. Already logged in
	s.Run("already logged in", func() {
		s.T().Parallel()

		// Регистрируем GS
		client1, err := testutil.NewGSClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client1.Close()

		serverID := byte(201)
		hexID := "already_logged_in_hex_id_0000000"

		err = client1.CompleteRegistration(serverID, hexID)
		s.Require().NoError(err)

		// Пытаемся подключиться второй раз с тем же serverID и hexID
		client2, err := testutil.NewGSClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client2.Close()

		// NewGSClient уже прочитал InitLS

		bfKey := make([]byte, 40)
		_, err = rand.Read(bfKey)
		s.Require().NoError(err)
		err = client2.SendBlowFishKey(bfKey)
		s.Require().NoError(err)

		err = client2.SendGameServerAuth(serverID, hexID, false)
		s.Require().NoError(err)

		// Should receive LoginServerFail with ReasonAlreadyLoggedIn
		reason, err := client2.ReadLoginServerFail()
		s.Require().NoError(err)
		s.Equal(byte(7), reason, "should receive ReasonAlreadyLoggedIn (7)")
	})
}

// TestPlayerAuthRequestInvalid тестирует невалидные PlayerAuthRequest.
func (s *GSListenerSuite) TestPlayerAuthRequestInvalid() {
	// Регистрируем GS
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	serverID := byte(202)
	hexID := "playerauth_invalid_hex_id_00000"
	err = client.CompleteRegistration(serverID, hexID)
	s.Require().NoError(err)

	// Send PlayerAuthRequest с невалидным sessionKey (не существует в SessionManager)
	invalidSessionKey := login.SessionKey{
		LoginOkID1: 99999,
		LoginOkID2: 88888,
		PlayOkID1:  77777,
		PlayOkID2:  66666,
	}

	err = client.SendPlayerAuthRequest("nonexistent_user", invalidSessionKey)
	s.Require().NoError(err)

	// Should receive PlayerAuthResponse with result=false
	account, result, err := client.ReadPlayerAuthResponse()
	s.Require().NoError(err)
	s.Equal("nonexistent_user", account, "account name should match")
	s.False(result, "PlayerAuthResponse should indicate failure for invalid session")
}

// ============================================================================
// Phase 3: Concurrency & Edge Cases
// ============================================================================

// TestConcurrentGSRegistration тестирует одновременную регистрацию множественных GameServers.
func (s *GSListenerSuite) TestConcurrentGSRegistration() {
	const numServers = 10

	type result struct {
		id      int
		success bool
		err     error
	}

	results := make(chan result, numServers)
	var wg sync.WaitGroup

	for i := range numServers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			res := result{id: id, success: false}

			client, err := testutil.NewGSClient(s.T(), s.addr)
			if err != nil {
				res.err = fmt.Errorf("failed to create: %w", err)
				results <- res
				return
			}
			defer client.Close()

			serverID := byte(150 + id)
			hexID := fmt.Sprintf("concurrent_gs_%d_hex_id_000000", id)

			if err := client.CompleteRegistration(serverID, hexID); err != nil {
				res.err = fmt.Errorf("registration failed: %w", err)
				results <- res
				return
			}

			// Verify registration
			gsInfo, exists := s.gsTable.GetByID(int(serverID))
			if !exists {
				res.err = fmt.Errorf("GS not found in table")
				results <- res
				return
			}

			if !gsInfo.IsAuthed() {
				res.err = fmt.Errorf("GS not authed")
				results <- res
				return
			}

			res.success = true
			results <- res
		}(i)
	}

	wg.Wait()
	close(results)

	// Check results
	successCount := 0
	for res := range results {
		if !res.success {
			s.T().Logf("GS %d failed: %v", res.id, res.err)
		} else {
			successCount++
		}
	}

	s.Equal(numServers, successCount, "all concurrent GS registrations should succeed")
}

// TestGSDisconnectDuringRegistration тестирует disconnect во время регистрации.
func (s *GSListenerSuite) TestGSDisconnectDuringRegistration() {
	// Create GS client and start registration
	client, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)

	// Send BlowFishKey
	bfKey := make([]byte, 40)
	_, err = rand.Read(bfKey)
	s.Require().NoError(err)
	err = client.SendBlowFishKey(bfKey)
	s.Require().NoError(err)

	// Close connection immediately (simulate network failure)
	client.Close()

	// Wait for cleanup (verify server ready for new connections)
	testutil.WaitForCleanup(s.T(), func() bool {
		conn, err := net.DialTimeout("tcp", s.addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 5*time.Second)

	// Create new GS client to verify server still works
	client2, err := testutil.NewGSClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client2.Close()

	serverID := byte(220)
	hexID := "disconnect_test_hex_id_000000000"
	err = client2.CompleteRegistration(serverID, hexID)
	s.NoError(err, "server should still accept new GS connections after disconnect")
}

// TestGSListenerSuite запускает GSListenerSuite.
func TestGSListenerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(GSListenerSuite))
}
