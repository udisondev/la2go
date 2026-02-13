package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestExEnchantSkillInfo is the 0xD0 sub-opcode 0x06.
const SubOpcodeRequestExEnchantSkillInfo int16 = 0x06

// SubOpcodeRequestExEnchantSkill is the 0xD0 sub-opcode 0x07.
const SubOpcodeRequestExEnchantSkill int16 = 0x07

// RequestExEnchantSkillInfo requests info about skill enchant cost/chance.
type RequestExEnchantSkillInfo struct {
	SkillID    int32
	SkillLevel int32
}

// ParseRequestExEnchantSkillInfo parses the packet from raw bytes (after sub-opcode).
func ParseRequestExEnchantSkillInfo(data []byte) (*RequestExEnchantSkillInfo, error) {
	r := packet.NewReader(data)

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillID: %w", err)
	}

	skillLevel, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillLevel: %w", err)
	}

	return &RequestExEnchantSkillInfo{
		SkillID:    skillID,
		SkillLevel: skillLevel,
	}, nil
}

// RequestExEnchantSkill requests to enchant a skill.
type RequestExEnchantSkill struct {
	SkillID    int32
	SkillLevel int32
}

// ParseRequestExEnchantSkill parses the packet from raw bytes (after sub-opcode).
func ParseRequestExEnchantSkill(data []byte) (*RequestExEnchantSkill, error) {
	r := packet.NewReader(data)

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillID: %w", err)
	}

	skillLevel, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillLevel: %w", err)
	}

	return &RequestExEnchantSkill{
		SkillID:    skillID,
		SkillLevel: skillLevel,
	}, nil
}
