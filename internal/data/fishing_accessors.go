package data

// FishTemplate is the exported view of fish data.
type FishTemplate struct {
	ID             int32
	ItemID         int32
	FishType       string // "swift","ugly","wide"
	Group          int32  // 0=normal,1=easy,2=hard
	Level          int32
	HP             int32
	HPRegen        float64
	CombatDuration int32 // seconds
	FishGrade      int32 // 0=easy,1=normal,2=hard
}

// FishingRodTemplate is the exported view of a fishing rod.
type FishingRodTemplate struct {
	ID     int32
	ItemID int32
	Level  int32
	Name   string
	Damage float64
}

// FishingMonsterTemplate is the exported view of a fishing monster entry.
type FishingMonsterTemplate struct {
	MinLevel  int32
	MaxLevel  int32
	MonsterID int32
	Chance    int32 // percent
}

// Fish grade constants matching L2J FishGrade values.
const (
	FishGradeEasy   int32 = 0
	FishGradeNormal int32 = 1
	FishGradeHard   int32 = 2
)

func fishDefToTemplate(d *fishDef) *FishTemplate {
	return &FishTemplate{
		ID:             d.id,
		ItemID:         d.itemID,
		FishType:       d.fishType,
		Group:          d.group,
		Level:          d.level,
		HP:             d.hp,
		HPRegen:        d.hpRegen,
		CombatDuration: d.combatDuration,
		FishGrade:      d.fishGrade,
	}
}

func rodDefToTemplate(d *fishingRodDef) *FishingRodTemplate {
	return &FishingRodTemplate{
		ID:     d.id,
		ItemID: d.itemID,
		Level:  d.level,
		Name:   d.name,
		Damage: d.damage,
	}
}

// GetFishTemplate returns the fish template by fish ID, or nil if not found.
func GetFishTemplate(fishID int32) *FishTemplate {
	d := FishTable[fishID]
	if d == nil {
		return nil
	}
	return fishDefToTemplate(d)
}

// GetFishByLevel returns all fish templates matching the given level.
func GetFishByLevel(level int32) []*FishTemplate {
	var result []*FishTemplate
	for _, d := range FishTable {
		if d.level == level {
			result = append(result, fishDefToTemplate(d))
		}
	}
	return result
}

// GetFishByLevelAndGrade returns fish templates matching level and grade.
func GetFishByLevelAndGrade(level, grade int32) []*FishTemplate {
	var result []*FishTemplate
	for _, d := range FishTable {
		if d.level == level && d.fishGrade == grade {
			result = append(result, fishDefToTemplate(d))
		}
	}
	return result
}

// GetFishByLevelAndType returns fish templates matching level and type.
func GetFishByLevelAndType(level int32, fishType string) []*FishTemplate {
	var result []*FishTemplate
	for _, d := range FishTable {
		if d.level == level && d.fishType == fishType {
			result = append(result, fishDefToTemplate(d))
		}
	}
	return result
}

// GetFishingRod returns the rod template by item ID, or nil if not found.
func GetFishingRod(itemID int32) *FishingRodTemplate {
	d := FishingRodTable[itemID]
	if d == nil {
		return nil
	}
	return rodDefToTemplate(d)
}

// GetAllFishingRods returns all rod templates.
func GetAllFishingRods() []*FishingRodTemplate {
	result := make([]*FishingRodTemplate, 0, len(FishingRodTable))
	for _, d := range FishingRodTable {
		result = append(result, rodDefToTemplate(d))
	}
	return result
}

// GetFishingMonster returns the fishing monster matching the player's level,
// or nil if the player level is out of range.
func GetFishingMonster(playerLevel int32) *FishingMonsterTemplate {
	for i := range FishingMonsterTable {
		d := &FishingMonsterTable[i]
		if playerLevel >= d.minLevel && playerLevel <= d.maxLevel {
			return &FishingMonsterTemplate{
				MinLevel:  d.minLevel,
				MaxLevel:  d.maxLevel,
				MonsterID: d.monsterID,
				Chance:    d.chance,
			}
		}
	}
	return nil
}

// GetAllFishingMonsters returns all fishing monster entries.
func GetAllFishingMonsters() []FishingMonsterTemplate {
	result := make([]FishingMonsterTemplate, len(FishingMonsterTable))
	for i := range FishingMonsterTable {
		d := &FishingMonsterTable[i]
		result[i] = FishingMonsterTemplate{
			MinLevel:  d.minLevel,
			MaxLevel:  d.maxLevel,
			MonsterID: d.monsterID,
			Chance:    d.chance,
		}
	}
	return result
}
