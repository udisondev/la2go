package data

// EnchantSkillRoute contains data for a single skill enchant level.
type EnchantSkillRoute struct {
	SkillID   int32
	Level     int32
	SpCost    int32 // SP required
	ExpCost   int64 // EXP required
	Rate      int32 // success rate 0-100
	BaseLevel int32 // level to reset to on failure
}

// enchantSkillRoutes maps skillID → level → route.
// TODO: Load from EnchantSkillTreeData.xml (14554 lines, 483 routes).
var enchantSkillRoutes map[int64]*EnchantSkillRoute

func init() {
	enchantSkillRoutes = make(map[int64]*EnchantSkillRoute)
}

// enchantKey returns a combined key for skillID+level.
func enchantKey(skillID, level int32) int64 {
	return int64(skillID)<<32 | int64(level)
}

// GetEnchantSkillRoute returns the enchant route for a skill at a specific enchant level.
// Returns nil if no enchant route exists.
func GetEnchantSkillRoute(skillID, level int32) *EnchantSkillRoute {
	return enchantSkillRoutes[enchantKey(skillID, level)]
}

// RegisterEnchantSkillRoute registers an enchant skill route (used by data loader).
func RegisterEnchantSkillRoute(route *EnchantSkillRoute) {
	enchantSkillRoutes[enchantKey(route.SkillID, route.Level)] = route
}
