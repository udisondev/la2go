package clientpackets

// OpcodeRequestSkillList is the opcode for RequestSkillList (C2S 0x3F).
// Java reference: ClientPackets.REQUEST_SKILL_LIST(0x3F).
const OpcodeRequestSkillList = 0x3F

// RequestSkillList has no body fields.
type RequestSkillList struct{}

// ParseRequestSkillList parses RequestSkillList packet (no fields).
func ParseRequestSkillList(_ []byte) (*RequestSkillList, error) {
	return &RequestSkillList{}, nil
}
