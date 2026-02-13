package hall

import (
	"sync"
	"testing"
	"time"
)

func newTestHall(t *testing.T) *ClanHall {
	t.Helper()
	return NewClanHall(22, "Moonstone Hall", TypeAuctionable, GradeD, "Gludio")
}

func TestNewClanHall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       int32
		hallName string
		hallType HallType
		grade    Grade
		location string
	}{
		{
			name:     "auctionable hall",
			id:       22,
			hallName: "Moonstone Hall",
			hallType: TypeAuctionable,
			grade:    GradeD,
			location: "Gludio",
		},
		{
			name:     "siegable hall",
			id:       21,
			hallName: "Fortress of Resistance",
			hallType: TypeSiegable,
			grade:    GradeNone,
			location: "Dion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := NewClanHall(tt.id, tt.hallName, tt.hallType, tt.grade, tt.location)

			if h.ID() != tt.id {
				t.Errorf("ID() = %d; want %d", h.ID(), tt.id)
			}
			if h.Name() != tt.hallName {
				t.Errorf("Name() = %q; want %q", h.Name(), tt.hallName)
			}
			if h.Type() != tt.hallType {
				t.Errorf("Type() = %d; want %d", h.Type(), tt.hallType)
			}
			if h.Grade() != tt.grade {
				t.Errorf("Grade() = %d; want %d", h.Grade(), tt.grade)
			}
			if h.Location() != tt.location {
				t.Errorf("Location() = %q; want %q", h.Location(), tt.location)
			}
			if h.OwnerClanID() != 0 {
				t.Errorf("OwnerClanID() = %d; want 0 (no owner)", h.OwnerClanID())
			}
			if h.HasOwner() {
				t.Error("HasOwner() = true; want false for new hall")
			}
		})
	}
}

func TestClanHall_SetOwnerClanID(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	h.SetOwnerClanID(100)
	if got := h.OwnerClanID(); got != 100 {
		t.Errorf("OwnerClanID() = %d; want 100", got)
	}

	h.SetOwnerClanID(200)
	if got := h.OwnerClanID(); got != 200 {
		t.Errorf("OwnerClanID() after update = %d; want 200", got)
	}

	h.SetOwnerClanID(0)
	if got := h.OwnerClanID(); got != 0 {
		t.Errorf("OwnerClanID() after clear = %d; want 0", got)
	}
}

func TestClanHall_HasOwner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		clanID  int32
		want    bool
	}{
		{name: "no owner (zero)", clanID: 0, want: false},
		{name: "has owner", clanID: 42, want: true},
		{name: "negative clan ID", clanID: -1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newTestHall(t)
			h.SetOwnerClanID(tt.clanID)
			if got := h.HasOwner(); got != tt.want {
				t.Errorf("HasOwner() = %v; want %v (clanID=%d)", got, tt.want, tt.clanID)
			}
		})
	}
}

func TestClanHall_Lease(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	if got := h.Lease(); got != 0 {
		t.Errorf("Lease() initial = %d; want 0", got)
	}

	h.SetLease(500_000)
	if got := h.Lease(); got != 500_000 {
		t.Errorf("Lease() after set = %d; want 500000", got)
	}

	h.SetLease(1_000_000)
	if got := h.Lease(); got != 1_000_000 {
		t.Errorf("Lease() after update = %d; want 1000000", got)
	}
}

func TestClanHall_PaidUntil(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	if got := h.PaidUntil(); !got.IsZero() {
		t.Errorf("PaidUntil() initial = %v; want zero", got)
	}

	deadline := time.Now().Add(7 * 24 * time.Hour)
	h.SetPaidUntil(deadline)
	if got := h.PaidUntil(); !got.Equal(deadline) {
		t.Errorf("PaidUntil() = %v; want %v", got, deadline)
	}
}

