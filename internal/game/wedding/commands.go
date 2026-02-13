package wedding

import "github.com/udisondev/la2go/internal/model"

// EngageRequest represents a pending engagement proposal.
type EngageRequest struct {
	FromObjectID int32
	ToObjectID   int32
}

// CanEngage checks whether two players can get engaged.
// Returns an error describing why they cannot, or nil.
func CanEngage(requester, target *model.Player, mgr *Manager) error {
	reqID := int32(requester.ObjectID())
	tgtID := int32(target.ObjectID())

	if reqID == tgtID {
		return ErrSelfEngage
	}

	if mgr.CoupleByPlayer(reqID) != nil {
		return ErrAlreadyEngaged
	}
	if mgr.CoupleByPlayer(tgtID) != nil {
		return ErrAlreadyEngaged
	}

	if target.IsEngageRequest() {
		return ErrAlreadyEngaged
	}

	return nil
}

// CanDivorce checks whether a player can divorce.
// Returns the couple or an error.
func CanDivorce(player *model.Player, mgr *Manager) (*model.Couple, error) {
	objectID := int32(player.ObjectID())

	c := mgr.CoupleByPlayer(objectID)
	if c == nil {
		return nil, ErrNotEngaged
	}

	return c, nil
}

// CalcDivorcePenalty calculates the adena amount transferred on divorce.
// Returns the amount deducted from the divorcing player's wallet.
// In Java: 20% of the other partner's adena, minimum 0.
func CalcDivorcePenalty(partnerAdena int64) int64 {
	penalty := partnerAdena * DivorcePenaltyPct / 100
	if penalty < 0 {
		return 0
	}
	return penalty
}

// CanTeleportToPartner checks whether a player can use .gotolove.
// Returns the partner ObjectID or an error.
func CanTeleportToPartner(player *model.Player, mgr *Manager) (int32, error) {
	objectID := int32(player.ObjectID())

	c := mgr.CoupleByPlayer(objectID)
	if c == nil {
		return 0, ErrNotEngaged
	}

	if !c.Married {
		return 0, ErrNotMarried
	}

	partnerID := c.PartnerOf(objectID)
	if partnerID == 0 {
		return 0, ErrCoupleNotFound
	}

	return partnerID, nil
}
