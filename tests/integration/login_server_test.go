package integration

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/testutil"
)

// LoginServerSuite тестирует LoginServer с реальными TCP соединениями.
type LoginServerSuite struct {
	IntegrationSuite
	server *login.Server
	cfg    config.LoginServer
	addr   string // адрес сервера (listener.Addr().String())
}

// SetupSuite инициализирует LoginServer.
func (s *LoginServerSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Создаём конфиг сервера с полными настройками для тестов
	s.cfg = config.LoginServer{
		BindAddress:        "127.0.0.1",
		Port:               0, // случайный порт
		AutoCreateAccounts: true,
		ShowLicence:        true, // Phase 1: тестируем с ShowLicence=true
		GameServers: []config.GameServerEntry{
			{ID: 1, Name: "Bartz"},
			{ID: 2, Name: "Sieghardt"},
		},
	}

	var err error
	s.server, err = login.NewServer(s.cfg, s.db)
	if err != nil {
		s.T().Fatalf("failed to create login server: %v", err)
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
			s.T().Logf("login server error: %v", err)
		}
	}()

	// Ждём запуска сервера (polling с timeout вместо sleep)
	if err := testutil.WaitForTCPReady(s.addr, 5*time.Second); err != nil {
		s.T().Fatalf("login server failed to start: %v", err)
	}
}

// TearDownSuite останавливает сервер.
func (s *LoginServerSuite) TearDownSuite() {
	s.IntegrationSuite.TearDownSuite()
}

// TestClientConnection тестирует подключение клиента и получение Init пакета.
func (s *LoginServerSuite) TestClientConnection() {
	addr := s.addr
	s.Require().NotEmpty(addr, "server address should be available")

	// Используем LoginClient для корректной расшифровки Init пакета
	client, err := testutil.NewLoginClient(s.T(), addr)
	s.Require().NoError(err, "failed to create login client")
	defer client.Close()

	// Init уже прочитан в NewLoginClient, проверяем что все поля установлены
	s.NotZero(client.SessionID(), "sessionID should be set after reading Init packet")
	s.NotNil(client.RSAModulus(), "RSA modulus should be set")
	s.Len(client.RSAModulus(), 128, "RSA modulus should be 128 bytes")
	s.NotNil(client.BlowfishKey(), "Blowfish key should be set")
	s.Len(client.BlowfishKey(), 16, "Blowfish key should be 16 bytes")
}

// TestMultipleClients тестирует одновременное подключение нескольких клиентов.
func (s *LoginServerSuite) TestMultipleClients() {
	addr := s.addr
	s.Require().NotEmpty(addr)

	const clients = 10

	for i := range clients {
		// Используем LoginClient для корректной расшифровки Init пакета
		client, err := testutil.NewLoginClient(s.T(), addr)
		s.Require().NoError(err, "client %d failed to create login client", i)

		// Init уже прочитан в NewLoginClient
		s.NotZero(client.SessionID(), "client %d: sessionID should be set", i)

		client.Close()
	}
}

// ============================================================================
// Phase 1: Critical Path Tests
// ============================================================================

// TestFullAuthFlow тестирует полный flow аутентификации:
// Init → AuthGameGuard → RequestAuthLogin → LoginOk
func (s *LoginServerSuite) TestFullAuthFlow() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err, "failed to create login client")
	defer client.Close()

	// Init уже прочитан в NewLoginClient
	s.NotZero(client.SessionID(), "sessionID should be set")

	// Send AuthGameGuard + Read GGAuth
	err = client.SendAuthGameGuard()
	s.Require().NoError(err, "failed to send AuthGameGuard")

	err = client.ReadGGAuth()
	s.Require().NoError(err, "failed to read GGAuth")

	// Send RequestAuthLogin + Read LoginOk
	err = client.SendRequestAuthLogin("testuser", "testpass")
	s.Require().NoError(err, "failed to send RequestAuthLogin")

	loginOkID1, loginOkID2, err := client.ReadLoginOk()
	s.Require().NoError(err, "failed to read LoginOk")
	s.NotZero(loginOkID1, "loginOkID1 should be non-zero")
	s.NotZero(loginOkID2, "loginOkID2 should be non-zero")

	// Verify account created in DB
	acc, err := s.db.GetAccount(s.ctx, "testuser")
	s.Require().NoError(err, "failed to get account from DB")
	s.NotNil(acc, "account should exist in DB")
	s.Equal("testuser", acc.Login, "account login should match")
}

// TestAuthGameGuardFlow тестирует GGAuth handshake.
func (s *LoginServerSuite) TestAuthGameGuardFlow() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Send AuthGameGuard
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)

	// Read GGAuth response
	err = client.ReadGGAuth()
	s.NoError(err, "server should respond with GGAuth")
}

