package data

// MaxPlayerLevel is the maximum achievable player level in Interlude.
const MaxPlayerLevel = 80

// ExperienceTable holds cumulative XP required to reach each level.
// Index = level (0-81). Level 0 and 1 require 0 XP.
// Data from L2J Mobius experience.xml (Interlude).
var ExperienceTable = [82]int64{
	0,          // 0 (unused)
	0,          // 1
	68,         // 2
	363,        // 3
	1168,       // 4
	2884,       // 5
	6038,       // 6
	11287,      // 7
	19423,      // 8
	31378,      // 9
	48229,      // 10
	71201,      // 11
	101676,     // 12
	141192,     // 13
	191452,     // 14
	254327,     // 15
	331864,     // 16
	426284,     // 17
	539995,     // 18
	675590,     // 19
	835854,     // 20
	1023775,    // 21
	1242536,    // 22
	1495531,    // 23
	1786365,    // 24
	2118860,    // 25
	2497059,    // 26
	2925229,    // 27
	3407873,    // 28
	3949727,    // 29
	4555766,    // 30
	5231213,    // 31
	5981539,    // 32
	6812472,    // 33
	7729999,    // 34
	8740372,    // 35
	9850111,    // 36
	11066012,   // 37
	12395149,   // 38
	13844879,   // 39
	15422851,   // 40
	17137002,   // 41
	18995573,   // 42
	21007109,   // 43
	23180476,   // 44
	25524859,   // 45
	28049776,   // 46
	30765073,   // 47
	33680933,   // 48
	36807883,   // 49
	40156799,   // 50
	43738914,   // 51
	47565824,   // 52
	51649497,   // 53
	56002282,   // 54
	60636913,   // 55
	65566520,   // 56
	70804633,   // 57
	76365186,   // 58
	82262524,   // 59
	88511413,   // 60
	95127046,   // 61
	102124950,  // 62
	109521094,  // 63
	117331800,  // 64
	125573854,  // 65
	134264511,  // 66
	143421503,  // 67
	153063052,  // 68
	163207876,  // 69
	173875199,  // 70
	185084664,  // 71
	196856353,  // 72
	209210793,  // 73
	222168975,  // 74
	235752477,  // 75
	249983468,  // 76
	264884712,  // 77
	280479584,  // 78
	296792080,  // 79
	313846832,  // 80
	331670128,  // 81 (theoretical cap for level 80 → 81 overflow)
}

// GetExpForLevel returns cumulative XP required to reach the given level.
// Returns 0 for level <= 1. Returns max XP for level > MaxPlayerLevel.
func GetExpForLevel(level int32) int64 {
	if level <= 1 {
		return 0
	}
	if level > MaxPlayerLevel+1 {
		level = MaxPlayerLevel + 1
	}
	return ExperienceTable[level]
}

// GetLevelForExp returns the level corresponding to the given cumulative XP.
// Scans upward from startLevel to find the highest level whose threshold is <= exp.
func GetLevelForExp(exp int64, startLevel int32) int32 {
	if startLevel < 1 {
		startLevel = 1
	}
	level := startLevel
	for level < MaxPlayerLevel {
		if ExperienceTable[level+1] > exp {
			break
		}
		level++
	}
	return level
}
