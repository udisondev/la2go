package hall

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewTable(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	const wantTotal = 44 // 38 auctionable + 6 siegable
	if got := tbl.HallCount(); got != wantTotal {
		t.Fatalf("HallCount() = %d; want %d", got, wantTotal)
	}

	// Проверяем что все холлы из knownHalls доступны по ID.
	for _, info := range knownHalls {
		h := tbl.Hall(info.ID)
		if h == nil {
			t.Errorf("Hall(%d) = nil; want non-nil for %q", info.ID, info.Name)
			continue
		}
		if h.Name() != info.Name {
			t.Errorf("Hall(%d).Name() = %q; want %q", info.ID, h.Name(), info.Name)
		}
		if h.Type() != info.HallType {
			t.Errorf("Hall(%d).Type() = %d; want %d", info.ID, h.Type(), info.HallType)
		}
		if info.HallType == TypeAuctionable && h.Lease() != info.Lease {
			t.Errorf("Hall(%d).Lease() = %d; want %d", info.ID, h.Lease(), info.Lease)
		}
	}
}

func TestTable_Hall(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	tests := []struct {
		name    string
		hallID  int32
		wantNil bool
	}{
		{name: "existing auctionable", hallID: 22, wantNil: false},
		{name: "existing siegable", hallID: 21, wantNil: false},
		{name: "non-existent", hallID: 9999, wantNil: true},
		{name: "zero ID", hallID: 0, wantNil: true},
		{name: "negative ID", hallID: -1, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tbl.Hall(tt.hallID)
			if tt.wantNil && got != nil {
				t.Errorf("Hall(%d) = %v; want nil", tt.hallID, got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("Hall(%d) = nil; want non-nil", tt.hallID)
			}
		})
	}
}

func TestTable_HallByOwner(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	// Без владельцев — nil.
	if got := tbl.HallByOwner(100); got != nil {
		t.Errorf("HallByOwner(100) with no owners = %v; want nil", got)
	}

	// Назначаем владельца.
	if err := tbl.SetOwner(22, 100); err != nil {
		t.Fatalf("SetOwner(22, 100): %v", err)
	}

	got := tbl.HallByOwner(100)
	if got == nil {
		t.Fatal("HallByOwner(100) = nil; want hall 22")
	}
	if got.ID() != 22 {
		t.Errorf("HallByOwner(100).ID() = %d; want 22", got.ID())
	}

	// Несуществующий владелец.
	if got := tbl.HallByOwner(999); got != nil {
		t.Errorf("HallByOwner(999) = %v; want nil", got)
	}
}

func TestTable_FreeHalls(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	free := tbl.FreeHalls()
	const wantAuctionable = 38
	if len(free) != wantAuctionable {
		t.Errorf("FreeHalls() initially = %d; want %d (all auctionable)", len(free), wantAuctionable)
	}

	// Все свободные холлы должны быть аукционными.
	for _, h := range free {
		if h.Type() != TypeAuctionable {
			t.Errorf("FreeHalls() contains non-auctionable hall %d (type=%d)", h.ID(), h.Type())
		}
	}

	// Назначаем владельца — количество свободных уменьшается.
	if err := tbl.SetOwner(22, 100); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	free = tbl.FreeHalls()
	if len(free) != wantAuctionable-1 {
		t.Errorf("FreeHalls() after one owner = %d; want %d", len(free), wantAuctionable-1)
	}
}

func TestTable_OwnedHalls(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	// Изначально пусто.
	if got := tbl.OwnedHalls(); len(got) != 0 {
		t.Errorf("OwnedHalls() initially = %d; want 0", len(got))
	}

	// Назначаем два холла.
	if err := tbl.SetOwner(22, 100); err != nil {
		t.Fatalf("SetOwner(22, 100): %v", err)
	}
	if err := tbl.SetOwner(36, 200); err != nil {
		t.Fatalf("SetOwner(36, 200): %v", err)
	}

	owned := tbl.OwnedHalls()
	if len(owned) != 2 {
		t.Errorf("OwnedHalls() = %d; want 2", len(owned))
	}

	ids := make(map[int32]bool, 2)
	for _, h := range owned {
		ids[h.ID()] = true
	}
	if !ids[22] || !ids[36] {
		t.Errorf("OwnedHalls() IDs = %v; want {22, 36}", ids)
	}
}

