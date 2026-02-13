package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodePledgeCrest is the S2C opcode 0x6C for clan crest data.
// Java reference: PledgeCrest.java (ServerPackets.PLEDGE_CREST = 0x6C).
const OpcodePledgeCrest byte = 0x6C

// PledgeCrest sends a clan crest image to the client.
//
// Packet structure (S2C 0x6C):
//   - opcode  byte    0x6C
//   - crestID int32   the crest identifier
//   - length  int32   data length in bytes (0 = no crest)
//   - data    []byte  raw crest image data (BMP format)
type PledgeCrest struct {
	CrestID int32
	Data    []byte // nil or empty = no crest
}

// Write serializes the PledgeCrest packet.
func (p *PledgeCrest) Write() ([]byte, error) {
	dataLen := len(p.Data)
	w := packet.NewWriter(1 + 4 + 4 + dataLen)
	w.WriteByte(OpcodePledgeCrest)
	w.WriteInt(p.CrestID)
	w.WriteInt(int32(dataLen))
	if dataLen > 0 {
		w.WriteBytes(p.Data)
	}
	return w.Bytes(), nil
}
