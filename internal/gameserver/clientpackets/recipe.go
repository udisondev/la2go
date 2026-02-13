package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Opcodes for recipe-related client packets.
// Java reference: ClientPackets.java
const (
	OpcodeRequestRecipeBookOpen     = 0xAC
	OpcodeRequestRecipeItemMakeInfo = 0xAE
	OpcodeRequestRecipeItemMakeSelf = 0xAF
)

// RequestRecipeBookOpen — client requests to open recipe book.
// Java reference: RequestRecipeBookOpen.java
//
// Format: [isDwarvenCraft:int32] (0 = dwarven, 1 = common)
type RequestRecipeBookOpen struct {
	IsDwarvenCraft bool
}

// ParseRequestRecipeBookOpen parses the RequestRecipeBookOpen packet.
func ParseRequestRecipeBookOpen(data []byte) (*RequestRecipeBookOpen, error) {
	r := packet.NewReader(data)
	isDwarven, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("read isDwarvenCraft: %w", err)
	}

	return &RequestRecipeBookOpen{
		IsDwarvenCraft: isDwarven == 0, // 0 = dwarven, 1 = common
	}, nil
}

// RequestRecipeItemMakeInfo — client requests recipe crafting info.
// Java reference: RequestRecipeItemMakeInfo.java
//
// Format: [recipeListId:int32]
type RequestRecipeItemMakeInfo struct {
	RecipeListID int32
}

// ParseRequestRecipeItemMakeInfo parses the RequestRecipeItemMakeInfo packet.
func ParseRequestRecipeItemMakeInfo(data []byte) (*RequestRecipeItemMakeInfo, error) {
	r := packet.NewReader(data)
	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("read recipeListId: %w", err)
	}

	return &RequestRecipeItemMakeInfo{
		RecipeListID: id,
	}, nil
}

// RequestRecipeItemMakeSelf — client requests to craft an item.
// Java reference: RequestRecipeItemMakeSelf.java
//
// Format: [recipeListId:int32]
type RequestRecipeItemMakeSelf struct {
	RecipeListID int32
}

// ParseRequestRecipeItemMakeSelf parses the RequestRecipeItemMakeSelf packet.
func ParseRequestRecipeItemMakeSelf(data []byte) (*RequestRecipeItemMakeSelf, error) {
	r := packet.NewReader(data)
	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("read recipeListId: %w", err)
	}

	return &RequestRecipeItemMakeSelf{
		RecipeListID: id,
	}, nil
}

// RequestRecipeBookDestroy — client requests to delete a recipe from book.
// Java reference: RequestRecipeBookDestroy.java (opcode 0xAD)
//
// Format: [recipeListId:int32]
type RequestRecipeBookDestroy struct {
	RecipeListID int32
}

// ParseRequestRecipeBookDestroy parses the RequestRecipeBookDestroy packet.
func ParseRequestRecipeBookDestroy(data []byte) (*RequestRecipeBookDestroy, error) {
	r := packet.NewReader(data)
	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("read recipeListId: %w", err)
	}

	return &RequestRecipeBookDestroy{
		RecipeListID: id,
	}, nil
}
