package model

import (
	"sync"
	"testing"
)

// newRecipeTestPlayer -- хелпер для создания тестового игрока для recipe тестов.
func newRecipeTestPlayer(t *testing.T) *Player {
	t.Helper()
	p, err := NewPlayer(1, 1, 1, "RecipeTester", 20, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	return p
}

func TestLearnRecipe_Dwarven(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if err := p.LearnRecipe(100, true); err != nil {
		t.Fatalf("LearnRecipe(100, dwarven) error: %v", err)
	}

	if !p.HasRecipe(100) {
		t.Error("HasRecipe(100) = false after LearnRecipe(100, dwarven)")
	}

	book := p.GetRecipeBook(true)
	if len(book) != 1 {
		t.Fatalf("GetRecipeBook(dwarven) len = %d, want 1", len(book))
	}
	if book[0] != 100 {
		t.Errorf("GetRecipeBook(dwarven)[0] = %d, want 100", book[0])
	}

	// Не должен быть в common книге
	commonBook := p.GetRecipeBook(false)
	if len(commonBook) != 0 {
		t.Errorf("GetRecipeBook(common) len = %d, want 0 (recipe was dwarven)", len(commonBook))
	}
}

func TestLearnRecipe_Common(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if err := p.LearnRecipe(200, false); err != nil {
		t.Fatalf("LearnRecipe(200, common) error: %v", err)
	}

	if !p.HasRecipe(200) {
		t.Error("HasRecipe(200) = false after LearnRecipe(200, common)")
	}

	book := p.GetRecipeBook(false)
	if len(book) != 1 {
		t.Fatalf("GetRecipeBook(common) len = %d, want 1", len(book))
	}
	if book[0] != 200 {
		t.Errorf("GetRecipeBook(common)[0] = %d, want 200", book[0])
	}

	// Не должен быть в dwarven книге
	dwarvenBook := p.GetRecipeBook(true)
	if len(dwarvenBook) != 0 {
		t.Errorf("GetRecipeBook(dwarven) len = %d, want 0 (recipe was common)", len(dwarvenBook))
	}
}

func TestLearnRecipe_Duplicate(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if err := p.LearnRecipe(100, true); err != nil {
		t.Fatalf("first LearnRecipe(100, dwarven): %v", err)
	}

	err := p.LearnRecipe(100, true)
	if err == nil {
		t.Fatal("second LearnRecipe(100, dwarven) = nil, want error for duplicate")
	}

	if p.RecipeCount() != 1 {
		t.Errorf("RecipeCount() = %d, want 1 after duplicate learn", p.RecipeCount())
	}
}

func TestForgetRecipe(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if err := p.LearnRecipe(100, true); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}
	if err := p.LearnRecipe(200, false); err != nil {
		t.Fatalf("LearnRecipe: %v", err)
	}

	if p.RecipeCount() != 2 {
		t.Fatalf("RecipeCount() = %d, want 2", p.RecipeCount())
	}

	if err := p.ForgetRecipe(100, true); err != nil {
		t.Fatalf("ForgetRecipe(100, dwarven): %v", err)
	}

	if p.HasRecipe(100) {
		t.Error("HasRecipe(100) = true after ForgetRecipe")
	}
	if p.RecipeCount() != 1 {
		t.Errorf("RecipeCount() = %d, want 1 after forget", p.RecipeCount())
	}

	// common рецепт всё ещё должен быть
	if !p.HasRecipe(200) {
		t.Error("HasRecipe(200) = false, should not be affected by forgetting dwarven recipe")
	}
}

func TestForgetRecipe_NotLearned(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	err := p.ForgetRecipe(999, true)
	if err == nil {
		t.Error("ForgetRecipe(999, dwarven) = nil, want error for non-existent recipe")
	}

	err = p.ForgetRecipe(999, false)
	if err == nil {
		t.Error("ForgetRecipe(999, common) = nil, want error for non-existent recipe")
	}
}

func TestGetRecipeBook_FreshPlayer(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	// Для свежего игрока книги nil (map не инициализирован)
	dwarven := p.GetRecipeBook(true)
	if dwarven != nil {
		t.Errorf("GetRecipeBook(dwarven) = %v, want nil for fresh player", dwarven)
	}

	common := p.GetRecipeBook(false)
	if common != nil {
		t.Errorf("GetRecipeBook(common) = %v, want nil for fresh player", common)
	}
}

func TestRecipeCount_BothTypes(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if p.RecipeCount() != 0 {
		t.Errorf("RecipeCount() = %d, want 0 for fresh player", p.RecipeCount())
	}

	recipes := []struct {
		id       int32
		dwarven  bool
	}{
		{100, true},
		{101, true},
		{200, false},
		{201, false},
		{202, false},
	}

	for _, r := range recipes {
		if err := p.LearnRecipe(r.id, r.dwarven); err != nil {
			t.Fatalf("LearnRecipe(%d, dwarven=%v): %v", r.id, r.dwarven, err)
		}
	}

	if got := p.RecipeCount(); got != 5 {
		t.Errorf("RecipeCount() = %d, want 5 (2 dwarven + 3 common)", got)
	}
}

