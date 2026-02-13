package hall

import (
	"log/slog"
	"sync"
)

// HallInfo describes a static clan hall definition.
type HallInfo struct {
	ID          int32
	Name        string
	HallType    HallType
	Grade       Grade
	Location    string
	Lease       int64 // Weekly lease cost (auctionable only)
	StartingBid int64 // Starting auction bid (auctionable only)
}

// 38 auctionable halls (IDs 22-61, excluding 34, 35 which are siegable).
// 6 siegable halls (IDs 21, 34, 35, 62, 63, 64).
var knownHalls = []HallInfo{
	// Grade D — Gludio/Gludin barracks
	{ID: 22, Name: "Moonstone Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludio", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 23, Name: "Onyx Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludio", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 24, Name: "Topaz Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludio", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 25, Name: "Ruby Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludio", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 26, Name: "Crystal Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludin", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 27, Name: "Onyx Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludin", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 28, Name: "Sapphire Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludin", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 29, Name: "Moonstone Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludin", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 30, Name: "Emerald Hall", HallType: TypeAuctionable, Grade: GradeD, Location: "Gludin", Lease: 500_000, StartingBid: 20_000_000},

	// Grade C — Dion barracks
	{ID: 31, Name: "The Atramental Barracks", HallType: TypeAuctionable, Grade: GradeC, Location: "Dion", Lease: 200_000, StartingBid: 8_000_000},
	{ID: 32, Name: "The Scarlet Barracks", HallType: TypeAuctionable, Grade: GradeC, Location: "Dion", Lease: 200_000, StartingBid: 8_000_000},
	{ID: 33, Name: "The Viridian Barracks", HallType: TypeAuctionable, Grade: GradeC, Location: "Dion", Lease: 200_000, StartingBid: 8_000_000},

	// Grade B — Aden premium halls
	{ID: 36, Name: "The Golden Chamber", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 37, Name: "The Silver Chamber", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 38, Name: "The Mithril Chamber", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 39, Name: "Silver Manor", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 40, Name: "Gold Manor", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 41, Name: "The Bronze Chamber", HallType: TypeAuctionable, Grade: GradeB, Location: "Aden", Lease: 1_000_000, StartingBid: 50_000_000},

	// Grade B — Giran
	{ID: 42, Name: "Luna Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Giran", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 43, Name: "Titan Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Giran", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 44, Name: "Royal Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Giran", Lease: 1_000_000, StartingBid: 50_000_000},

	// Grade B — Goddard
	{ID: 45, Name: "Imperial Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Goddard", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 46, Name: "Western Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Goddard", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 47, Name: "Eastern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Goddard", Lease: 1_000_000, StartingBid: 50_000_000},

	// Grade B — Rune
	{ID: 48, Name: "Northern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Rune", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 49, Name: "Southern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Rune", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 50, Name: "Western Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Rune", Lease: 1_000_000, StartingBid: 50_000_000},

	// Grade B — Oren
	{ID: 51, Name: "Orbis Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 52, Name: "Tower of Honor", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 53, Name: "Hall of Victory", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 54, Name: "Eastern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 55, Name: "Western Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 56, Name: "Northern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},
	{ID: 57, Name: "Southern Hall", HallType: TypeAuctionable, Grade: GradeB, Location: "Oren", Lease: 1_000_000, StartingBid: 50_000_000},

	// Grade D — Schuttgart
	{ID: 58, Name: "Partisan Hideaway", HallType: TypeAuctionable, Grade: GradeD, Location: "Schuttgart", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 59, Name: "Orc Barracks", HallType: TypeAuctionable, Grade: GradeD, Location: "Schuttgart", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 60, Name: "Dwarf Barracks", HallType: TypeAuctionable, Grade: GradeD, Location: "Schuttgart", Lease: 500_000, StartingBid: 20_000_000},
	{ID: 61, Name: "Elite Barracks", HallType: TypeAuctionable, Grade: GradeD, Location: "Schuttgart", Lease: 500_000, StartingBid: 20_000_000},

	// Siegable halls
	{ID: 21, Name: "Fortress of Resistance", HallType: TypeSiegable, Grade: GradeNone, Location: "Dion"},
	{ID: 34, Name: "Devastated Castle", HallType: TypeSiegable, Grade: GradeNone, Location: "Aden"},
	{ID: 35, Name: "Bandit Stronghold", HallType: TypeSiegable, Grade: GradeNone, Location: "Oren"},
	{ID: 62, Name: "Rainbow Springs", HallType: TypeSiegable, Grade: GradeNone, Location: "Goddard"},
	{ID: 63, Name: "Beast Farm", HallType: TypeSiegable, Grade: GradeNone, Location: "Rune"},
	{ID: 64, Name: "Fortress of the Dead", HallType: TypeSiegable, Grade: GradeNone, Location: "Rune"},
}

