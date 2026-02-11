package data

import (
	"testing"
)

// TestLoadSkillTrees_Count tests that all trees from XML are loaded.
func TestLoadSkillTrees_Count(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	if len(ClassSkillTrees) < 85 {
		t.Errorf("ClassSkillTrees should have >= 85 classes, got %d", len(ClassSkillTrees))
	}

	var totalEntries int
	for _, skills := range ClassSkillTrees {
		totalEntries += len(skills)
	}
	if totalEntries < 13000 {
		t.Errorf("total class entries should be >= 13000, got %d", totalEntries)
	}

	// Special trees
	if len(SpecialSkillTrees) < 4 {
		t.Errorf("SpecialSkillTrees should have >= 4 types, got %d", len(SpecialSkillTrees))
	}
	for _, treeType := range []string{"fishingSkillTree", "heroSkillTree", "nobleSkillTree", "pledgeSkillTree"} {
		if _, ok := SpecialSkillTrees[treeType]; !ok {
			t.Errorf("special tree %q not found", treeType)
		}
	}
}

// TestLoadSkillTrees_HumanFighter tests Human Fighter (classID=0) tree.
func TestLoadSkillTrees_HumanFighter(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	skills, ok := ClassSkillTrees[0]
	if !ok {
		t.Fatal("Human Fighter (classID=0) not found")
	}
	if len(skills) == 0 {
		t.Fatal("Human Fighter has no skills")
	}

	// Lucky (id=194) should be autoGet at level 1
	found := false
	for _, sl := range skills {
		if sl.SkillID == 194 && sl.SkillLevel == 1 {
			found = true
			if sl.MinLevel != 1 {
				t.Errorf("Lucky minLevel: got %d, want 1", sl.MinLevel)
			}
			if !sl.AutoGet {
				t.Error("Lucky should be autoGet")
			}
		}
	}
	if !found {
		t.Error("Lucky (id=194) not found for Human Fighter")
	}

	// Power Strike (id=3) should exist as learnedByNpc
	found = false
	for _, sl := range skills {
		if sl.SkillID == 3 && sl.SkillLevel == 1 {
			found = true
			if !sl.LearnedByNpc {
				t.Error("Power Strike L1 should be learnedByNpc")
			}
		}
	}
	if !found {
		t.Error("Power Strike L1 not found for Human Fighter")
	}
}

// TestLoadSkillTrees_SortedByMinLevel tests that skills are sorted by MinLevel.
func TestLoadSkillTrees_SortedByMinLevel(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	for classID, skills := range ClassSkillTrees {
		for i := 1; i < len(skills); i++ {
			if skills[i].MinLevel < skills[i-1].MinLevel {
				t.Errorf("classID=%d: skills not sorted by MinLevel at index %d (%d < %d)",
					classID, i, skills[i].MinLevel, skills[i-1].MinLevel)
			}
		}
	}
}

// TestGetAutoGetSkills tests retrieval of auto-get skills.
func TestGetAutoGetSkills(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	// Human Fighter at level 1: should get Lucky (id=194)
	autoSkills := GetAutoGetSkills(0, 1)
	if len(autoSkills) == 0 {
		t.Fatal("no auto-get skills for Human Fighter at level 1")
	}

	found := false
	for _, sl := range autoSkills {
		if sl.SkillID == 194 && sl.SkillLevel == 1 {
			found = true
		}
	}
	if !found {
		t.Error("Lucky should be auto-get for Human Fighter at level 1")
	}
}

// TestGetAutoGetSkills_UnknownClass tests unknown class returns nil.
func TestGetAutoGetSkills_UnknownClass(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	autoSkills := GetAutoGetSkills(999, 1)
	if autoSkills != nil {
		t.Errorf("unknown class should return nil, got %d skills", len(autoSkills))
	}
}

// TestMultipleClasses tests that multiple class trees are loaded.
func TestMultipleClasses(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	// Should have starting classes (0, 10, 18, 25, 31, 38, 44, 49, 53)
	for _, classID := range []int32{0, 10, 18, 25, 31, 38, 44, 49, 53} {
		if _, ok := ClassSkillTrees[classID]; !ok {
			t.Errorf("classID=%d not found in ClassSkillTrees", classID)
		}
	}
}

// TestLoadSkillTrees_FishingItems tests fishing tree has item requirements.
func TestLoadSkillTrees_FishingItems(t *testing.T) {
	if err := LoadSkillTrees(); err != nil {
		t.Fatalf("LoadSkillTrees() failed: %v", err)
	}

	fishing, ok := SpecialSkillTrees["fishingSkillTree"]
	if !ok {
		t.Fatal("fishingSkillTree not found")
	}
	if len(fishing) == 0 {
		t.Fatal("fishingSkillTree has no entries")
	}

	// Fishing skill should have item requirement (Adena, id=57)
	found := false
	for _, sl := range fishing {
		if sl.SkillID == 1312 && len(sl.Items) > 0 {
			found = true
			if sl.Items[0].ItemID != 57 {
				t.Errorf("Fishing item id: got %d, want 57 (Adena)", sl.Items[0].ItemID)
			}
		}
	}
	if !found {
		t.Error("Fishing skill (id=1312) with items not found")
	}
}
