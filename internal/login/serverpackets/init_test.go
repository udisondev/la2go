package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/constants"
)

func TestInit(t *testing.T) {
	sessionID := int32(0x12345678)
	scrambledModulus := make([]byte, constants.RSA1024ModulusSize)
	for i := range scrambledModulus {
		scrambledModulus[i] = byte(i)
	}
	blowfishKey := []byte{
		0x04, 0xa1, 0xc3, 0x42, 0xad, 0xaa, 0xf2, 0x34,
		0x30, 0x78, 0x9f, 0x61, 0xb8, 0x92, 0x53, 0x32,
	}

	buf := make([]byte, constants.TestInitPacketBufSize)
	n := Init(buf, sessionID, scrambledModulus, blowfishKey)

	// Проверка размера
	if n != constants.InitPacketTotalSize {
		t.Errorf("Init returned %d bytes, expected %d", n, constants.InitPacketTotalSize)
	}

	// Проверка opcode
	if buf[constants.InitPacketOpcodeOffset] != InitOpcode {
		t.Errorf("opcode = 0x%02X, expected 0x00", buf[constants.InitPacketOpcodeOffset])
	}

	// Проверка sessionID
	gotSessionID := int32(binary.LittleEndian.Uint32(buf[constants.InitPacketSessionIDOffset:]))
	if gotSessionID != sessionID {
		t.Errorf("sessionID = 0x%08X, expected 0x%08X", gotSessionID, sessionID)
	}

	// Проверка protocolRevision
	gotProtocolRev := binary.LittleEndian.Uint32(buf[constants.InitPacketProtocolRevOffset:])
	if gotProtocolRev != constants.ProtocolRevisionInit {
		t.Errorf("protocolRevision = 0x%08X, expected 0x%08X", gotProtocolRev, constants.ProtocolRevisionInit)
	}

	// Проверка scrambled modulus
	for i := range constants.RSA1024ModulusSize {
		if buf[constants.InitPacketModulusOffset+i] != scrambledModulus[i] {
			t.Errorf("scrambledModulus[%d] = 0x%02X, expected 0x%02X", i, buf[constants.InitPacketModulusOffset+i], scrambledModulus[i])
			break
		}
	}

	// Проверка GG constants
	ggConstants := []uint32{constants.GGConst1, constants.GGConst2, constants.GGConst3, constants.GGConst4}
	for i, expected := range ggConstants {
		offset := constants.InitPacketGGConstantsOffset + i*4
		got := binary.LittleEndian.Uint32(buf[offset:])
		if got != expected {
			t.Errorf("ggData[%d] at offset %d = 0x%08X, expected 0x%08X", i, offset, got, expected)
		}
	}

	// Проверка blowfish key
	for i := range constants.BlowfishKeySize {
		if buf[constants.InitPacketBlowfishKeyOffset+i] != blowfishKey[i] {
			t.Errorf("blowfishKey[%d] at offset %d = 0x%02X, expected 0x%02X", i, constants.InitPacketBlowfishKeyOffset+i, buf[constants.InitPacketBlowfishKeyOffset+i], blowfishKey[i])
		}
	}

	// Проверка null terminator
	if buf[constants.InitPacketNullTerminatorOffset] != 0x00 {
		t.Errorf("null terminator at offset %d = 0x%02X, expected 0x00", constants.InitPacketNullTerminatorOffset, buf[constants.InitPacketNullTerminatorOffset])
	}

	t.Logf("Init packet structure validated successfully")
	t.Logf("  opcode: 0x%02X", buf[0])
	t.Logf("  sessionID: 0x%08X", gotSessionID)
	t.Logf("  protocolRev: 0x%08X", gotProtocolRev)
	t.Logf("  scrambledModulus: %d bytes at offset %d", len(scrambledModulus), constants.InitPacketModulusOffset)
	t.Logf("  ggData: 4×4 bytes at offset %d", constants.InitPacketGGConstantsOffset)
	t.Logf("  blowfishKey: %d bytes at offset %d", constants.BlowfishKeySize, constants.InitPacketBlowfishKeyOffset)
	t.Logf("  null terminator: 1 byte at offset %d", constants.InitPacketNullTerminatorOffset)
	t.Logf("  total: %d bytes", n)
}
