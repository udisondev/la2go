package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeShowMemberListDelete removes a member from the list (S2C 0x56).
//
// Java reference: PledgeShowMemberListDelete.java (opcode 0x56).
const OpcodePledgeShowMemberListDelete = 0x56

// PledgeShowMemberListDelete removes a single clan member from the client list.
type PledgeShowMemberListDelete struct {
	Name string
}

// Write serializes the packet.
func (p *PledgeShowMemberListDelete) Write() ([]byte, error) {
	w := packet.NewWriter(32)

	w.WriteByte(OpcodePledgeShowMemberListDelete)
	w.WriteString(p.Name)

	return w.Bytes(), nil
}
