package fishing

// CalcDamage computes fishing damage based on rod damage, expertise bonus,
// skill power, rod level (for grade bonus), and fishing shot multiplier.
//
// Formula (from Reeling.java / Pumping.java):
//
//	baseDmg  = rodDamage + expertise + skillPower
//	gradeBonus = rodLevel * 0.1
//	dmg = int(baseDmg * gradeBonus * shotMultiplier)
func CalcDamage(rodDamage float64, expertise float64, skillPower float64,
	rodLevel int32, useFishShot bool) int32 {

	base := rodDamage + expertise + skillPower
	gradeBonus := float64(rodLevel) * 0.1
	shotMul := 1.0
	if useFishShot {
		shotMul = 2.0
	}
	return int32(base * gradeBonus * shotMul)
}

// CalcPenalty returns the penalty subtracted from damage when the player's
// fishing expertise skill level is too low relative to the action skill level.
//
// If expertiseLevel <= (skillLevel - 2), penalty = dmg * 0.05.
func CalcPenalty(dmg int32, expertiseLevel, skillLevel int32) int32 {
	if expertiseLevel <= (skillLevel - 2) {
		return int32(float64(dmg) * 0.05)
	}
	return 0
}
