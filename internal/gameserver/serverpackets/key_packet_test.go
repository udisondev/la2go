package serverpackets

import (
	"testing"
)

func TestKeyPacket_Write(t *testing.T) {
	blowfishKey := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
	}

	pkt := NewKeyPacket(blowfishKey)

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("KeyPacket.Write failed: %v", err)
	}

	// Expected length: 1 (opcode) + 1 (protocol version) + 16 (blowfish key) = 18 bytes
	expectedLen := 18
	if len(data) != expectedLen {
		t.Fatalf("expected length %d, got %d", expectedLen, len(data))
	}

	// Verify opcode (0x2E)
	if data[0] != OpcodeKeyPacket {
		t.Errorf("expected opcode 0x%02X, got 0x%02X", OpcodeKeyPacket, data[0])
	}

	// Verify protocol version (0x01 for Interlude)
	if data[1] != 0x01 {
		t.Errorf("expected protocol version 0x01, got 0x%02X", data[1])
	}

	// Verify Blowfish key
	for i, expected := range blowfishKey {
		offset := 2 + i
		if data[offset] != expected {
			t.Errorf("at key byte %d: expected 0x%02X, got 0x%02X", i, expected, data[offset])
		}
	}
}

func TestKeyPacket_Write_ZeroKey(t *testing.T) {
	blowfishKey := make([]byte, 16) // all zeros

	pkt := NewKeyPacket(blowfishKey)

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("KeyPacket.Write failed: %v", err)
	}

	// Should still work (even though zeros are not recommended)
	if len(data) != 18 {
		t.Fatalf("expected length 18, got %d", len(data))
	}

	// Verify all key bytes are zero
	for i := range 16 {
		offset := 2 + i
		if data[offset] != 0x00 {
			t.Errorf("at key byte %d: expected 0x00, got 0x%02X", i, data[offset])
		}
	}
}
