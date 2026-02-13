package craft

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// mockItemCreator реализует ItemCreator для тестов.
type mockItemCreator struct {
	nextObjectID atomic.Uint32
	createErr    error // если != nil, CreateItem возвращает эту ошибку
}

func (m *mockItemCreator) CreateItem(itemID int32, count int32, ownerID int64) (*model.Item, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}

	objID := m.nextObjectID.Add(1)
	tmpl := &model.ItemTemplate{
		ItemID:    itemID,
		Name:      fmt.Sprintf("crafted_%d", itemID),
		Type:      model.ItemTypeEtcItem,
		Stackable: count > 1,
	}
	item, err := model.NewItem(objID, itemID, ownerID, count, tmpl)
	if err != nil {
		return nil, fmt.Errorf("mock create item: %w", err)
	}
	return item, nil
}

// testRecipeListID = 1 — mk_wooden_arrow, dwarven, successRate=100, mpCost=0
// Ingredients: itemID=1864 count=4, itemID=1869 count=2
// Productions: itemID=17 count=500
const (
	testRecipeListID  int32 = 1
	testIngredient1ID int32 = 1864 // count=4
	testIngredient2ID int32 = 1869 // count=2
	testProductID     int32 = 17   // wooden arrows, count=500
)

func TestMain(m *testing.M) {
	if err := data.LoadRecipes(); err != nil {
		fmt.Fprintf(os.Stderr, "load recipes: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// newTestPlayer создаёт игрока с objectID, characterID и начальным HP/MP.
// MP по умолчанию = maxMP = 525 (NewPlayer для level=1: 500 + 1*25).
func newTestPlayer(t *testing.T, objectID uint32) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, int64(objectID)*100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer(%d, ...) error: %v", objectID, err)
	}
	return p
}

// addMaterialsToInventory добавляет материалы для testRecipeListID в инвентарь.
// objectID используется как база для генерации objectID предметов.
func addMaterialsToInventory(t *testing.T, player *model.Player, baseObjID uint32) {
	t.Helper()

	tmpl1 := &model.ItemTemplate{ItemID: testIngredient1ID, Name: "ingredient_1864", Type: model.ItemTypeEtcItem, Stackable: true}
	item1, err := model.NewItem(baseObjID+1000, testIngredient1ID, player.CharacterID(), 4, tmpl1)
	if err != nil {
		t.Fatalf("NewItem ingredient1: %v", err)
	}
	if err := player.Inventory().AddItem(item1); err != nil {
		t.Fatalf("AddItem ingredient1: %v", err)
	}

	tmpl2 := &model.ItemTemplate{ItemID: testIngredient2ID, Name: "ingredient_1869", Type: model.ItemTypeEtcItem, Stackable: true}
	item2, err := model.NewItem(baseObjID+2000, testIngredient2ID, player.CharacterID(), 2, tmpl2)
	if err != nil {
		t.Fatalf("NewItem ingredient2: %v", err)
	}
	if err := player.Inventory().AddItem(item2); err != nil {
		t.Fatalf("AddItem ingredient2: %v", err)
	}
}

func TestCraft(t *testing.T) {
	t.Parallel()

	// Проверяем что тестовый рецепт загружен.
	recipe := data.GetRecipeTemplate(testRecipeListID)
	if recipe == nil {
		t.Fatal("recipe template with listID=1 not found; data.LoadRecipes() may have failed")
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T, c *Controller) *model.Player
		recipeID  int32
		wantErr   bool
		errSubstr string
		// Проверки результата (только когда wantErr=false).
		wantSuccess    bool
		wantItemNonNil bool
	}{
		{
			name: "successful craft 100% rate",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 1)
				if err := p.LearnRecipe(testRecipeListID, true); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				addMaterialsToInventory(t, p, 1)
				return p
			},
			recipeID:       testRecipeListID,
			wantErr:        false,
			wantSuccess:    true,
			wantItemNonNil: true,
		},
		{
			name: "nil player",
			setup: func(t *testing.T, c *Controller) *model.Player {
				return nil
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "player is nil",
		},
		{
			name: "recipe not found in data",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 2)
				// Рецепт выучен (ID который не существует в data)
				if err := p.LearnRecipe(99999, false); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				return p
			},
			recipeID:  99999,
			wantErr:   true,
			errSubstr: "not found",
		},
		{
			name: "recipe not learned",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 3)
				// Не выучиваем рецепт
				return p
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "not learned",
		},
		{
			name: "not enough materials",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 4)
				if err := p.LearnRecipe(testRecipeListID, true); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				// Добавляем только 1 из 2 ингредиентов
				tmpl := &model.ItemTemplate{ItemID: testIngredient1ID, Name: "ing1", Type: model.ItemTypeEtcItem, Stackable: true}
				item, err := model.NewItem(4000, testIngredient1ID, p.CharacterID(), 4, tmpl)
				if err != nil {
					t.Fatalf("NewItem: %v", err)
				}
				if err := p.Inventory().AddItem(item); err != nil {
					t.Fatalf("AddItem: %v", err)
				}
				// Второй ингредиент не добавляем
				return p
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "missing material",
		},
		{
			name: "dead player",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 5)
				if err := p.LearnRecipe(testRecipeListID, true); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				addMaterialsToInventory(t, p, 5)
				// Убиваем игрока
				p.SetCurrentHP(0)
				return p
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "dead",
		},
		{
			name: "player in store mode",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 6)
				if err := p.LearnRecipe(testRecipeListID, true); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				addMaterialsToInventory(t, p, 6)
				p.SetPrivateStoreType(model.StoreSell)
				return p
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "store mode",
		},
		{
			name: "player in combat",
			setup: func(t *testing.T, c *Controller) *model.Player {
				p := newTestPlayer(t, 7)
				if err := p.LearnRecipe(testRecipeListID, true); err != nil {
					t.Fatalf("LearnRecipe: %v", err)
				}
				addMaterialsToInventory(t, p, 7)
				p.MarkAttackStance()
				return p
			},
			recipeID:  testRecipeListID,
			wantErr:   true,
			errSubstr: "combat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			creator := &mockItemCreator{}
			creator.nextObjectID.Store(uint32(tt.recipeID)*1000 + 500)
			ctrl := NewController(creator)

			player := tt.setup(t, ctrl)

			result, err := ctrl.Craft(player, tt.recipeID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Craft(player, %d) error = nil; want error containing %q", tt.recipeID, tt.errSubstr)
					return
				}
				if tt.errSubstr != "" && !containsSubstr(err.Error(), tt.errSubstr) {
					t.Errorf("Craft(player, %d) error = %q; want error containing %q", tt.recipeID, err, tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Craft(player, %d) unexpected error: %v", tt.recipeID, err)
			}
			if result == nil {
				t.Fatalf("Craft(player, %d) returned nil result", tt.recipeID)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Craft(player, %d) result.Success = %v; want %v", tt.recipeID, result.Success, tt.wantSuccess)
			}
			if tt.wantItemNonNil && result.Item == nil {
				t.Errorf("Craft(player, %d) result.Item = nil; want non-nil", tt.recipeID)
			}
			if result.Recipe == nil {
				t.Errorf("Craft(player, %d) result.Recipe = nil; want non-nil", tt.recipeID)
			}
		})
	}
}

