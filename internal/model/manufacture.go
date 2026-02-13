package model

// ManufactureItem represents a recipe offered in a manufacture (crafting) shop.
//
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: ManufactureItem.java
type ManufactureItem struct {
	RecipeID  int32 // recipeListID from RecipeTable
	Cost      int64 // price the crafter charges for this recipe
	IsDwarven bool  // true = dwarven recipe, false = common
}
