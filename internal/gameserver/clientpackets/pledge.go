package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Clan-related client packet opcodes (C2S).
const (
	OpcodeRequestJoinPledge              = 0x24 // Invite player to clan
	OpcodeRequestAnswerJoinPledge        = 0x25 // Accept/deny clan invite
	OpcodeRequestWithdrawalPledge        = 0x26 // Leave clan
	OpcodeRequestOustPledgeMember        = 0x27 // Kick from clan
	OpcodeRequestPledgeCrest      = 0x68 // Request clan crest image
	OpcodeRequestPledgeInfo       = 0x66 // Request clan info
	OpcodeRequestPledgeMemberList = 0x3C // Request member list
	OpcodeRequestPledgePower      = 0xC0 // Set rank privileges

	// Extended sub-opcodes (0xD0 prefix) for pledge packets.
	SubOpcodeRequestPledgeMemberInfo        int16 = 0x1D // Request specific member info
	SubOpcodeRequestPledgeSetMemberPowerGrade int16 = 0x1C // Set member rank
	SubOpcodeRequestPledgeWarList           int16 = 0x1E // Get war list
	SubOpcodeRequestPledgeReorganizeMember  int16 = 0x24 // Move member to sub-pledge
)

// RequestJoinPledge — invite a player to the clan.
type RequestJoinPledge struct {
	ObjectID   int32 // Target player objectID
	PledgeType int32 // Sub-pledge to join (0=main, -1=academy, etc.)
}

// ParseRequestJoinPledge parses the join pledge request packet.
func ParseRequestJoinPledge(data []byte) (*RequestJoinPledge, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	pledgeType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading pledgeType: %w", err)
	}

	return &RequestJoinPledge{
		ObjectID:   objectID,
		PledgeType: pledgeType,
	}, nil
}

// RequestAnswerJoinPledge — accept/deny clan invite.
type RequestAnswerJoinPledge struct {
	Answer int32 // 1 = accept, 0 = deny
}

// ParseRequestAnswerJoinPledge parses the answer packet.
func ParseRequestAnswerJoinPledge(data []byte) (*RequestAnswerJoinPledge, error) {
	r := packet.NewReader(data)

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading answer: %w", err)
	}

	return &RequestAnswerJoinPledge{Answer: answer}, nil
}

// RequestOustPledgeMember — kick a member from the clan.
type RequestOustPledgeMember struct {
	Name string // Character name to kick
}

// ParseRequestOustPledgeMember parses the oust packet.
func ParseRequestOustPledgeMember(data []byte) (*RequestOustPledgeMember, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	return &RequestOustPledgeMember{Name: name}, nil
}

// RequestPledgeInfo — request clan info by clan ID.
type RequestPledgeInfo struct {
	ClanID int32
}

// ParseRequestPledgeInfo parses the pledge info request.
func ParseRequestPledgeInfo(data []byte) (*RequestPledgeInfo, error) {
	r := packet.NewReader(data)

	clanID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading clanID: %w", err)
	}

	return &RequestPledgeInfo{ClanID: clanID}, nil
}

// RequestPledgeCrest — request clan crest by crest ID.
type RequestPledgeCrest struct {
	CrestID int32
}

// ParseRequestPledgeCrest parses the crest request.
func ParseRequestPledgeCrest(data []byte) (*RequestPledgeCrest, error) {
	r := packet.NewReader(data)

	crestID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading crestID: %w", err)
	}

	return &RequestPledgeCrest{CrestID: crestID}, nil
}

// RequestPledgeSetMemberPowerGrade — set a member's rank.
type RequestPledgeSetMemberPowerGrade struct {
	MemberName string
	PowerGrade int32
}

// ParseRequestPledgeSetMemberPowerGrade parses the power grade request.
func ParseRequestPledgeSetMemberPowerGrade(data []byte) (*RequestPledgeSetMemberPowerGrade, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading memberName: %w", err)
	}

	grade, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading powerGrade: %w", err)
	}

	return &RequestPledgeSetMemberPowerGrade{
		MemberName: name,
		PowerGrade: grade,
	}, nil
}

// RequestPledgeReorganizeMember — move member to another sub-pledge.
type RequestPledgeReorganizeMember struct {
	MemberName    string
	NewPledgeType int32
}

// ParseRequestPledgeReorganizeMember parses the reorganize request.
func ParseRequestPledgeReorganizeMember(data []byte) (*RequestPledgeReorganizeMember, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading memberName: %w", err)
	}

	pledgeType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading newPledgeType: %w", err)
	}

	return &RequestPledgeReorganizeMember{
		MemberName:    name,
		NewPledgeType: pledgeType,
	}, nil
}

// RequestPledgePower — set privileges for a rank.
type RequestPledgePower struct {
	PowerGrade int32
	Privileges int32 // Bitflag mask
}

// ParseRequestPledgePower parses the pledge power request.
func ParseRequestPledgePower(data []byte) (*RequestPledgePower, error) {
	r := packet.NewReader(data)

	grade, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading powerGrade: %w", err)
	}

	privs, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading privileges: %w", err)
	}

	return &RequestPledgePower{
		PowerGrade: grade,
		Privileges: privs,
	}, nil
}

// RequestPledgeMemberInfo — request detailed info about a specific clan member.
// Extended packet 0xD0:0x1D.
type RequestPledgeMemberInfo struct {
	MemberName string
}

// ParseRequestPledgeMemberInfo parses the pledge member info request.
func ParseRequestPledgeMemberInfo(data []byte) (*RequestPledgeMemberInfo, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading memberName: %w", err)
	}

	return &RequestPledgeMemberInfo{MemberName: name}, nil
}

// RequestPledgeWarList — request list of clan wars.
type RequestPledgeWarList struct {
	Page int32 // 0 = wars we declared, 1 = wars declared on us
	Tab  int32
}

// ParseRequestPledgeWarList parses the war list request.
func ParseRequestPledgeWarList(data []byte) (*RequestPledgeWarList, error) {
	r := packet.NewReader(data)

	page, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading page: %w", err)
	}

	tab, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading tab: %w", err)
	}

	return &RequestPledgeWarList{
		Page: page,
		Tab:  tab,
	}, nil
}
