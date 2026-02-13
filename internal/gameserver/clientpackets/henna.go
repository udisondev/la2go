package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Henna system opcodes (Phase 13).
const (
	// OpcodeRequestHennaItemList is the client packet for requesting available hennas (C2S 0xBA).
	// Java: RequestHennaItemList.java
	OpcodeRequestHennaItemList = 0xBA

	// OpcodeRequestHennaItemInfo is the client packet for henna draw info (C2S 0xBB).
	// Java: RequestHennaItemInfo.java
	OpcodeRequestHennaItemInfo = 0xBB

	// OpcodeRequestHennaEquip is the client packet for equipping a henna (C2S 0xBC).
	// Java: RequestHennaEquip.java
	OpcodeRequestHennaEquip = 0xBC

	// OpcodeRequestHennaRemoveList is the client packet for requesting removable hennas (C2S 0xBD).
	// Java: RequestHennaRemoveList.java
	OpcodeRequestHennaRemoveList = 0xBD

	// OpcodeRequestHennaItemRemoveInfo is the client packet for henna remove info (C2S 0xBE).
	// Java: RequestHennaItemRemoveInfo.java
	OpcodeRequestHennaItemRemoveInfo = 0xBE

	// OpcodeRequestHennaRemove is the client packet for removing a henna (C2S 0xBF).
	// Java: RequestHennaRemove.java
	OpcodeRequestHennaRemove = 0xBF
)

// RequestHennaItemList — request list of available hennas for equip.
type RequestHennaItemList struct {
	Unknown int32 // unknown field (always read, never used in Java)
}

// ParseRequestHennaItemList parses the packet.
func ParseRequestHennaItemList(data []byte) (*RequestHennaItemList, error) {
	r := packet.NewReader(data)
	val, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading unknown field: %w", err)
	}
	return &RequestHennaItemList{Unknown: val}, nil
}

// RequestHennaItemInfo — request detailed info about a henna before equipping.
type RequestHennaItemInfo struct {
	SymbolID int32
}

// ParseRequestHennaItemInfo parses the packet.
func ParseRequestHennaItemInfo(data []byte) (*RequestHennaItemInfo, error) {
	r := packet.NewReader(data)
	symbolID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading symbolID: %w", err)
	}
	return &RequestHennaItemInfo{SymbolID: symbolID}, nil
}

// RequestHennaEquip — equip a henna.
type RequestHennaEquip struct {
	SymbolID int32
}

// ParseRequestHennaEquip parses the packet.
func ParseRequestHennaEquip(data []byte) (*RequestHennaEquip, error) {
	r := packet.NewReader(data)
	symbolID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading symbolID: %w", err)
	}
	return &RequestHennaEquip{SymbolID: symbolID}, nil
}

// RequestHennaRemoveList — request list of equipped hennas for removal.
type RequestHennaRemoveList struct {
	Unknown int32
}

// ParseRequestHennaRemoveList parses the packet.
func ParseRequestHennaRemoveList(data []byte) (*RequestHennaRemoveList, error) {
	r := packet.NewReader(data)
	val, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading unknown field: %w", err)
	}
	return &RequestHennaRemoveList{Unknown: val}, nil
}

// RequestHennaItemRemoveInfo — request detailed info about a henna before removing.
type RequestHennaItemRemoveInfo struct {
	SymbolID int32
}

// ParseRequestHennaItemRemoveInfo parses the packet.
func ParseRequestHennaItemRemoveInfo(data []byte) (*RequestHennaItemRemoveInfo, error) {
	r := packet.NewReader(data)
	symbolID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading symbolID: %w", err)
	}
	return &RequestHennaItemRemoveInfo{SymbolID: symbolID}, nil
}

// RequestHennaRemove — remove a henna by symbol ID.
type RequestHennaRemove struct {
	SymbolID int32
}

// ParseRequestHennaRemove parses the packet.
func ParseRequestHennaRemove(data []byte) (*RequestHennaRemove, error) {
	r := packet.NewReader(data)
	symbolID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading symbolID: %w", err)
	}
	return &RequestHennaRemove{SymbolID: symbolID}, nil
}