// TestRequestAuthLoginSuccess тестирует успешную аутентификацию.
func (s *LoginServerSuite) TestRequestAuthLoginSuccess() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Complete AuthGameGuard first
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)
	err = client.ReadGGAuth()
	s.Require().NoError(err)

	// Send RequestAuthLogin
	err = client.SendRequestAuthLogin("user1", "password123")
	s.Require().NoError(err)

	// Read LoginOk
	loginOkID1, loginOkID2, err := client.ReadLoginOk()
	s.Require().NoError(err)
	s.NotZero(loginOkID1)
	s.NotZero(loginOkID2)

	// Verify last_active updated
	acc, err := s.db.GetAccount(s.ctx, "user1")
	s.Require().NoError(err)
	s.NotNil(acc)
	s.False(acc.LastActive.IsZero(), "last_active should be set")
}

// TestRequestServerListFlow тестирует получение списка серверов.
func (s *LoginServerSuite) TestRequestServerListFlow() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Complete auth first
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)
	err = client.ReadGGAuth()
	s.Require().NoError(err)

	err = client.SendRequestAuthLogin("user2", "password123")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := client.ReadLoginOk()
	s.Require().NoError(err)

	// Send RequestServerList
	err = client.SendRequestServerList(loginOkID1, loginOkID2)
	s.Require().NoError(err)

	// Read ServerList packet
	serverList, err := client.ReadServerList()
	s.Require().NoError(err)
	s.NotNil(serverList)

	// Verify ServerList opcode
	s.Equal(byte(0x04), serverList[0], "should be ServerList packet")

	// Parse server count (byte at offset 1)
	s.Require().Greater(len(serverList), 2, "packet should have at least opcode + count + lastServer")
	serverCount := int(serverList[1])
	s.Equal(len(s.cfg.GameServers), serverCount, "server count should match config")

	// Verify lastServer (byte at offset 2)
	lastServer := serverList[2]
	s.NotZero(lastServer, "lastServer should be set")
}

// TestRequestServerLoginFlow тестирует выбор сервера и получение PlayOk.
func (s *LoginServerSuite) TestRequestServerLoginFlow() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Complete auth
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)
	err = client.ReadGGAuth()
	s.Require().NoError(err)

	err = client.SendRequestAuthLogin("user3", "password123")
	s.Require().NoError(err)

	loginOkID1, loginOkID2, err := client.ReadLoginOk()
	s.Require().NoError(err)

	// Request server list
	err = client.SendRequestServerList(loginOkID1, loginOkID2)
	s.Require().NoError(err)
	_, err = client.ReadServerList()
	s.Require().NoError(err)

	// Send RequestServerLogin (выбираем server ID 1)
	err = client.SendRequestServerLogin(loginOkID1, loginOkID2, 1)
	s.Require().NoError(err)

	// Read PlayOk
	playOkID1, playOkID2, err := client.ReadPlayOk()
	s.Require().NoError(err)
	s.NotZero(playOkID1, "playOkID1 should be non-zero")
	s.NotZero(playOkID2, "playOkID2 should be non-zero")
}

// TestStateTransitions проверяет переходы состояний клиента.
func (s *LoginServerSuite) TestStateTransitions() {
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// State: CONNECTED → AUTHED_GG
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)
	err = client.ReadGGAuth()
	s.NoError(err, "should transition to AUTHED_GG")

	// State: AUTHED_GG → AUTHED_LOGIN
	err = client.SendRequestAuthLogin("user4", "password123")
	s.Require().NoError(err)
	_, _, err = client.ReadLoginOk()
	s.NoError(err, "should transition to AUTHED_LOGIN")
}

// TestAutoCreateAccount тестирует автоматическое создание аккаунта.
func (s *LoginServerSuite) TestAutoCreateAccount() {
	newLogin := "autocreated"

	// Verify account doesn't exist
	acc, err := s.db.GetAccount(s.ctx, newLogin)
	s.Require().NoError(err)
	s.Nil(acc, "account should not exist before test")

	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client.Close()

	// Complete auth with new credentials
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)
	err = client.ReadGGAuth()
	s.Require().NoError(err)

	err = client.SendRequestAuthLogin(newLogin, "newpass")
	s.Require().NoError(err)

	_, _, err = client.ReadLoginOk()
	s.Require().NoError(err, "auto-create should succeed")

	// Verify account created
	acc, err = s.db.GetAccount(s.ctx, newLogin)
	s.Require().NoError(err)
	s.NotNil(acc, "account should be auto-created")
	s.Equal(newLogin, acc.Login)
}

// ============================================================================
// Phase 2: Error Handling Tests
// ============================================================================

