package data

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := LoadNpcTemplates(); err != nil {
		panic("load npc templates: " + err.Error())
	}
	if err := LoadRecipes(); err != nil {
		panic("load recipes: " + err.Error())
	}
	if err := LoadHennaTemplates(); err != nil {
		panic("load henna templates: " + err.Error())
	}
	if err := LoadPetData(); err != nil {
		panic("load pet data: " + err.Error())
	}
	if err := LoadSeeds(); err != nil {
		panic("load seeds: " + err.Error())
	}
	if err := LoadFishingData(); err != nil {
		panic("load fishing data: " + err.Error())
	}
	if err := LoadFishingRods(); err != nil {
		panic("load fishing rods: " + err.Error())
	}
	if err := LoadFishingMonsters(); err != nil {
		panic("load fishing monsters: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestGetRecipeTemplate_Valid(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	// Берём первый доступный ID из таблицы
	var wantID int32
	for id := range RecipeTable {
		wantID = id
		break
	}

	got := GetRecipeTemplate(wantID)
	if got == nil {
		t.Fatalf("GetRecipeTemplate(%d) = nil, want non-nil", wantID)
	}
	if got.ID != wantID {
		t.Errorf("GetRecipeTemplate(%d).ID = %d, want %d", wantID, got.ID, wantID)
	}
}

func TestGetRecipeTemplate_Invalid(t *testing.T) {
	t.Parallel()

	got := GetRecipeTemplate(-99999)
	if got != nil {
		t.Errorf("GetRecipeTemplate(-99999) = %+v, want nil", got)
	}
}

func TestGetRecipeByRecipeID_Valid(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	// Берём recipeID из первого определения
	var wantRecipeID int32
	var wantListID int32
	for _, def := range RecipeTable {
		wantRecipeID = def.recipeID
		wantListID = def.id
		break
	}

	got := GetRecipeByRecipeID(wantRecipeID)
	if got == nil {
		t.Fatalf("GetRecipeByRecipeID(%d) = nil, want non-nil", wantRecipeID)
	}
	if got.RecipeID != wantRecipeID {
		t.Errorf("GetRecipeByRecipeID(%d).RecipeID = %d, want %d", wantRecipeID, got.RecipeID, wantRecipeID)
	}
	if got.ID != wantListID {
		t.Errorf("GetRecipeByRecipeID(%d).ID = %d, want %d", wantRecipeID, got.ID, wantListID)
	}
}

func TestGetRecipeByRecipeID_Invalid(t *testing.T) {
	t.Parallel()

	got := GetRecipeByRecipeID(-99999)
	if got != nil {
		t.Errorf("GetRecipeByRecipeID(-99999) = %+v, want nil", got)
	}
}

func TestGetAllRecipeTemplates(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	all := GetAllRecipeTemplates()
	if len(all) == 0 {
		t.Fatal("GetAllRecipeTemplates() returned empty slice")
	}
	if len(all) != len(RecipeTable) {
		t.Errorf("GetAllRecipeTemplates() len = %d, want %d", len(all), len(RecipeTable))
	}
}

func TestRecipeTemplateFields(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	// Находим рецепт с ингредиентами (craft рецепт, не материал)
	var found *RecipeTemplate
	for _, def := range RecipeTable {
		tmpl := recipeDefToTemplate(def)
		if len(tmpl.Ingredients) > 0 && len(tmpl.Productions) > 0 {
			found = tmpl
			break
		}
	}

	if found == nil {
		t.Fatal("no recipe with ingredients and productions found")
	}

	if found.ID == 0 {
		t.Error("recipe ID = 0, want non-zero")
	}
	if found.Name == "" {
		t.Errorf("recipe %d: Name is empty", found.ID)
	}
	if found.SuccessRate < 0 || found.SuccessRate > 100 {
		t.Errorf("recipe %d: SuccessRate = %d, want 0-100", found.ID, found.SuccessRate)
	}
	if len(found.Ingredients) == 0 {
		t.Errorf("recipe %d: Ingredients is empty", found.ID)
	}
	if len(found.Productions) == 0 {
		t.Errorf("recipe %d: Productions is empty", found.ID)
	}

	// Проверяем что ингредиенты имеют валидные поля
	for i, ing := range found.Ingredients {
		if ing.ItemID <= 0 {
			t.Errorf("recipe %d: Ingredients[%d].ItemID = %d, want > 0", found.ID, i, ing.ItemID)
		}
		if ing.Count <= 0 {
			t.Errorf("recipe %d: Ingredients[%d].Count = %d, want > 0", found.ID, i, ing.Count)
		}
	}

	for i, prod := range found.Productions {
		if prod.ItemID <= 0 {
			t.Errorf("recipe %d: Productions[%d].ItemID = %d, want > 0", found.ID, i, prod.ItemID)
		}
		if prod.Count <= 0 {
			t.Errorf("recipe %d: Productions[%d].Count = %d, want > 0", found.ID, i, prod.Count)
		}
	}
}

func TestRecipeTemplate_IsDwarven(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	var hasDwarven, hasCommon bool
	for _, def := range RecipeTable {
		tmpl := recipeDefToTemplate(def)
		if tmpl.IsDwarven {
			hasDwarven = true
		} else {
			hasCommon = true
		}
		if hasDwarven && hasCommon {
			break
		}
	}

	if !hasDwarven {
		t.Error("no dwarven recipes found in loaded data")
	}
	if !hasCommon {
		t.Error("no common recipes found in loaded data")
	}
}

func TestGetRecipeTemplate_ConsistentWithTable(t *testing.T) {
	t.Parallel()

	if len(RecipeTable) == 0 {
		t.Skip("no recipe data loaded")
	}

	// Каждый элемент RecipeTable должен быть доступен через GetRecipeTemplate
	for id, def := range RecipeTable {
		tmpl := GetRecipeTemplate(id)
		if tmpl == nil {
			t.Errorf("GetRecipeTemplate(%d) = nil, but RecipeTable has entry", id)
			continue
		}
		if tmpl.Name != def.name {
			t.Errorf("GetRecipeTemplate(%d).Name = %q, want %q", id, tmpl.Name, def.name)
		}
		if tmpl.RecipeID != def.recipeID {
			t.Errorf("GetRecipeTemplate(%d).RecipeID = %d, want %d", id, tmpl.RecipeID, def.recipeID)
		}
	}
}
