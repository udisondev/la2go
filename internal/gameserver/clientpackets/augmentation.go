package clientpackets

import (
	"encoding/binary"
	"fmt"
)

// Sub-opcodes for augmentation C2S packets (0xD0:0x29-0x2E).
// Java reference: RequestConfirmTargetItem, RequestConfirmRefinerItem, etc.
const (
	SubOpcodeRequestConfirmTargetItem  int16 = 0x29
	SubOpcodeRequestConfirmRefinerItem int16 = 0x2A
	SubOpcodeRequestConfirmGemStone    int16 = 0x2B
	SubOpcodeRequestRefine             int16 = 0x2C
	SubOpcodeRequestConfirmCancelItem  int16 = 0x2D
	SubOpcodeRequestRefineCancel       int16 = 0x2E
)

// RequestConfirmTargetItem — player selects weapon to augment (C2S 0xD0:0x29).
type RequestConfirmTargetItem struct {
	ObjectID int32
}

// ParseRequestConfirmTargetItem parses RequestConfirmTargetItem from raw bytes.
func ParseRequestConfirmTargetItem(data []byte) (*RequestConfirmTargetItem, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("RequestConfirmTargetItem too short: %d bytes", len(data))
	}
	return &RequestConfirmTargetItem{
		ObjectID: int32(binary.LittleEndian.Uint32(data[:4])),
	}, nil
}

// RequestConfirmRefinerItem — player selects Life Stone (C2S 0xD0:0x2A).
type RequestConfirmRefinerItem struct {
	TargetObjectID  int32
	RefinerObjectID int32
}

// ParseRequestConfirmRefinerItem parses RequestConfirmRefinerItem from raw bytes.
func ParseRequestConfirmRefinerItem(data []byte) (*RequestConfirmRefinerItem, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("RequestConfirmRefinerItem too short: %d bytes", len(data))
	}
	return &RequestConfirmRefinerItem{
		TargetObjectID:  int32(binary.LittleEndian.Uint32(data[:4])),
		RefinerObjectID: int32(binary.LittleEndian.Uint32(data[4:8])),
	}, nil
}

// RequestConfirmGemStone — player confirms gemstones (C2S 0xD0:0x2B).
type RequestConfirmGemStone struct {
	TargetObjectID  int32
	RefinerObjectID int32
	GemStoneObjectID int32
	GemStoneCount   int64
}

// ParseRequestConfirmGemStone parses RequestConfirmGemStone from raw bytes.
func ParseRequestConfirmGemStone(data []byte) (*RequestConfirmGemStone, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("RequestConfirmGemStone too short: %d bytes", len(data))
	}
	return &RequestConfirmGemStone{
		TargetObjectID:   int32(binary.LittleEndian.Uint32(data[:4])),
		RefinerObjectID:  int32(binary.LittleEndian.Uint32(data[4:8])),
		GemStoneObjectID: int32(binary.LittleEndian.Uint32(data[8:12])),
		GemStoneCount:    int64(binary.LittleEndian.Uint64(data[12:20])),
	}, nil
}

// RequestRefine — player clicks "Augment" button (C2S 0xD0:0x2C).
type RequestRefine struct {
	TargetObjectID   int32
	RefinerObjectID  int32
	GemStoneObjectID int32
	GemStoneCount    int64
}

// ParseRequestRefine parses RequestRefine from raw bytes.
func ParseRequestRefine(data []byte) (*RequestRefine, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("RequestRefine too short: %d bytes", len(data))
	}
	return &RequestRefine{
		TargetObjectID:   int32(binary.LittleEndian.Uint32(data[:4])),
		RefinerObjectID:  int32(binary.LittleEndian.Uint32(data[4:8])),
		GemStoneObjectID: int32(binary.LittleEndian.Uint32(data[8:12])),
		GemStoneCount:    int64(binary.LittleEndian.Uint64(data[12:20])),
	}, nil
}

// RequestConfirmCancelItem — player selects augmented weapon to cancel (C2S 0xD0:0x2D).
type RequestConfirmCancelItem struct {
	ObjectID int32
}

// ParseRequestConfirmCancelItem parses RequestConfirmCancelItem from raw bytes.
func ParseRequestConfirmCancelItem(data []byte) (*RequestConfirmCancelItem, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("RequestConfirmCancelItem too short: %d bytes", len(data))
	}
	return &RequestConfirmCancelItem{
		ObjectID: int32(binary.LittleEndian.Uint32(data[:4])),
	}, nil
}

// RequestRefineCancel — player confirms augmentation removal (C2S 0xD0:0x2E).
type RequestRefineCancel struct {
	ObjectID int32
}

// ParseRequestRefineCancel parses RequestRefineCancel from raw bytes.
func ParseRequestRefineCancel(data []byte) (*RequestRefineCancel, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("RequestRefineCancel too short: %d bytes", len(data))
	}
	return &RequestRefineCancel{
		ObjectID: int32(binary.LittleEndian.Uint32(data[:4])),
	}, nil
}
