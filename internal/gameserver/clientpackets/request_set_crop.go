package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestSetCrop is the sub-opcode for RequestSetCrop (C2S 0xD0:0x0B).
const SubOpcodeRequestSetCrop int16 = 0x0B

// CropEntry represents a single crop configuration in a RequestSetCrop packet.
type CropEntry struct {
	CropID      int32
	StartAmount int32
	Price       int32
	RewardType  byte
}

// RequestSetCrop is a client packet to set crop procurement for the next manor period.
//
// Packet structure (after 0xD0 sub-opcode stripped):
//   - manorID (int32) -- castle ID
//   - count (int32) -- number of crop entries
//   - per entry:
//   - cropID (int32)
//   - startAmount (int32)
//   - price (int32)
//   - rewardType (byte) -- 1 or 2
type RequestSetCrop struct {
	ManorID int32
	Crops   []CropEntry
}

// ParseRequestSetCrop parses RequestSetCrop from raw bytes.
// Sub-opcode already stripped by the extended opcode dispatcher.
func ParseRequestSetCrop(rawData []byte) (*RequestSetCrop, error) {
	r := packet.NewReader(rawData)

	manorID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading manorID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	if count < 0 || count > 500 {
		return nil, fmt.Errorf("invalid crop count %d", count)
	}

	crops := make([]CropEntry, 0, count)
	for range count {
		cropID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading cropID: %w", err)
		}

		amount, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading amount: %w", err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading price: %w", err)
		}

		rewardType, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("reading rewardType: %w", err)
		}

		crops = append(crops, CropEntry{
			CropID:      cropID,
			StartAmount: amount,
			Price:       price,
			RewardType:  rewardType,
		})
	}

	return &RequestSetCrop{
		ManorID: manorID,
		Crops:   crops,
	}, nil
}
