package data

// AugmentInfo is an exported view of augmentation data.
// Phase 28: Augmentation System.
type AugmentInfo struct {
	ID         int32
	SkillID    int32
	SkillLevel int32
	AugType    string // "blue","red","yellow","purple"
}

// GetAugmentInfo returns augmentation info by ID.
// Returns nil if not found.
func GetAugmentInfo(id int32) *AugmentInfo {
	def := AugmentationTable[id]
	if def == nil {
		return nil
	}
	return &AugmentInfo{
		ID:         def.id,
		SkillID:    def.skillID,
		SkillLevel: def.skillLevel,
		AugType:    def.augType,
	}
}

// AugmentHasSkill returns true if augmentation ID maps to a skill.
func AugmentHasSkill(id int32) bool {
	def := AugmentationTable[id]
	return def != nil && def.skillID > 0
}

// AugmentSkill returns (skillID, skillLevel) for an augmentation.
// Returns (0, 0) if no skill.
func AugmentSkill(id int32) (int32, int32) {
	def := AugmentationTable[id]
	if def == nil || def.skillID == 0 {
		return 0, 0
	}
	return def.skillID, def.skillLevel
}
