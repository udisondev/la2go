package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestProcureCropList is the 0xD0 sub-opcode 0x09.
const SubOpcodeRequestProcureCropList int16 = 0x09

// CropSaleEntry is a single crop sale entry.
type CropSaleEntry struct {
	ObjectID int32
	ItemID   int32
	ManorID  int32
	Count    int32
}

// RequestProcureCropList represents a crop sale to Manor.
type RequestProcureCropList struct {
	Items []CropSaleEntry
}

// ParseRequestProcureCropList parses the packet from raw bytes (after sub-opcode).
func ParseRequestProcureCropList(data []byte) (*RequestProcureCropList, error) {
	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid crop count: %d", count)
	}

	items := make([]CropSaleEntry, 0, count)
	for range count {
		objID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading ObjectID: %w", err)
		}

		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading ItemID: %w", err)
		}

		manorID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading ManorID: %w", err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading Count: %w", err)
		}

		if cnt <= 0 {
			continue
		}

		items = append(items, CropSaleEntry{
			ObjectID: objID,
			ItemID:   itemID,
			ManorID:  manorID,
			Count:    cnt,
		})
	}

	return &RequestProcureCropList{
		Items: items,
	}, nil
}
