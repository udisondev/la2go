package manor

import "sync/atomic"

// CropProcure represents a single crop procurement entry for a castle.
// Amount is modified atomically during gameplay (players sell crops).
type CropProcure struct {
	cropID      int32
	amount      atomic.Int32
	price       int64
	startAmount int32
	rewardType  int32 // 1 = reward1, 2 = reward2
}

// NewCropProcure creates a new crop procurement entry.
func NewCropProcure(cropID int32, amount int32, rewardType int32, startAmount int32, price int64) *CropProcure {
	cp := &CropProcure{
		cropID:      cropID,
		price:       price,
		startAmount: startAmount,
		rewardType:  rewardType,
	}
	cp.amount.Store(amount)
	return cp
}

// CropID returns the crop item ID.
func (cp *CropProcure) CropID() int32 { return cp.cropID }

// Amount returns the current remaining amount.
func (cp *CropProcure) Amount() int32 { return cp.amount.Load() }

// Price returns the price per crop.
func (cp *CropProcure) Price() int64 { return cp.price }

// StartAmount returns the initial amount for this period.
func (cp *CropProcure) StartAmount() int32 { return cp.startAmount }

// RewardType returns the reward type (1 = reward1, 2 = reward2).
func (cp *CropProcure) RewardType() int32 { return cp.rewardType }

// SetAmount sets the remaining amount directly.
func (cp *CropProcure) SetAmount(amount int32) { cp.amount.Store(amount) }

// DecreaseAmount atomically decreases the amount by val.
// Returns true if the amount was decreased successfully (CAS loop).
// Returns false if the current amount is less than val.
func (cp *CropProcure) DecreaseAmount(val int32) bool {
	for {
		current := cp.amount.Load()
		if current < val {
			return false
		}
		if cp.amount.CompareAndSwap(current, current-val) {
			return true
		}
	}
}
