package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestBuySeed is the C2S opcode 0xC4.
const OpcodeRequestBuySeed byte = 0xC4

// SeedPurchase is a single seed purchase entry.
type SeedPurchase struct {
	ItemID int32
	Count  int32
}

// RequestBuySeed represents a manor seed purchase request.
type RequestBuySeed struct {
	ManorID int32
	Items   []SeedPurchase
}

// ParseRequestBuySeed parses the packet from raw bytes.
func ParseRequestBuySeed(data []byte) (*RequestBuySeed, error) {
	r := packet.NewReader(data)

	manorID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ManorID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]SeedPurchase, 0, count)
	for range count {
		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading ItemID: %w", err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading Count: %w", err)
		}

		if cnt <= 0 {
			continue
		}

		items = append(items, SeedPurchase{
			ItemID: itemID,
			Count:  cnt,
		})
	}

	return &RequestBuySeed{
		ManorID: manorID,
		Items:   items,
	}, nil
}
