package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodePartySpelled is the S2C opcode 0xEE.
// Sends buff/debuff info for party member display.
const OpcodePartySpelled byte = 0xEE

// PartyEffect represents a single active effect on a party member.
type PartyEffect struct {
	SkillID    int32
	SkillLevel int16
	Duration   int32 // remaining seconds
}

// PartySpelled sends the effect list for a party member.
// Java reference: PartySpelled.java
type PartySpelled struct {
	Type     int32 // 0=player, 1=pet, 2=servitor
	ObjectID int32
	Effects  []PartyEffect
}

// Write serializes the packet.
func (p *PartySpelled) Write() ([]byte, error) {
	// 1 + 4 + 4 + 4 + effects*(4+2+4)
	w := packet.NewWriter(13 + len(p.Effects)*10)
	w.WriteByte(OpcodePartySpelled)
	w.WriteInt(p.Type)
	w.WriteInt(p.ObjectID)
	w.WriteInt(int32(len(p.Effects)))

	for _, e := range p.Effects {
		w.WriteInt(e.SkillID)
		w.WriteShort(e.SkillLevel)
		w.WriteInt(e.Duration)
	}

	return w.Bytes(), nil
}
