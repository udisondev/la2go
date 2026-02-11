package data

// Shared constants and profiles for player class templates.
// Extracted from playerTemplateDefs to eliminate duplication (DRY).
// Values 1:1 from L2J Mobius CT 0 Interlude XML templates.

// --- MDef: identical for ALL 58 classes ---

const (
	BaseMDefRear    int32 = 9
	BaseMDefLear    int32 = 9
	BaseMDefRFinger int32 = 5
	BaseMDefLFinger int32 = 5
	BaseMDefNeck    int32 = 13
)

// --- PDef: 2 profiles (Fighter / Mystic) + 5 shared ---

const (
	BasePDefHead      int32 = 12
	BasePDefFeet      int32 = 7
	BasePDefGloves    int32 = 8
	BasePDefUnderwear int32 = 3
	BasePDefCloak     int32 = 1
)

const (
	FighterPDefChest int32 = 31
	FighterPDefLegs  int32 = 18
	MysticPDefChest  int32 = 15
	MysticPDefLegs   int32 = 8
)

// --- Combat stats: identical for all 58 (except basePAtk, baseAtkRange) ---

const (
	BasePAtkSpd   int32 = 300
	BaseMAtkSpd   int32 = 0
	BaseCritRate  int32 = 4
	BaseRandomDmg int32 = 10
	BaseMAtk      int32 = 6
)

const (
	FighterBasePAtk int32 = 4
	MysticBasePAtk  int32 = 3
)

const (
	BaseAtkRange    int32 = 20 // 49 classes (Human, Elf, DarkElf, Dwarf)
	OrcBaseAtkRange int32 = 25 // 9 Orc classes
)

// --- Regen arrays: identical for all 58 classes ---
// Read-only, shared across all templates (safe: convertTemplateDef copies slice header, not data).

var baseHPRegen = []float64{2, 2.05, 2.1, 2.15, 2.2, 2.25, 2.3, 2.35, 2.4, 2.45, 2.5, 2.6, 2.7, 2.8, 2.9, 3, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 4, 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9, 5, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9, 6, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 7, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 8, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8, 8.9, 9, 9.1, 9.2, 9.3, 9.4}

var baseMPRegen = []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.8, 1.8, 1.8, 1.8, 1.8, 1.8, 1.8, 1.8, 1.8, 1.8, 2.1, 2.1, 2.1, 2.1, 2.1, 2.1, 2.1, 2.1, 2.1, 2.1, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.7, 2.7, 2.7, 2.7, 2.7, 2.7, 2.7, 2.7, 2.7, 2.7, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}

var baseCPRegen = []float64{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2.5, 2.5, 2.5, 2.5, 2.5, 2.5, 2.5, 2.5, 2.5, 2.5, 3.5, 3.5, 3.5, 3.5, 3.5, 3.5, 3.5, 3.5, 3.5, 3.5, 4.5, 4.5, 4.5, 4.5, 4.5, 4.5, 4.5, 4.5, 4.5, 4.5, 5.5, 5.5, 5.5, 5.5, 5.5, 5.5, 5.5, 5.5, 5.5, 5.5, 6.5, 6.5, 6.5, 6.5, 6.5, 6.5, 6.5, 6.5, 6.5, 6.5, 7.5, 7.5, 7.5, 7.5, 7.5, 7.5, 7.5, 7.5, 7.5, 7.5, 8.5, 8.5, 8.5, 8.5, 8.5, 8.5, 8.5, 8.5, 8.5, 8.5}

// --- Creation points: 6 racial spawn locations ---

var humanFighterSpawn = []Location{{-71338, 258271, -3104}, {-71417, 258270, -3104}, {-71453, 258305, -3104}, {-71467, 258378, -3104}}

var humanMysticSpawn = []Location{{-90875, 248162, -3570}, {-90954, 248118, -3570}, {-90918, 248070, -3570}, {-90890, 248027, -3570}}

var elfSpawn = []Location{{46045, 41251, -3440}, {46117, 41247, -3440}, {46182, 41198, -3440}, {46115, 41141, -3440}, {46048, 41141, -3440}, {45978, 41196, -3440}}

var darkElfSpawn = []Location{{28295, 11063, -4224}, {28302, 11008, -4224}, {28377, 10916, -4224}, {28456, 10997, -4224}, {28461, 11044, -4224}, {28395, 11127, -4224}}

var orcSpawn = []Location{{-56733, -113459, -690}, {-56686, -113470, -690}, {-56728, -113610, -690}, {-56693, -113610, -690}, {-56743, -113757, -690}, {-56682, -113730, -690}}

var dwarfSpawn = []Location{{108644, -173947, -400}, {108678, -174002, -400}, {108505, -173964, -400}, {108512, -174026, -400}, {108549, -174075, -400}, {108576, -174122, -400}}

// --- Collision profiles: 7 unique race+type combinations ---

type collisionProfile struct {
	radiusMale, heightMale, radiusFemale, heightFemale float32
}

var (
	collHumanFighter = collisionProfile{9, 23, 8, 23.5}
	collHumanMystic  = collisionProfile{7.5, 22.8, 6.5, 22.5}
	collElf          = collisionProfile{7.5, 24, 7.5, 23}
	collDarkElf      = collisionProfile{7.5, 24, 7, 23.5}
	collOrcFighter   = collisionProfile{11, 28, 7, 27}
	collOrcMystic    = collisionProfile{7, 27.5, 8, 25.5}
	collDwarf        = collisionProfile{9, 18, 5, 19}
)

// --- Base stat profiles: 9 unique race+type combinations ---

type baseStatProfile struct {
	str, con, dex, int_, wit, men uint8
}

var (
	statsHumanFighter   = baseStatProfile{40, 43, 30, 21, 11, 25}
	statsHumanMystic    = baseStatProfile{22, 27, 21, 41, 20, 39}
	statsElfFighter     = baseStatProfile{36, 36, 35, 23, 14, 26}
	statsElfMystic      = baseStatProfile{21, 25, 24, 37, 23, 40}
	statsDarkElfFighter = baseStatProfile{41, 32, 34, 25, 12, 26}
	statsDarkElfMystic  = baseStatProfile{23, 24, 23, 44, 19, 37}
	statsOrcFighter     = baseStatProfile{40, 47, 26, 18, 12, 27}
	statsOrcMystic      = baseStatProfile{27, 31, 24, 31, 15, 42}
	statsDwarf          = baseStatProfile{39, 45, 29, 20, 10, 27}
)