func TestTable_AllHalls(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	all := tbl.AllHalls()
	if len(all) != len(knownHalls) {
		t.Errorf("AllHalls() = %d; want %d", len(all), len(knownHalls))
	}
}

func TestTable_SiegableHalls(t *testing.T) {
	t.Parallel()

	tbl := NewTable()

	siegable := tbl.SiegableHalls()
	const wantSiegable = 6
	if len(siegable) != wantSiegable {
		t.Errorf("SiegableHalls() = %d; want %d", len(siegable), wantSiegable)
	}

	expectedIDs := map[int32]bool{21: true, 34: true, 35: true, 62: true, 63: true, 64: true}
	for _, h := range siegable {
		if h.Type() != TypeSiegable {
			t.Errorf("SiegableHalls() contains non-siegable hall %d (type=%d)", h.ID(), h.Type())
		}
		if !expectedIDs[h.ID()] {
			t.Errorf("SiegableHalls() unexpected hall ID %d", h.ID())
		}
	}
}

func TestTable_SetOwner(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		if err := tbl.SetOwner(22, 100); err != nil {
			t.Fatalf("SetOwner(22, 100) = %v; want nil", err)
		}
		h := tbl.Hall(22)
		if h.OwnerClanID() != 100 {
			t.Errorf("Hall(22).OwnerClanID() = %d; want 100", h.OwnerClanID())
		}
	})

	t.Run("hall not found", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		err := tbl.SetOwner(9999, 100)
		if !errors.Is(err, ErrHallNotFound) {
			t.Errorf("SetOwner(9999, 100) = %v; want ErrHallNotFound", err)
		}
	})

	t.Run("already owned", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		if err := tbl.SetOwner(22, 100); err != nil {
			t.Fatalf("SetOwner first: %v", err)
		}

		err := tbl.SetOwner(22, 200)
		if !errors.Is(err, ErrHallAlreadyOwned) {
			t.Errorf("SetOwner(22, 200) = %v; want ErrHallAlreadyOwned", err)
		}

		// Владелец не изменился.
		if got := tbl.Hall(22).OwnerClanID(); got != 100 {
			t.Errorf("OwnerClanID() after rejected SetOwner = %d; want 100", got)
		}
	})
}

func TestTable_FreeHall(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		if err := tbl.SetOwner(22, 100); err != nil {
			t.Fatalf("SetOwner: %v", err)
		}
		if err := tbl.FreeHall(22); err != nil {
			t.Fatalf("FreeHall(22) = %v; want nil", err)
		}
		if tbl.Hall(22).HasOwner() {
			t.Error("Hall(22).HasOwner() after FreeHall = true; want false")
		}
	})

	t.Run("hall not found", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		err := tbl.FreeHall(9999)
		if !errors.Is(err, ErrHallNotFound) {
			t.Errorf("FreeHall(9999) = %v; want ErrHallNotFound", err)
		}
	})

	t.Run("free unowned hall", func(t *testing.T) {
		t.Parallel()
		tbl := NewTable()

		// Освобождение холла без владельца не должно вызывать ошибку.
		if err := tbl.FreeHall(22); err != nil {
			t.Errorf("FreeHall(22) unowned = %v; want nil", err)
		}
	})
}

func TestTable_Auction(t *testing.T) {
	t.Parallel()

	tbl := NewTable()
	endDate := time.Now().Add(7 * 24 * time.Hour)

	// Начально аукциона нет.
	if got := tbl.Auction(22); got != nil {
		t.Errorf("Auction(22) initially = %v; want nil", got)
	}

	// Создаем аукцион.
	a := NewAuction(1, 22, 20_000_000, endDate)
	if err := tbl.StartAuction(a); err != nil {
		t.Fatalf("StartAuction: %v", err)
	}

	// Получаем аукцион.
	got := tbl.Auction(22)
	if got == nil {
		t.Fatal("Auction(22) after start = nil; want non-nil")
	}
	if got.ID() != 1 {
		t.Errorf("Auction.ID() = %d; want 1", got.ID())
	}
	if got.HallID() != 22 {
		t.Errorf("Auction.HallID() = %d; want 22", got.HallID())
	}
	if got.StartingBid() != 20_000_000 {
		t.Errorf("Auction.StartingBid() = %d; want 20000000", got.StartingBid())
	}

	// Удаляем аукцион.
	tbl.RemoveAuction(22)
	if got := tbl.Auction(22); got != nil {
		t.Errorf("Auction(22) after remove = %v; want nil", got)
	}
}

