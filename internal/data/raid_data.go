package data

// IsRaidBoss returns true if NPC with given templateID is a raid boss.
func IsRaidBoss(templateID int32) bool {
	def := GetNpcDef(templateID)
	if def == nil {
		return false
	}
	return def.npcType == "raid_boss"
}

// IsGrandBoss returns true if NPC with given templateID is a grand boss.
func IsGrandBoss(templateID int32) bool {
	def := GetNpcDef(templateID)
	if def == nil {
		return false
	}
	return def.npcType == "grand_boss"
}

// IsAnyBoss returns true if NPC is either a raid boss or grand boss.
func IsAnyBoss(templateID int32) bool {
	def := GetNpcDef(templateID)
	if def == nil {
		return false
	}
	return def.npcType == "raid_boss" || def.npcType == "grand_boss"
}

// GetAllRaidBossIDs returns template IDs of all raid bosses in NpcTable.
func GetAllRaidBossIDs() []int32 {
	if NpcTable == nil {
		return nil
	}
	ids := make([]int32, 0, 256)
	for id, def := range NpcTable {
		if def.npcType == "raid_boss" {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetAllGrandBossIDs returns template IDs of all grand bosses in NpcTable.
func GetAllGrandBossIDs() []int32 {
	if NpcTable == nil {
		return nil
	}
	ids := make([]int32, 0, 16)
	for id, def := range NpcTable {
		if def.npcType == "grand_boss" {
			ids = append(ids, id)
		}
	}
	return ids
}

// RaidBoss respawn constants (seconds).
const (
	RaidRespawnMinDefault = 43200  // 12 hours
	RaidRespawnMaxDefault = 86400  // 24 hours
	RaidRespawnRandBase   = 43200  // random addition base (12h)

	// Grand boss respawn: typically much longer (managed per-boss by GrandBossManager)
	GrandBossRespawnDefault = 172800 // 48 hours
)

// Raid boss point constants.
const (
	RaidPointsPerLevel  = 1   // base points per raid boss level
	RaidPointsMinLevel  = 20  // minimum raid boss level for points
	RaidPointsWeekReset = 7   // days between point resets (Monday)
)
