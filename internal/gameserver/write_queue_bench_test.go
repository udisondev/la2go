package gameserver

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/testutil"
)

// BenchmarkEncryptToPooled measures overhead of BytePool.EncryptToPooled().
// Called for EACH client during broadcast — critical hot path (Phase 7.0).
// Expected: 0 allocs/op in steady state (pool reuse).
func BenchmarkEncryptToPooled(b *testing.B) {
	pool := NewBytePool(128)
	conn := testutil.NewMockConn()
	gc, err := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
	if err != nil {
		b.Fatalf("NewGameClient: %v", err)
	}

	payload := []byte{0x01, 0x02, 0x03, 0x04}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		enc, err := pool.EncryptToPooled(gc.Encryption(), payload, len(payload))
		if err != nil {
			b.Fatal(err)
		}
		pool.Put(enc)
	}
}

// BenchmarkEncryptToPooled_PayloadSizes measures EncryptToPooled for different payload sizes.
// Covers typical packet sizes: CharMoveToLocation ~28B, StatusUpdate ~57B, UserInfo ~500B.
func BenchmarkEncryptToPooled_PayloadSizes(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"small_28B", 28},  // CharMoveToLocation
		{"medium_57B", 57}, // StatusUpdate (6 attrs)
		{"large_500B", 500}, // UserInfo
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			pool := NewBytePool(1024)
			conn := testutil.NewMockConn()
			gc, err := NewGameClient(conn, make([]byte, 16), pool, 16, 0)
			if err != nil {
				b.Fatalf("NewGameClient: %v", err)
			}

			payload := make([]byte, s.size)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				enc, err := pool.EncryptToPooled(gc.Encryption(), payload, len(payload))
				if err != nil {
					b.Fatal(err)
				}
				pool.Put(enc)
			}
		})
	}
}

// BenchmarkSend_NonBlocking measures overhead of non-blocking Send().
// Hot path: select/default on buffered channel for each outgoing packet.
// Expected: <100ns, 0 allocs/op.
func BenchmarkSend_NonBlocking(b *testing.B) {
	pool := NewBytePool(128)
	conn := testutil.NewMockConn()
	gc, err := NewGameClient(conn, make([]byte, 16), pool, 2048, 0)
	if err != nil {
		b.Fatalf("NewGameClient: %v", err)
	}

	go gc.writePump()
	defer gc.CloseAsync()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		pkt := pool.Get(4)
		_ = gc.Send(pkt)
	}
}

// newBenchClient creates a GameClient for benchmark tests (avoids *testing.T requirement).
func newBenchClient(conn net.Conn, pool *BytePool, queueSize int) *GameClient {
	client := &GameClient{
		conn:              conn,
		ip:                "bench",
		sendCh:            make(chan []byte, queueSize),
		closeCh:           make(chan struct{}),
		writePool:         pool,
		writeTimeout:      5 * time.Second,
		selectedCharacter: -1,
	}
	client.state.Store(int32(ClientStateConnected))
	return client
}

// BenchmarkWritePump_Throughput measures writePump drain + batch write throughput.
// Uses net.Pipe() for realistic I/O with a drain reader goroutine.
func BenchmarkWritePump_Throughput(b *testing.B) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	pool := NewBytePool(128)
	gc := newBenchClient(client, pool, 4096)

	go gc.writePump()
	defer gc.CloseAsync()

	// Drain reader goroutine
	go func() {
		_, _ = io.Copy(io.Discard, server)
	}()

	pkt := []byte{0x01, 0x02, 0x03, 0x04}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = gc.Send(pkt)
	}
}

// BenchmarkWritePump_Throughput_LargePacket measures writePump with ~500B packets (UserInfo size).
func BenchmarkWritePump_Throughput_LargePacket(b *testing.B) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	pool := NewBytePool(1024)
	gc := newBenchClient(client, pool, 4096)

	go gc.writePump()
	defer gc.CloseAsync()

	go func() {
		_, _ = io.Copy(io.Discard, server)
	}()

	pkt := make([]byte, 500)
	for i := range pkt {
		pkt[i] = byte(i % 256)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = gc.Send(pkt)
	}
}
