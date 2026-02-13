package siege

import (
	"sync"
	"sync/atomic"
	"time"
)

// TaxRate limits.
const (
	MinTaxRate = 0
	MaxTaxRate = 25 // Interlude max tax rate 25%
)

// Castle represents a castle that can be sieged and owned by a clan.
// Thread-safe: all mutable fields protected by mu.
type Castle struct {
	mu sync.RWMutex

	id              int32
	name            string
	ownerClanID     int32 // 0 = no owner
	taxRate         int32 // 0-25%
	treasury        int64
	siegeDate       time.Time // Next siege date
	timeRegOver     bool      // Whether siege time registration is complete
	maxMercenaries  int32     // Max mercenaries for this castle

	// Активная осада (nil если осады нет).
	siege atomic.Pointer[Siege]
}

// NewCastle creates a castle with the given properties.
func NewCastle(id int32, name string, maxMerc int32) *Castle {
	return &Castle{
		id:             id,
		name:           name,
		maxMercenaries: maxMerc,
	}
}

// ID returns the castle ID.
func (c *Castle) ID() int32 { return c.id }

// Name returns the castle name.
func (c *Castle) Name() string { return c.name }

// OwnerClanID returns the clan ID that owns this castle (0 = no owner).
func (c *Castle) OwnerClanID() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ownerClanID
}

// SetOwnerClanID sets the owning clan ID.
func (c *Castle) SetOwnerClanID(clanID int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ownerClanID = clanID
}

// TaxRate returns the current tax rate (0-25).
func (c *Castle) TaxRate() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.taxRate
}

// SetTaxRate sets the tax rate (clamped to 0-25).
func (c *Castle) SetTaxRate(rate int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if rate < MinTaxRate {
		rate = MinTaxRate
	}
	if rate > MaxTaxRate {
		rate = MaxTaxRate
	}
	c.taxRate = rate
}

// Treasury returns the current treasury amount.
func (c *Castle) Treasury() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.treasury
}

// AddToTreasury adds adena to the treasury (can be negative).
func (c *Castle) AddToTreasury(amount int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.treasury += amount
	if c.treasury < 0 {
		c.treasury = 0
	}
	return c.treasury
}

// SetTreasury sets the treasury value directly.
func (c *Castle) SetTreasury(amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.treasury = amount
}

// SiegeDate returns the next scheduled siege date.
func (c *Castle) SiegeDate() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.siegeDate
}

// SetSiegeDate sets the next siege date.
func (c *Castle) SetSiegeDate(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.siegeDate = t
}

// IsTimeRegistrationOver returns whether time registration is complete.
func (c *Castle) IsTimeRegistrationOver() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeRegOver
}

// SetTimeRegistrationOver marks time registration as complete/incomplete.
func (c *Castle) SetTimeRegistrationOver(over bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeRegOver = over
}

// MaxMercenaries returns the max mercenary count for this castle.
func (c *Castle) MaxMercenaries() int32 { return c.maxMercenaries }

// Siege returns the active siege for this castle, or nil.
func (c *Castle) Siege() *Siege {
	return c.siege.Load()
}

// SetSiege sets or clears the active siege.
func (c *Castle) SetSiege(s *Siege) {
	c.siege.Store(s)
}

// HasOwner returns true if the castle has an owner clan.
func (c *Castle) HasOwner() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ownerClanID > 0
}

// TaxAmount calculates tax for the given price.
func (c *Castle) TaxAmount(price int64) int64 {
	c.mu.RLock()
	rate := c.taxRate
	c.mu.RUnlock()
	if rate <= 0 {
		return 0
	}
	return price * int64(rate) / 100
}
