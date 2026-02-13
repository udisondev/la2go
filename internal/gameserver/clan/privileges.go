package clan

// Privilege represents a single clan privilege bitflag.
type Privilege int32

// Clan privilege constants — bit positions match Java ClanAccess enum ordinals.
// Each privilege is 1 << ordinal. Client interprets bitmask by these positions.
// Java reference: ClanAccess.java (ordinal 0-23)
const (
	PrivNone              Privilege = 0
	PrivCLJoinClan        Privilege = 1 << 1  // INVITE_MEMBER (ordinal 1)
	PrivCLGiveTitles      Privilege = 1 << 2  // ASSIGN_TITLE (ordinal 2)
	PrivCLViewWarehouse   Privilege = 1 << 3  // ACCESS_WAREHOUSE (ordinal 3)
	PrivCLManageRanks     Privilege = 1 << 4  // MODIFY_RANKS (ordinal 4)
	PrivCLPledgeWar       Privilege = 1 << 5  // WAR_DECLARATION (ordinal 5)
	PrivCLDismiss         Privilege = 1 << 6  // REMOVE_MEMBER (ordinal 6)
	PrivCLRegisterCrest   Privilege = 1 << 7  // CHANGE_CREST (ordinal 7)
	PrivCLApprentice      Privilege = 1 << 8  // DISMISS_MENTEE (ordinal 8)
	PrivCLMemberFame      Privilege = 1 << 9  // MEMBER_FAME (ordinal 9)
	PrivCLAccessAirship   Privilege = 1 << 10 // ACCESS_AIRSHIP (ordinal 10)
	PrivCHOpenDoor        Privilege = 1 << 11 // HALL_OPEN_DOOR (ordinal 11)
	PrivCHSetFunctions    Privilege = 1 << 12 // HALL_FUNCTIONS (ordinal 12)
	PrivCHAuction         Privilege = 1 << 13 // HALL_AUCTION (ordinal 13)
	PrivCHDismiss         Privilege = 1 << 14 // HALL_BANISH (ordinal 14)
	PrivCHManageFunctions Privilege = 1 << 15 // HALL_MANAGE_FUNCTIONS (ordinal 15)
	PrivCSOpenDoor        Privilege = 1 << 16 // CASTLE_OPEN_DOOR (ordinal 16)
	PrivCSManorAdmin      Privilege = 1 << 17 // CASTLE_MANOR (ordinal 17)
	PrivCSManageSiege     Privilege = 1 << 18 // CASTLE_SIEGE (ordinal 18)
	PrivCSUseFunctions    Privilege = 1 << 19 // CASTLE_FUNCTIONS (ordinal 19)
	PrivCSDismiss         Privilege = 1 << 20 // CASTLE_BANISH (ordinal 20)
	PrivCSVault           Privilege = 1 << 21 // CASTLE_VAULT (ordinal 21)
	PrivCSMercenaries     Privilege = 1 << 22 // CASTLE_MERCENARIES (ordinal 22)
	PrivCSManageFunctions Privilege = 1 << 23 // CASTLE_MANAGE_FUNCTIONS (ordinal 23)

	// PrivAll combines all privileges (24 bits, ordinals 0-23).
	PrivAll Privilege = (1 << 24) - 1
)

// DefaultRankPrivileges returns the default privilege mask for a given power grade.
// Power grade 1 = leader (all), higher grades = fewer privileges.
func DefaultRankPrivileges(powerGrade int32) Privilege {
	switch powerGrade {
	case 1:
		return PrivAll
	case 2:
		return PrivCLJoinClan | PrivCLGiveTitles | PrivCLViewWarehouse |
			PrivCLManageRanks | PrivCLPledgeWar | PrivCLDismiss |
			PrivCLRegisterCrest
	case 3:
		return PrivCLJoinClan | PrivCLGiveTitles | PrivCLViewWarehouse |
			PrivCLDismiss
	case 4:
		return PrivCLViewWarehouse
	default:
		return PrivNone
	}
}

// Has checks if the privilege mask contains the given privilege.
func (p Privilege) Has(priv Privilege) bool {
	return p&priv == priv
}

// Add adds a privilege to the mask.
func (p Privilege) Add(priv Privilege) Privilege {
	return p | priv
}

// Remove removes a privilege from the mask.
func (p Privilege) Remove(priv Privilege) Privilege {
	return p &^ priv
}
