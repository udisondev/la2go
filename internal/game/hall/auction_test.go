package hall

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func futureDate() time.Time { return time.Now().Add(24 * time.Hour) }
func pastDate() time.Time   { return time.Now().Add(-24 * time.Hour) }

func TestNewAuction(t *testing.T) {
	t.Parallel()

	end := futureDate()
	a := NewAuction(10, 20, 5000, end)

	if a == nil {
		t.Fatal("NewAuction returned nil")
	}
	if got := a.ID(); got != 10 {
		t.Errorf("ID() = %d; want 10", got)
	}
	if got := a.HallID(); got != 20 {
		t.Errorf("HallID() = %d; want 20", got)
	}
	if got := a.StartingBid(); got != 5000 {
		t.Errorf("StartingBid() = %d; want 5000", got)
	}
	if got := a.EndDate(); !got.Equal(end) {
		t.Errorf("EndDate() = %v; want %v", got, end)
	}
	if got := a.BidderCount(); got != 0 {
		t.Errorf("BidderCount() = %d; want 0", got)
	}
	clanID, amount := a.HighestBidder()
	if clanID != 0 || amount != 0 {
		t.Errorf("HighestBidder() = (%d, %d); want (0, 0)", clanID, amount)
	}
}

func TestAuction_ID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   int32
	}{
		{"positive", 1},
		{"large", 9999},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewAuction(tt.id, 1, 100, futureDate())
			if got := a.ID(); got != tt.id {
				t.Errorf("ID() = %d; want %d", got, tt.id)
			}
		})
	}
}

func TestAuction_HallID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		hallID int32
	}{
		{"normal", 42},
		{"zero", 0},
		{"large", 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewAuction(1, tt.hallID, 100, futureDate())
			if got := a.HallID(); got != tt.hallID {
				t.Errorf("HallID() = %d; want %d", got, tt.hallID)
			}
		})
	}
}

func TestAuction_StartingBid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		bid  int64
	}{
		{"small", 100},
		{"large", 1_000_000},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewAuction(1, 1, tt.bid, futureDate())
			if got := a.StartingBid(); got != tt.bid {
				t.Errorf("StartingBid() = %d; want %d", got, tt.bid)
			}
		})
	}
}

func TestAuction_EndDate(t *testing.T) {
	t.Parallel()

	end := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	a := NewAuction(1, 1, 100, end)

	if got := a.EndDate(); !got.Equal(end) {
		t.Errorf("EndDate() = %v; want %v", got, end)
	}
}

func TestAuction_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		end    time.Time
		active bool
	}{
		{"future end date is active", futureDate(), true},
		{"past end date is inactive", pastDate(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewAuction(1, 1, 100, tt.end)
			if got := a.IsActive(); got != tt.active {
				t.Errorf("IsActive() = %v; want %v", got, tt.active)
			}
		})
	}
}

func TestAuction_PlaceBid_FirstBid(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 1000, futureDate())
	deduct, err := a.PlaceBid(100, 1000)
	if err != nil {
		t.Fatalf("PlaceBid() error = %v; want nil", err)
	}
	if deduct != 1000 {
		t.Errorf("PlaceBid() deductAmount = %d; want 1000", deduct)
	}

	if got := a.BidderCount(); got != 1 {
		t.Errorf("BidderCount() = %d; want 1", got)
	}

	clanID, amount := a.HighestBidder()
	if clanID != 100 {
		t.Errorf("HighestBidder() clanID = %d; want 100", clanID)
	}
	if amount != 1000 {
		t.Errorf("HighestBidder() amount = %d; want 1000", amount)
	}

	b := a.Bidder(100)
	if b == nil {
		t.Fatal("Bidder(100) = nil; want non-nil")
	}
	if b.ClanID != 100 {
		t.Errorf("Bidder.ClanID = %d; want 100", b.ClanID)
	}
	if b.MaxBid != 1000 {
		t.Errorf("Bidder.MaxBid = %d; want 1000", b.MaxBid)
	}
	if b.CurBid != 1000 {
		t.Errorf("Bidder.CurBid = %d; want 1000", b.CurBid)
	}
}

