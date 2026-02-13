// Package augment implements the weapon augmentation (variation) system.
//
// Phase 28: Augmentation System.
// Java reference: Augmentation.java, AbstractRefinePacket.java, RequestRefine.java.
package augment

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// Life Stone item IDs.
const (
	LifeStoneMin int32 = 8723 // Life Stone: level 46
	LifeStoneMax int32 = 8762 // Top-Grade Life Stone: level 76
)

// Life Stone grades.
const (
	GradeNone int32 = 0
	GradeMid  int32 = 1
	GradeHigh int32 = 2
	GradeTop  int32 = 3
)

// Gemstone item IDs required for augmentation.
// Java reference: AbstractRefinePacket.java — GEMSTONE_D, GEMSTONE_C.
const (
	GemstoneD int32 = 2130 // Gemstone D — used for C/B grade weapons
	GemstoneC int32 = 2131 // Gemstone C — used for A/S grade weapons
)

// Errors.
var (
	ErrNotWeapon        = errors.New("item is not a weapon")
	ErrAlreadyAugmented = errors.New("weapon already augmented")
	ErrNotAugmented     = errors.New("weapon is not augmented")
	ErrInvalidLifeStone = errors.New("invalid life stone")
	ErrWeaponEquipped   = errors.New("weapon must not be equipped")
)

// Service handles augmentation operations.
type Service struct {
	mu sync.Mutex
}

// NewService creates a new augmentation service.
func NewService() *Service {
	return &Service{}
}

// ValidateTarget checks if item is valid for augmentation.
func (s *Service) ValidateTarget(item *model.Item) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}
	if !item.IsWeapon() {
		return ErrNotWeapon
	}
	if item.IsEquipped() {
		return ErrWeaponEquipped
	}
	if item.AugmentationID() != 0 {
		return ErrAlreadyAugmented
	}
	return nil
}

// ValidateRefiner checks if life stone item is valid.
func (s *Service) ValidateRefiner(item *model.Item) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}
	if !IsLifeStone(item.ItemID()) {
		return ErrInvalidLifeStone
	}
	return nil
}

// GemstoneRequirement returns (gemItemID, gemCount) for augmentation based on weapon crystal grade.
// Java reference: AbstractRefinePacket.getGemStoneId / getGemStoneCount.
func GemstoneRequirement(grade model.CrystalType) (int32, int32) {
	switch grade {
	case model.CrystalC:
		return GemstoneD, 20
	case model.CrystalB:
		return GemstoneD, 30
	case model.CrystalA:
		return GemstoneC, 20
	case model.CrystalS:
		return GemstoneC, 25
	default:
		return 0, 0
	}
}

// LifeStoneGrade returns the grade of a life stone (0-3).
func LifeStoneGrade(itemID int32) int32 {
	if itemID < LifeStoneMin || itemID > LifeStoneMax {
		return -1
	}
	offset := itemID - LifeStoneMin
	return offset / 10 // 0-9=None, 10-19=Mid, 20-29=High, 30-39=Top
}

// LifeStoneLevel returns the level tier of a life stone (0-9 within grade).
func LifeStoneLevel(itemID int32) int32 {
	if itemID < LifeStoneMin || itemID > LifeStoneMax {
		return -1
	}
	offset := itemID - LifeStoneMin
	return offset % 10
}

// IsLifeStone returns true if itemID is a life stone.
func IsLifeStone(itemID int32) bool {
	return itemID >= LifeStoneMin && itemID <= LifeStoneMax
}

// Augment performs the augmentation, generating a random augmentation ID.
// Returns the augmentationID on success.
func (s *Service) Augment(weapon *model.Item, lifeStoneID int32) (int32, error) {
	if err := s.ValidateTarget(weapon); err != nil {
		return 0, fmt.Errorf("validate target: %w", err)
	}

	grade := LifeStoneGrade(lifeStoneID)
	if grade < 0 {
		return 0, ErrInvalidLifeStone
	}

	augID := generateAugmentID(grade)
	weapon.SetAugmentationID(augID)
	return augID, nil
}

// RemoveAugmentation removes augmentation from a weapon.
// Returns the old augmentationID.
func (s *Service) RemoveAugmentation(weapon *model.Item) (int32, error) {
	if weapon == nil {
		return 0, fmt.Errorf("weapon is nil")
	}
	if !weapon.IsWeapon() {
		return 0, ErrNotWeapon
	}

	oldID := weapon.AugmentationID()
	if oldID == 0 {
		return 0, ErrNotAugmented
	}

	weapon.SetAugmentationID(0)
	return oldID, nil
}

// generateAugmentID generates a random augmentation ID based on life stone grade.
//
// ID ranges by color (from Java AugmentationData):
//   - Blue (stat): 1-14440
//   - Purple (skill): 14441-24440
//   - Yellow (skill): 24441-33440
//   - Red (skill): 33441-38440
//
// Higher grade life stones give better chances of skill augmentations.
func generateAugmentID(grade int32) int32 {
	const (
		blueMax   = 14440
		purpleMax = 24440
		yellowMax = 33440
		redMax    = 38440
	)

	roll := cryptoRandInt(100)
	var augRange int32

	switch {
	case grade >= 3 && roll < 35:
		colorRoll := cryptoRandInt(100)
		switch {
		case colorRoll < 10:
			augRange = redMax
		case colorRoll < 40:
			augRange = yellowMax
		default:
			augRange = purpleMax
		}
	case grade >= 2 && roll < 20:
		colorRoll := cryptoRandInt(100)
		if colorRoll < 20 {
			augRange = yellowMax
		} else {
			augRange = purpleMax
		}
	case grade >= 1 && roll < 10:
		augRange = purpleMax
	default:
		augRange = blueMax
	}

	return cryptoRandInt(augRange) + 1
}

// cryptoRandInt returns a cryptographically random int in [0, max).
func cryptoRandInt(max int32) int32 {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int32(n.Int64())
}
