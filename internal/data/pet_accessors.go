package data

// PetInfo — exported view of pet template data.
// Phase 19: Pets/Summons System.
type PetInfo struct {
	NpcID  int32
	ItemID int32 // control item (collar)
}

// PetLevelInfo — exported view of pet stats at a specific level.
// Phase 19: Pets/Summons System.
type PetLevelInfo struct {
	Level    int32
	Exp      int64
	MaxHP    float64
	MaxMP    float64
	PAtk     float64
	PDef     float64
	MAtk     float64
	MDef     float64
	MaxFeed  int32
	FeedRate float64
}

// PetSkillInfo — exported view of a pet skill.
type PetSkillInfo struct {
	SkillID int32
	Level   int32
}

// GetPetInfo returns exported pet info by NPC ID.
// Returns nil if not found.
func GetPetInfo(npcID int32) *PetInfo {
	def, ok := PetTable[npcID]
	if !ok {
		return nil
	}
	return &PetInfo{
		NpcID:  def.npcID,
		ItemID: def.itemID,
	}
}

// GetPetByControlItem returns pet info by control item ID (collar).
// Returns nil if no pet uses this item.
func GetPetByControlItem(itemID int32) *PetInfo {
	for _, def := range PetTable {
		if def.itemID == itemID {
			return &PetInfo{
				NpcID:  def.npcID,
				ItemID: def.itemID,
			}
		}
	}
	return nil
}

// GetPetLevelInfo returns pet stats for a specific NPC at a specific level.
// Returns nil if not found.
func GetPetLevelInfo(npcID, level int32) *PetLevelInfo {
	def, ok := PetTable[npcID]
	if !ok {
		return nil
	}
	for i := range def.levels {
		if def.levels[i].level == level {
			l := &def.levels[i]
			return &PetLevelInfo{
				Level:    l.level,
				Exp:      l.exp,
				MaxHP:    l.hp,
				MaxMP:    l.mp,
				PAtk:     l.pAtk,
				PDef:     l.pDef,
				MAtk:     l.mAtk,
				MDef:     l.mDef,
				MaxFeed:  l.maxFeed,
				FeedRate: l.feedRate,
			}
		}
	}
	return nil
}

// GetPetMaxLevel returns maximum level for a pet type.
// Returns 0 if not found.
func GetPetMaxLevel(npcID int32) int32 {
	def, ok := PetTable[npcID]
	if !ok {
		return 0
	}
	if len(def.levels) == 0 {
		return 0
	}
	return def.levels[len(def.levels)-1].level
}

// GetPetExpForLevel returns experience needed for a specific level.
// Returns -1 if not found.
func GetPetExpForLevel(npcID, level int32) int64 {
	info := GetPetLevelInfo(npcID, level)
	if info == nil {
		return -1
	}
	return info.Exp
}

// GetPetSkills returns skills for a pet type.
func GetPetSkills(npcID int32) []PetSkillInfo {
	def, ok := PetTable[npcID]
	if !ok {
		return nil
	}
	if len(def.skills) == 0 {
		return nil
	}
	result := make([]PetSkillInfo, len(def.skills))
	for i, s := range def.skills {
		result[i] = PetSkillInfo{SkillID: s.skillID, Level: s.level}
	}
	return result
}

// IsPetNpc checks if an NPC ID is a pet type.
func IsPetNpc(npcID int32) bool {
	_, ok := PetTable[npcID]
	return ok
}

// IsPetControlItem checks if an item ID is a pet control item.
func IsPetControlItem(itemID int32) bool {
	return GetPetByControlItem(itemID) != nil
}
