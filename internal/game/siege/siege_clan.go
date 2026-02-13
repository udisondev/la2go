package siege

// SiegeClanType defines the role of a clan in a siege.
type SiegeClanType int32

// Clan types in a siege (matches Java/DB convention).
const (
	ClanTypeOwner              SiegeClanType = -1 // Castle owner (auto-defender)
	ClanTypeDefender           SiegeClanType = 0  // Registered defender
	ClanTypeAttacker           SiegeClanType = 1  // Registered attacker
	ClanTypeDefenderNotApproved SiegeClanType = 2  // Defender pending approval
)

// SiegeClan represents a clan participating in a siege.
type SiegeClan struct {
	ClanID   int32
	ClanName string
	Type     SiegeClanType
	// Флаг атакующего клана (ObjectID в мире, 0 если нет).
	FlagObjectID int32
}

// NewSiegeClan creates a siege clan entry.
func NewSiegeClan(clanID int32, clanName string, clanType SiegeClanType) *SiegeClan {
	return &SiegeClan{
		ClanID:   clanID,
		ClanName: clanName,
		Type:     clanType,
	}
}

// IsAttacker returns true if the clan is an attacker.
func (sc *SiegeClan) IsAttacker() bool {
	return sc.Type == ClanTypeAttacker
}

// IsDefender returns true if the clan is a defender (approved or owner).
func (sc *SiegeClan) IsDefender() bool {
	return sc.Type == ClanTypeDefender || sc.Type == ClanTypeOwner
}

// IsPending returns true if the clan is a pending (unapproved) defender.
func (sc *SiegeClan) IsPending() bool {
	return sc.Type == ClanTypeDefenderNotApproved
}
