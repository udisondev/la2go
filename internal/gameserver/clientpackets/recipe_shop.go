package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Recipe Shop opcodes.
//
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: ClientPackets.java
const (
	// OpcodeRequestRecipeShopMessageSet is 0xB1 — set crafting shop message.
	OpcodeRequestRecipeShopMessageSet byte = 0xB1
	// OpcodeRequestRecipeShopListSet is 0xB2 — open crafting shop.
	OpcodeRequestRecipeShopListSet byte = 0xB2
	// OpcodeRequestRecipeShopManageQuit is 0xB3 — close crafting shop.
	OpcodeRequestRecipeShopManageQuit byte = 0xB3
	// OpcodeRequestRecipeShopMakeInfo is 0xB5 — recipe shop item info.
	OpcodeRequestRecipeShopMakeInfo byte = 0xB5
	// OpcodeRequestRecipeShopMakeItem is 0xB6 — craft item from recipe shop.
	OpcodeRequestRecipeShopMakeItem byte = 0xB6
	// OpcodeRequestRecipeShopManagePrev is 0xB7 — return to recipe management.
	OpcodeRequestRecipeShopManagePrev byte = 0xB7
)

// maxRecipeShopMsgLength is the max allowed manufacture shop message length (L2 Interlude).
const maxRecipeShopMsgLength = 29

// --- RequestRecipeShopMessageSet (0xB1) ---

// RequestRecipeShopMessageSet represents the client packet for setting manufacture shop message.
//
// Packet structure (body after opcode):
//   - message (string): UTF-16LE null-terminated shop title
//
// Java reference: RequestRecipeShopMessageSet.java
type RequestRecipeShopMessageSet struct {
	Message string
}

// ParseRequestRecipeShopMessageSet parses RequestRecipeShopMessageSet packet.
func ParseRequestRecipeShopMessageSet(data []byte) (*RequestRecipeShopMessageSet, error) {
	r := packet.NewReader(data)

	msg, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading shop message: %w", err)
	}

	if len(msg) > maxRecipeShopMsgLength {
		msg = msg[:maxRecipeShopMsgLength]
	}

	return &RequestRecipeShopMessageSet{Message: msg}, nil
}

// --- RequestRecipeShopListSet (0xB2) ---

// RecipeShopEntry represents a single recipe in the manufacture shop setup.
type RecipeShopEntry struct {
	RecipeID int32 // recipeListID
	Cost     int64 // price in Adena
}

// RequestRecipeShopListSet represents the client packet for setting up manufacture shop.
//
// Packet structure (body after opcode):
//   - count (int32): number of recipes
//   - for each recipe (8 bytes): recipeID (int32), cost (int32)
//
// Java reference: RequestRecipeShopListSet.java
type RequestRecipeShopListSet struct {
	Items []RecipeShopEntry
}

// ParseRequestRecipeShopListSet parses RequestRecipeShopListSet packet.
func ParseRequestRecipeShopListSet(data []byte) (*RequestRecipeShopListSet, error) {
	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading recipe count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid recipe count: %d", count)
	}

	// Empty list = close shop
	if count == 0 {
		return &RequestRecipeShopListSet{Items: nil}, nil
	}

	items := make([]RecipeShopEntry, count)
	for i := range count {
		recipeID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading recipe[%d] recipeID: %w", i, err)
		}

		cost, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading recipe[%d] cost: %w", i, err)
		}

		if cost < 0 {
			return nil, fmt.Errorf("invalid cost for recipe[%d]: %d", i, cost)
		}

		items[i] = RecipeShopEntry{
			RecipeID: recipeID,
			Cost:     int64(cost),
		}
	}

	return &RequestRecipeShopListSet{Items: items}, nil
}

// --- RequestRecipeShopMakeInfo (0xB5) ---

// RequestRecipeShopMakeInfo represents the client packet requesting recipe info from a shop.
//
// Packet structure (body after opcode):
//   - shopObjectID (int32): ObjectID of the manufacturer
//   - recipeID (int32): recipeListID to query
//
// Java reference: RequestRecipeShopMakeInfo.java
type RequestRecipeShopMakeInfo struct {
	ShopObjectID int32
	RecipeID     int32
}

// ParseRequestRecipeShopMakeInfo parses RequestRecipeShopMakeInfo packet.
func ParseRequestRecipeShopMakeInfo(data []byte) (*RequestRecipeShopMakeInfo, error) {
	r := packet.NewReader(data)

	shopObjID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading shopObjectID: %w", err)
	}

	recipeID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading recipeID: %w", err)
	}

	return &RequestRecipeShopMakeInfo{
		ShopObjectID: shopObjID,
		RecipeID:     recipeID,
	}, nil
}

// --- RequestRecipeShopMakeItem (0xB6) ---

// RequestRecipeShopMakeItem represents the client packet for ordering craft from a shop.
//
// Packet structure (body after opcode):
//   - manufacturerID (int32): ObjectID of the crafter
//   - recipeID (int32): recipeListID to craft
//   - unknown (int32): not used
//
// Java reference: RequestRecipeShopMakeItem.java
type RequestRecipeShopMakeItem struct {
	ManufacturerID int32
	RecipeID       int32
}

// ParseRequestRecipeShopMakeItem parses RequestRecipeShopMakeItem packet.
func ParseRequestRecipeShopMakeItem(data []byte) (*RequestRecipeShopMakeItem, error) {
	r := packet.NewReader(data)

	manuID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading manufacturerID: %w", err)
	}

	recipeID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading recipeID: %w", err)
	}

	// Skip unknown int32 (always ignored in Java)
	if _, err := r.ReadInt(); err != nil {
		// Not fatal — some clients might not send the trailing field
	}

	return &RequestRecipeShopMakeItem{
		ManufacturerID: manuID,
		RecipeID:       recipeID,
	}, nil
}
