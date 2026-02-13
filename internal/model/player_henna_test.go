package model

import (
	"sync"
	"testing"

	"github.com/udisondev/la2go/internal/data"
)

var hennaOnce sync.Once

// setupHennaTestData loads henna templates once for testing.
// classID=11 has access to dyeIDs 1-6 (STR/CON/DEX hennas).
func setupHennaTestData(t *testing.T) {
	t.Helper()
	hennaOnce.Do(func() {
		if err := data.LoadHennaTemplates(); err != nil {
			t.Fatalf("LoadHennaTemplates() error: %v", err)
		}
	})
}

func newTestPlayerForHenna(t *testing.T, classID int32) *Player {
	t.Helper()
	p, err := NewPlayer(1, 100, 1, "TestHenna", 40, 0, classID)
	if err != nil {
		t.Fatalf("NewPlayer() error: %v", err)
	}
	return p
}

func TestAddHenna(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	tests := []struct {
		name     string
		classID  int32
		dyeID   int32
		wantSlot int
		wantErr  bool
	}{
		{
			name:     "equip first henna",
			classID:  11, // classID=11 is allowed for dyeID=1
			dyeID:   1,
			wantSlot: 1,
		},
		{
			name:    "henna not found",
			classID: 11,
			dyeID:  99999,
			wantErr: true,
		},
		{
			name:    "class not allowed",
			classID: 999, // не в списке допустимых классов
			dyeID:  7,    // dyeID=7 allowed only for classIDs: 11, 26, 39
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := newTestPlayerForHenna(t, tt.classID)

			slot, err := p.AddHenna(tt.dyeID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("AddHenna() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("AddHenna() error: %v", err)
			}

			if slot != tt.wantSlot {
				t.Errorf("AddHenna() slot = %d; want %d", slot, tt.wantSlot)
			}
		})
	}
}

func TestAddHenna_FillAllSlots(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// dyeIDs 1, 2, 3 все доступны для classID=11
	for i, dyeID := range []int32{1, 2, 3} {
		slot, err := p.AddHenna(dyeID)
		if err != nil {
			t.Fatalf("AddHenna(dyeID=%d) error: %v", dyeID, err)
		}
		if slot != i+1 {
			t.Errorf("AddHenna(dyeID=%d) slot = %d; want %d", dyeID, slot, i+1)
		}
	}

	// Четвёртая хенна — ошибка, слотов нет
	_, err := p.AddHenna(4)
	if err == nil {
		t.Fatal("AddHenna() expected error for 4th henna, got nil")
	}

	if p.GetHennaEmptySlots() != 0 {
		t.Errorf("GetHennaEmptySlots() = %d; want 0", p.GetHennaEmptySlots())
	}
}

func TestRemoveHenna(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	tests := []struct {
		name    string
		slot    int
		setup   func(p *Player) // настройка хенн до удаления
		wantDye int32
		wantErr bool
	}{
		{
			name: "remove from slot 1",
			slot: 1,
			setup: func(p *Player) {
				if _, err := p.AddHenna(1); err != nil {
					panic(err)
				}
			},
			wantDye: 1,
		},
		{
			name:    "remove from empty slot",
			slot:    1,
			setup:   func(p *Player) {},
			wantErr: true,
		},
		{
			name:    "invalid slot 0",
			slot:    0,
			setup:   func(p *Player) {},
			wantErr: true,
		},
		{
			name:    "invalid slot 4",
			slot:    4,
			setup:   func(p *Player) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := newTestPlayerForHenna(t, 11)
			tt.setup(p)

			dyeID, err := p.RemoveHenna(tt.slot)
			if tt.wantErr {
				if err == nil {
					t.Fatal("RemoveHenna() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("RemoveHenna() error: %v", err)
			}

			if dyeID != tt.wantDye {
				t.Errorf("RemoveHenna() dyeID = %d; want %d", dyeID, tt.wantDye)
			}

			// Слот должен быть пустым после удаления
			if p.GetHenna(tt.slot) != nil {
				t.Error("slot should be empty after RemoveHenna()")
			}
		})
	}
}

func TestSetHenna(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// SetHenna — восстановление из БД, без валидации класса
	if err := p.SetHenna(1, 42); err != nil {
		t.Fatalf("SetHenna() error: %v", err)
	}

	h := p.GetHenna(1)
	if h == nil {
		t.Fatal("GetHenna(1) = nil after SetHenna")
	}
	if h.DyeID != 42 {
		t.Errorf("GetHenna(1).DyeID = %d; want 42", h.DyeID)
	}

	// Невалидный слот
	if err := p.SetHenna(0, 1); err == nil {
		t.Error("SetHenna(0, 1) expected error, got nil")
	}
	if err := p.SetHenna(4, 1); err == nil {
		t.Error("SetHenna(4, 1) expected error, got nil")
	}
}

func TestGetHennaEmptySlots(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	if got := p.GetHennaEmptySlots(); got != 3 {
		t.Errorf("GetHennaEmptySlots() = %d; want 3 (fresh player)", got)
	}

	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}
	if got := p.GetHennaEmptySlots(); got != 2 {
		t.Errorf("GetHennaEmptySlots() = %d; want 2", got)
	}

	if _, err := p.AddHenna(2); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddHenna(3); err != nil {
		t.Fatal(err)
	}
	if got := p.GetHennaEmptySlots(); got != 0 {
		t.Errorf("GetHennaEmptySlots() = %d; want 0", got)
	}
}