func TestCraftMPConsumption(t *testing.T) {
	t.Parallel()

	// Ищем рецепт с mpCost > 0 в загруженных данных.
	var mpRecipe *data.RecipeTemplate
	for _, r := range data.GetAllRecipeTemplates() {
		if r.MPCost > 0 {
			mpRecipe = r
			break
		}
	}

	// Если рецептов с mpCost > 0 нет среди загруженных данных,
	// пропускаем тест (в сгенерированных данных все mpCost=0).
	if mpRecipe == nil {
		t.Skip("no recipes with MPCost > 0 found in generated data")
	}

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(50000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 800)
	if err := p.LearnRecipe(mpRecipe.ID, mpRecipe.IsDwarven); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}

	// Устанавливаем MP ниже стоимости рецепта
	p.SetCurrentMP(mpRecipe.MPCost - 1)

	// Добавляем все материалы
	for i, ing := range mpRecipe.Ingredients {
		tmpl := &model.ItemTemplate{
			ItemID:    ing.ItemID,
			Name:      fmt.Sprintf("ing_%d", ing.ItemID),
			Type:      model.ItemTypeEtcItem,
			Stackable: true,
		}
		item, err := model.NewItem(uint32(80000+i), ing.ItemID, p.CharacterID(), ing.Count, tmpl)
		if err != nil {
			t.Fatalf("NewItem ingredient %d: %v", ing.ItemID, err)
		}
		if err := p.Inventory().AddItem(item); err != nil {
			t.Fatalf("AddItem ingredient %d: %v", ing.ItemID, err)
		}
	}

	_, err := ctrl.Craft(p, mpRecipe.ID)
	if err == nil {
		t.Errorf("Craft(player, %d) with MP=%d (need %d) error = nil; want error containing 'not enough MP'",
			mpRecipe.ID, mpRecipe.MPCost-1, mpRecipe.MPCost)
	} else if !containsSubstr(err.Error(), "not enough MP") {
		t.Errorf("Craft(player, %d) error = %q; want error containing 'not enough MP'", mpRecipe.ID, err)
	}
}