func TestAuction_PlaceBid_BelowMinimum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		startBid   int64
		bidAmount  int64
		setupBid   int64 // если > 0, сначала делаем ставку для поднятия минимума
		setupClan  int32
	}{
		{"below starting bid", 1000, 999, 0, 0},
		{"below highest bid + 1", 1000, 1500, 1500, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewAuction(1, 1, tt.startBid, futureDate())

			if tt.setupBid > 0 {
				if _, err := a.PlaceBid(tt.setupClan, tt.setupBid); err != nil {
					t.Fatalf("setup PlaceBid() error = %v", err)
				}
			}

			_, err := a.PlaceBid(300, tt.bidAmount)
			if !errors.Is(err, ErrBidTooLow) {
				t.Errorf("PlaceBid() error = %v; want %v", err, ErrBidTooLow)
			}
		})
	}
}

func TestAuction_PlaceBid_UpdateExistingBid(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 1000, futureDate())

	// Первая ставка.
	deduct1, err := a.PlaceBid(100, 1000)
	if err != nil {
		t.Fatalf("first PlaceBid() error = %v", err)
	}
	if deduct1 != 1000 {
		t.Errorf("first deductAmount = %d; want 1000", deduct1)
	}

	// Обновление ставки — должна вернуться только разница.
	deduct2, err := a.PlaceBid(100, 2000)
	if err != nil {
		t.Fatalf("second PlaceBid() error = %v", err)
	}
	if deduct2 != 1000 {
		t.Errorf("second deductAmount = %d; want 1000 (difference)", deduct2)
	}

	b := a.Bidder(100)
	if b == nil {
		t.Fatal("Bidder(100) = nil; want non-nil")
	}
	if b.MaxBid != 2000 {
		t.Errorf("Bidder.MaxBid = %d; want 2000", b.MaxBid)
	}
	if b.CurBid != 2000 {
		t.Errorf("Bidder.CurBid = %d; want 2000", b.CurBid)
	}
}

func TestAuction_PlaceBid_ExistingBidTooLow(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 1000, futureDate())

	if _, err := a.PlaceBid(100, 2000); err != nil {
		t.Fatalf("initial PlaceBid() error = %v", err)
	}

	// Попытка обновить ставку на ту же или меньшую сумму.
	_, err := a.PlaceBid(100, 1500)
	if !errors.Is(err, ErrBidTooLow) {
		t.Errorf("PlaceBid() error = %v; want %v", err, ErrBidTooLow)
	}

	_, err = a.PlaceBid(100, 2000)
	if !errors.Is(err, ErrBidTooLow) {
		t.Errorf("PlaceBid(same amount) error = %v; want %v", err, ErrBidTooLow)
	}
}

func TestAuction_PlaceBid_HighestBidderTracking(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if _, err := a.PlaceBid(10, 100); err != nil {
		t.Fatalf("PlaceBid(10) error = %v", err)
	}
	clanID, amount := a.HighestBidder()
	if clanID != 10 || amount != 100 {
		t.Errorf("after first bid: HighestBidder() = (%d, %d); want (10, 100)", clanID, amount)
	}

	if _, err := a.PlaceBid(20, 200); err != nil {
		t.Fatalf("PlaceBid(20) error = %v", err)
	}
	clanID, amount = a.HighestBidder()
	if clanID != 20 || amount != 200 {
		t.Errorf("after second bid: HighestBidder() = (%d, %d); want (20, 200)", clanID, amount)
	}

	if _, err := a.PlaceBid(30, 500); err != nil {
		t.Fatalf("PlaceBid(30) error = %v", err)
	}
	clanID, amount = a.HighestBidder()
	if clanID != 30 || amount != 500 {
		t.Errorf("after third bid: HighestBidder() = (%d, %d); want (30, 500)", clanID, amount)
	}
}

func TestAuction_PlaceBid_AuctionClosed(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, pastDate())

	_, err := a.PlaceBid(10, 100)
	if !errors.Is(err, ErrAuctionClosed) {
		t.Errorf("PlaceBid() error = %v; want %v", err, ErrAuctionClosed)
	}
}

func TestAuction_BidderCount(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if got := a.BidderCount(); got != 0 {
		t.Errorf("initial BidderCount() = %d; want 0", got)
	}

	for i := range int32(5) {
		clanID := i + 1
		bid := int64(100 + (i+1)*100)
		if _, err := a.PlaceBid(clanID, bid); err != nil {
			t.Fatalf("PlaceBid(%d, %d) error = %v", clanID, bid, err)
		}
	}

	if got := a.BidderCount(); got != 5 {
		t.Errorf("BidderCount() = %d; want 5", got)
	}
}