// Table manages all clan halls and their auctions.
// Thread-safe: protected by mu.
type Table struct {
	mu    sync.RWMutex
	halls map[int32]*ClanHall // hallID → ClanHall

	auctionMu sync.RWMutex
	auctions  map[int32]*Auction // hallID → Auction (active auctions only)
}

// NewTable creates a clan hall table and initializes all known halls.
func NewTable() *Table {
	t := &Table{
		halls:    make(map[int32]*ClanHall, len(knownHalls)),
		auctions: make(map[int32]*Auction, 40),
	}
	for _, info := range knownHalls {
		h := NewClanHall(info.ID, info.Name, info.HallType, info.Grade, info.Location)
		h.SetLease(info.Lease)
		t.halls[info.ID] = h
	}
	slog.Info("clan hall table initialized", "halls", len(t.halls))
	return t
}

// Hall returns a clan hall by ID, or nil.
func (t *Table) Hall(id int32) *ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.halls[id]
}

// HallByOwner returns the clan hall owned by a clan, or nil.
func (t *Table) HallByOwner(clanID int32) *ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, h := range t.halls {
		if h.OwnerClanID() == clanID {
			return h
		}
	}
	return nil
}

// FreeHalls returns all unowned auctionable halls.
func (t *Table) FreeHalls() []*ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*ClanHall, 0, 20)
	for _, h := range t.halls {
		if h.Type() == TypeAuctionable && !h.HasOwner() {
			result = append(result, h)
		}
	}
	return result
}

// OwnedHalls returns all owned auctionable halls.
func (t *Table) OwnedHalls() []*ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*ClanHall, 0, 20)
	for _, h := range t.halls {
		if h.Type() == TypeAuctionable && h.HasOwner() {
			result = append(result, h)
		}
	}
	return result
}

// AllHalls returns a snapshot of all halls.
func (t *Table) AllHalls() []*ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*ClanHall, 0, len(t.halls))
	for _, h := range t.halls {
		result = append(result, h)
	}
	return result
}

// SiegableHalls returns all siegable halls.
func (t *Table) SiegableHalls() []*ClanHall {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*ClanHall, 0, 6)
	for _, h := range t.halls {
		if h.Type() == TypeSiegable {
			result = append(result, h)
		}
	}
	return result
}

// SetOwner assigns a hall to a clan with initial lease.
// Returns ErrHallNotFound if hall does not exist,
// ErrHallAlreadyOwned if already owned.
func (t *Table) SetOwner(hallID, clanID int32) error {
	h := t.Hall(hallID)
	if h == nil {
		return ErrHallNotFound
	}
	if h.HasOwner() {
		return ErrHallAlreadyOwned
	}
	h.SetOwner(clanID)
	slog.Info("clan hall assigned",
		"hall_id", hallID, "hall", h.Name(), "clan_id", clanID)
	return nil
}

// FreeHall releases a hall from its owner.
// Returns ErrHallNotFound if hall does not exist.
func (t *Table) FreeHall(hallID int32) error {
	h := t.Hall(hallID)
	if h == nil {
		return ErrHallNotFound
	}
	h.Free()
	return nil
}

// --- Auction management ---

// Auction returns an active auction for a hall, or nil.
func (t *Table) Auction(hallID int32) *Auction {
	t.auctionMu.RLock()
	defer t.auctionMu.RUnlock()
	return t.auctions[hallID]
}

// StartAuction creates an auction for a hall.
// Returns ErrHallNotFound if hall does not exist,
// ErrHallAlreadyOwned if already owned.
func (t *Table) StartAuction(a *Auction) error {
	h := t.Hall(a.HallID())
	if h == nil {
		return ErrHallNotFound
	}
	t.auctionMu.Lock()
	t.auctions[a.HallID()] = a
	t.auctionMu.Unlock()

	slog.Info("auction started",
		"hall_id", a.HallID(), "hall", h.Name(),
		"starting_bid", a.StartingBid())
	return nil
}

// RemoveAuction removes an auction.
func (t *Table) RemoveAuction(hallID int32) {
	t.auctionMu.Lock()
	delete(t.auctions, hallID)
	t.auctionMu.Unlock()
}

// ActiveAuctions returns a snapshot of all active auctions.
func (t *Table) ActiveAuctions() []*Auction {
	t.auctionMu.RLock()
	defer t.auctionMu.RUnlock()
	result := make([]*Auction, 0, len(t.auctions))
	for _, a := range t.auctions {
		result = append(result, a)
	}
	return result
}

// HallCount returns the total number of halls.
func (t *Table) HallCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.halls)
}

// GetHallInfo returns static info for a known hall by ID, or nil.
func GetHallInfo(id int32) *HallInfo {
	for i := range knownHalls {
		if knownHalls[i].ID == id {
			return &knownHalls[i]
		}
	}
	return nil
}
