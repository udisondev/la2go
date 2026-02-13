package model

import "time"

// Couple represents a marriage/engagement between two players.
// Phase 33: Marriage System.
type Couple struct {
	ID          int32
	Player1ID   int32 // always min(p1, p2) for consistency
	Player2ID   int32 // always max(p1, p2)
	Married     bool
	AffiancedAt time.Time
	MarriedAt   *time.Time // nil until ceremony completes
}

// Contains returns true if the given objectID is one of the couple members.
func (c *Couple) Contains(objectID int32) bool {
	return c.Player1ID == objectID || c.Player2ID == objectID
}

// PartnerOf returns the ObjectID of the other member, or 0 if not found.
func (c *Couple) PartnerOf(objectID int32) int32 {
	if c.Player1ID == objectID {
		return c.Player2ID
	}
	if c.Player2ID == objectID {
		return c.Player1ID
	}
	return 0
}