func TestAuction_Bidder(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	// Нет ставок — nil.
	if b := a.Bidder(999); b != nil {
		t.Errorf("Bidder(999) = %+v; want nil", b)
	}

	if _, err := a.PlaceBid(42, 100); err != nil {
		t.Fatalf("PlaceBid() error = %v", err)
	}

	b := a.Bidder(42)
	if b == nil {
		t.Fatal("Bidder(42) = nil; want non-nil")
	}
	if b.ClanID != 42 {
		t.Errorf("Bidder.ClanID = %d; want 42", b.ClanID)
	}
	if b.MaxBid != 100 {
		t.Errorf("Bidder.MaxBid = %d; want 100", b.MaxBid)
	}
}

func TestAuction_Bidders(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if got := a.Bidders(); len(got) != 0 {
		t.Errorf("Bidders() len = %d; want 0", len(got))
	}

	clans := []int32{10, 20, 30}
	for i, c := range clans {
		bid := int64(100 + (i+1)*100)
		if _, err := a.PlaceBid(c, bid); err != nil {
			t.Fatalf("PlaceBid(%d) error = %v", c, err)
		}
	}

	bidders := a.Bidders()
	if len(bidders) != 3 {
		t.Fatalf("Bidders() len = %d; want 3", len(bidders))
	}

	// Проверяем, что все кланы присутствуют.
	seen := make(map[int32]bool, 3)
	for _, b := range bidders {
		seen[b.ClanID] = true
	}
	for _, c := range clans {
		if !seen[c] {
			t.Errorf("Bidders() missing clan %d", c)
		}
	}
}

func TestAuction_CancelBid(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if _, err := a.PlaceBid(10, 1000); err != nil {
		t.Fatalf("PlaceBid() error = %v", err)
	}

	refund, err := a.CancelBid(10)
	if err != nil {
		t.Fatalf("CancelBid() error = %v", err)
	}

	// 1000 - 10% = 900
	wantRefund := int64(900)
	if refund != wantRefund {
		t.Errorf("CancelBid() refund = %d; want %d", refund, wantRefund)
	}

	if got := a.BidderCount(); got != 0 {
		t.Errorf("BidderCount() = %d; want 0 after cancel", got)
	}

	if b := a.Bidder(10); b != nil {
		t.Errorf("Bidder(10) = %+v; want nil after cancel", b)
	}
}

func TestAuction_CancelBid_NoBid(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	_, err := a.CancelBid(999)
	if !errors.Is(err, ErrNoBid) {
		t.Errorf("CancelBid() error = %v; want %v", err, ErrNoBid)
	}
}

func TestAuction_CancelBid_RecalculatesHighest(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if _, err := a.PlaceBid(10, 100); err != nil {
		t.Fatalf("PlaceBid(10) error = %v", err)
	}
	if _, err := a.PlaceBid(20, 200); err != nil {
		t.Fatalf("PlaceBid(20) error = %v", err)
	}
	if _, err := a.PlaceBid(30, 500); err != nil {
		t.Fatalf("PlaceBid(30) error = %v", err)
	}

	// Лидер — clan 30 с 500.
	clanID, amount := a.HighestBidder()
	if clanID != 30 || amount != 500 {
		t.Fatalf("before cancel: HighestBidder() = (%d, %d); want (30, 500)", clanID, amount)
	}

	// Отменяем ставку лидера.
	if _, err := a.CancelBid(30); err != nil {
		t.Fatalf("CancelBid(30) error = %v", err)
	}

	// Новый лидер — clan 20 с 200.
	clanID, amount = a.HighestBidder()
	if clanID != 20 || amount != 200 {
		t.Errorf("after cancel: HighestBidder() = (%d, %d); want (20, 200)", clanID, amount)
	}

	if got := a.BidderCount(); got != 2 {
		t.Errorf("BidderCount() = %d; want 2", got)
	}
}

func TestAuction_EndAuction_WithWinner(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	if _, err := a.PlaceBid(10, 100); err != nil {
		t.Fatalf("PlaceBid(10) error = %v", err)
	}
	if _, err := a.PlaceBid(20, 300); err != nil {
		t.Fatalf("PlaceBid(20) error = %v", err)
	}
	if _, err := a.PlaceBid(30, 500); err != nil {
		t.Fatalf("PlaceBid(30) error = %v", err)
	}

	winner, bid := a.EndAuction()
	if winner != 30 {
		t.Errorf("EndAuction() winnerClanID = %d; want 30", winner)
	}
	if bid != 500 {
		t.Errorf("EndAuction() winningBid = %d; want 500", bid)
	}
}

