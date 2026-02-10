package protocol

import (
	"bytes"
	"testing"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
)

// TestEncryptInPlace verifies that EncryptInPlace produces the same result as WritePacket.
func TestEncryptInPlace(t *testing.T) {
	// Create encryption with dynamic key
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	enc1, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		t.Fatalf("NewLoginEncryption failed: %v", err)
	}

	enc2, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		t.Fatalf("NewLoginEncryption failed: %v", err)
	}

	// Skip firstPacket for both (sendVisibleObjectsInfo calls after authentication)
	dummyBuf := make([]byte, 1024)
	_, _ = enc1.EncryptPacket(dummyBuf, constants.PacketHeaderSize, 8) // firstPacket consumed
	_, _ = enc2.EncryptPacket(dummyBuf, constants.PacketHeaderSize, 8) // firstPacket consumed

	// Test payload
	payload := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	payloadLen := len(payload)

	// Method 1: EncryptInPlace
	buf1 := make([]byte, 1024)
	copy(buf1[constants.PacketHeaderSize:], payload)
	encSize1, err := EncryptInPlace(enc1, buf1, payloadLen)
	if err != nil {
		t.Fatalf("EncryptInPlace failed: %v", err)
	}

	// Method 2: WritePacket
	buf2 := make([]byte, 1024)
	copy(buf2[constants.PacketHeaderSize:], payload)
	var output bytes.Buffer
	if err := WritePacket(&output, enc2, buf2, payloadLen); err != nil {
		t.Fatalf("WritePacket failed: %v", err)
	}

	// Compare results
	result1 := buf1[:encSize1]
	result2 := output.Bytes()

	if !bytes.Equal(result1, result2) {
		t.Errorf("EncryptInPlace result differs from WritePacket\nEncryptInPlace: %x\nWritePacket: %x",
			result1, result2)
	}
}

// TestEncryptInPlace_BufferTooSmall verifies error handling for small buffers.
func TestEncryptInPlace_BufferTooSmall(t *testing.T) {
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	enc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		t.Fatalf("NewLoginEncryption failed: %v", err)
	}

	// Buffer too small
	buf := make([]byte, 10)
	_, err = EncryptInPlace(enc, buf, 100)
	if err == nil {
		t.Error("EncryptInPlace should fail with small buffer, got nil error")
	}
}

// TestWriteEncrypted verifies that WriteEncrypted successfully sends pre-encrypted data.
func TestWriteEncrypted(t *testing.T) {
	// Pre-encrypted data (simulated)
	encryptedData := []byte{0x00, 0x10, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}

	var output bytes.Buffer
	err := WriteEncrypted(&output, encryptedData, len(encryptedData))
	if err != nil {
		t.Fatalf("WriteEncrypted failed: %v", err)
	}

	if !bytes.Equal(output.Bytes(), encryptedData) {
		t.Errorf("WriteEncrypted output mismatch\nExpected: %x\nGot: %x",
			encryptedData, output.Bytes())
	}
}

// TestWriteEncrypted_PartialBuffer verifies that WriteEncrypted respects encryptedSize.
func TestWriteEncrypted_PartialBuffer(t *testing.T) {
	// Buffer with padding
	buf := make([]byte, 100)
	for i := range 10 {
		buf[i] = byte(i)
	}

	var output bytes.Buffer
	err := WriteEncrypted(&output, buf, 10) // Only write first 10 bytes
	if err != nil {
		t.Fatalf("WriteEncrypted failed: %v", err)
	}

	expected := buf[:10]
	if !bytes.Equal(output.Bytes(), expected) {
		t.Errorf("WriteEncrypted wrote wrong data\nExpected: %x\nGot: %x",
			expected, output.Bytes())
	}

	if output.Len() != 10 {
		t.Errorf("WriteEncrypted wrote wrong size: expected 10, got %d", output.Len())
	}
}

