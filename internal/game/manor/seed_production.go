package manor

import "sync/atomic"

// SeedProduction represents a single seed production entry for a castle.
// Amount is modified atomically during gameplay (players buy seeds).
type SeedProduction struct {
	seedID      int32
	amount      atomic.Int32
	price       int64
	startAmount int32
}

// NewSeedProduction creates a new seed production entry.
func NewSeedProduction(seedID int32, amount int32, price int64, startAmount int32) *SeedProduction {
	sp := &SeedProduction{
		seedID:      seedID,
		price:       price,
		startAmount: startAmount,
	}
	sp.amount.Store(amount)
	return sp
}

// SeedID returns the seed item ID.
func (sp *SeedProduction) SeedID() int32 { return sp.seedID }

// Amount returns the current remaining amount.
func (sp *SeedProduction) Amount() int32 { return sp.amount.Load() }

// Price returns the price per seed.
func (sp *SeedProduction) Price() int64 { return sp.price }

// StartAmount returns the initial amount for this period.
func (sp *SeedProduction) StartAmount() int32 { return sp.startAmount }

// SetAmount sets the remaining amount directly.
func (sp *SeedProduction) SetAmount(amount int32) { sp.amount.Store(amount) }

// DecreaseAmount atomically decreases the amount by val.
// Returns true if the amount was decreased successfully (CAS loop).
// Returns false if the current amount is less than val.
func (sp *SeedProduction) DecreaseAmount(val int32) bool {
	for {
		current := sp.amount.Load()
		if current < val {
			return false
		}
		if sp.amount.CompareAndSwap(current, current-val) {
			return true
		}
	}
}
