package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// ExPledgeEmblem is the S2C extended packet 0xFE:0x28 for large clan crest data.
// Java reference: ExPledgeCrestLarge.java
//
// Packet structure (S2C 0xFE sub 0x28):
//   - opcode     byte    0xFE
//   - subOpcode  int16   0x28
//   - unk        int32   always 0
//   - crestID    int32   the large crest identifier
//   - length     int32   data length in bytes (0 = no crest)
//   - data       []byte  raw crest image data (BMP format)
type ExPledgeEmblem struct {
	CrestID int32
	Data    []byte // nil or empty = no crest
}

// Write serializes the ExPledgeEmblem packet.
func (p *ExPledgeEmblem) Write() ([]byte, error) {
	dataLen := len(p.Data)
	w := packet.NewWriter(1 + 2 + 4 + 4 + 4 + dataLen)
	w.WriteByte(0xFE)
	w.WriteShort(0x28)
	w.WriteInt(0) // unk
	w.WriteInt(p.CrestID)
	w.WriteInt(int32(dataLen))
	if dataLen > 0 {
		w.WriteBytes(p.Data)
	}
	return w.Bytes(), nil
}
