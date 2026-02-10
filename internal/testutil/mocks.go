package testutil

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// MockDB — in-memory имплементация database для unit тестов.
// Не требует реального PostgreSQL.
type MockDB struct {
	mu       sync.RWMutex
	accounts map[string]*model.Account
}

// NewMockDB создаёт новый MockDB экземпляр.
func NewMockDB() *MockDB {
	return &MockDB{
		accounts: make(map[string]*model.Account),
	}
}

// GetAccount получает аккаунт по логину.
func (m *MockDB) GetAccount(ctx context.Context, login string) (*model.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	acc, exists := m.accounts[login]
	if !exists {
		return nil, nil // account not found
	}

	// Возвращаем копию чтобы избежать race conditions
	return &model.Account{
		Login:        acc.Login,
		PasswordHash: acc.PasswordHash,
		LastIP:       acc.LastIP,
		AccessLevel:  acc.AccessLevel,
	}, nil
}

// CreateAccount создаёт новый аккаунт.
func (m *MockDB) CreateAccount(ctx context.Context, login, passwordHash, lastIP string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.accounts[login]; exists {
		return fmt.Errorf("account %q already exists", login)
	}

	m.accounts[login] = &model.Account{
		Login:        login,
		PasswordHash: passwordHash,
		LastIP:       lastIP,
		AccessLevel:  0,
	}

	return nil
}

// UpdateLastLogin обновляет время последнего логина и IP.
func (m *MockDB) UpdateLastLogin(ctx context.Context, login, lastIP string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, exists := m.accounts[login]
	if !exists {
		return fmt.Errorf("account %q not found", login)
	}

	acc.LastIP = lastIP
	return nil
}

// BanAccount банит аккаунт.
func (m *MockDB) BanAccount(ctx context.Context, login string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, exists := m.accounts[login]
	if !exists {
		return fmt.Errorf("account %q not found", login)
	}

	acc.AccessLevel = -100 // banned
	return nil
}

// Close закрывает MockDB (no-op для in-memory).
func (m *MockDB) Close() error {
	return nil
}

// Reset очищает все данные MockDB.
func (m *MockDB) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts = make(map[string]*model.Account)
}

// AccountCount возвращает количество аккаунтов в MockDB.
func (m *MockDB) AccountCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.accounts)
}

// MockConn — mock для net.Conn, используется в unit тестах.
type MockConn struct {
	readBuf    []byte
	writeBuf   []byte
	writeCount int // Phase 4.16: track number of Write() calls for broadcast tests
}

// NewMockConn создаёт новый MockConn экземпляр.
func NewMockConn() *MockConn {
	return &MockConn{
		readBuf:  make([]byte, 0),
		writeBuf: make([]byte, 0),
	}
}

// Read читает данные из readBuf.
func (m *MockConn) Read(b []byte) (int, error) {
	n := copy(b, m.readBuf)
	m.readBuf = m.readBuf[n:]
	return n, nil
}

// Write записывает данные в writeBuf.
func (m *MockConn) Write(b []byte) (int, error) {
	m.writeBuf = append(m.writeBuf, b...)
	m.writeCount++ // Phase 4.16: increment write counter
	return len(b), nil
}

// WriteCount returns the number of Write() calls since creation or last reset.
// Phase 4.16: Used for broadcast packet reduction tests.
func (m *MockConn) WriteCount() int {
	return m.writeCount
}

// ResetWriteCount resets the write counter to zero.
// Phase 4.16: Called between broadcast tests to isolate measurements.
func (m *MockConn) ResetWriteCount() {
	m.writeCount = 0
}

// Close закрывает соединение (no-op).
func (m *MockConn) Close() error {
	return nil
}

// LocalAddr возвращает локальный адрес (mock).
func (m *MockConn) LocalAddr() net.Addr {
	return &mockAddr{network: "tcp", address: "127.0.0.1:7777"}
}

// RemoteAddr возвращает удалённый адрес (mock).
func (m *MockConn) RemoteAddr() net.Addr {
	return &mockAddr{network: "tcp", address: "192.168.1.100:12345"}
}

// SetDeadline устанавливает deadline (no-op).
func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline устанавливает read deadline (no-op).
func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline устанавливает write deadline (no-op).
func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// mockAddr — mock для net.Addr.
type mockAddr struct {
	network string
	address string
}

// Network возвращает имя сети.
func (a *mockAddr) Network() string {
	return a.network
}

// String возвращает строковое представление адреса.
func (a *mockAddr) String() string {
	return a.address
}
