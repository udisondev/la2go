package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// ExSendManorList sends the list of all manor castle names.
// S2C opcode 0xFE, sub-opcode 0x1B.
type ExSendManorList struct {
	CastleNames []string // castle names indexed by manor ID
}

// Write serializes ExSendManorList packet to binary format.
func (p *ExSendManorList) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(0xFE)
	w.WriteShort(0x1B)
	w.WriteInt(int32(len(p.CastleNames)))
	for i, name := range p.CastleNames {
		w.WriteInt(int32(i + 1)) // castleId (1-based)
		w.WriteString(name)
	}
	return w.Bytes(), nil
}