func TestAuction_EndAuction_NoBids(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	winner, bid := a.EndAuction()
	if winner != 0 {
		t.Errorf("EndAuction() winnerClanID = %d; want 0", winner)
	}
	if bid != 0 {
		t.Errorf("EndAuction() winningBid = %d; want 0", bid)
	}
}

func TestAuction_LoserBids(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	// Нет ставок — пустой список.
	if losers := a.LoserBids(); len(losers) != 0 {
		t.Errorf("LoserBids() len = %d; want 0 (no bids)", len(losers))
	}

	if _, err := a.PlaceBid(10, 100); err != nil {
		t.Fatalf("PlaceBid(10) error = %v", err)
	}
	if _, err := a.PlaceBid(20, 300); err != nil {
		t.Fatalf("PlaceBid(20) error = %v", err)
	}
	if _, err := a.PlaceBid(30, 500); err != nil {
		t.Fatalf("PlaceBid(30) error = %v", err)
	}

	losers := a.LoserBids()
	if len(losers) != 2 {
		t.Fatalf("LoserBids() len = %d; want 2", len(losers))
	}

	loserIDs := make(map[int32]bool, 2)
	for _, l := range losers {
		loserIDs[l.ClanID] = true
	}

	if loserIDs[30] {
		t.Error("LoserBids() contains winner clan 30")
	}
	if !loserIDs[10] {
		t.Error("LoserBids() missing loser clan 10")
	}
	if !loserIDs[20] {
		t.Error("LoserBids() missing loser clan 20")
	}
}

func TestAuction_HighestBidder(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 100, futureDate())

	// До ставок.
	clanID, amount := a.HighestBidder()
	if clanID != 0 || amount != 0 {
		t.Errorf("initial HighestBidder() = (%d, %d); want (0, 0)", clanID, amount)
	}

	// После одной ставки.
	if _, err := a.PlaceBid(77, 100); err != nil {
		t.Fatalf("PlaceBid() error = %v", err)
	}
	clanID, amount = a.HighestBidder()
	if clanID != 77 || amount != 100 {
		t.Errorf("HighestBidder() = (%d, %d); want (77, 100)", clanID, amount)
	}

	// Перебивают ставку.
	if _, err := a.PlaceBid(88, 200); err != nil {
		t.Fatalf("PlaceBid() error = %v", err)
	}
	clanID, amount = a.HighestBidder()
	if clanID != 88 || amount != 200 {
		t.Errorf("HighestBidder() = (%d, %d); want (88, 200)", clanID, amount)
	}
}

func TestAuction_ConcurrentBids(t *testing.T) {
	t.Parallel()

	a := NewAuction(1, 1, 1, futureDate())

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range int32(goroutines) {
		clanID := i + 1
		go func() {
			defer wg.Done()
			// Каждая горутина ставит уникальную сумму.
			// Ставка = clanID * 100, чтобы гарантировать уникальность и рост.
			bid := int64(clanID) * 100
			// Ошибки ожидаемы (ErrBidTooLow) из-за гонки —
			// важно отсутствие паник и data races.
			a.PlaceBid(clanID, bid)
		}()
	}
	wg.Wait()

	// Проверяем инвариант: highestBidder соответствует реальным данным.
	hClan, hAmount := a.HighestBidder()
	if hClan == 0 {
		t.Fatal("HighestBidder() clanID = 0 after concurrent bids; want non-zero")
	}

	b := a.Bidder(hClan)
	if b == nil {
		t.Fatalf("Bidder(%d) = nil for highest bidder", hClan)
	}
	if b.MaxBid != hAmount {
		t.Errorf("highest bidder MaxBid = %d; HighestBidder amount = %d; mismatch", b.MaxBid, hAmount)
	}

	// Все зарегистрированные участники должны иметь корректные ставки.
	for _, bd := range a.Bidders() {
		if bd.MaxBid <= 0 {
			t.Errorf("Bidder %d has MaxBid = %d; want > 0", bd.ClanID, bd.MaxBid)
		}
		if bd.CurBid <= 0 {
			t.Errorf("Bidder %d has CurBid = %d; want > 0", bd.ClanID, bd.CurBid)
		}
	}

	if got := a.BidderCount(); got == 0 {
		t.Error("BidderCount() = 0 after concurrent bids; want > 0")
	}
}
