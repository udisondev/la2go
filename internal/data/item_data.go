package data

// itemDef — определение предмета для Go-литералов (generated).
type itemDef struct {
	id       int32
	name     string
	itemType string // "Weapon","Armor","EtcItem"

	// Common
	icon          string
	defaultAction string
	material      string
	weight        int32
	price         int64
	stackable     bool
	tradeable     bool
	droppable     bool
	sellable      bool
	depositable   bool
	questItem     bool

	// Weapon
	weaponType   string // "SWORD","BLUNT","DAGGER","BOW","POLE","DUAL","DUALFIST","FIST","NONE"
	bodyPart     string // "rhand","lrhand","lhand","chest","legs","head","feet","gloves","onepiece","neck","rear","lfinger","rfinger","under","back","hair","alldress"
	randomDamage int32
	attackRange  int32
	soulshots    int32
	spiritshots  int32
	magicWeapon  bool

	// Armor
	armorType string // "HEAVY","LIGHT","MAGIC","NONE","SIGIL","PET"

	// EtcItem
	etcItemType string // "MATERIAL","RECIPE","POTION","SCROLL","ARROW","QUEST"
	handler     string

	// Item skill (e.g., potions, scrolls): skillID-level from XML "item_skill" attr.
	itemSkillID    int32
	itemSkillLevel int32
	reuseDelay     int32 // milliseconds
	olyRestricted  bool  // cannot be used in Olympiad
	forNpc         bool  // NPC-only item

	// Stats
	pAtk     int32
	mAtk     int32
	pDef     int32
	mDef     int32
	pAtkSpd  int32
	mAtkSpd  int32
	critRate int32

	// Enchant stats (bonus per +1 enchant)
	enchantable bool

	// Crystal type / Grade (NONE=0, D=1, C=2, B=3, A=4, S=5)
	// Java reference: ItemTemplate.java CrystalType enum
	crystalType string // "NONE","D","C","B","A","S"

	// Conditions
	condMsgId int32
}

// itemStatDef — stat modifier on an item.
type itemStatDef struct {
	stat string // "pAtk","mAtk","pDef","critRate","pAtkSpd"
	val  float64
}
