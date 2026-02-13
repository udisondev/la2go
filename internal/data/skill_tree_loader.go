package data

import (
	"log/slog"
	"sort"
)

// ClassSkillTrees maps classID → list of SkillLearn sorted by MinLevel.
// Загружается через LoadSkillTrees() при старте сервера.
var ClassSkillTrees map[int32][]*SkillLearn

// SpecialSkillTrees maps treeType → list of SkillLearn.
// Types: "fishingSkillTree", "heroSkillTree", "nobleSkillTree", "pledgeSkillTree", etc.
var SpecialSkillTrees map[string][]*SkillLearn

// GetAutoGetSkills returns auto-get skills for a class at a specific level.
// Returns skills where AutoGet=true AND MinLevel <= playerLevel.
func GetAutoGetSkills(classID, playerLevel int32) []*SkillLearn {
	skills, ok := ClassSkillTrees[classID]
	if !ok {
		return nil
	}

	var result []*SkillLearn
	for _, sl := range skills {
		if sl.AutoGet && sl.MinLevel <= playerLevel {
			result = append(result, sl)
		}
	}
	return result
}

// GetAvailableSkills returns all skills available for learning at a specific level.
// Returns skills where MinLevel <= playerLevel (both auto-get and manual).
func GetAvailableSkills(classID, playerLevel int32) []*SkillLearn {
	skills, ok := ClassSkillTrees[classID]
	if !ok {
		return nil
	}

	var result []*SkillLearn
	for _, sl := range skills {
		if sl.MinLevel <= playerLevel {
			result = append(result, sl)
		}
	}
	return result
}

// GetNewAutoGetSkills returns auto-get skills that should be granted at exactly this level.
// Used on level-up to grant only the new skills (MinLevel == playerLevel).
func GetNewAutoGetSkills(classID, playerLevel int32) []*SkillLearn {
	skills, ok := ClassSkillTrees[classID]
	if !ok {
		return nil
	}

	var result []*SkillLearn
	for _, sl := range skills {
		if sl.AutoGet && sl.MinLevel == playerLevel {
			result = append(result, sl)
		}
	}
	return result
}

// GetLearnableSkills returns class skills available for learning that the player doesn't yet know.
// knownSkills maps skillID→currentLevel. Returns only skills where:
// - MinLevel <= playerLevel
// - SkillLevel > current known level (or skill is unknown)
// - AutoGet == false (manual learning only)
func GetLearnableSkills(classID, playerLevel int32, knownSkills map[int32]int32) []*SkillLearn {
	skills, ok := ClassSkillTrees[classID]
	if !ok {
		return nil
	}

	var result []*SkillLearn
	for _, sl := range skills {
		if sl.AutoGet {
			continue // skip auto-get skills
		}
		if sl.MinLevel > playerLevel {
			continue // player too low level
		}
		currentLvl := knownSkills[sl.SkillID]
		if sl.SkillLevel <= currentLvl {
			continue // already learned this level or higher
		}
		// Only show the next level (not skip levels)
		if currentLvl > 0 && sl.SkillLevel != currentLvl+1 {
			continue
		}
		result = append(result, sl)
	}
	return result
}

// GetSkillLearn returns a specific SkillLearn entry or nil.
func GetSkillLearn(classID, skillID, skillLevel int32) *SkillLearn {
	skills, ok := ClassSkillTrees[classID]
	if !ok {
		return nil
	}
	for _, sl := range skills {
		if sl.SkillID == skillID && sl.SkillLevel == skillLevel {
			return sl
		}
	}
	return nil
}

// GetSpecialSkillLearn returns a specific SkillLearn entry from a special skill tree.
// treeType examples: "fishingSkillTree", "pledgeSkillTree", "heroSkillTree", "nobleSkillTree".
func GetSpecialSkillLearn(treeType string, skillID, skillLevel int32) *SkillLearn {
	skills, ok := SpecialSkillTrees[treeType]
	if !ok {
		return nil
	}
	for _, sl := range skills {
		if sl.SkillID == skillID && sl.SkillLevel == skillLevel {
			return sl
		}
	}
	return nil
}

// GetLearnableSpecialSkills returns skills from a special tree available for learning.
// knownSkills maps skillID→currentLevel. Filters same as GetLearnableSkills.
func GetLearnableSpecialSkills(treeType string, playerLevel int32, knownSkills map[int32]int32) []*SkillLearn {
	skills, ok := SpecialSkillTrees[treeType]
	if !ok {
		return nil
	}

	var result []*SkillLearn
	for _, sl := range skills {
		if sl.AutoGet {
			continue
		}
		if sl.MinLevel > playerLevel {
			continue
		}
		currentLvl := knownSkills[sl.SkillID]
		if sl.SkillLevel <= currentLvl {
			continue
		}
		if currentLvl > 0 && sl.SkillLevel != currentLvl+1 {
			continue
		}
		result = append(result, sl)
	}
	return result
}

// LoadSkillTrees строит ClassSkillTrees и SpecialSkillTrees из Go-литералов (skillTreeDefs).
// Вызывается при старте сервера после LoadSkills().
func LoadSkillTrees() error {
	ClassSkillTrees = make(map[int32][]*SkillLearn)
	SpecialSkillTrees = make(map[string][]*SkillLearn)

	for _, treeDef := range skillTreeDefs {
		for _, entry := range treeDef.skills {
			sl := &SkillLearn{
				SkillID:      entry.skillID,
				SkillLevel:   entry.skillLevel,
				MinLevel:     entry.minLevel,
				SpCost:       entry.spCost,
				AutoGet:      entry.autoGet,
				ClassID:      treeDef.classID,
				LearnedByNpc: entry.learnedByNpc,
			}

			// Convert items
			for _, item := range entry.items {
				sl.Items = append(sl.Items, ItemReq{
					ItemID: item.itemID,
					Count:  item.count,
				})
			}

			if treeDef.treeType == "classSkillTree" {
				ClassSkillTrees[treeDef.classID] = append(ClassSkillTrees[treeDef.classID], sl)
			} else {
				SpecialSkillTrees[treeDef.treeType] = append(SpecialSkillTrees[treeDef.treeType], sl)
			}
		}
	}

	// Sort each class tree by MinLevel
	for classID := range ClassSkillTrees {
		sort.Slice(ClassSkillTrees[classID], func(i, j int) bool {
			return ClassSkillTrees[classID][i].MinLevel < ClassSkillTrees[classID][j].MinLevel
		})
	}

	// Sort special trees by MinLevel
	for treeType := range SpecialSkillTrees {
		sort.Slice(SpecialSkillTrees[treeType], func(i, j int) bool {
			return SpecialSkillTrees[treeType][i].MinLevel < SpecialSkillTrees[treeType][j].MinLevel
		})
	}

	var totalEntries int
	for _, skills := range ClassSkillTrees {
		totalEntries += len(skills)
	}
	var specialEntries int
	for _, skills := range SpecialSkillTrees {
		specialEntries += len(skills)
	}

	slog.Info("loaded skill trees",
		"classes", len(ClassSkillTrees),
		"class_entries", totalEntries,
		"special_trees", len(SpecialSkillTrees),
		"special_entries", specialEntries,
	)
	return nil
}
