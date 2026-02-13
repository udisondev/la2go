package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Clan War opcodes (C2S).
const (
	// OpcodeRequestStartPledgeWar is 0x4D — declare clan war.
	OpcodeRequestStartPledgeWar byte = 0x4D
	// OpcodeRequestReplyStartPledgeWar is 0x4E — reply to war declaration.
	OpcodeRequestReplyStartPledgeWar byte = 0x4E
	// OpcodeRequestStopPledgeWar is 0x4F — request to stop war.
	OpcodeRequestStopPledgeWar byte = 0x4F
	// OpcodeRequestReplyStopPledgeWar is 0x50 — reply to stop war request.
	OpcodeRequestReplyStopPledgeWar byte = 0x50
	// OpcodeRequestSurrenderPledgeWar is 0x51 — surrender in war.
	OpcodeRequestSurrenderPledgeWar byte = 0x51
	// OpcodeRequestReplySurrenderPledgeWar is 0x52 — reply to war surrender.
	OpcodeRequestReplySurrenderPledgeWar byte = 0x52
)

// RequestStartPledgeWar — declare war on another clan.
// Packet: string (clan name).
type RequestStartPledgeWar struct {
	ClanName string
}

// ParseRequestStartPledgeWar parses the start war request.
func ParseRequestStartPledgeWar(data []byte) (*RequestStartPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading clanName: %w", err)
	}

	return &RequestStartPledgeWar{ClanName: name}, nil
}

// RequestReplyStartPledgeWar — reply to war declaration dialog.
// Packet: string (requester name), int32 (answer: 1=accept, 0=deny).
type RequestReplyStartPledgeWar struct {
	Name   string
	Answer int32
}

// ParseRequestReplyStartPledgeWar parses the reply packet.
func ParseRequestReplyStartPledgeWar(data []byte) (*RequestReplyStartPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading answer: %w", err)
	}

	return &RequestReplyStartPledgeWar{Name: name, Answer: answer}, nil
}

// RequestStopPledgeWar — request to stop war with a clan.
// Packet: string (clan name).
type RequestStopPledgeWar struct {
	ClanName string
}

// ParseRequestStopPledgeWar parses the stop war request.
func ParseRequestStopPledgeWar(data []byte) (*RequestStopPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading clanName: %w", err)
	}

	return &RequestStopPledgeWar{ClanName: name}, nil
}

// RequestReplyStopPledgeWar — reply to stop war dialog.
// Packet: string (requester name), int32 (answer: 1=accept, 0=deny).
type RequestReplyStopPledgeWar struct {
	Name   string
	Answer int32
}

// ParseRequestReplyStopPledgeWar parses the reply packet.
func ParseRequestReplyStopPledgeWar(data []byte) (*RequestReplyStopPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading answer: %w", err)
	}

	return &RequestReplyStopPledgeWar{Name: name, Answer: answer}, nil
}

// RequestSurrenderPledgeWar — surrender in war.
// Packet: string (clan name).
type RequestSurrenderPledgeWar struct {
	ClanName string
}

// ParseRequestSurrenderPledgeWar parses the surrender request.
func ParseRequestSurrenderPledgeWar(data []byte) (*RequestSurrenderPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading clanName: %w", err)
	}

	return &RequestSurrenderPledgeWar{ClanName: name}, nil
}

// RequestReplySurrenderPledgeWar — reply to surrender dialog.
// Packet: string (requester name), int32 (answer: 1=accept, 0=deny).
type RequestReplySurrenderPledgeWar struct {
	Name   string
	Answer int32
}

// ParseRequestReplySurrenderPledgeWar parses the reply packet.
func ParseRequestReplySurrenderPledgeWar(data []byte) (*RequestReplySurrenderPledgeWar, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading answer: %w", err)
	}

	return &RequestReplySurrenderPledgeWar{Name: name, Answer: answer}, nil
}
