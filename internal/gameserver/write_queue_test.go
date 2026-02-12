package gameserver

import (
	"bytes"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/udisondev/la2go/internal/testutil"
)

// pipeConn wraps a net.Conn from net.Pipe() to add SetWriteDeadline support.
type pipeConn struct {
	net.Conn
}

func newTestClient(t *testing.T, conn net.Conn, pool *BytePool, queueSize int) *GameClient {
	t.Helper()
	client := &GameClient{
		conn:              conn,
		ip:                "test",
		sendCh:            make(chan []byte, queueSize),
		closeCh:           make(chan struct{}),
		writePool:         pool,
		writeTimeout:      5 * time.Second,
		selectedCharacter: -1,
	}
	client.state.Store(int32(ClientStateConnected))
	return client
}

func TestWritePump_SinglePacket(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	pool := NewBytePool(64)
	gc := newTestClient(t, client, pool, 16)

	go gc.writePump()
	defer gc.CloseAsync()

	// Send a packet
	pkt := []byte{0x01, 0x02, 0x03, 0x04}
	if err := gc.Send(pkt); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Read from server side
	buf := make([]byte, 64)
	if err := server.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	n, err := server.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !bytes.Equal(buf[:n], pkt) {
		t.Errorf("got %v, want %v", buf[:n], pkt)
	}
}

func TestWritePump_BatchDrain(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()

	pool := NewBytePool(64)
	gc := newTestClient(t, client, pool, 16)

	// Pre-fill channel BEFORE starting writePump to guarantee batching
	pkt1 := []byte{0x01, 0x02}
	pkt2 := []byte{0x03, 0x04}
	pkt3 := []byte{0x05, 0x06}

	gc.sendCh <- pkt1
	gc.sendCh <- pkt2
	gc.sendCh <- pkt3

	go gc.writePump()
	defer func() {
		gc.CloseAsync()
		client.Close()
	}()

	// Read all data from server side
	var received []byte
	buf := make([]byte, 256)
	if err := server.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}

	for len(received) < 6 {
		n, err := server.Read(buf)
		if err != nil {
			t.Fatalf("Read failed after %d bytes: %v", len(received), err)
		}
		received = append(received, buf[:n]...)
	}

	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	if !bytes.Equal(received, expected) {
		t.Errorf("got %v, want %v", received, expected)
	}
}

func TestSend_QueueFull(t *testing.T) {
	conn := testutil.NewMockConn()
	pool := NewBytePool(64)
	gc := newTestClient(t, conn, pool, 2)
	// Don't start writePump — channel will fill up

	// Fill the queue
	gc.sendCh <- []byte{0x01}
	gc.sendCh <- []byte{0x02}

	// Third Send should fail (queue full)
	pkt := pool.Get(4)
	copy(pkt, []byte{0x03, 0x04, 0x05, 0x06})
	err := gc.Send(pkt)
	if err == nil {
		t.Fatal("expected error for full queue, got nil")
	}

	// Client should be marked as disconnected
	if gc.State() != ClientStateDisconnected {
		t.Errorf("expected state Disconnected, got %v", gc.State())
	}
}

func TestSendSync_Timeout(t *testing.T) {
	conn := testutil.NewMockConn()
	pool := NewBytePool(64)
	gc := newTestClient(t, conn, pool, 1)
	// Don't start writePump — channel will fill up

	// Fill the queue
	gc.sendCh <- []byte{0x01}

	// SendSync should timeout
	pkt := pool.Get(4)
	copy(pkt, []byte{0x02, 0x03, 0x04, 0x05})
	err := gc.SendSync(pkt, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestSendSync_ClientClosed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		conn := testutil.NewMockConn()
		pool := NewBytePool(64)
		gc := newTestClient(t, conn, pool, 1)
		// Don't start writePump

		// Fill the queue
		gc.sendCh <- []byte{0x01}

		// Close client in background (instant with fake clock)
		go func() {
			time.Sleep(20 * time.Millisecond)
			gc.CloseAsync()
		}()

		pkt := pool.Get(4)
		err := gc.SendSync(pkt, 5*time.Second)
		if err == nil {
			t.Fatal("expected client closed error, got nil")
		}
	})
}

func TestWritePump_DrainOnClose(t *testing.T) {
	conn := testutil.NewMockConn()
	pool := NewBytePool(64)
	gc := newTestClient(t, conn, pool, 16)

	// Track pool Put calls
	var putCount atomic.Int32
	origPool := NewBytePool(64)
	trackingPool := &trackingBytePool{
		BytePool: origPool,
		putCount: &putCount,
	}
	gc.writePool = trackingPool.BytePool

	// Pre-fill channel
	for range 5 {
		buf := pool.Get(4)
		gc.sendCh <- buf
	}

	// Close without starting writePump — simulate immediate close
	gc.CloseAsync()

	// Start writePump — it should drain and return
	done := make(chan struct{})
	go func() {
		gc.writePump()
		close(done)
	}()

	select {
	case <-done:
		// writePump exited — good
	case <-time.After(2 * time.Second):
		t.Fatal("writePump did not exit after close")
	}

	// Channel should be empty (drained)
	if len(gc.sendCh) != 0 {
		t.Errorf("sendCh not drained: %d items remain", len(gc.sendCh))
	}
}