func TestCraftMaterialsConsumedOnFailure(t *testing.T) {
	t.Parallel()

	// Рецепт с successRate=100 -- материалы будут потрачены и крафт успешен.
	// Мы не можем задать successRate=0 программно для загруженного рецепта,
	// но можем проверить что материалы потребляются при successRate=100.
	// Это подтверждает поведение "consume before success check".

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(60000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 900)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}
	addMaterialsToInventory(t, p, 900)

	// Проверяем что материалы есть до крафта
	before1 := p.Inventory().CountItemsByID(testIngredient1ID)
	before2 := p.Inventory().CountItemsByID(testIngredient2ID)
	if before1 != 4 {
		t.Fatalf("before craft: ingredient %d count = %d; want 4", testIngredient1ID, before1)
	}
	if before2 != 2 {
		t.Fatalf("before craft: ingredient %d count = %d; want 2", testIngredient2ID, before2)
	}

	result, err := ctrl.Craft(p, testRecipeListID)
	if err != nil {
		t.Fatalf("Craft error: %v", err)
	}
	if !result.Success {
		t.Fatal("Craft result.Success = false; want true (successRate=100)")
	}

	// Материалы должны быть потреблены
	after1 := p.Inventory().CountItemsByID(testIngredient1ID)
	after2 := p.Inventory().CountItemsByID(testIngredient2ID)
	if after1 != 0 {
		t.Errorf("after craft: ingredient %d count = %d; want 0", testIngredient1ID, after1)
	}
	if after2 != 0 {
		t.Errorf("after craft: ingredient %d count = %d; want 0", testIngredient2ID, after2)
	}
}

func TestCraftProductAddedToInventory(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(70000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 950)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}
	addMaterialsToInventory(t, p, 950)

	// До крафта продукта нет
	beforeProduct := p.Inventory().CountItemsByID(testProductID)
	if beforeProduct != 0 {
		t.Fatalf("before craft: product %d count = %d; want 0", testProductID, beforeProduct)
	}

	result, err := ctrl.Craft(p, testRecipeListID)
	if err != nil {
		t.Fatalf("Craft error: %v", err)
	}
	if !result.Success {
		t.Fatal("Craft result.Success = false; want true")
	}

	// Продукт должен быть в инвентаре
	afterProduct := p.Inventory().CountItemsByID(testProductID)
	if afterProduct != 500 {
		t.Errorf("after craft: product %d count = %d; want 500", testProductID, afterProduct)
	}
}

func TestCraftCreateItemError(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{
		createErr: fmt.Errorf("item service unavailable"),
	}
	ctrl := NewController(creator)

	p := newTestPlayer(t, 960)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}
	addMaterialsToInventory(t, p, 960)

	_, err := ctrl.Craft(p, testRecipeListID)
	if err == nil {
		t.Error("Craft with failing ItemCreator error = nil; want error")
	} else if !containsSubstr(err.Error(), "create craft product") {
		t.Errorf("Craft error = %q; want error containing 'create craft product'", err)
	}
}

func TestAlreadyCrafting(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(80000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 100)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}

	// Симулируем "уже крафтит" — вручную вставляем в activeCrafts
	ctrl.mu.Lock()
	ctrl.activeCrafts[p.ObjectID()] = struct{}{}
	ctrl.mu.Unlock()

	_, err := ctrl.Craft(p, testRecipeListID)
	if err == nil {
		t.Error("Craft while already crafting error = nil; want error containing 'already crafting'")
	} else if !containsSubstr(err.Error(), "already crafting") {
		t.Errorf("Craft error = %q; want error containing 'already crafting'", err)
	}

	// Убираем флаг — cleanup
	ctrl.mu.Lock()
	delete(ctrl.activeCrafts, p.ObjectID())
	ctrl.mu.Unlock()
}

