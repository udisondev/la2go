package integration

import (
	"context"
	"fmt"
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

// CrossServerSuite тестирует взаимодействие LoginServer и gslistener
// через shared SessionManager и GameServerTable.
type CrossServerSuite struct {
	IntegrationSuite
	loginServer *login.Server
	gsListener  *gslistener.Server
	sessionMgr  *login.SessionManager
	gsTable     *gameserver.GameServerTable
	cfg         config.LoginServer
	lsAddr      string // адрес LoginServer
	gsAddr      string // адрес gslistener
}

// SetupSuite инициализирует оба сервера с shared зависимостями.
func (s *CrossServerSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Shared зависимости
	s.sessionMgr = login.NewSessionManager()
	s.gsTable = gameserver.NewGameServerTable(s.db)

	// Конфиг для обоих серверов
	s.cfg = config.LoginServer{
		BindAddress:        "127.0.0.1",
		Port:               0, // случайный порт
		GSListenHost:       "127.0.0.1",
		GSListenPort:       0, // случайный порт
		AutoCreateAccounts: true,
		ShowLicence:        true, // Phase 1: тестируем с ShowLicence=true
		GameServers: []config.GameServerEntry{
			{ID: 1, Name: "TestServer1", Host: "127.0.0.1", Port: 7777},
			{ID: 2, Name: "TestServer2", Host: "127.0.0.1", Port: 7777},
			{ID: 3, Name: "TestServer3", Host: "127.0.0.1", Port: 7777},
			// Дополнительные серверы для TestConcurrentPlayerAuthRequests
			{ID: 101, Name: "ConcurrentTestServer1", Host: "127.0.0.1", Port: 7777},
			{ID: 102, Name: "ConcurrentTestServer2", Host: "127.0.0.1", Port: 7777},
			{ID: 103, Name: "ConcurrentTestServer3", Host: "127.0.0.1", Port: 7777},
		},
	}

	// Создаём LoginServer с shared SessionManager
	var err error
	s.loginServer, err = login.NewServer(s.cfg, s.db, login.WithSessionManager(s.sessionMgr))
	if err != nil {
		s.T().Fatalf("failed to create login server: %v", err)
	}

	// Создаём gslistener
	s.gsListener, err = gslistener.NewServer(s.cfg, s.db, s.gsTable, s.sessionMgr)
	if err != nil {
		s.T().Fatalf("failed to create gslistener: %v", err)
	}

	// Запускаем LoginServer
	lsListener, lsAddr := testutil.ListenTCP(s.T())
	s.lsAddr = lsAddr
	lsCtx, lsCancel := context.WithTimeout(context.Background(), 30*time.Second)
	s.T().Cleanup(lsCancel)
	s.T().Cleanup(func() {
		_ = s.loginServer.Close()
	})

	go func() {
		if err := s.loginServer.Serve(lsCtx, lsListener); err != nil && err != context.Canceled {
			s.T().Logf("login server error: %v", err)
		}
	}()

	// Запускаем gslistener
	gsListener, gsAddr := testutil.ListenTCP(s.T())
	s.gsAddr = gsAddr
	gsCtx, gsCancel := context.WithTimeout(context.Background(), 30*time.Second)
	s.T().Cleanup(gsCancel)
	s.T().Cleanup(func() {
		_ = s.gsListener.Close()
	})

	go func() {
		if err := s.gsListener.Serve(gsCtx, gsListener); err != nil && err != context.Canceled {
			s.T().Logf("gslistener error: %v", err)
		}
	}()

	// Ждём запуска обоих серверов (polling с timeout вместо sleep)
	if err := testutil.WaitForTCPReady(s.lsAddr, 5*time.Second); err != nil {
		s.T().Fatalf("login server failed to start: %v", err)
	}
	if err := testutil.WaitForTCPReady(s.gsAddr, 5*time.Second); err != nil {
		s.T().Fatalf("gslistener failed to start: %v", err)
	}
}

// TearDownSuite останавливает оба сервера.
func (s *CrossServerSuite) TearDownSuite() {
	s.IntegrationSuite.TearDownSuite()
}

// ============================================================================
// Phase 1: Critical Path Tests
// ============================================================================

// TestClientToGSSessionRelay тестирует полный flow:
// 1. Client auth на LoginServer → получает sessionKey
// 2. Client выбирает server → получает PlayOk
// 3. GameServer регистрируется на gslistener
// 4. GameServer отправляет PlayerAuthRequest с sessionKey
// 5. LoginServer валидирует session через SessionManager
// 6. GameServer получает PlayerAuthResponse(success=true)
func (s *CrossServerSuite) TestClientToGSSessionRelay() {
	// 1. Client auth на LoginServer
	lsClient, err := testutil.NewLoginClient(s.T(), s.lsAddr)
	s.Require().NoError(err, "failed to create login client")
	defer lsClient.Close()

	// Complete auth flow
	err = lsClient.SendAuthGameGuard()
	s.Require().NoError(err)
	err = lsClient.ReadGGAuth()
	s.Require().NoError(err)

	account := "crosstest1"
	err = lsClient.SendRequestAuthLogin(account, "password")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := lsClient.ReadLoginOk()
	s.Require().NoError(err)

	// 2. Client запрашивает server list и выбирает сервер
	err = lsClient.SendRequestServerList(loginOkID1, loginOkID2)
	s.Require().NoError(err)
	_, err = lsClient.ReadServerList()
	s.Require().NoError(err)

	// Выбираем server ID 1
	err = lsClient.SendRequestServerLogin(loginOkID1, loginOkID2, 1)
	s.Require().NoError(err)

	playOkID1, playOkID2, err := lsClient.ReadPlayOk()
	s.Require().NoError(err)

	// Verify session создана в SessionManager
	sessionCount := s.sessionMgr.Count()
	s.Equal(1, sessionCount, "session should be stored in SessionManager")

	// 3. GameServer регистрируется на gslistener
	gsClient, err := testutil.NewGSClient(s.T(), s.gsAddr)
	s.Require().NoError(err, "failed to create GS client")
	defer gsClient.Close()

	serverID := byte(1)
	hexID := "cross_test_hex_id_000000000000"
	err = gsClient.CompleteRegistration(serverID, hexID)
	s.Require().NoError(err, "GS registration failed")

	// 4. GameServer отправляет PlayerAuthRequest
	sessionKey := login.SessionKey{
		LoginOkID1: loginOkID1,
		LoginOkID2: loginOkID2,
		PlayOkID1:  playOkID1,
		PlayOkID2:  playOkID2,
	}

	err = gsClient.SendPlayerAuthRequest(account, sessionKey)
	s.Require().NoError(err)

	// 5. & 6. Читаем PlayerAuthResponse
	respAccount, result, err := gsClient.ReadPlayerAuthResponse()
	s.Require().NoError(err)
	s.Equal(account, respAccount, "account name should match")
	s.True(result, "PlayerAuthResponse should be successful - LS validated session")

	// Verify session удалена из SessionManager после валидации
	sessionCount = s.sessionMgr.Count()
	s.Equal(0, sessionCount, "session should be removed after GS validation")
}

// TestConcurrentClientAndGSAuth тестирует одновременную работу LS и GS.
// Несколько клиентов и несколько GameServers работают параллельно.
func (s *CrossServerSuite) TestConcurrentClientAndGSAuth() {
	// Регистрируем GameServer
	gsClient, err := testutil.NewGSClient(s.T(), s.gsAddr)
	s.Require().NoError(err)
	defer gsClient.Close()

	serverID := byte(2)
	err = gsClient.CompleteRegistration(serverID, "concurrent_gs_hex_id_0000000")
	s.Require().NoError(err)

	// Создаём несколько клиентов одновременно
	const numClients = 5

	for i := range numClients {
		// Каждый клиент проходит full auth
		lsClient, err := testutil.NewLoginClient(s.T(), s.lsAddr)
		s.Require().NoError(err, "client %d: failed to create", i)
		defer lsClient.Close()

		err = lsClient.SendAuthGameGuard()
		s.Require().NoError(err, "client %d: AuthGameGuard failed", i)
		err = lsClient.ReadGGAuth()
		s.Require().NoError(err, "client %d: ReadGGAuth failed", i)

		account := testutil.Fixtures.ValidAccount + string(rune('0'+i))
		err = lsClient.SendRequestAuthLogin(account, "pass")
		s.Require().NoError(err, "client %d: RequestAuthLogin failed", i)

		loginOkID1, loginOkID2, err := lsClient.ReadLoginOk()
		s.Require().NoError(err, "client %d: ReadLoginOk failed", i)

		// Request server list
		err = lsClient.SendRequestServerList(loginOkID1, loginOkID2)
		s.Require().NoError(err, "client %d: RequestServerList failed", i)
		_, err = lsClient.ReadServerList()
		s.Require().NoError(err, "client %d: ReadServerList failed", i)

		// Select server
		err = lsClient.SendRequestServerLogin(loginOkID1, loginOkID2, serverID)
		s.Require().NoError(err, "client %d: RequestServerLogin failed", i)
		playOkID1, playOkID2, err := lsClient.ReadPlayOk()
		s.Require().NoError(err, "client %d: ReadPlayOk failed", i)

		// GS validates session
		sessionKey := login.SessionKey{
			LoginOkID1: loginOkID1,
			LoginOkID2: loginOkID2,
			PlayOkID1:  playOkID1,
			PlayOkID2:  playOkID2,
		}

		err = gsClient.SendPlayerAuthRequest(account, sessionKey)
		s.Require().NoError(err, "client %d: PlayerAuthRequest failed", i)

		respAccount, result, err := gsClient.ReadPlayerAuthResponse()
		s.Require().NoError(err, "client %d: PlayerAuthResponse failed", i)
		s.Equal(account, respAccount, "client %d: account mismatch", i)
		s.True(result, "client %d: auth should succeed", i)
	}

	// Verify все sessions удалены
	sessionCount := s.sessionMgr.Count()
	s.Equal(0, sessionCount, "all sessions should be removed")
}

// TestSessionCleanupAfterGSValidation проверяет что SessionManager
// корректно удаляет session после валидации через GS.
func (s *CrossServerSuite) TestSessionCleanupAfterGSValidation() {
	// Регистрируем GS
	gsClient, err := testutil.NewGSClient(s.T(), s.gsAddr)
	s.Require().NoError(err)
	defer gsClient.Close()

	serverID := byte(3)
	err = gsClient.CompleteRegistration(serverID, "cleanup_test_hex_id_00000000")
	s.Require().NoError(err)

	// Client auth
	lsClient, err := testutil.NewLoginClient(s.T(), s.lsAddr)
	s.Require().NoError(err)
	defer lsClient.Close()

	err = lsClient.SendAuthGameGuard()
	s.Require().NoError(err)
	err = lsClient.ReadGGAuth()
	s.Require().NoError(err)

	account := "cleanuptest"
	err = lsClient.SendRequestAuthLogin(account, "password")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := lsClient.ReadLoginOk()
	s.Require().NoError(err)

	err = lsClient.SendRequestServerList(loginOkID1, loginOkID2)
	s.Require().NoError(err)
	_, err = lsClient.ReadServerList()
	s.Require().NoError(err)

	err = lsClient.SendRequestServerLogin(loginOkID1, loginOkID2, serverID)
	s.Require().NoError(err)

	playOkID1, playOkID2, err := lsClient.ReadPlayOk()
	s.Require().NoError(err)

	// Verify session в SessionManager BEFORE GS validation
	countBefore := s.sessionMgr.Count()
	s.Equal(1, countBefore, "session should exist before GS validation")

	// GS validates
	sessionKey := login.SessionKey{
		LoginOkID1: loginOkID1,
		LoginOkID2: loginOkID2,
		PlayOkID1:  playOkID1,
		PlayOkID2:  playOkID2,
	}

	err = gsClient.SendPlayerAuthRequest(account, sessionKey)
	s.Require().NoError(err)

	_, result, err := gsClient.ReadPlayerAuthResponse()
	s.Require().NoError(err)
	s.True(result, "validation should succeed")

	// Verify session удалена AFTER GS validation
	countAfter := s.sessionMgr.Count()
	s.Equal(0, countAfter, "session should be removed after GS validation")

	// Try to validate same session again — should fail
	err = gsClient.SendPlayerAuthRequest(account, sessionKey)
	s.Require().NoError(err)

	_, result2, err := gsClient.ReadPlayerAuthResponse()
	s.Require().NoError(err)
	s.False(result2, "second validation with same session should fail")
}

// ============================================================================
// Phase 3: Concurrency Tests
// ============================================================================

// TestConcurrentPlayerAuthRequests тестирует race condition при множественных PlayerAuthRequests.
// Симулирует ситуацию когда несколько GameServers пытаются валидировать одну и ту же session
// (например, игрок пытается подключиться к нескольким серверам одновременно).
func (s *CrossServerSuite) TestConcurrentPlayerAuthRequests() {
	// Регистрируем несколько GameServers (используем уникальные IDs 101-103)
	const numServers = 3
	gsClients := make([]*testutil.GSClient, numServers)
	for i := range numServers {
		client, err := testutil.NewGSClient(s.T(), s.gsAddr)
		s.Require().NoError(err)
		defer client.Close()

		serverID := byte(101 + i) // IDs: 101, 102, 103 (чтобы не конфликтовать с другими тестами)
		hexID := fmt.Sprintf("concurrent_gs_%d_hex_id", i)
		err = client.CompleteRegistration(serverID, hexID)
		s.Require().NoError(err)

		gsClients[i] = client
	}

	// Client auth на LoginServer (выбирает server 1)
	lsClient, err := testutil.NewLoginClient(s.T(), s.lsAddr)
	s.Require().NoError(err)
	defer lsClient.Close()

	err = lsClient.SendAuthGameGuard()
	s.Require().NoError(err)
	err = lsClient.ReadGGAuth()
	s.Require().NoError(err)

	account := "conctest"
	err = lsClient.SendRequestAuthLogin(account, "password")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := lsClient.ReadLoginOk()
	s.Require().NoError(err)

	err = lsClient.SendRequestServerList(loginOkID1, loginOkID2)
	s.Require().NoError(err)
	_, err = lsClient.ReadServerList()
	s.Require().NoError(err)

	err = lsClient.SendRequestServerLogin(loginOkID1, loginOkID2, 101)
	s.Require().NoError(err)
	playOkID1, playOkID2, err := lsClient.ReadPlayOk()
	s.Require().NoError(err)

	sessionKey := login.SessionKey{
		LoginOkID1: loginOkID1,
		LoginOkID2: loginOkID2,
		PlayOkID1:  playOkID1,
		PlayOkID2:  playOkID2,
	}

	// Simulate race: каждый GameServer пытается валидировать одну и ту же session
	type result struct {
		serverIdx int
		success   bool
		err       error
	}
	results := make(chan result, numServers)
	var wg sync.WaitGroup

	for i := range numServers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			client := gsClients[idx]
			err := client.SendPlayerAuthRequest(account, sessionKey)
			if err != nil {
				results <- result{serverIdx: idx, success: false, err: err}
				return
			}

			respAccount, success, err := client.ReadPlayerAuthResponse()
			if err != nil {
				results <- result{serverIdx: idx, success: false, err: err}
				return
			}

			if respAccount != account {
				results <- result{serverIdx: idx, success: false, err: fmt.Errorf("account mismatch")}
				return
			}

			results <- result{serverIdx: idx, success: success, err: nil}
		}(i)
	}

	wg.Wait()
	close(results)

	// Only ONE should succeed (first validates and removes session)
	successCount := 0
	for res := range results {
		if res.err != nil {
			s.T().Logf("server %d: request error: %v", res.serverIdx, res.err)
		} else if res.success {
			successCount++
			s.T().Logf("server %d: auth SUCCESS", res.serverIdx)
		} else {
			s.T().Logf("server %d: auth FAIL (expected - session already validated)", res.serverIdx)
		}
	}

	s.Equal(1, successCount, "only one PlayerAuthRequest should succeed (race winner)")

	// Verify session removed
	count := s.sessionMgr.Count()
	s.Equal(0, count, "session should be removed after validation")
}

// TestCrossServerSuite запускает CrossServerSuite.
func TestCrossServerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(CrossServerSuite))
}