func TestHasHennas(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	if p.HasHennas() {
		t.Error("HasHennas() = true; want false for fresh player")
	}

	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}
	if !p.HasHennas() {
		t.Error("HasHennas() = false; want true after AddHenna")
	}

	if _, err := p.RemoveHenna(1); err != nil {
		t.Fatal(err)
	}
	if p.HasHennas() {
		t.Error("HasHennas() = true; want false after RemoveHenna")
	}
}

func TestGetHennaList(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddHenna(2); err != nil {
		t.Fatal(err)
	}

	list := p.GetHennaList()
	if list[0] == nil || list[0].DyeID != 1 {
		t.Errorf("GetHennaList()[0].DyeID = %v; want 1", list[0])
	}
	if list[1] == nil || list[1].DyeID != 2 {
		t.Errorf("GetHennaList()[1].DyeID = %v; want 2", list[1])
	}
	if list[2] != nil {
		t.Errorf("GetHennaList()[2] = %v; want nil", list[2])
	}
}

func TestRecalcHennaStats(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// dyeID=1: STR+1, CON-3
	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}

	if got := p.HennaStatSTR(); got != 1 {
		t.Errorf("HennaStatSTR() = %d; want 1", got)
	}
	// Отрицательные бонусы (CON-3) — addHennaStat суммирует любые значения
	if got := p.HennaStatCON(); got != -3 {
		t.Errorf("HennaStatCON() = %d; want -3", got)
	}

	// dyeID=2: STR+1, DEX-3
	if _, err := p.AddHenna(2); err != nil {
		t.Fatal(err)
	}

	if got := p.HennaStatSTR(); got != 2 {
		t.Errorf("HennaStatSTR() = %d; want 2", got)
	}
	if got := p.HennaStatDEX(); got != -3 {
		t.Errorf("HennaStatDEX() = %d; want -3", got)
	}
}

func TestRecalcHennaStats_Cap(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// Заполним 3 слота хеннами с STR+1 каждая: dyeID=1(STR+1), dyeID=2(STR+1), dyeID=13(STR+1)
	for _, dyeID := range []int32{1, 2, 13} {
		if _, err := p.AddHenna(dyeID); err != nil {
			t.Fatal(err)
		}
	}

	// 3 хенны по STR+1 = 3 (меньше кэпа 5)
	if got := p.HennaStatSTR(); got != 3 {
		t.Errorf("HennaStatSTR() = %d; want 3 (3 hennas, each +1)", got)
	}
}

func TestRecalcHennaStats_CapAt5(t *testing.T) {
	t.Parallel()

	// Тест кэпа через addHennaStat напрямую
	tests := []struct {
		name    string
		current int32
		bonus   int32
		want    int32
	}{
		{"no cap", 0, 3, 3},
		{"exact cap", 2, 3, 5},
		{"over cap", 3, 4, 5},
		{"already capped", 5, 1, 5},
		{"negative bonus", 0, -2, -2},
		{"negative then positive", -2, 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := addHennaStat(tt.current, tt.bonus)
			if got != tt.want {
				t.Errorf("addHennaStat(%d, %d) = %d; want %d", tt.current, tt.bonus, got, tt.want)
			}
		})
	}
}

func TestRecalcHennaStats_INTMENWITStats(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// dyeID=7: INT+1, MEN-3 (allowed for classID=11)
	if _, err := p.AddHenna(7); err != nil {
		t.Fatal(err)
	}

	if got := p.HennaStatINT(); got != 1 {
		t.Errorf("HennaStatINT() = %d; want 1", got)
	}
	if got := p.HennaStatMEN(); got != -3 {
		t.Errorf("HennaStatMEN() = %d; want -3", got)
	}

	// dyeID=11: WIT+1, INT-3 (allowed for classID=11)
	if _, err := p.AddHenna(11); err != nil {
		t.Fatal(err)
	}

	if got := p.HennaStatWIT(); got != 1 {
		t.Errorf("HennaStatWIT() = %d; want 1", got)
	}
	// INT: 1 + (-3) = -2
	if got := p.HennaStatINT(); got != -2 {
		t.Errorf("HennaStatINT() = %d; want -2", got)
	}
}

func TestRecalcHennaStats_PublicMethod(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// Используем SetHenna (без RecalcHennaStats) + явный RecalcHennaStats
	if err := p.SetHenna(1, 1); err != nil { // dyeID=1: STR+1, CON-3
		t.Fatal(err)
	}

	// SetHenna не пересчитывает статы
	if got := p.HennaStatSTR(); got != 0 {
		t.Errorf("HennaStatSTR() = %d; want 0 before RecalcHennaStats", got)
	}

	// Явный вызов RecalcHennaStats
	p.RecalcHennaStats()

	if got := p.HennaStatSTR(); got != 1 {
		t.Errorf("HennaStatSTR() = %d; want 1 after RecalcHennaStats", got)
	}
}

func TestRecalcHennaStats_AfterRemove(t *testing.T) {
	t.Parallel()
	setupHennaTestData(t)

	p := newTestPlayerForHenna(t, 11)

	// Добавляем dyeID=1 (STR+1, CON-3)
	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}
	if got := p.HennaStatSTR(); got != 1 {
		t.Errorf("HennaStatSTR() = %d; want 1 before remove", got)
	}

	// Удаляем
	if _, err := p.RemoveHenna(1); err != nil {
		t.Fatal(err)
	}

	// Статы должны вернуться к 0
	if got := p.HennaStatSTR(); got != 0 {
		t.Errorf("HennaStatSTR() = %d; want 0 after remove", got)
	}
	if got := p.HennaStatCON(); got != 0 {
		t.Errorf("HennaStatCON() = %d; want 0 after remove", got)
	}
}
