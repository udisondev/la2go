package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeSetSummonRemainTime is the opcode for SetSummonRemainTime (S2C 0xD1).
const OpcodeSetSummonRemainTime byte = 0xD1

// SetSummonRemainTime informs the client about servitor lifetime remaining.
// Phase 52: Pet/Summon system gaps.
// Java reference: SetSummonRemainTime.java
type SetSummonRemainTime struct {
	MaxTime       int32 // Total lifetime in seconds
	RemainingTime int32 // Remaining lifetime in seconds
}

// Write serializes SetSummonRemainTime packet.
func (p *SetSummonRemainTime) Write() ([]byte, error) {
	w := packet.NewWriter(9) // opcode(1) + 2*int32(8)
	w.WriteByte(OpcodeSetSummonRemainTime)
	w.WriteInt(p.MaxTime)
	w.WriteInt(p.RemainingTime)
	return w.Bytes(), nil
}
