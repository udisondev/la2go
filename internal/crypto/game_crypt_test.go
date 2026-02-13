package crypto

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestGameCrypt_FirstEncryptSkipped(t *testing.T) {
	gc := NewGameCrypt()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	gc.SetKey(key)

	original := []byte{0x01, 0x02, 0x03, 0x04}
	data := make([]byte, len(original))
	copy(data, original)

	gc.Encrypt(data)

	// First call must be skipped — data unchanged
	if !bytes.Equal(data, original) {
		t.Fatalf("first Encrypt must be no-op: got %x, want %x", data, original)
	}
	if !gc.IsEnabled() {
		t.Fatal("IsEnabled must be true after first Encrypt")
	}
}

func TestGameCrypt_DecryptBeforeEnableIsNoop(t *testing.T) {
	gc := NewGameCrypt()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	gc.SetKey(key)

	original := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	data := make([]byte, len(original))
	copy(data, original)

	gc.Decrypt(data)

	if !bytes.Equal(data, original) {
		t.Fatalf("Decrypt before enable must be no-op: got %x, want %x", data, original)
	}
}

func TestGameCrypt_EncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i*7 + 3)
	}

	// Server-side crypt (encrypts outgoing, decrypts incoming)
	server := NewGameCrypt()
	server.SetKey(key)

	// Client-side crypt (decrypts server's outgoing, encrypts own outgoing)
	client := NewGameCrypt()
	client.SetKey(key)

	// Skip first encrypt on both sides (Init packet)
	server.Encrypt([]byte{0x00})
	client.Encrypt([]byte{0x00})

	// Server encrypts a packet → client decrypts it
	original := []byte("Hello L2 World! This is a test packet from server.")
	packet := make([]byte, len(original))
	copy(packet, original)

	server.Encrypt(packet)

	// Packet must be modified
	if bytes.Equal(packet, original) {
		t.Fatal("encrypted data must differ from original")
	}

	client.Decrypt(packet)

	if !bytes.Equal(packet, original) {
		t.Fatalf("round-trip failed: got %x, want %x", packet, original)
	}
}

func TestGameCrypt_MultiplePacketsRoundTrip(t *testing.T) {
	key := []byte{
		0x94, 0x35, 0x00, 0x00, 0xa1, 0x6c, 0x54, 0x87,
		0x09, 0x12, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45,
	}

	server := NewGameCrypt()
	server.SetKey(key)
	client := NewGameCrypt()
	client.SetKey(key)

	// Skip first encrypt
	server.Encrypt([]byte{0x00})
	client.Encrypt([]byte{0x00})

	packets := [][]byte{
		{0x04, 0x01, 0x02, 0x03}, // short
		make([]byte, 100),         // medium
		make([]byte, 1000),        // large
	}
	for i := range packets[1] {
		packets[1][i] = byte(i)
	}
	for i := range packets[2] {
		packets[2][i] = byte(i * 3)
	}

	for idx, original := range packets {
		pkt := make([]byte, len(original))
		copy(pkt, original)

		server.Encrypt(pkt)
		client.Decrypt(pkt)

		if !bytes.Equal(pkt, original) {
			t.Fatalf("packet %d round-trip failed at byte %d", idx, firstDiff(pkt, original))
		}
	}
}

func TestGameCrypt_KeyShift(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i)
	}

	gc := NewGameCrypt()
	gc.SetKey(key)

	// Skip first
	gc.Encrypt([]byte{0x00})

	// Read initial key[8:12] value
	initialVal := binary.LittleEndian.Uint32(gc.outKey[8:12])

	// Encrypt 10 bytes
	data := make([]byte, 10)
	gc.Encrypt(data)

	newVal := binary.LittleEndian.Uint32(gc.outKey[8:12])
	if newVal != initialVal+10 {
		t.Fatalf("key shift: got %d, want %d", newVal, initialVal+10)
	}

	// Encrypt another 20 bytes
	data2 := make([]byte, 20)
	gc.Encrypt(data2)

	newVal2 := binary.LittleEndian.Uint32(gc.outKey[8:12])
	if newVal2 != initialVal+30 {
		t.Fatalf("key shift after 2nd encrypt: got %d, want %d", newVal2, initialVal+30)
	}
}

func TestGameCrypt_KnownVector(t *testing.T) {
	// Manually compute expected values to verify the XOR rolling algorithm.
	// Key: 0x00..0x0F, encrypt data: {0x41, 0x42, 0x43, 0x44}
	key := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}

	gc := NewGameCrypt()
	gc.SetKey(key)
	gc.Encrypt([]byte{0x00}) // skip first

	data := []byte{0x41, 0x42, 0x43, 0x44}

	// Manual calculation:
	// i=0: prev=0, encrypted = 0x41 ^ key[0]=0x00 ^ 0x00 = 0x41, prev=0x41
	// i=1: encrypted = 0x42 ^ key[1]=0x01 ^ 0x41 = 0x02, prev=0x02
	// i=2: encrypted = 0x43 ^ key[2]=0x02 ^ 0x02 = 0x43, prev=0x43
	// i=3: encrypted = 0x44 ^ key[3]=0x03 ^ 0x43 = 0x04, prev=0x04
	expected := []byte{0x41, 0x02, 0x43, 0x04}

	gc.Encrypt(data)

	if !bytes.Equal(data, expected) {
		t.Fatalf("known vector encrypt: got %x, want %x", data, expected)
	}

	// Now decrypt with fresh crypt (same key)
	gc2 := NewGameCrypt()
	gc2.SetKey(key)
	gc2.Encrypt([]byte{0x00}) // enable

	gc2.Decrypt(data)
	original := []byte{0x41, 0x42, 0x43, 0x44}
	if !bytes.Equal(data, original) {
		t.Fatalf("known vector decrypt: got %x, want %x", data, original)
	}
}

func TestGameCrypt_BidirectionalCommunication(t *testing.T) {
	// Simulate full bidirectional communication:
	// Server and client exchange packets in both directions.
	key := []byte{
		0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
	}

	serverCrypt := NewGameCrypt()
	serverCrypt.SetKey(key)
	clientCrypt := NewGameCrypt()
	clientCrypt.SetKey(key)

	// Both skip first encrypt
	serverCrypt.Encrypt([]byte{0x00})
	clientCrypt.Encrypt([]byte{0x00})

	// Server → Client: 3 packets
	for i := range 3 {
		original := []byte{byte(i), 0x10, 0x20, 0x30, 0x40, 0x50}
		pkt := make([]byte, len(original))
		copy(pkt, original)

		serverCrypt.Encrypt(pkt)
		clientCrypt.Decrypt(pkt)

		if !bytes.Equal(pkt, original) {
			t.Fatalf("server→client packet %d failed", i)
		}
	}

	// Client → Server: 3 packets
	for i := range 3 {
		original := []byte{byte(i + 100), 0xAA, 0xBB, 0xCC}
		pkt := make([]byte, len(original))
		copy(pkt, original)

		clientCrypt.Encrypt(pkt)
		serverCrypt.Decrypt(pkt)

		if !bytes.Equal(pkt, original) {
			t.Fatalf("client→server packet %d failed", i)
		}
	}
}

func TestGameCrypt_EmptyData(t *testing.T) {
	gc := NewGameCrypt()
	gc.SetKey(make([]byte, 16))
	gc.Encrypt([]byte{0x00}) // enable

	// Empty slice should not panic
	gc.Encrypt([]byte{})
	gc.Decrypt([]byte{})
}

func firstDiff(a, b []byte) int {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i
		}
	}
	return len(a)
}
