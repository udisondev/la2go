package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestSetSeed is the sub-opcode for RequestSetSeed (C2S 0xD0:0x0A).
const SubOpcodeRequestSetSeed int16 = 0x0A

// SeedEntry represents a single seed configuration in a RequestSetSeed packet.
type SeedEntry struct {
	SeedID      int32
	StartAmount int32
	Price       int32
}

// RequestSetSeed is a client packet to set seed production for the next manor period.
//
// Packet structure (after 0xD0 sub-opcode stripped):
//   - manorID (int32) -- castle ID
//   - count (int32) -- number of seed entries
//   - per entry:
//   - seedID (int32)
//   - startAmount (int32)
//   - price (int32)
type RequestSetSeed struct {
	ManorID int32
	Seeds   []SeedEntry
}

// ParseRequestSetSeed parses RequestSetSeed from raw bytes.
// Sub-opcode already stripped by the extended opcode dispatcher.
func ParseRequestSetSeed(rawData []byte) (*RequestSetSeed, error) {
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
		return nil, fmt.Errorf("invalid seed count %d", count)
	}

	seeds := make([]SeedEntry, 0, count)
	for range count {
		seedID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading seedID: %w", err)
		}

		amount, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading amount: %w", err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading price: %w", err)
		}

		seeds = append(seeds, SeedEntry{
			SeedID:      seedID,
			StartAmount: amount,
			Price:       price,
		})
	}

	return &RequestSetSeed{
		ManorID: manorID,
		Seeds:   seeds,
	}, nil
}
