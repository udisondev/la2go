package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAcquireSkillInfo is the C2S opcode 0x6B.
const OpcodeRequestAcquireSkillInfo byte = 0x6B

// OpcodeRequestAcquireSkill is the C2S opcode 0x6C.
const OpcodeRequestAcquireSkill byte = 0x6C

// AcquireSkillType represents the type of skill being learned.
type AcquireSkillType int32

const (
	AcquireSkillTypeClass   AcquireSkillType = 0
	AcquireSkillTypeFishing AcquireSkillType = 1
	AcquireSkillTypePledge  AcquireSkillType = 2
)

// RequestAcquireSkillInfo represents a request for skill learning details.
type RequestAcquireSkillInfo struct {
	SkillID   int32
	Level     int32
	SkillType AcquireSkillType
}

// ParseRequestAcquireSkillInfo parses the packet from raw bytes.
func ParseRequestAcquireSkillInfo(data []byte) (*RequestAcquireSkillInfo, error) {
	r := packet.NewReader(data)

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillID: %w", err)
	}

	level, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Level: %w", err)
	}

	skillType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillType: %w", err)
	}

	return &RequestAcquireSkillInfo{
		SkillID:   skillID,
		Level:     level,
		SkillType: AcquireSkillType(skillType),
	}, nil
}

// RequestAcquireSkill represents a request to learn a skill.
type RequestAcquireSkill struct {
	SkillID   int32
	Level     int32
	SkillType AcquireSkillType
}

// ParseRequestAcquireSkill parses the packet from raw bytes.
func ParseRequestAcquireSkill(data []byte) (*RequestAcquireSkill, error) {
	r := packet.NewReader(data)

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillID: %w", err)
	}

	level, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Level: %w", err)
	}

	skillType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillType: %w", err)
	}

	return &RequestAcquireSkill{
		SkillID:   skillID,
		Level:     level,
		SkillType: AcquireSkillType(skillType),
	}, nil
}
