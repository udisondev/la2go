package testutil

import (
	"net"
	"testing"
	"time"
)

// PipeConn создаёт пару net.Conn соединений через net.Pipe для тестирования.
// Автоматически закрывает соединения при завершении теста.
func PipeConn(t testing.TB) (client, server net.Conn) {
	t.Helper()

	server, client = net.Pipe()

	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
	})

	return client, server
}

// FakeAddr реализует net.Addr для тестов.
type FakeAddr struct {
	NetworkName string
	AddrString  string
}

func (f FakeAddr) Network() string { return f.NetworkName }
func (f FakeAddr) String() string  { return f.AddrString }

// NewFakeAddr создаёт FakeAddr.
func NewFakeAddr(network, addr string) FakeAddr {
	return FakeAddr{
		NetworkName: network,
		AddrString:  addr,
	}
}

// TCPAddr создаёт FakeAddr для TCP соединения.
func TCPAddr(addr string) FakeAddr {
	return NewFakeAddr("tcp", addr)
}

// ConnWithDeadline оборачивает net.Conn и автоматически устанавливает deadline для read/write.
type ConnWithDeadline struct {
	net.Conn
	deadline time.Duration
}

// NewConnWithDeadline создаёт обёртку с автоматическим deadline.
func NewConnWithDeadline(conn net.Conn, deadline time.Duration) *ConnWithDeadline {
	return &ConnWithDeadline{
		Conn:     conn,
		deadline: deadline,
	}
}

func (c *ConnWithDeadline) Read(b []byte) (int, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.deadline)); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *ConnWithDeadline) Write(b []byte) (int, error) {
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.deadline)); err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

// ListenTCP создаёт TCP listener на случайном порту для тестов.
// Возвращает listener и адрес в формате "host:port".
// Автоматически закрывает listener при завершении теста.
func ListenTCP(t testing.TB) (net.Listener, string) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create TCP listener: %v", err)
	}

	t.Cleanup(func() {
		_ = listener.Close()
	})

	return listener, listener.Addr().String()
}
