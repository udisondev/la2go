package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const OpcodeKeyPacket = 0x00

// KeyPacket is the first packet sent to the client after TCP connection.
// Contains the Blowfish key for encrypting subsequent packets.
//
// Structure (Interlude):
// - byte: opcode (0x00) — ServerPackets.KEY_PACKET
// - byte: result (0x01 = protocol ok, 0x00 = wrong protocol)
// - byte[8]: Blowfish key (Java: 8 bytes; Go: 16 bytes for extended key)
//
// Java reference: KeyPacket.java — also writes encryption flag, serverID, obfuscation key
type KeyPacket struct {
	blowfishKey []byte // 16 bytes
}

// NewKeyPacket creates a KeyPacket with the given Blowfish key.
func NewKeyPacket(blowfishKey []byte) KeyPacket {
	return KeyPacket{
		blowfishKey: blowfishKey,
	}
}

// Write serializes the KeyPacket.
func (p *KeyPacket) Write() ([]byte, error) {
	w := packet.NewWriter(32) // 18 bytes + padding

	if err := w.WriteByte(OpcodeKeyPacket); err != nil {
		return nil, err
	}

	// Protocol version (0x01 for Interlude)
	if err := w.WriteByte(0x01); err != nil {
		return nil, err
	}

	// Blowfish key (16 bytes)
	w.WriteBytes(p.blowfishKey)

	return w.Bytes(), nil
}