func TestClanHall_IsLeaseExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		clanID    int32
		paidUntil time.Time
		want      bool
	}{
		{
			name:      "no owner - not expired",
			clanID:    0,
			paidUntil: time.Now().Add(-time.Hour),
			want:      false,
		},
		{
			name:      "has owner - lease active",
			clanID:    10,
			paidUntil: time.Now().Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "has owner - lease expired",
			clanID:    10,
			paidUntil: time.Now().Add(-time.Hour),
			want:      true,
		},
		{
			name:      "has owner - zero paidUntil (expired)",
			clanID:    10,
			paidUntil: time.Time{},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newTestHall(t)
			h.SetOwnerClanID(tt.clanID)
			h.SetPaidUntil(tt.paidUntil)
			if got := h.IsLeaseExpired(); got != tt.want {
				t.Errorf("IsLeaseExpired() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestClanHall_SetFunction(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	f1 := &Function{Type: FuncRestoreHP, Level: 3, Lease: 10_000}
	h.SetFunction(f1)

	got := h.Function(FuncRestoreHP)
	if got == nil {
		t.Fatal("Function(FuncRestoreHP) = nil; want non-nil")
	}
	if got.Level != 3 {
		t.Errorf("Function(FuncRestoreHP).Level = %d; want 3", got.Level)
	}

	// Замена существующей функции.
	f2 := &Function{Type: FuncRestoreHP, Level: 5, Lease: 21_000}
	h.SetFunction(f2)

	got = h.Function(FuncRestoreHP)
	if got == nil {
		t.Fatal("Function(FuncRestoreHP) after replace = nil; want non-nil")
	}
	if got.Level != 5 {
		t.Errorf("Function(FuncRestoreHP).Level after replace = %d; want 5", got.Level)
	}
}

func TestClanHall_RemoveFunction(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	h.SetFunction(&Function{Type: FuncRestoreHP, Level: 1})
	h.SetFunction(&Function{Type: FuncRestoreMP, Level: 2})

	h.RemoveFunction(FuncRestoreHP)

	if got := h.Function(FuncRestoreHP); got != nil {
		t.Errorf("Function(FuncRestoreHP) after remove = %v; want nil", got)
	}

	// MP должна остаться.
	if got := h.Function(FuncRestoreMP); got == nil {
		t.Error("Function(FuncRestoreMP) = nil; want non-nil (should not be removed)")
	}

	// Повторное удаление несуществующей функции не паникует.
	h.RemoveFunction(FuncRestoreHP)
}

func TestClanHall_Functions(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	// Пустой snapshot.
	if got := h.Functions(); len(got) != 0 {
		t.Errorf("Functions() on empty = %d items; want 0", len(got))
	}

	h.SetFunction(&Function{Type: FuncRestoreHP, Level: 1})
	h.SetFunction(&Function{Type: FuncRestoreMP, Level: 2})
	h.SetFunction(&Function{Type: FuncTeleport, Level: 1})

	fns := h.Functions()
	if len(fns) != 3 {
		t.Errorf("Functions() = %d items; want 3", len(fns))
	}

	// Проверяем что все типы присутствуют.
	types := make(map[FunctionType]bool, 3)
	for _, f := range fns {
		types[f.Type] = true
	}
	for _, ft := range []FunctionType{FuncRestoreHP, FuncRestoreMP, FuncTeleport} {
		if !types[ft] {
			t.Errorf("Functions() missing type %d", ft)
		}
	}
}

func TestClanHall_FunctionLevel(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	// Отсутствующая функция возвращает 0.
	if got := h.FunctionLevel(FuncSupport); got != 0 {
		t.Errorf("FunctionLevel(FuncSupport) missing = %d; want 0", got)
	}

	h.SetFunction(&Function{Type: FuncSupport, Level: 7})
	if got := h.FunctionLevel(FuncSupport); got != 7 {
		t.Errorf("FunctionLevel(FuncSupport) = %d; want 7", got)
	}
}

func TestClanHall_Free(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)
	h.SetOwner(100)
	h.SetFunction(&Function{Type: FuncRestoreHP, Level: 5})
	h.SetFunction(&Function{Type: FuncTeleport, Level: 2})

	h.Free()

	if h.HasOwner() {
		t.Error("HasOwner() after Free() = true; want false")
	}
	if got := h.OwnerClanID(); got != 0 {
		t.Errorf("OwnerClanID() after Free() = %d; want 0", got)
	}
	if got := h.PaidUntil(); !got.IsZero() {
		t.Errorf("PaidUntil() after Free() = %v; want zero", got)
	}
	if got := h.Functions(); len(got) != 0 {
		t.Errorf("Functions() after Free() = %d items; want 0", len(got))
	}
}

func TestClanHall_SetOwner(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)
	before := time.Now()
	h.SetOwner(42)
	after := time.Now()

	if got := h.OwnerClanID(); got != 42 {
		t.Errorf("OwnerClanID() after SetOwner = %d; want 42", got)
	}

	paid := h.PaidUntil()
	expectedMin := before.Add(FeeRate)
	expectedMax := after.Add(FeeRate)
	if paid.Before(expectedMin) || paid.After(expectedMax) {
		t.Errorf("PaidUntil() = %v; want between %v and %v", paid, expectedMin, expectedMax)
	}

	// Функции очищены при смене владельца.
	if got := h.Functions(); len(got) != 0 {
		t.Errorf("Functions() after SetOwner = %d items; want 0 (cleared)", len(got))
	}
}

func TestClanHall_SiegeFields(t *testing.T) {
	t.Parallel()

	h := NewClanHall(21, "Fortress of Resistance", TypeSiegable, GradeNone, "Dion")

	siegeDate := time.Date(2026, 3, 1, 18, 0, 0, 0, time.UTC)
	h.SetNextSiege(siegeDate)
	if got := h.NextSiege(); !got.Equal(siegeDate) {
		t.Errorf("NextSiege() = %v; want %v", got, siegeDate)
	}

	siegeLen := 2 * time.Hour
	h.SetSiegeLength(siegeLen)
	if got := h.SiegeLength(); got != siegeLen {
		t.Errorf("SiegeLength() = %v; want %v", got, siegeLen)
	}
}

func TestClanHall_ZoneAndDescription(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)

	// Начальные значения — нулевые.
	if got := h.ZoneID(); got != 0 {
		t.Errorf("ZoneID() initial = %d; want 0", got)
	}
	if got := h.Description(); got != "" {
		t.Errorf("Description() initial = %q; want empty", got)
	}

	h.SetZoneID(12010)
	if got := h.ZoneID(); got != 12010 {
		t.Errorf("ZoneID() = %d; want 12010", got)
	}

	h.SetDescription("A grand hall in Gludio castle town.")
	if got := h.Description(); got != "A grand hall in Gludio castle town." {
		t.Errorf("Description() = %q; want %q", got, "A grand hall in Gludio castle town.")
	}
}

func TestClanHall_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	h := newTestHall(t)
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 4)

	// Конкурентные записи и чтения.
	for range goroutines {
		go func() {
			defer wg.Done()
			h.SetOwnerClanID(42)
		}()
		go func() {
			defer wg.Done()
			_ = h.OwnerClanID()
		}()
		go func() {
			defer wg.Done()
			h.SetFunction(&Function{Type: FuncRestoreHP, Level: 3})
		}()
		go func() {
			defer wg.Done()
			_ = h.Functions()
		}()
	}

	wg.Wait()

	// Если дошли сюда без data race — тест пройден.
}

func TestFunction_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		endTime time.Time
		want    bool
	}{
		{
			name:    "active - future end time",
			endTime: time.Now().Add(24 * time.Hour),
			want:    true,
		},
		{
			name:    "expired - past end time",
			endTime: time.Now().Add(-time.Hour),
			want:    false,
		},
		{
			name:    "expired - zero time",
			endTime: time.Time{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &Function{
				Type:    FuncRestoreHP,
				Level:   1,
				EndTime: tt.endTime,
			}
			if got := f.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v; want %v (endTime=%v)", got, tt.want, tt.endTime)
			}
		})
	}
}
