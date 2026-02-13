package data

// TestNpcSkill is an exported type for cross-package test setup of NPC skill definitions.
type TestNpcSkill struct {
	SkillID int32
	Level   int32
}

// SetTestNpcDef populates NpcTable with a test NPC definition.
// Intended for tests from other packages that need NPC data setup.
func SetTestNpcDef(templateID int32, skills []TestNpcSkill, clans []string) {
	if NpcTable == nil {
		NpcTable = make(map[int32]*npcDef, 8)
	}
	npcSkills := make([]npcSkillDef, len(skills))
	for i, s := range skills {
		npcSkills[i] = npcSkillDef{skillID: s.SkillID, level: s.Level}
	}
	NpcTable[templateID] = &npcDef{
		id:            templateID,
		name:          "TestNpc",
		clans:         clans,
		clanHelpRange: 300,
		skills:        npcSkills,
	}
}

// ClearTestNpcTable resets NpcTable for test isolation.
func ClearTestNpcTable() {
	NpcTable = make(map[int32]*npcDef, 8)
}

// DeleteTestNpcDef removes a single entry from NpcTable.
func DeleteTestNpcDef(templateID int32) {
	delete(NpcTable, templateID)
}
