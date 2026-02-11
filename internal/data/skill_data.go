package data

// skillDef — определение скилла с per-level массивами.
// Для полей-слайсов: индекс = level-1. Если массив короче levels — берётся последний элемент.
// Одноэлементный слайс = одинаковое значение для всех уровней.
type skillDef struct {
	id            int32
	name          string
	levels        int32
	operateType   string
	isMagic       bool
	targetType    string
	hitTime       int32
	coolTime      int32
	reuseDelay    int32
	castRange     int32
	effectRange   int32
	magicLevel    int32
	abnormalType  string
	abnormalLevel int32

	// per-level arrays (index = level-1)
	power        []float64
	mpConsume    []int32
	hpConsume    []int32
	abnormalTime []int32

	// boolean flags
	isDebuff         bool
	overHit          bool
	ignoreShld       bool
	nextActionAttack bool
	isSuicideAttack  bool
	stayAfterDeath   bool
	staticReuse      bool
	removedOnDamage  bool

	// scalar attributes
	trait             string // "BLEED","BOSS","SHOCK","DERANGEMENT"...
	basicProperty     string // "CON","MEN","WIT"...
	element           int32  // 0=None,1=Fire,...6=Dark
	elementPower      int32
	affectRange       int32
	affectLimit       string // "5-12" format
	flyType           string
	flyRadius         int32
	fanRange          string
	itemConsumeId     int32
	itemConsumeCount  int32
	chargeConsume    int32
	blowChance       int32
	activateRate      int32
	lvlBonusRate      int32
	baseCritRate      int32
	condMsgId         int32
	condAddName       int32

	// per-level scalar overrides
	magicLevelByLvl  []int32
	effectPoint      []int32
	abnormalLevelTbl []int32
	mpInitialConsume []int32

	// enchant routes
	enchantGroup1   int32
	enchantGroup2   int32
	enchant1        []enchantOverride // route 1 attribute overrides
	enchant2        []enchantOverride // route 2 attribute overrides
	enchant1Effects []effectDef       // effects for route 1 (if different)
	enchant2Effects []effectDef       // effects for route 2

	// conditions
	conditions []conditionDef

	// effects
	effects []effectDef
}

// effectDef — определение эффекта скилла.
// params — скалярные параметры (одинаковые для всех уровней).
// perLvl — per-level параметры (индекс = level-1, последний элемент если массив короче).
// statMods — <add>/<mul>/<sub>/<set> внутри эффекта.
type effectDef struct {
	name     string
	params   map[string]string
	perLvl   map[string][]string
	statMods []statModDef
}

// statModDef — stat modifier inside an effect (<add>, <mul>, <sub>, <set>).
type statModDef struct {
	op   string // "add", "mul", "sub", "set"
	stat string // "pDef", "runSpd", "critRate"...
	val  string // может быть "#tableRef" — резолвится генератором
}

// enchantOverride — per-enchant-level attribute overrides.
type enchantOverride struct {
	attr   string   // "power", "mpConsume", "magicLevel"...
	values []string // per-enchant-level values (strings — parsed by loader)
}

// conditionDef — condition for skill usage.
type conditionDef struct {
	typ      string            // "using", "player", "target", "and", "or"
	params   map[string]string // kind="SWORD,BLUNT", Charges="2", race="UNDEAD"...
	children []conditionDef    // for and/or — nested conditions
}
