package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeSkillCoolTime is the S2C opcode 0xC1.
// Sent at login to sync skill cooldown timers.
const OpcodeSkillCoolTime byte = 0xC1

// SkillCoolDown represents a single skill on cooldown.
type SkillCoolDown struct {
	SkillID       int32
	SkillLevel    int32
	ReuseTime     int32 // total reuse time in seconds
	RemainingTime int32 // remaining time in seconds (min 1)
}

// SkillCoolTime sends the list of skills currently on cooldown.
// Java reference: SkillCoolTime.java
type SkillCoolTime struct {
	CoolDowns []SkillCoolDown
}

// Write serializes the packet.
func (p *SkillCoolTime) Write() ([]byte, error) {
	w := packet.NewWriter(5 + len(p.CoolDowns)*16)
	w.WriteByte(OpcodeSkillCoolTime)
	w.WriteInt(int32(len(p.CoolDowns)))

	for _, cd := range p.CoolDowns {
		w.WriteInt(cd.SkillID)
		w.WriteInt(cd.SkillLevel)
		w.WriteInt(cd.ReuseTime)
		w.WriteInt(cd.RemainingTime)
	}

	return w.Bytes(), nil
}