func TestHasRecipe_ChecksBothBooks(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	if err := p.LearnRecipe(100, true); err != nil {
		t.Fatalf("LearnRecipe(100, dwarven): %v", err)
	}
	if err := p.LearnRecipe(200, false); err != nil {
		t.Fatalf("LearnRecipe(200, common): %v", err)
	}

	// HasRecipe ищет в обеих книгах
	if !p.HasRecipe(100) {
		t.Error("HasRecipe(100) = false, want true (dwarven)")
	}
	if !p.HasRecipe(200) {
		t.Error("HasRecipe(200) = false, want true (common)")
	}
	if p.HasRecipe(999) {
		t.Error("HasRecipe(999) = true, want false")
	}
}

func TestLearnForgetRecipe_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		learnIDs []int32
		dwarven  bool
		forgetID int32
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "learn one dwarven, forget it",
			learnIDs: []int32{10},
			dwarven:  true,
			forgetID: 10,
			wantErr:  false,
			wantLen:  0,
		},
		{
			name:     "learn multiple common, forget one",
			learnIDs: []int32{20, 21, 22},
			dwarven:  false,
			forgetID: 21,
			wantErr:  false,
			wantLen:  2,
		},
		{
			name:     "forget non-existent dwarven",
			learnIDs: []int32{30},
			dwarven:  true,
			forgetID: 999,
			wantErr:  true,
			wantLen:  1,
		},
		{
			name:     "forget non-existent common",
			learnIDs: []int32{40},
			dwarven:  false,
			forgetID: 888,
			wantErr:  true,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := newRecipeTestPlayer(t)

			for _, id := range tt.learnIDs {
				if err := p.LearnRecipe(id, tt.dwarven); err != nil {
					t.Fatalf("LearnRecipe(%d, dwarven=%v): %v", id, tt.dwarven, err)
				}
			}

			err := p.ForgetRecipe(tt.forgetID, tt.dwarven)
			if (err != nil) != tt.wantErr {
				t.Errorf("ForgetRecipe(%d) error = %v, wantErr = %v", tt.forgetID, err, tt.wantErr)
			}

			book := p.GetRecipeBook(tt.dwarven)
			gotLen := len(book)
			if gotLen != tt.wantLen {
				t.Errorf("GetRecipeBook(dwarven=%v) len = %d, want %d", tt.dwarven, gotLen, tt.wantLen)
			}
		})
	}
}

func TestRecipe_ConcurrentLearnForget(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // learn + forget горутины

	// Параллельное добавление уникальных рецептов
	for i := range goroutines {
		go func(id int32) {
			defer wg.Done()
			// Игнорируем ошибку — могут быть дубликаты при параллельном выполнении
			_ = p.LearnRecipe(id, true)
		}(int32(i))
	}

	// Параллельное добавление common рецептов
	for i := range goroutines {
		go func(id int32) {
			defer wg.Done()
			_ = p.LearnRecipe(id+1000, false)
		}(int32(i))
	}

	wg.Wait()

	// Проверяем что все рецепты выучены
	dwarvenCount := len(p.GetRecipeBook(true))
	commonCount := len(p.GetRecipeBook(false))
	total := p.RecipeCount()

	if total != dwarvenCount+commonCount {
		t.Errorf("RecipeCount() = %d, want %d (dwarven=%d + common=%d)",
			total, dwarvenCount+commonCount, dwarvenCount, commonCount)
	}

	if dwarvenCount != goroutines {
		t.Errorf("dwarven recipes count = %d, want %d", dwarvenCount, goroutines)
	}
	if commonCount != goroutines {
		t.Errorf("common recipes count = %d, want %d", commonCount, goroutines)
	}

	// Параллельное удаление
	var wg2 sync.WaitGroup
	wg2.Add(goroutines)
	for i := range goroutines {
		go func(id int32) {
			defer wg2.Done()
			_ = p.ForgetRecipe(id, true)
		}(int32(i))
	}
	wg2.Wait()

	dwarvenAfter := len(p.GetRecipeBook(true))
	if dwarvenAfter != 0 {
		t.Errorf("dwarven recipes after concurrent forget = %d, want 0", dwarvenAfter)
	}

	// Common рецепты не затронуты
	commonAfter := len(p.GetRecipeBook(false))
	if commonAfter != goroutines {
		t.Errorf("common recipes after dwarven forget = %d, want %d", commonAfter, goroutines)
	}
}

func TestRecipe_ConcurrentHasRecipe(t *testing.T) {
	t.Parallel()

	p := newRecipeTestPlayer(t)

	// Выучим несколько рецептов
	for i := range 10 {
		if err := p.LearnRecipe(int32(i), true); err != nil {
			t.Fatalf("LearnRecipe(%d): %v", i, err)
		}
	}

	// Параллельное чтение
	var wg sync.WaitGroup
	wg.Add(100)
	for range 100 {
		go func() {
			defer wg.Done()
			for i := range 10 {
				p.HasRecipe(int32(i))
			}
			p.GetRecipeBook(true)
			p.RecipeCount()
		}()
	}
	wg.Wait()

	// Если тест дошёл сюда без data race — горутинная безопасность подтверждена
}
