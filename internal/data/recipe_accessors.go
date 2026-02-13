package data

// RecipeTemplate — exported view of a recipe for use outside the data package.
// Phase 15: Recipe/Craft System.
type RecipeTemplate struct {
	ID          int32
	RecipeID    int32 // itemID книги рецепта
	Name        string
	CraftLevel  int32
	IsDwarven   bool
	SuccessRate int32 // 0-100
	MPCost      int32
	Ingredients []RecipeIngredient
	Productions []RecipeIngredient
}

// RecipeIngredient — exported view of a recipe ingredient/product.
type RecipeIngredient struct {
	ItemID int32
	Count  int32
}

// GetRecipeTemplate returns an exported recipe by listID.
// Returns nil if not found.
// Phase 15: Recipe/Craft System.
func GetRecipeTemplate(listID int32) *RecipeTemplate {
	def := RecipeTable[listID]
	if def == nil {
		return nil
	}
	return recipeDefToTemplate(def)
}

// GetRecipeByRecipeID finds a recipe by its recipeID (book item ID).
// Returns nil if not found.
// Phase 15: Recipe/Craft System.
func GetRecipeByRecipeID(recipeID int32) *RecipeTemplate {
	for _, def := range RecipeTable {
		if def.recipeID == recipeID {
			return recipeDefToTemplate(def)
		}
	}
	return nil
}

// GetAllRecipeTemplates returns all loaded recipes as exported structs.
func GetAllRecipeTemplates() []*RecipeTemplate {
	result := make([]*RecipeTemplate, 0, len(RecipeTable))
	for _, def := range RecipeTable {
		result = append(result, recipeDefToTemplate(def))
	}
	return result
}

func recipeDefToTemplate(def *recipeDef) *RecipeTemplate {
	ings := make([]RecipeIngredient, len(def.ingredients))
	for i, ing := range def.ingredients {
		ings[i] = RecipeIngredient{
			ItemID: ing.itemID,
			Count:  ing.count,
		}
	}

	prods := make([]RecipeIngredient, len(def.productions))
	for i, prod := range def.productions {
		prods[i] = RecipeIngredient{
			ItemID: prod.itemID,
			Count:  prod.count,
		}
	}

	return &RecipeTemplate{
		ID:          def.id,
		RecipeID:    def.recipeID,
		Name:        def.name,
		CraftLevel:  def.craftLevel,
		IsDwarven:   def.recipeType == "dwarven",
		SuccessRate: def.successRate,
		MPCost:      def.mpCost,
		Ingredients: ings,
		Productions: prods,
	}
}
