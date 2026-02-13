package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestTeleportToLocation_Write(t *testing.T) {
	pkt := NewTeleportToLocation(12345, 100, 200, 300)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// 1 opcode + 6 * 4 bytes = 25
	if len(data) != 25 {
		t.Fatalf("packet length = %d, want 25", len(data))
	}

	// Verify opcode
	if data[0] != OpcodeTeleportToLocation {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeTeleportToLocation)
	}

	// Verify objectID
	objectID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objectID != 12345 {
		t.Errorf("objectID = %d, want 12345", objectID)
	}

	// Verify X
	x := int32(binary.LittleEndian.Uint32(data[5:9]))
	if x != 100 {
		t.Errorf("X = %d, want 100", x)
	}

	// Verify Y
	y := int32(binary.LittleEndian.Uint32(data[9:13]))
	if y != 200 {
		t.Errorf("Y = %d, want 200", y)
	}

	// Verify Z
	z := int32(binary.LittleEndian.Uint32(data[13:17]))
	if z != 300 {
		t.Errorf("Z = %d, want 300", z)
	}

	// Verify fade (default 0)
	fade := int32(binary.LittleEndian.Uint32(data[17:21]))
	if fade != 0 {
		t.Errorf("fade = %d, want 0", fade)
	}

	// Verify heading (default 0)
	heading := int32(binary.LittleEndian.Uint32(data[21:25]))
	if heading != 0 {
		t.Errorf("heading = %d, want 0", heading)
	}
}

func TestTeleportToLocation_Write_WithHeading(t *testing.T) {
	pkt := TeleportToLocation{
		ObjectID: 999,
		X:        -80749,
		Y:        149834,
		Z:        -3043,
		Fade:     1, // instant
		Heading:  32000,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	objectID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objectID != 999 {
		t.Errorf("objectID = %d, want 999", objectID)
	}

	x := int32(binary.LittleEndian.Uint32(data[5:9]))
	if x != -80749 {
		t.Errorf("X = %d, want -80749", x)
	}

	y := int32(binary.LittleEndian.Uint32(data[9:13]))
	if y != 149834 {
		t.Errorf("Y = %d, want 149834", y)
	}

	z := int32(binary.LittleEndian.Uint32(data[13:17]))
	if z != -3043 {
		t.Errorf("Z = %d, want -3043", z)
	}

	fade := int32(binary.LittleEndian.Uint32(data[17:21]))
	if fade != 1 {
		t.Errorf("fade = %d, want 1 (instant)", fade)
	}

	heading := int32(binary.LittleEndian.Uint32(data[21:25]))
	if heading != 32000 {
		t.Errorf("heading = %d, want 32000", heading)
	}
}

func TestTeleportToLocation_Write_NegativeCoords(t *testing.T) {
	pkt := NewTeleportToLocation(1, -12694, -178112, -916)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	x := int32(binary.LittleEndian.Uint32(data[5:9]))
	if x != -12694 {
		t.Errorf("X = %d, want -12694", x)
	}

	y := int32(binary.LittleEndian.Uint32(data[9:13]))
	if y != -178112 {
		t.Errorf("Y = %d, want -178112", y)
	}

	z := int32(binary.LittleEndian.Uint32(data[13:17]))
	if z != -916 {
		t.Errorf("Z = %d, want -916", z)
	}
}
