package serverpackets

import (
	"encoding/binary"

	"github.com/udisondev/la2go/internal/constants"
)

const (
	opcodeInitLS = 0x00
)

// InitLS [0x00] — LS → GS initial packet
// Отправляется сразу после подключения GameServer
//
// Format:
//   [opcodeInitLS]                    // opcode
//   [revision int32 LE]               // protocol revision
//   [keySize int32 LE]                // RSA-512 modulus size
//   [rsaModulus constants.RSA512ModulusSize bytes] // raw modulus (NO scrambling for GS)
//
// Returns: number of bytes written to buf
func InitLS(buf []byte, revision int32, rsaModulus []byte) int {
	pos := 0

	// Opcode
	buf[pos] = opcodeInitLS
	pos++

	// Revision (LE)
	binary.LittleEndian.PutUint32(buf[pos:], uint32(revision))
	pos += 4

	// Key size (always constants.RSA512ModulusSize for RSA-512)
	binary.LittleEndian.PutUint32(buf[pos:], constants.RSA512ModulusSize)
	pos += 4

	// RSA modulus (pad if needed)
	modulus := make([]byte, constants.RSA512ModulusSize)
	if len(rsaModulus) >= constants.RSA512ModulusSize {
		copy(modulus, rsaModulus[:constants.RSA512ModulusSize])
	} else {
		// Left-pad with zeros
		copy(modulus[constants.RSA512ModulusSize-len(rsaModulus):], rsaModulus)
	}
	copy(buf[pos:], modulus)
	pos += constants.RSA512ModulusSize

	return pos
}
