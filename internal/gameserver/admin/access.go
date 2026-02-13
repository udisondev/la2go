// Package admin provides admin and user command handling for game server.
// Phase 17: Admin Commands System.
//
// Java reference: AdminData.java, AccessLevel.java, AdminCommandHandler.java
package admin

// AccessLevel defines a GM access level with associated permissions.
// Level 0 = normal player, 1+ = GM, 100+ = full admin.
//
// Java reference: AccessLevel.java
type AccessLevel struct {
	Level               int32
	Name                string
	IsGM                bool
	AllowPeaceAttack    bool
	AllowFixedRes       bool
	AllowTransaction    bool
	AllowAltG           bool
	CanBan              bool
	CanUseAdminCommands bool
	GiveDamage          bool
	TakeAggro           bool
	GainExp             bool
}

// DefaultAccessLevels returns the predefined access levels.
// Java reference: AccessLevels.xml
var defaultAccessLevels = map[int32]*AccessLevel{
	0: {
		Level:               0,
		Name:                "User",
		IsGM:                false,
		AllowPeaceAttack:    false,
		AllowFixedRes:       false,
		AllowTransaction:    true,
		AllowAltG:           false,
		CanBan:              false,
		CanUseAdminCommands: false,
		GiveDamage:          true,
		TakeAggro:           true,
		GainExp:             true,
	},
	1: {
		Level:               1,
		Name:                "Moderator",
		IsGM:                true,
		AllowPeaceAttack:    false,
		AllowFixedRes:       true,
		AllowTransaction:    true,
		AllowAltG:           true,
		CanBan:              true,
		CanUseAdminCommands: true,
		GiveDamage:          true,
		TakeAggro:           false,
		GainExp:             false,
	},
	2: {
		Level:               2,
		Name:                "Game Master",
		IsGM:                true,
		AllowPeaceAttack:    true,
		AllowFixedRes:       true,
		AllowTransaction:    true,
		AllowAltG:           true,
		CanBan:              true,
		CanUseAdminCommands: true,
		GiveDamage:          true,
		TakeAggro:           false,
		GainExp:             false,
	},
	100: {
		Level:               100,
		Name:                "Administrator",
		IsGM:                true,
		AllowPeaceAttack:    true,
		AllowFixedRes:       true,
		AllowTransaction:    true,
		AllowAltG:           true,
		CanBan:              true,
		CanUseAdminCommands: true,
		GiveDamage:          true,
		TakeAggro:           false,
		GainExp:             false,
	},
}

// GetAccessLevel returns AccessLevel for the given level value.
// Unknown levels inherit from the highest matching known level below them.
// Negative levels (banned) return nil.
func GetAccessLevel(level int32) *AccessLevel {
	if level < 0 {
		return nil
	}

	if al, ok := defaultAccessLevels[level]; ok {
		return al
	}

	// For unknown positive levels, find the highest known level <= given level.
	var best *AccessLevel
	for _, al := range defaultAccessLevels {
		if al.Level <= level && (best == nil || al.Level > best.Level) {
			best = al
		}
	}
	return best
}
