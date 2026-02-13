package hall

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Auction errors.
var (
	ErrAuctionNotFound    = errors.New("auction not found")
	ErrBidTooLow          = errors.New("bid too low")
	ErrAlreadyHighBidder  = errors.New("already highest bidder")
	ErrNoBid              = errors.New("no bid to cancel")
	ErrAuctionClosed      = errors.New("auction is closed")
	ErrNotHallOwner       = errors.New("not clan hall owner")
	ErrHallAlreadyOwned   = errors.New("clan hall already owned")
	ErrHallNotFound       = errors.New("clan hall not found")
	ErrClanAlreadyBidding = errors.New("clan already bidding on another hall")
)

// CancelFeeRate is the 10% fee for canceling a bid.
const CancelFeeRate = 10

// Bidder represents a clan's bid on a clan hall auction.
type Bidder struct {
	ClanID  int32
	BidAt   time.Time
	MaxBid  int64 // Highest bid placed
	CurBid  int64 // Current committed amount
}

// Auction represents an auction for an auctionable clan hall.
// Thread-safe.
type Auction struct {
	mu sync.RWMutex

	id         int32 // Typically matches hall ID
	hallID     int32
	startingBid int64
	endDate    time.Time

	bidders          map[int32]*Bidder // clanID → Bidder
	highestBidderID  int32
	highestBidAmount int64
}

// NewAuction creates a new auction for a hall.
func NewAuction(id, hallID int32, startingBid int64, endDate time.Time) *Auction {
	return &Auction{
		id:          id,
		hallID:      hallID,
		startingBid: startingBid,
		endDate:     endDate,
		bidders:     make(map[int32]*Bidder, 8),
	}
}

// ID returns the auction ID.
func (a *Auction) ID() int32 { return a.id }

// HallID returns the hall being auctioned.
func (a *Auction) HallID() int32 { return a.hallID }

// StartingBid returns the minimum bid.
func (a *Auction) StartingBid() int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.startingBid
}

// EndDate returns when the auction ends.
func (a *Auction) EndDate() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.endDate
}

// IsActive returns true if the auction is still open.
func (a *Auction) IsActive() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.endDate.After(time.Now())
}

// HighestBidder returns the current highest bidder info.
func (a *Auction) HighestBidder() (clanID int32, amount int64) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.highestBidderID, a.highestBidAmount
}

// BidderCount returns the number of bidders.
func (a *Auction) BidderCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.bidders)
}

// Bidder returns a bidder by clan ID, or nil.
func (a *Auction) Bidder(clanID int32) *Bidder {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.bidders[clanID]
}

// Bidders returns a snapshot of all bidders.
func (a *Auction) Bidders() []*Bidder {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]*Bidder, 0, len(a.bidders))
	for _, b := range a.bidders {
		result = append(result, b)
	}
	return result
}

// PlaceBid places or updates a bid. Returns the additional amount to deduct from clan warehouse.
func (a *Auction) PlaceBid(clanID int32, bidAmount int64) (deductAmount int64, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.endDate.Before(time.Now()) {
		return 0, ErrAuctionClosed
	}

	minBid := a.startingBid
	if a.highestBidAmount > 0 {
		minBid = a.highestBidAmount + 1
	}
	if bidAmount < minBid {
		return 0, ErrBidTooLow
	}

	existing, ok := a.bidders[clanID]
	if ok {
		// Доплата разницы.
		deductAmount = bidAmount - existing.MaxBid
		if deductAmount <= 0 {
			return 0, ErrBidTooLow
		}
		existing.MaxBid = bidAmount
		existing.CurBid = bidAmount
		existing.BidAt = time.Now()
	} else {
		deductAmount = bidAmount
		a.bidders[clanID] = &Bidder{
			ClanID: clanID,
			BidAt:  time.Now(),
			MaxBid: bidAmount,
			CurBid: bidAmount,
		}
	}

	if bidAmount > a.highestBidAmount {
		a.highestBidderID = clanID
		a.highestBidAmount = bidAmount
	}

	slog.Info("auction bid placed",
		"auction_id", a.id, "hall_id", a.hallID,
		"clan_id", clanID, "bid", bidAmount)

	return deductAmount, nil
}

// CancelBid cancels a bid and returns the refund amount (minus 10% fee).
func (a *Auction) CancelBid(clanID int32) (refund int64, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.bidders[clanID]
	if !ok {
		return 0, ErrNoBid
	}

	refund = b.CurBid - (b.CurBid * CancelFeeRate / 100)
	delete(a.bidders, clanID)

	// Пересчитываем лидера.
	a.highestBidderID = 0
	a.highestBidAmount = 0
	for _, bid := range a.bidders {
		if bid.MaxBid > a.highestBidAmount {
			a.highestBidderID = bid.ClanID
			a.highestBidAmount = bid.MaxBid
		}
	}

	slog.Info("auction bid canceled",
		"auction_id", a.id, "hall_id", a.hallID,
		"clan_id", clanID, "refund", refund)

	return refund, nil
}

// EndAuction finalizes the auction. Returns winner clan ID and bid (0 if no winner).
func (a *Auction) EndAuction() (winnerClanID int32, winningBid int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.highestBidderID == 0 {
		slog.Info("auction ended with no bids",
			"auction_id", a.id, "hall_id", a.hallID)
		return 0, 0
	}

	winnerClanID = a.highestBidderID
	winningBid = a.highestBidAmount

	// Возвращаем проигравшим их ставки (будет обработано вызывающим кодом).
	slog.Info("auction ended",
		"auction_id", a.id, "hall_id", a.hallID,
		"winner_clan", winnerClanID, "bid", winningBid)

	return winnerClanID, winningBid
}

// LoserBids returns bids that should be refunded (all except winner).
func (a *Auction) LoserBids() []*Bidder {
	a.mu.RLock()
	defer a.mu.RUnlock()

	losers := make([]*Bidder, 0, len(a.bidders))
	for _, b := range a.bidders {
		if b.ClanID != a.highestBidderID {
			losers = append(losers, b)
		}
	}
	return losers
}