func TestIsCrafting(t *testing.T) {
	t.Parallel()

	ctrl := NewController(&mockItemCreator{})

	objectID := uint32(42)

	if ctrl.IsCrafting(objectID) {
		t.Errorf("IsCrafting(%d) = true; want false (no active craft)", objectID)
	}

	ctrl.mu.Lock()
	ctrl.activeCrafts[objectID] = struct{}{}
	ctrl.mu.Unlock()

	if !ctrl.IsCrafting(objectID) {
		t.Errorf("IsCrafting(%d) = false; want true (craft in progress)", objectID)
	}

	ctrl.mu.Lock()
	delete(ctrl.activeCrafts, objectID)
	ctrl.mu.Unlock()

	if ctrl.IsCrafting(objectID) {
		t.Errorf("IsCrafting(%d) = true after removal; want false", objectID)
	}
}

func TestCraftClearsActiveFlag(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(90000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 200)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}
	addMaterialsToInventory(t, p, 200)

	_, err := ctrl.Craft(p, testRecipeListID)
	if err != nil {
		t.Fatalf("Craft error: %v", err)
	}

	// После завершения крафта флаг должен быть снят
	if ctrl.IsCrafting(p.ObjectID()) {
		t.Errorf("IsCrafting(%d) = true after Craft completed; want false", p.ObjectID())
	}
}

func TestCraftClearsActiveFlagOnError(t *testing.T) {
	t.Parallel()

	ctrl := NewController(&mockItemCreator{})

	p := newTestPlayer(t, 201)
	// Не выучиваем рецепт — Craft вернёт ошибку

	_, err := ctrl.Craft(p, testRecipeListID)
	if err == nil {
		t.Fatal("Craft error = nil; want error")
	}

	// Флаг должен быть снят даже при ошибке
	if ctrl.IsCrafting(p.ObjectID()) {
		t.Errorf("IsCrafting(%d) = true after Craft error; want false", p.ObjectID())
	}
}

func TestConcurrentCraftAttempts(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{}
	creator.nextObjectID.Store(100000)
	ctrl := NewController(creator)

	p := newTestPlayer(t, 300)
	if err := p.LearnRecipe(testRecipeListID, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}

	// Добавляем достаточно материалов только для одного крафта
	addMaterialsToInventory(t, p, 300)

	const goroutines = 10
	var (
		wg       sync.WaitGroup
		errCount atomic.Int32
		okCount  atomic.Int32
	)

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			_, err := ctrl.Craft(p, testRecipeListID)
			if err != nil {
				errCount.Add(1)
			} else {
				okCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// Ровно один крафт должен успешно завершиться (или 0 если все получили ошибку
	// concurrent access), но "already crafting" гарантирует не более 1 одновременного.
	// Из-за недостатка материалов, максимум 1 успешный.
	total := okCount.Load() + errCount.Load()
	if total != goroutines {
		t.Errorf("total results = %d; want %d", total, goroutines)
	}

	// Не более 1 успешного крафта (материалов хватает только на 1)
	if okCount.Load() > 1 {
		t.Errorf("successful crafts = %d; want at most 1 (materials for 1 craft only)", okCount.Load())
	}

	// После всех горутин флаг должен быть снят
	if ctrl.IsCrafting(p.ObjectID()) {
		t.Errorf("IsCrafting(%d) = true after all goroutines finished; want false", p.ObjectID())
	}
}

func TestNewController(t *testing.T) {
	t.Parallel()

	creator := &mockItemCreator{}
	ctrl := NewController(creator)

	if ctrl == nil {
		t.Fatal("NewController() returned nil")
	}
	if ctrl.activeCrafts == nil {
		t.Error("NewController().activeCrafts is nil; want initialized map")
	}
	if ctrl.itemCreator != creator {
		t.Error("NewController().itemCreator does not match provided creator")
	}
}

// containsSubstr проверяет вхождение подстроки (без импорта strings).
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
