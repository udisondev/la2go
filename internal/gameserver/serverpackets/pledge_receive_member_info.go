package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// PledgeReceiveMemberInfo sends detailed info about a single clan member.
// S2C opcode 0xFE, sub-opcode 0x3D.
//
// Java reference: PledgeReceiveMemberInfo.java
type PledgeReceiveMemberInfo struct {
	PledgeType int32  // sub-pledge type
	Name       string // member name
	Title      string // member title
	PowerGrade int32  // rank (1-9)
	Level      int32  // character level
	ClassID    int32  // character class
	Gender     int32  // 0=male, 1=female
	ObjectID   int32  // character objectID
	Online     bool   // online status
	SponsorID  int32  // sponsor player objectID (for academy members)
}

// Write serializes the PledgeReceiveMemberInfo packet.
func (p *PledgeReceiveMemberInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(0xFE)
	w.WriteShort(0x3D)
	w.WriteInt(p.PledgeType)
	w.WriteString(p.Name)
	w.WriteString(p.Title)
	w.WriteInt(p.PowerGrade)

	onlineVal := int32(0)
	if p.Online {
		onlineVal = 1
	}
	w.WriteInt(onlineVal)
	w.WriteInt(p.Level)
	w.WriteInt(p.ClassID)
	w.WriteInt(p.Gender)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.SponsorID)

	return w.Bytes(), nil
}
