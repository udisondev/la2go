package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeAllyCrest is the S2C opcode 0xAE for alliance crest data.
// Java reference: AllyCrest.java (ServerPackets.ALLIANCE_CREST = 0xAE).
const OpcodeAllyCrest byte = 0xAE

// AllyCrest sends an alliance crest image to the client.
//
// Packet structure (S2C 0xAE):
//   - opcode  byte    0xAE
//   - crestID int32   the crest identifier
//   - length  int32   data length in bytes (0 = no crest)
//   - data    []byte  raw crest image data (BMP format)
type AllyCrest struct {
	CrestID int32
	Data    []byte // nil or empty = no crest
}

// Write serializes the AllyCrest packet.
func (p *AllyCrest) Write() ([]byte, error) {
	dataLen := len(p.Data)
	w := packet.NewWriter(1 + 4 + 4 + dataLen)
	w.WriteByte(OpcodeAllyCrest)
	w.WriteInt(p.CrestID)
	w.WriteInt(int32(dataLen))
	if dataLen > 0 {
		w.WriteBytes(p.Data)
	}
	return w.Bytes(), nil
}
