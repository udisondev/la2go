package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (recipes) ---

type xmlRecipeList struct {
	XMLName xml.Name    `xml:"list"`
	Items   []xmlRecipe `xml:"item"`
}

type xmlRecipe struct {
	ID          int32              `xml:"id,attr"`
	RecipeID    int32              `xml:"recipeId,attr"`
	Name        string             `xml:"name,attr"`
	CraftLevel  int32              `xml:"craftLevel,attr"`
	Type        string             `xml:"type,attr"`
	SuccessRate int32              `xml:"successRate,attr"`
	MPCost      int32              `xml:"mpCost,attr"`
	Ingredients []xmlRecipeElement `xml:"ingredient"`
	Productions []xmlRecipeElement `xml:"production"`
}

type xmlRecipeElement struct {
	ID    int32 `xml:"id,attr"`
	Count int32 `xml:"count,attr"`
}

// --- Parsed structures (recipes) ---

type parsedRecipe struct {
	id          int32
	recipeID    int32
	name        string
	craftLevel  int32
	recipeType  string // "dwarven","common"
	successRate int32
	mpCost      int32
	ingredients []parsedRecipeIng
	productions []parsedRecipeIng
}

type parsedRecipeIng struct {
	itemID int32
	count  int32
}

func generateRecipes(javaDir, outDir string) error {
	recipesFile := filepath.Join(javaDir, "Recipes.xml")
	recipes, err := parseRecipeFile(recipesFile)
	if err != nil {
		return fmt.Errorf("parse recipes: %w", err)
	}

	sort.Slice(recipes, func(i, j int) bool { return recipes[i].id < recipes[j].id })

	outPath := filepath.Join(outDir, "recipe_data_generated.go")
	if err := generateRecipesGoFile(recipes, outPath); err != nil {
		return fmt.Errorf("generate recipes: %w", err)
	}

	fmt.Printf("  Generated %s: %d recipes\n", outPath, len(recipes))
	return nil
}

func parseRecipeFile(path string) ([]parsedRecipe, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlRecipeList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedRecipe, 0, len(list.Items))
	for _, xr := range list.Items {
		result = append(result, convertRecipe(xr))
	}
	return result, nil
}

func convertRecipe(xr xmlRecipe) parsedRecipe {
	ingredients := make([]parsedRecipeIng, 0, len(xr.Ingredients))
	for _, ing := range xr.Ingredients {
		ingredients = append(ingredients, parsedRecipeIng{
			itemID: ing.ID,
			count:  ing.Count,
		})
	}

	productions := make([]parsedRecipeIng, 0, len(xr.Productions))
	for _, prod := range xr.Productions {
		productions = append(productions, parsedRecipeIng{
			itemID: prod.ID,
			count:  prod.Count,
		})
	}

	return parsedRecipe{
		id:          xr.ID,
		recipeID:    xr.RecipeID,
		name:        xr.Name,
		craftLevel:  xr.CraftLevel,
		recipeType:  xr.Type,
		successRate: xr.SuccessRate,
		mpCost:      xr.MPCost,
		ingredients: ingredients,
		productions: productions,
	}
}

// --- Code generation (recipes) ---

func generateRecipesGoFile(recipes []parsedRecipe, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "recipes")
	buf.WriteString("var recipeDefs = []recipeDef{\n")

	for i := range recipes {
		writeRecipeDef(&buf, &recipes[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeRecipeDef(buf *bytes.Buffer, r *parsedRecipe) {
	fmt.Fprintf(buf, "{id: %d, recipeID: %d, name: %q, craftLevel: %d, recipeType: %q, successRate: %d, mpCost: %d",
		r.id, r.recipeID, r.name, r.craftLevel, r.recipeType, r.successRate, r.mpCost)

	buf.WriteString(", ingredients: []recipeIngDef{")
	for i, ing := range r.ingredients {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{itemID: %d, count: %d}", ing.itemID, ing.count)
	}

	buf.WriteString("}, productions: []recipeIngDef{")
	for i, prod := range r.productions {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{itemID: %d, count: %d}", prod.itemID, prod.count)
	}

	buf.WriteString("}},\n")
}
