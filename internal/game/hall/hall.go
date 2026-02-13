package hall

import (
	"log/slog"
	"sync"
	"time"
)

// HallType distinguishes auctionable from siegable clan halls.
type HallType int32

const (
	TypeAuctionable HallType = 0 // Won via auction, weekly lease
	TypeSiegable    HallType = 1 // Won via clan hall siege
)

// Grade represents the clan hall grade (affects available functions).
type Grade int32

const (
	GradeNone Grade = -1
	GradeS    Grade = 0
	GradeA    Grade = 1
	GradeB    Grade = 2
	GradeC    Grade = 3
	GradeD    Grade = 4
)

// FeeRate is the weekly clan hall maintenance period (7 days).
const FeeRate = 7 * 24 * time.Hour

// ClanHall represents a clan hall that can be owned by a clan.
// Thread-safe: all mutable fields protected by mu.
type ClanHall struct {
	mu sync.RWMutex

	id          int32
	name        string
	hallType    HallType
	ownerClanID int32 // 0 = no owner
	grade       Grade
	location    string // Zone / area name

	// Аукционный клан-холл.
	lease     int64     // Weekly lease cost
	paidUntil time.Time // Lease expiration

	// Осаждаемый клан-холл.
	nextSiege   time.Time
	siegeLength time.Duration

	// Функции клан-холла (type → Function).
	functions map[FunctionType]*Function

	// Описание (для NPC-диалогов).
	description string

	// Координаты зоны.
	zoneID int32
}

// NewClanHall creates a new clan hall.
func NewClanHall(id int32, name string, hallType HallType, grade Grade, location string) *ClanHall {
	return &ClanHall{
		id:        id,
		name:      name,
		hallType:  hallType,
		grade:     grade,
		location:  location,
		functions: make(map[FunctionType]*Function, 6),
	}
}

// ID returns the clan hall ID.
func (h *ClanHall) ID() int32 { return h.id }

// Name returns the clan hall name.
func (h *ClanHall) Name() string { return h.name }

// Type returns the hall type (auctionable or siegable).
func (h *ClanHall) Type() HallType { return h.hallType }

// Grade returns the hall grade.
func (h *ClanHall) Grade() Grade { return h.grade }

// Location returns the hall location description.
func (h *ClanHall) Location() string { return h.location }

// OwnerClanID returns the owning clan ID (0 if unowned).
func (h *ClanHall) OwnerClanID() int32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ownerClanID
}

// SetOwnerClanID sets the owner clan ID.
func (h *ClanHall) SetOwnerClanID(clanID int32) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ownerClanID = clanID
}

// HasOwner returns true if the hall has an owner.
func (h *ClanHall) HasOwner() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ownerClanID > 0
}

// Lease returns the weekly lease cost.
func (h *ClanHall) Lease() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lease
}

// SetLease sets the weekly lease cost.
func (h *ClanHall) SetLease(amount int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lease = amount
}

// PaidUntil returns when the lease expires.
func (h *ClanHall) PaidUntil() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.paidUntil
}

// SetPaidUntil sets the lease expiration.
func (h *ClanHall) SetPaidUntil(t time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paidUntil = t
}

// IsLeaseExpired returns true if the lease has expired.
func (h *ClanHall) IsLeaseExpired() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ownerClanID > 0 && h.paidUntil.Before(time.Now())
}

// NextSiege returns the next siege date (for siegable halls).
func (h *ClanHall) NextSiege() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.nextSiege
}

// SetNextSiege sets the next siege date.
func (h *ClanHall) SetNextSiege(t time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextSiege = t
}

// SiegeLength returns the siege duration.
func (h *ClanHall) SiegeLength() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.siegeLength
}

// SetSiegeLength sets the siege duration.
func (h *ClanHall) SetSiegeLength(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.siegeLength = d
}

// ZoneID returns the associated zone ID.
func (h *ClanHall) ZoneID() int32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.zoneID
}

// SetZoneID sets the zone ID.
func (h *ClanHall) SetZoneID(id int32) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.zoneID = id
}

// Description returns the hall description.
func (h *ClanHall) Description() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.description
}

// SetDescription sets the description.
func (h *ClanHall) SetDescription(desc string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.description = desc
}

// --- Functions ---

// SetFunction adds or replaces a function.
func (h *ClanHall) SetFunction(f *Function) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.functions[f.Type] = f
}

// RemoveFunction removes a function by type.
func (h *ClanHall) RemoveFunction(ft FunctionType) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.functions, ft)
}

// Function returns a function by type, or nil.
func (h *ClanHall) Function(ft FunctionType) *Function {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.functions[ft]
}

// Functions returns a snapshot of all active functions.
func (h *ClanHall) Functions() []*Function {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]*Function, 0, len(h.functions))
	for _, f := range h.functions {
		result = append(result, f)
	}
	return result
}

// FunctionLevel returns the level of a function, or 0 if not present.
func (h *ClanHall) FunctionLevel(ft FunctionType) int32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	f, ok := h.functions[ft]
	if !ok {
		return 0
	}
	return f.Level
}

// --- Ownership transfer ---

// Free releases the clan hall from its owner, removing all functions.
func (h *ClanHall) Free() {
	h.mu.Lock()
	defer h.mu.Unlock()

	prevOwner := h.ownerClanID
	h.ownerClanID = 0
	h.paidUntil = time.Time{}
	h.functions = make(map[FunctionType]*Function, 6)

	if prevOwner > 0 {
		slog.Info("clan hall freed", "hall_id", h.id, "prev_owner", prevOwner)
	}
}

// SetOwner assigns the hall to a clan.
func (h *ClanHall) SetOwner(clanID int32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ownerClanID = clanID
	h.paidUntil = time.Now().Add(FeeRate)
	h.functions = make(map[FunctionType]*Function, 6)

	slog.Info("clan hall owner set", "hall_id", h.id, "clan_id", clanID)
}
