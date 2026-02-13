package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestDie_Write(t *testing.T) {
	t.Parallel()

	pkt := &Die{
		ObjectID:    12345,
		CanTeleport: true,
		ToHideaway:  false,
		ToCastle:    true,
		ToSiegeHQ:   false,
		Sweepable:   false,
		FixedRes:    true,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != OpcodeDie {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeDie)
	}

	objID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objID != 12345 {
		t.Errorf("ObjectID = %d; want 12345", objID)
	}

	canTeleport := int32(binary.LittleEndian.Uint32(data[5:9]))
	if canTeleport != 1 {
		t.Errorf("CanTeleport = %d; want 1", canTeleport)
	}

	toHideaway := int32(binary.LittleEndian.Uint32(data[9:13]))
	if toHideaway != 0 {
		t.Errorf("ToHideaway = %d; want 0", toHideaway)
	}

	toCastle := int32(binary.LittleEndian.Uint32(data[13:17]))
	if toCastle != 1 {
		t.Errorf("ToCastle = %d; want 1", toCastle)
	}

	toSiegeHQ := int32(binary.LittleEndian.Uint32(data[17:21]))
	if toSiegeHQ != 0 {
		t.Errorf("ToSiegeHQ = %d; want 0", toSiegeHQ)
	}

	sweepable := int32(binary.LittleEndian.Uint32(data[21:25]))
	if sweepable != 0 {
		t.Errorf("Sweepable = %d; want 0", sweepable)
	}

	fixedRes := int32(binary.LittleEndian.Uint32(data[25:29]))
	if fixedRes != 1 {
		t.Errorf("FixedRes = %d; want 1", fixedRes)
	}

	// Total size: 1 + 7*4 = 29 bytes
	if len(data) != 29 {
		t.Errorf("data length = %d; want 29", len(data))
	}
}

func TestDie_WriteAllFalse(t *testing.T) {
	t.Parallel()

	pkt := &Die{ObjectID: 1}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// All booleans should be 0
	for i := 5; i < 29; i += 4 {
		val := int32(binary.LittleEndian.Uint32(data[i : i+4]))
		if val != 0 {
			t.Errorf("field at offset %d = %d; want 0", i, val)
		}
	}
}

func TestRevive_Write(t *testing.T) {
	t.Parallel()

	pkt := &Revive{ObjectID: 9999}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != OpcodeRevive {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeRevive)
	}

	objID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objID != 9999 {
		t.Errorf("ObjectID = %d; want 9999", objID)
	}

	if len(data) != 5 {
		t.Errorf("data length = %d; want 5", len(data))
	}
}
