package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeGmList is the S2C opcode 0x96 for GM list response.
// Java reference: GmList.java (ServerPackets.GM_LIST = 0x96).
const OpcodeGmList byte = 0x96

// GmList sends a list of visible (non-hidden) GM names to the client.
//
// Packet structure (S2C 0x96):
//   - opcode byte   0x96
//   - count  int32  number of visible GMs
//   - for each GM:
//   - name   string GM character name
type GmList struct {
	Names []string
}

// Write serializes the GmList packet.
func (p *GmList) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 4 + len(p.Names)*64)
	w.WriteByte(OpcodeGmList)
	w.WriteInt(int32(len(p.Names)))
	for _, name := range p.Names {
		w.WriteString(name)
	}
	return w.Bytes(), nil
}
