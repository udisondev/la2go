package serverpackets

import (
	"encoding/binary"

	"github.com/udisondev/la2go/internal/constants"
)

const InitOpcode = 0x00

// Init writes the Init packet (opcode 0x00) into buf.
// Contains: sessionId, protocol version, scrambled RSA modulus, GG constants, blowfish key.
// Returns the number of bytes written.
func Init(buf []byte, sessionID int32, scrambledModulus, blowfishKey []byte) int {
	buf[constants.InitPacketOpcodeOffset] = InitOpcode
	binary.LittleEndian.PutUint32(buf[constants.InitPacketSessionIDOffset:], uint32(sessionID))
	binary.LittleEndian.PutUint32(buf[constants.InitPacketProtocolRevOffset:], constants.ProtocolRevisionInit)

	copy(buf[constants.InitPacketModulusOffset:], scrambledModulus) // 128 bytes

	// GG constants (4×int32 = 16 bytes)
	binary.LittleEndian.PutUint32(buf[constants.InitPacketGGConstantsOffset:], constants.GGConst1)
	binary.LittleEndian.PutUint32(buf[constants.InitPacketGGConstantsOffset+4:], constants.GGConst2)
	binary.LittleEndian.PutUint32(buf[constants.InitPacketGGConstantsOffset+8:], constants.GGConst3)
	binary.LittleEndian.PutUint32(buf[constants.InitPacketGGConstantsOffset+12:], constants.GGConst4)

	copy(buf[constants.InitPacketBlowfishKeyOffset:], blowfishKey) // 16 bytes
	buf[constants.InitPacketNullTerminatorOffset] = 0x00           // null terminator

	return constants.InitPacketTotalSize
}