// TestAuthenticationErrors тестирует различные error cases при аутентификации.
func (s *LoginServerSuite) TestAuthenticationErrors() {
	testCases := []struct {
		name       string
		login      string
		password   string
		setup      func() // опциональная подготовка (например, создание banned account)
		wantOpcode byte
		wantReason byte
	}{
		{
			name:       "wrong password for existing account",
			login:      "wrongpass",
			password:   "rightpass",
			wantOpcode: 0x01, // LoginFail
			wantReason: 0x02, // ReasonUserOrPassWrong
			setup: func() {
				// Создаём аккаунт с правильным паролем (wrongpassword)
				hash := db.HashPassword("wrongpassword")
				err := s.db.CreateAccount(s.ctx, "wrongpass", hash, "127.0.0.1")
				s.Require().NoError(err)
			},
		},
		{
			name:       "empty login",
			login:      "",
			password:   "password",
			wantOpcode: 0x01, // LoginFail
			wantReason: 0x02, // ReasonUserOrPassWrong
		},
		{
			name:       "empty password",
			login:      "user",
			password:   "",
			wantOpcode: 0x01, // LoginFail
			wantReason: 0x02, // ReasonUserOrPassWrong
		},
		{
			name:       "banned account",
			login:      "banneduser",
			password:   "testpass",
			wantOpcode: 0x02, // AccountKicked
			wantReason: 0x20, // ReasonPermanentlyBanned
			setup: func() {
				// Создаём banned account (access_level < 0)
				// Используем правильный hash (base64, не hex)
				hash := db.HashPassword("testpass")
				err := s.db.CreateAccount(s.ctx, "banneduser", hash, "127.0.0.1")
				s.Require().NoError(err)
				// Обновляем access_level на -100
				_, err = s.db.Pool().Exec(s.ctx, "UPDATE accounts SET access_level = -100 WHERE login = $1", "banneduser")
				s.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			client, err := testutil.NewLoginClient(s.T(), s.addr)
			s.Require().NoError(err)
			defer client.Close()

			// Complete AuthGameGuard
			err = client.SendAuthGameGuard()
			s.Require().NoError(err)
			err = client.ReadGGAuth()
			s.Require().NoError(err)

			// Send RequestAuthLogin with test credentials
			err = client.SendRequestAuthLogin(tc.login, tc.password)
			s.Require().NoError(err)

			// Read error packet
			payload, err := client.ReadPacket()
			s.Require().NoError(err)
			s.Require().GreaterOrEqual(len(payload), 2, "error packet should have at least opcode + reason")

			s.Equal(tc.wantOpcode, payload[0], "should receive correct error opcode")
			s.Equal(tc.wantReason, payload[1], "should receive correct error reason")
		})
	}
}

// TestRequestServerListErrors тестирует error cases для RequestServerList.
func (s *LoginServerSuite) TestRequestServerListErrors() {
	// 1. Wrong sessionKey
	s.Run("wrong sessionKey", func() {
		s.T().Parallel()

		client, err := testutil.NewLoginClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client.Close()

		// Complete auth
		err = client.SendAuthGameGuard()
		s.Require().NoError(err)
		err = client.ReadGGAuth()
		s.Require().NoError(err)

		err = client.SendRequestAuthLogin("serverlist_wrong_key", "password")
		s.Require().NoError(err)

		loginOkID1, loginOkID2, err := client.ReadLoginOk()
		s.Require().NoError(err)

		// Send RequestServerList with WRONG sessionKey
		err = client.SendRequestServerList(loginOkID1+999, loginOkID2+999)
		s.Require().NoError(err)

		// Should receive LoginFail with ReasonAccessFailed
		reason, err := client.ReadLoginFail()
		s.Require().NoError(err)
		s.Equal(byte(0x15), reason, "should receive ReasonAccessFailed")
	})
}

