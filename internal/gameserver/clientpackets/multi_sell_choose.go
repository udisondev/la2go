package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMultiSellChoose is the client packet opcode for multisell exchange.
// Client sends this when player confirms a multisell exchange.
//
// Phase 8: Trade System Foundation.
// Java reference: MultiSellChoose.java
const OpcodeMultiSellChoose = 0xA7

// MultiSellChoose represents the client's multisell exchange request.
//
// Packet structure:
//   - listId (int32): multisell list ID
//   - entryId (int32): entry ID within the list (1-indexed)
//   - amount (int32): number of exchanges
//
// Phase 8: Trade System Foundation.
type MultiSellChoose struct {
	ListID  int32
	EntryID int32
	Amount  int32
}

// ParseMultiSellChoose parses a MultiSellChoose packet from raw bytes.
func ParseMultiSellChoose(data []byte) (*MultiSellChoose, error) {
	r := packet.NewReader(data)

	listID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading listId: %w", err)
	}

	entryID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading entryId: %w", err)
	}

	amount, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading amount: %w", err)
	}

	if amount <= 0 || amount > 5000 {
		return nil, fmt.Errorf("invalid amount: %d (must be 1..5000)", amount)
	}

	return &MultiSellChoose{
		ListID:  listID,
		EntryID: entryID,
		Amount:  amount,
	}, nil
}