// TestEncryptInPlace_AfterAuthentication simulates sendVisibleObjectsInfo scenario.
func TestEncryptInPlace_AfterAuthentication(t *testing.T) {
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	enc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		t.Fatalf("NewLoginEncryption failed: %v", err)
	}

	// Simulate firstPacket consumption (authentication flow)
	dummyBuf := make([]byte, 1024)
	_, _ = enc.EncryptPacket(dummyBuf, constants.PacketHeaderSize, 8)

	// Now encrypt multiple packets (like sendVisibleObjectsInfo)
	for i := range 10 {
		payload := []byte{byte(i), 0xAA, 0xBB, 0xCC}
		buf := make([]byte, 1024)
		copy(buf[constants.PacketHeaderSize:], payload)

		encSize, err := EncryptInPlace(enc, buf, len(payload))
		if err != nil {
			t.Fatalf("EncryptInPlace[%d] failed: %v", i, err)
		}

		if encSize < len(payload) {
			t.Errorf("EncryptInPlace[%d] returned invalid size: %d < %d", i, encSize, len(payload))
		}
	}
}

// TestWriteBatch verifies that WriteBatch correctly sends multiple packets.
func TestWriteBatch(t *testing.T) {
	// Create 3 packets
	packet1 := []byte{0x00, 0x08, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	packet2 := []byte{0x00, 0x06, 0x11, 0x22, 0x33, 0x44}
	packet3 := []byte{0x00, 0x04, 0x55, 0x66}

	packets := [][]byte{packet1, packet2, packet3}

	var output bytes.Buffer
	err := WriteBatch(&output, packets)
	if err != nil {
		t.Fatalf("WriteBatch failed: %v", err)
	}

	// Expected: concatenation of all packets
	expected := append(append([]byte{}, packet1...), packet2...)
	expected = append(expected, packet3...)

	if !bytes.Equal(output.Bytes(), expected) {
		t.Errorf("WriteBatch output mismatch\nExpected: %x\nGot: %x",
			expected, output.Bytes())
	}
}

// TestWriteBatch_Empty verifies that WriteBatch handles empty packet list.
func TestWriteBatch_Empty(t *testing.T) {
	var output bytes.Buffer
	err := WriteBatch(&output, nil)
	if err != nil {
		t.Fatalf("WriteBatch with empty packets should succeed, got error: %v", err)
	}

	if output.Len() != 0 {
		t.Errorf("WriteBatch with empty packets should write nothing, got %d bytes", output.Len())
	}
}

// TestWriteBatch_Single verifies that WriteBatch works with single packet.
func TestWriteBatch_Single(t *testing.T) {
	packet := []byte{0x00, 0x08, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}

	var output bytes.Buffer
	err := WriteBatch(&output, [][]byte{packet})
	if err != nil {
		t.Fatalf("WriteBatch with single packet failed: %v", err)
	}

	if !bytes.Equal(output.Bytes(), packet) {
		t.Errorf("WriteBatch single packet mismatch\nExpected: %x\nGot: %x",
			packet, output.Bytes())
	}
}

// TestWriteBatch_Large verifies that WriteBatch handles many packets.
func TestWriteBatch_Large(t *testing.T) {
	// Create 450 packets (like sendVisibleObjectsInfo)
	packets := make([][]byte, 450)
	expectedSize := 0

	for i := range 450 {
		// Packet: [size][opcode][payload...]
		packet := make([]byte, 10)
		packet[0] = 0x00
		packet[1] = 0x0A // 10 bytes
		packet[2] = byte(i % 256)
		packets[i] = packet
		expectedSize += len(packet)
	}

	var output bytes.Buffer
	err := WriteBatch(&output, packets)
	if err != nil {
		t.Fatalf("WriteBatch with 450 packets failed: %v", err)
	}

	if output.Len() != expectedSize {
		t.Errorf("WriteBatch wrote wrong total size: expected %d, got %d",
			expectedSize, output.Len())
	}
}