// TestRequestServerLoginErrors тестирует error cases для RequestServerLogin.
func (s *LoginServerSuite) TestRequestServerLoginErrors() {
	// 1. Wrong sessionKey
	s.Run("wrong sessionKey", func() {
		s.T().Parallel()

		client, err := testutil.NewLoginClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client.Close()

		// Complete auth and server list
		err = client.SendAuthGameGuard()
		s.Require().NoError(err)
		err = client.ReadGGAuth()
		s.Require().NoError(err)

		err = client.SendRequestAuthLogin("serverlogin_wrong_key", "password")
		s.Require().NoError(err)

		loginOkID1, loginOkID2, err := client.ReadLoginOk()
		s.Require().NoError(err)

		err = client.SendRequestServerList(loginOkID1, loginOkID2)
		s.Require().NoError(err)
		_, err = client.ReadServerList()
		s.Require().NoError(err)

		// Send RequestServerLogin with WRONG sessionKey
		err = client.SendRequestServerLogin(loginOkID1+999, loginOkID2+999, 1)
		s.Require().NoError(err)

		// Should receive LoginFail with ReasonAccessFailed
		reason, err := client.ReadLoginFail()
		s.Require().NoError(err)
		s.Equal(byte(0x15), reason, "should receive ReasonAccessFailed")
	})

	// 2. Unknown server ID
	s.Run("unknown server ID", func() {
		s.T().Parallel()

		client, err := testutil.NewLoginClient(s.T(), s.addr)
		s.Require().NoError(err)
		defer client.Close()

		// Complete auth
		err = client.SendAuthGameGuard()
		s.Require().NoError(err)
		err = client.ReadGGAuth()
		s.Require().NoError(err)

		err = client.SendRequestAuthLogin("serverlogin_unknown_server", "password")
		s.Require().NoError(err)

		loginOkID1, loginOkID2, err := client.ReadLoginOk()
		s.Require().NoError(err)

		err = client.SendRequestServerList(loginOkID1, loginOkID2)
		s.Require().NoError(err)
		_, err = client.ReadServerList()
		s.Require().NoError(err)

		// Send RequestServerLogin with non-existent server ID (99)
		err = client.SendRequestServerLogin(loginOkID1, loginOkID2, 99)
		s.Require().NoError(err)

		// Should receive PlayFail with ReasonServerOverloaded
		reason, err := client.ReadPlayFail()
		s.Require().NoError(err)
		s.Equal(byte(0x0F), reason, "should receive ReasonServerOverloaded for unknown server")
	})
}

// ============================================================================
// Phase 3: Concurrency & Edge Cases
// ============================================================================

// TestConcurrentClientAuth тестирует одновременную аутентификацию множественных клиентов.
func (s *LoginServerSuite) TestConcurrentClientAuth() {
	const numClients = 20

	type result struct {
		id      int
		success bool
		err     error
	}

	results := make(chan result, numClients)
	var wg sync.WaitGroup

	for i := range numClients {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			res := result{id: id, success: false}

			client, err := testutil.NewLoginClient(s.T(), s.addr)
			if err != nil {
				res.err = fmt.Errorf("failed to create: %w", err)
				results <- res
				return
			}
			defer client.Close()

			// Complete auth flow
			if err := client.SendAuthGameGuard(); err != nil {
				res.err = fmt.Errorf("SendAuthGameGuard: %w", err)
				results <- res
				return
			}

			if err := client.ReadGGAuth(); err != nil {
				res.err = fmt.Errorf("ReadGGAuth: %w", err)
				results <- res
				return
			}

			login := fmt.Sprintf("concurrent_user_%d", id)
			if err := client.SendRequestAuthLogin(login, "password"); err != nil {
				res.err = fmt.Errorf("SendRequestAuthLogin: %w", err)
				results <- res
				return
			}

			loginOkID1, loginOkID2, err := client.ReadLoginOk()
			if err != nil {
				res.err = fmt.Errorf("ReadLoginOk: %w", err)
				results <- res
				return
			}

			if loginOkID1 == 0 || loginOkID2 == 0 {
				res.err = fmt.Errorf("invalid session keys")
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
			s.T().Logf("client %d failed: %v", res.id, res.err)
		} else {
			successCount++
		}
	}

	s.Equal(numClients, successCount, "all concurrent auths should succeed")

	// Note: We don't verify DB state here as CreateAccount may be async.
	// The important part is that all auth requests succeeded without errors or race conditions.
}

// TestClientDisconnectDuringAuth тестирует disconnect между пакетами.
func (s *LoginServerSuite) TestClientDisconnectDuringAuth() {
	// Create client and start auth
	client, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)

	// Send AuthGameGuard but disconnect before reading response
	err = client.SendAuthGameGuard()
	s.Require().NoError(err)

	// Close connection immediately (simulate network failure)
	client.Close()

	// Server should handle gracefully (no crash, connection cleaned up)
	// Wait for cleanup (verify server ready for new connections)
	testutil.WaitForCleanup(s.T(), func() bool {
		conn, err := net.DialTimeout("tcp", s.addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 5*time.Second)

	// Create new client to verify server still works
	client2, err := testutil.NewLoginClient(s.T(), s.addr)
	s.Require().NoError(err)
	defer client2.Close()

	err = client2.SendAuthGameGuard()
	s.NoError(err, "server should still accept new connections after client disconnect")
}

// TestLoginServerSuite запускает LoginServerSuite.
func TestLoginServerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(LoginServerSuite))
}
