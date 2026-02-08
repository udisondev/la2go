package login

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// MockAccountRepository мок для AccountRepository в unit тестах.
type MockAccountRepository struct {
	GetAccountFunc         func(ctx context.Context, login string) (*model.Account, error)
	CreateAccountFunc      func(ctx context.Context, login, passwordHash, ip string) error
	GetOrCreateAccountFunc func(ctx context.Context, login, passwordHash, ip string) (*model.Account, error)
	UpdateLastLoginFunc    func(ctx context.Context, login, ip string) error
}

func (m *MockAccountRepository) GetAccount(ctx context.Context, login string) (*model.Account, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, login)
	}
	return nil, nil
}

func (m *MockAccountRepository) CreateAccount(ctx context.Context, login, passwordHash, ip string) error {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, login, passwordHash, ip)
	}
	return nil
}

func (m *MockAccountRepository) GetOrCreateAccount(ctx context.Context, login, passwordHash, ip string) (*model.Account, error) {
	if m.GetOrCreateAccountFunc != nil {
		return m.GetOrCreateAccountFunc(ctx, login, passwordHash, ip)
	}
	// Default: создаём новый аккаунт
	return &model.Account{
		Login:        login,
		PasswordHash: passwordHash,
		AccessLevel:  0,
	}, nil
}

func (m *MockAccountRepository) UpdateLastLogin(ctx context.Context, login, ip string) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, login, ip)
	}
	return nil
}

// buildAuthGameGuardPacket создаёт тестовый пакет AuthGameGuard.
func buildAuthGameGuardPacket(sessionID int32) []byte {
	packet := make([]byte, 5)
	packet[0] = OpcodeAuthGameGuard
	binary.LittleEndian.PutUint32(packet[1:], uint32(sessionID))
	return packet
}

func TestHandler_HandleAuthGameGuard_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockAccountRepository{}
	cfg := config.DefaultLoginServer()
	sm := NewSessionManager()
	handler := NewHandler(mockRepo, cfg, sm)

	rsaKeyPair, _ := crypto.GenerateRSAKeyPair()
	client := &Client{
		sessionID:  12345,
		rsaKeyPair: rsaKeyPair,
		state:      StateConnected,
		ip:         "127.0.0.1",
	}

	packet := buildAuthGameGuardPacket(client.SessionID())
	buf := make([]byte, 1024)

	// Act
	n, ok, err := handler.HandlePacket(context.Background(), client, packet, buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected connection to stay open")
	}
	if n == 0 {
		t.Error("expected response packet")
	}
	if client.State() != StateAuthedGG {
		t.Errorf("expected state AUTHED_GG, got %v", client.State())
	}

	// Проверяем ответ GGAuth
	if buf[0] != 0x0B {
		t.Errorf("expected GGAuth opcode 0x0B, got 0x%02X", buf[0])
	}
}

func TestHandler_HandleAuthGameGuard_WrongSessionID(t *testing.T) {
	// Arrange
	mockRepo := &MockAccountRepository{}
	cfg := config.DefaultLoginServer()
	sm := NewSessionManager()
	handler := NewHandler(mockRepo, cfg, sm)

	rsaKeyPair, _ := crypto.GenerateRSAKeyPair()
	client := &Client{
		sessionID:  12345,
		rsaKeyPair: rsaKeyPair,
		state:      StateConnected,
		ip:         "127.0.0.1",
	}

	// Неправильный session ID
	packet := buildAuthGameGuardPacket(99999)
	buf := make([]byte, 1024)

	// Act
	_, ok, err := handler.HandlePacket(context.Background(), client, packet, buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected connection to close")
	}

	// Проверяем LoginFail packet
	if buf[0] != 0x01 {
		t.Errorf("expected LoginFail opcode 0x01, got 0x%02X", buf[0])
	}
	if buf[1] != serverpackets.ReasonAccessFailed {
		t.Errorf("expected ReasonAccessFailed, got 0x%02X", buf[1])
	}
}

func TestHandler_HandleAuthGameGuard_WrongState(t *testing.T) {
	// Arrange
	mockRepo := &MockAccountRepository{}
	cfg := config.DefaultLoginServer()
	sm := NewSessionManager()
	handler := NewHandler(mockRepo, cfg, sm)

	rsaKeyPair, _ := crypto.GenerateRSAKeyPair()
	client := &Client{
		sessionID:  12345,
		rsaKeyPair: rsaKeyPair,
		state:      StateAuthedGG, // wrong state
		ip:         "127.0.0.1",
	}

	packet := buildAuthGameGuardPacket(client.SessionID())
	buf := make([]byte, 1024)

	// Act
	n, ok, err := handler.HandlePacket(context.Background(), client, packet, buf)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected connection to stay open")
	}
	if n != 0 {
		t.Error("expected no response (wrong state)")
	}
}

// TestMockAccountRepository_GetAccount тестирует MockAccountRepository.
func TestMockAccountRepository_GetAccount(t *testing.T) {
	// Arrange
	expectedAccount := &model.Account{
		Login:        "testuser",
		PasswordHash: "hash",
		AccessLevel:  0,
	}

	mock := &MockAccountRepository{
		GetAccountFunc: func(ctx context.Context, login string) (*model.Account, error) {
			if login == "testuser" {
				return expectedAccount, nil
			}
			return nil, fmt.Errorf("user not found")
		},
	}

	// Act
	acc, err := mock.GetAccount(context.Background(), "testuser")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc == nil {
		t.Fatal("expected account to be returned")
	}
	if acc.Login != "testuser" {
		t.Errorf("expected login 'testuser', got %q", acc.Login)
	}
}

// TestMockAccountRepository_CreateAccount тестирует MockAccountRepository.CreateAccount.
func TestMockAccountRepository_CreateAccount(t *testing.T) {
	// Arrange
	var capturedLogin, capturedHash, capturedIP string

	mock := &MockAccountRepository{
		CreateAccountFunc: func(ctx context.Context, login, passwordHash, ip string) error {
			capturedLogin = login
			capturedHash = passwordHash
			capturedIP = ip
			return nil
		},
	}

	// Act
	err := mock.CreateAccount(context.Background(), "newuser", "hashvalue", "192.168.1.1")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLogin != "newuser" {
		t.Errorf("expected login 'newuser', got %q", capturedLogin)
	}
	if capturedHash != "hashvalue" {
		t.Errorf("expected hash 'hashvalue', got %q", capturedHash)
	}
	if capturedIP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", capturedIP)
	}
}

// TestMockAccountRepository_DatabaseError тестирует обработку DB ошибок.
func TestMockAccountRepository_DatabaseError(t *testing.T) {
	// Arrange
	mock := &MockAccountRepository{
		GetAccountFunc: func(ctx context.Context, login string) (*model.Account, error) {
			return nil, fmt.Errorf("connection lost")
		},
	}

	// Act
	acc, err := mock.GetAccount(context.Background(), "anyuser")

	// Assert
	if err == nil {
		t.Error("expected error, got nil")
	}
	if acc != nil {
		t.Error("expected nil account on error")
	}
	if err.Error() != "connection lost" {
		t.Errorf("expected error 'connection lost', got %q", err.Error())
	}
}

