package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Private Store client packet opcodes.
//
// Phase 8.1: Private Store System.
// Java reference: ClientPackets.java
const (
	OpcodeRequestPrivateStoreManageSell = 0x73
	OpcodeSetPrivateStoreListSell       = 0x74
	OpcodeRequestPrivateStoreQuitSell   = 0x76
	OpcodeSetPrivateStoreMsgSell        = 0x77
	OpcodeRequestPrivateStoreBuy        = 0x79
	OpcodeRequestPrivateStoreManageBuy  = 0x90
	OpcodeSetPrivateStoreListBuy        = 0x91
	OpcodeRequestPrivateStoreQuitBuy    = 0x93
	OpcodeSetPrivateStoreMsgBuy         = 0x94
	OpcodeRequestPrivateStoreSell       = 0x96
)

// --- SetPrivateStoreListSell (0x74) ---

// SellListEntry represents a single item in a sell list setup.
type SellListEntry struct {
	ObjectID int32 // ObjectID of item from inventory
	Count    int32 // Quantity to sell
	Price    int64 // Price per unit (Adena)
}

// SetPrivateStoreListSell represents the client packet for setting up a sell store.
//
// Packet structure (body after opcode):
//   - packageSale (int32): 1 = package sell, 0 = normal
//   - count (int32): number of items
//   - for each item: objectID (int32), count (int32), price (int32)
//
// BATCH_LENGTH = 12 bytes per item (3x int32)
type SetPrivateStoreListSell struct {
	PackageSale bool
	Items       []SellListEntry
}

// ParseSetPrivateStoreListSell parses SetPrivateStoreListSell packet.
func ParseSetPrivateStoreListSell(data []byte) (*SetPrivateStoreListSell, error) {
	r := packet.NewReader(data)

	pkgFlag, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading packageSale flag: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]SellListEntry, count)
	for i := range count {
		objID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] objectID: %w", i, err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] price: %w", i, err)
		}

		if cnt <= 0 {
			return nil, fmt.Errorf("invalid count for item[%d]: %d", i, cnt)
		}
		if price < 0 {
			return nil, fmt.Errorf("invalid price for item[%d]: %d", i, price)
		}

		items[i] = SellListEntry{
			ObjectID: objID,
			Count:    cnt,
			Price:    int64(price),
		}
	}

	return &SetPrivateStoreListSell{
		PackageSale: pkgFlag == 1,
		Items:       items,
	}, nil
}

// --- SetPrivateStoreListBuy (0x91) ---

// BuyListEntry represents a single item in a buy list setup.
type BuyListEntry struct {
	ItemID int32 // Template ID of item to buy
	Count  int32 // Quantity to buy
	Price  int64 // Price per unit (Adena)
}

// SetPrivateStoreListBuy represents the client packet for setting up a buy store.
//
// Packet structure (body after opcode):
//   - count (int32): number of items
//   - for each item: itemID (int32), enchant (int16), reserved (int16), count (int32), price (int32)
//
// BATCH_LENGTH = 16 bytes per item (int32 + 2x int16 + int32 + int32)
type SetPrivateStoreListBuy struct {
	Items []BuyListEntry
}

// ParseSetPrivateStoreListBuy parses SetPrivateStoreListBuy packet.
func ParseSetPrivateStoreListBuy(data []byte) (*SetPrivateStoreListBuy, error) {
	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]BuyListEntry, count)
	for i := range count {
		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] itemID: %w", i, err)
		}

		// Java: readShort() x2 — enchant level + reserved (4 bytes total)
		if _, err := r.ReadShort(); err != nil {
			return nil, fmt.Errorf("reading item[%d] enchant: %w", i, err)
		}
		if _, err := r.ReadShort(); err != nil {
			return nil, fmt.Errorf("reading item[%d] reserved: %w", i, err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] price: %w", i, err)
		}

		if cnt <= 0 {
			return nil, fmt.Errorf("invalid count for item[%d]: %d", i, cnt)
		}
		if price < 0 {
			return nil, fmt.Errorf("invalid price for item[%d]: %d", i, price)
		}

		items[i] = BuyListEntry{
			ItemID: itemID,
			Count:  cnt,
			Price:  int64(price),
		}
	}

	return &SetPrivateStoreListBuy{
		Items: items,
	}, nil
}

// --- RequestPrivateStoreBuy (0x79) ---

// PrivateStoreBuyEntry represents a single item purchase from a sell store.
type PrivateStoreBuyEntry struct {
	ObjectID int32 // ObjectID of item in seller's store
	Count    int32 // Quantity to buy
	Price    int64 // Expected price per unit
}

