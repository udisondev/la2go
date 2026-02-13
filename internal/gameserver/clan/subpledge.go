package clan

// SubPledge type constants matching L2J Interlude.
const (
	PledgeMain    int32 = 0    // Main clan
	PledgeAcademy int32 = -1   // Clan Academy (level 5+)
	PledgeRoyal1  int32 = 100  // Royal Guard 1 (level 6+)
	PledgeRoyal2  int32 = 200  // Royal Guard 2 (level 6+)
	PledgeKnight1 int32 = 1001 // Order of Knights 1 (level 7+)
	PledgeKnight2 int32 = 1002 // Order of Knights 2 (level 7+)
	PledgeKnight3 int32 = 2001 // Order of Knights 3 (level 8)
	PledgeKnight4 int32 = 2002 // Order of Knights 4 (level 8)
)

// SubPledge represents a sub-pledge (regiment) within a clan.
type SubPledge struct {
	ID       int32  // PledgeAcademy, PledgeRoyal1, etc.
	Name     string // Display name
	LeaderID int64  // Object ID of the sub-pledge leader (0 if none)
}

// MaxSubPledgeMembers returns the max member count for a sub-pledge type.
func MaxSubPledgeMembers(pledgeType int32) int32 {
	switch pledgeType {
	case PledgeAcademy:
		return 20
	case PledgeRoyal1, PledgeRoyal2:
		return 20
	case PledgeKnight1, PledgeKnight2, PledgeKnight3, PledgeKnight4:
		return 10
	default:
		return 0
	}
}

// IsSubPledge returns true if the type represents a sub-pledge (not main).
func IsSubPledge(pledgeType int32) bool {
	switch pledgeType {
	case PledgeAcademy, PledgeRoyal1, PledgeRoyal2,
		PledgeKnight1, PledgeKnight2, PledgeKnight3, PledgeKnight4:
		return true
	default:
		return false
	}
}

// SubPledgeRequiredLevel returns the minimum clan level to create a sub-pledge.
func SubPledgeRequiredLevel(pledgeType int32) int32 {
	switch pledgeType {
	case PledgeAcademy:
		return 5
	case PledgeRoyal1, PledgeRoyal2:
		return 6
	case PledgeKnight1, PledgeKnight2:
		return 7
	case PledgeKnight3, PledgeKnight4:
		return 8
	default:
		return 0
	}
}