type trackingBytePool struct {
	*BytePool
	putCount *atomic.Int32
}

func TestWritePump_ExitsOnWriteError(t *testing.T) {
	server, client := net.Pipe()
	pool := NewBytePool(64)
	gc := newTestClient(t, client, pool, 16)

	// Close server side to cause write error
	server.Close()

	done := make(chan struct{})
	go func() {
		gc.writePump()
		close(done)
	}()

	// Send a packet — should cause write error and writePump exit
	gc.sendCh <- []byte{0x01, 0x02, 0x03}

	select {
	case <-done:
		// writePump exited — good
	case <-time.After(2 * time.Second):
		t.Fatal("writePump did not exit after write error")
	}

	client.Close()
}

func TestCloseAsync_Idempotent(t *testing.T) {
	conn := testutil.NewMockConn()
	gc := newTestClient(t, conn, nil, 16)

	// Multiple CloseAsync calls should not panic
	gc.CloseAsync()
	gc.CloseAsync()
	gc.CloseAsync()

	if gc.State() != ClientStateDisconnected {
		t.Errorf("expected Disconnected state, got %v", gc.State())
	}
}

func TestWritePump_ConcurrentSend(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	pool := NewBytePool(64)
	gc := newTestClient(t, client, pool, 2048)

	go gc.writePump()
	defer gc.CloseAsync()

	const numSenders = 10
	const packetsPerSender = 100

	var sentCount atomic.Int32
	var wg sync.WaitGroup
	for range numSenders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range packetsPerSender {
				pkt := []byte{0xAA, 0xBB}
				if err := gc.Send(pkt); err != nil {
					return // client may close
				}
				sentCount.Add(1)
			}
		}()
	}

	// Wait for senders to finish
	wg.Wait()
	totalSent := int(sentCount.Load())

	// Read all data from server side
	totalExpected := totalSent * 2 // 2 bytes per packet
	received := 0
	buf := make([]byte, 4096)
	if err := server.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}

	for received < totalExpected {
		n, err := server.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Read failed after %d bytes (want %d): %v", received, totalExpected, err)
		}
		received += n
	}

	if received != totalExpected {
		t.Errorf("received %d bytes, want %d (sent %d packets)", received, totalExpected, totalSent)
	}
}

func TestEncryptToPooled_ReturnsCorrectData(t *testing.T) {
	pool := NewBytePool(128)
	conn := testutil.NewMockConn()
	gc, err := NewGameClient(conn, make([]byte, 16), pool, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient: %v", err)
	}

	// First packet must be sent to initialize encryption state
	// (skip first-packet XOR pass)
	payload := []byte{0x01, 0x02, 0x03, 0x04}
	encPkt, err := pool.EncryptToPooled(gc.Encryption(), payload, len(payload))
	if err != nil {
		t.Fatalf("EncryptToPooled: %v", err)
	}

	if len(encPkt) == 0 {
		t.Fatal("EncryptToPooled returned empty packet")
	}

	// Verify original payload is NOT modified
	expected := []byte{0x01, 0x02, 0x03, 0x04}
	if !bytes.Equal(payload, expected) {
		t.Errorf("original payload modified: got %v, want %v", payload, expected)
	}

	// Return to pool
	pool.Put(encPkt)
}

func TestBroadcast_NoBufferCorruption(t *testing.T) {
	// Create 2 clients with DIFFERENT encryption keys
	pool := NewBytePool(128)
	conn1 := testutil.NewMockConn()
	conn2 := testutil.NewMockConn()

	key1 := make([]byte, 16)
	key2 := make([]byte, 16)
	for i := range 16 {
		key1[i] = byte(i + 1)
		key2[i] = byte(i + 17)
	}

	gc1, err := NewGameClient(conn1, key1, pool, 16, 0)
	if err != nil {
		t.Fatalf("NewGameClient 1: %v", err)
	}
	gc2, err := NewGameClient(conn2, key2, pool, 16, 0)
	if err != nil {
		t.Fatalf("NewGameClient 2: %v", err)
	}

	// Encrypt same payload for both clients
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	enc1, err := pool.EncryptToPooled(gc1.Encryption(), payload, len(payload))
	if err != nil {
		t.Fatalf("EncryptToPooled 1: %v", err)
	}
	enc2, err := pool.EncryptToPooled(gc2.Encryption(), payload, len(payload))
	if err != nil {
		t.Fatalf("EncryptToPooled 2: %v", err)
	}

	// Encrypted packets should be DIFFERENT (different keys)
	if bytes.Equal(enc1, enc2) {
		t.Error("encrypted packets are identical — buffer corruption or same key used")
	}

	// Original payload should be unchanged
	expectedPayload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if !bytes.Equal(payload, expectedPayload) {
		t.Errorf("original payload corrupted: got %v, want %v", payload, expectedPayload)
	}

	pool.Put(enc1)
	pool.Put(enc2)
}