func TestTable_StartAuction_HallNotFound(t *testing.T) {
	t.Parallel()

	tbl := NewTable()
	a := NewAuction(1, 9999, 100, time.Now().Add(time.Hour))

	err := tbl.StartAuction(a)
	if !errors.Is(err, ErrHallNotFound) {
		t.Errorf("StartAuction with bad hallID = %v; want ErrHallNotFound", err)
	}
}

func TestTable_ActiveAuctions(t *testing.T) {
	t.Parallel()

	tbl := NewTable()
	endDate := time.Now().Add(7 * 24 * time.Hour)

	// Пусто изначально.
	if got := tbl.ActiveAuctions(); len(got) != 0 {
		t.Errorf("ActiveAuctions() initially = %d; want 0", len(got))
	}

	// Добавляем два аукциона.
	a1 := NewAuction(1, 22, 20_000_000, endDate)
	a2 := NewAuction(2, 23, 20_000_000, endDate)
	if err := tbl.StartAuction(a1); err != nil {
		t.Fatalf("StartAuction a1: %v", err)
	}
	if err := tbl.StartAuction(a2); err != nil {
		t.Fatalf("StartAuction a2: %v", err)
	}

	auctions := tbl.ActiveAuctions()
	if len(auctions) != 2 {
		t.Errorf("ActiveAuctions() = %d; want 2", len(auctions))
	}

	// Удаляем один — остается один.
	tbl.RemoveAuction(22)
	auctions = tbl.ActiveAuctions()
	if len(auctions) != 1 {
		t.Errorf("ActiveAuctions() after remove = %d; want 1", len(auctions))
	}
}

func TestTable_HallCount(t *testing.T) {
	t.Parallel()

	tbl := NewTable()
	if got := tbl.HallCount(); got != len(knownHalls) {
		t.Errorf("HallCount() = %d; want %d", got, len(knownHalls))
	}
}

func TestGetHallInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       int32
		wantNil  bool
		wantName string
	}{
		{name: "auctionable found", id: 22, wantNil: false, wantName: "Moonstone Hall"},
		{name: "siegable found", id: 21, wantNil: false, wantName: "Fortress of Resistance"},
		{name: "last hall", id: 64, wantNil: false, wantName: "Fortress of the Dead"},
		{name: "not found", id: 9999, wantNil: true},
		{name: "zero ID", id: 0, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetHallInfo(tt.id)
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetHallInfo(%d) = %v; want nil", tt.id, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetHallInfo(%d) = nil; want non-nil", tt.id)
			}
			if got.Name != tt.wantName {
				t.Errorf("GetHallInfo(%d).Name = %q; want %q", tt.id, got.Name, tt.wantName)
			}
		})
	}
}

func TestTable_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tbl := NewTable()
	const goroutines = 50

	// Используем холлы 22-30 (9 штук) для конкурентных операций.
	hallIDs := []int32{22, 23, 24, 25, 26, 27, 28, 29, 30}

	var wg sync.WaitGroup
	wg.Add(goroutines * 4)

	for i := range goroutines {
		hallID := hallIDs[i%len(hallIDs)]
		clanID := int32(1000 + i)

		// Конкурентные SetOwner (некоторые будут возвращать ошибку — это нормально).
		go func() {
			defer wg.Done()
			_ = tbl.SetOwner(hallID, clanID)
		}()

		// Конкурентные FreeHall.
		go func() {
			defer wg.Done()
			_ = tbl.FreeHall(hallID)
		}()

		// Конкурентные чтения.
		go func() {
			defer wg.Done()
			_ = tbl.FreeHalls()
		}()

		go func() {
			defer wg.Done()
			_ = tbl.OwnedHalls()
		}()
	}

	wg.Wait()
	// Если дошли сюда без data race — тест пройден.
}
