package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeReceiveWarList sends the list of clan wars (S2C extended 0xFE:0x3E).
//
// Java reference: PledgeReceiveWarList.java.
const OpcodePledgeReceiveWarList = 0xFE

// SubOpcodePledgeReceiveWarList is the extended sub-opcode.
const SubOpcodePledgeReceiveWarList int16 = 0x3E

// PledgeWarEntry represents a single war entry.
type PledgeWarEntry struct {
	ClanName string
}

// PledgeReceiveWarList sends the war list to the client.
//
// Packet structure (S2C 0xCD):
//   - opcode    byte
//   - tab       int32  (0=wars we declared, 1=wars against us)
//   - count     int32
//   - [entries] repeated: clanName string
type PledgeReceiveWarList struct {
	Tab     int32
	Entries []PledgeWarEntry
}

// Write serializes the packet.
func (p *PledgeReceiveWarList) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePledgeReceiveWarList)
	w.WriteShort(SubOpcodePledgeReceiveWarList)
	w.WriteInt(p.Tab)
	w.WriteInt(int32(len(p.Entries)))

	for i := range p.Entries {
		w.WriteString(p.Entries[i].ClanName)
	}

	return w.Bytes(), nil
}
