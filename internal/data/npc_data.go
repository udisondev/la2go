package data

// npcDef — определение NPC для Go-литералов (generated).
// Содержит все данные из L2J XML для одного NPC.
type npcDef struct {
	id      int32
	name    string
	title   string
	npcType string // "monster","merchant","grand_boss","folk","guard" (snake_case)
	level   int32
	race    string // "FAIRY","ANIMAL","HUMANOID","UNDEAD","BEAST","HUMAN","DIVINE","PLANT","DRAGON","BUG","GIANT","DEMONIC","ELF","DARK_ELF","ORC","DWARF","NONE"
	sex     string // "MALE","FEMALE"

	// Stats
	str, intel, dex, wit, con, men int32
	hp, mp                         float64
	hpRegen, mpRegen               float64
	pAtk, mAtk                     float64
	pDef, mDef                     float64
	accuracy                       float64
	critical                       int32
	randomDamage                   int32
	evasion                        float64
	atkSpeed                       int32
	atkType                        string // "SWORD","BLUNT","DAGGER","BOW","POLE","FIST","DUAL","DUALFIST"
	atkRange                       int32
	atkDistance                     int32
	atkWidth                       int32
	walkSpeed                      int32
	runSpeed                       int32
	hitTime                        int32

	// AI
	aggroRange    int32
	clanHelpRange int32
	isAggressive  bool
	clans         []string
	ignoreNpcIds  []int32

	// Rewards
	baseExp int64
	baseSP  int32

	// Drop
	drops  []dropGroupDef
	spoils []dropItemDef

	// Skills
	skills []npcSkillDef

	// Collision
	collisionRadius float64
	collisionHeight float64

	// Equipment
	rhand int32
	lhand int32
	chest int32

	// Status flags
	undying    bool
	attackable bool
	talkable   bool
	canBeSown  bool

	corpseTime int32
	displayId  int32

	// Minions
	minions []minionDef
}

type dropGroupDef struct {
	chance float64
	items  []dropItemDef
}

type dropItemDef struct {
	itemID int32
	min    int32
	max    int32
	chance float64
}

type npcSkillDef struct {
	skillID int32
	level   int32
}

type minionDef struct {
	npcID int32
	count int32
}