// RequestPrivateStoreBuy represents the client packet for buying from a sell store.
//
// Packet structure (body after opcode):
//   - storePlayerID (int32): ObjectID of the seller
//   - count (int32): number of items
//   - for each item: objectID (int32), count (int32), price (int32)
type RequestPrivateStoreBuy struct {
	StorePlayerID int32
	Items         []PrivateStoreBuyEntry
}

// ParseRequestPrivateStoreBuy parses RequestPrivateStoreBuy packet.
func ParseRequestPrivateStoreBuy(data []byte) (*RequestPrivateStoreBuy, error) {
	r := packet.NewReader(data)

	storePlayerID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading storePlayerID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]PrivateStoreBuyEntry, count)
	for i := range count {
		objID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] objectID: %w", i, err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] price: %w", i, err)
		}

		items[i] = PrivateStoreBuyEntry{
			ObjectID: objID,
			Count:    cnt,
			Price:    int64(price),
		}
	}

	return &RequestPrivateStoreBuy{
		StorePlayerID: storePlayerID,
		Items:         items,
	}, nil
}

// --- RequestPrivateStoreSell (0x96) ---

// PrivateStoreSellEntry represents a single item sale to a buy store.
type PrivateStoreSellEntry struct {
	ObjectID int32 // ObjectID of item from seller's inventory
	ItemID   int32 // Template ID
	Count    int32 // Quantity to sell
	Price    int64 // Expected price per unit
}

// RequestPrivateStoreSell represents the client packet for selling to a buy store.
//
// Packet structure (body after opcode):
//   - storePlayerID (int32): ObjectID of the buyer (store owner)
//   - count (int32): number of items
//   - for each item: objectID (int32), itemID (int32), reserved (2x int16), count (int32), price (int32)
//
// BATCH_LENGTH = 20 bytes per item
type RequestPrivateStoreSell struct {
	StorePlayerID int32
	Items         []PrivateStoreSellEntry
}

// ParseRequestPrivateStoreSell parses RequestPrivateStoreSell packet.
func ParseRequestPrivateStoreSell(data []byte) (*RequestPrivateStoreSell, error) {
	r := packet.NewReader(data)

	storePlayerID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading storePlayerID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]PrivateStoreSellEntry, count)
	for i := range count {
		objID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] objectID: %w", i, err)
		}

		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] itemID: %w", i, err)
		}

		// Skip 2 reserved shorts
		if _, err := r.ReadShort(); err != nil {
			return nil, fmt.Errorf("reading item[%d] reserved1: %w", i, err)
		}
		if _, err := r.ReadShort(); err != nil {
			return nil, fmt.Errorf("reading item[%d] reserved2: %w", i, err)
		}

		cnt, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		price, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] price: %w", i, err)
		}

		items[i] = PrivateStoreSellEntry{
			ObjectID: objID,
			ItemID:   itemID,
			Count:    cnt,
			Price:    int64(price),
		}
	}

	return &RequestPrivateStoreSell{
		StorePlayerID: storePlayerID,
		Items:         items,
	}, nil
}

// --- SetPrivateStoreMsgSell (0x77) ---

// SetPrivateStoreMsgSell represents the client packet for setting sell store message.
//
// Packet structure (body after opcode):
//   - message (string): UTF-16LE null-terminated store title
type SetPrivateStoreMsgSell struct {
	Message string
}

// ParseSetPrivateStoreMsgSell parses SetPrivateStoreMsgSell packet.
func ParseSetPrivateStoreMsgSell(data []byte) (*SetPrivateStoreMsgSell, error) {
	r := packet.NewReader(data)

	msg, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading store message: %w", err)
	}

	return &SetPrivateStoreMsgSell{Message: msg}, nil
}

// --- SetPrivateStoreMsgBuy (0x94) ---

// SetPrivateStoreMsgBuy represents the client packet for setting buy store message.
//
// Packet structure (body after opcode):
//   - message (string): UTF-16LE null-terminated store title
type SetPrivateStoreMsgBuy struct {
	Message string
}

// ParseSetPrivateStoreMsgBuy parses SetPrivateStoreMsgBuy packet.
func ParseSetPrivateStoreMsgBuy(data []byte) (*SetPrivateStoreMsgBuy, error) {
	r := packet.NewReader(data)

	msg, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading store message: %w", err)
	}

	return &SetPrivateStoreMsgBuy{Message: msg}, nil
}
