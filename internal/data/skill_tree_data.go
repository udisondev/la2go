package data

// skillTreeDef — определение дерева скиллов (class, fishing, hero, noble, pledge, etc.).
type skillTreeDef struct {
	treeType      string // "classSkillTree", "fishingSkillTree", etc.
	classID       int32  // only for classSkillTree
	parentClassID int32  // class inheritance (not used in Interlude)
	skills        []skillTreeEntry
}

// skillTreeEntry — одна запись в дереве скиллов.
type skillTreeEntry struct {
	skillID      int32
	skillLevel   int32
	minLevel     int32
	spCost       int64
	autoGet      bool
	learnedByNpc bool
	items        []itemReq
	socialClass  string // for pledgeSkillTree
	race         string // for race-specific skills
}

// itemReq — required item for learning a skill.
type itemReq struct {
	itemID int32
	count  int32
}
