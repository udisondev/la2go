package data

import (
	"testing"
)

// init ensures SkillTable is loaded for benchmarks.
func init() {
	if SkillTable == nil {
		if err := LoadSkills(); err != nil {
			panic("failed to load skills for benchmarks: " + err.Error())
		}
	}
}

// findExistingSkillID finds a real skill ID from loaded SkillTable for benchmarks.
func findExistingSkillID() (int32, int32) {
	for skillID, levels := range SkillTable {
		for level := range levels {
			return skillID, level
		}
	}
	return 0, 0
}

// BenchmarkGetSkillTemplate_Hit benchmarks lookup for existing skill (cache hit path).
// Expected: ~10-30ns (map[int32] lookup × 2).
func BenchmarkGetSkillTemplate_Hit(b *testing.B) {
	b.ReportAllocs()
	skillID, level := findExistingSkillID()
	if skillID == 0 {
		b.Skip("no skills loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = GetSkillTemplate(skillID, level)
	}
}

// BenchmarkGetSkillTemplate_Miss benchmarks lookup for non-existing skill.
// Expected: ~5-15ns (first map lookup misses).
func BenchmarkGetSkillTemplate_Miss(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		_ = GetSkillTemplate(999999, 1)
	}
}

// BenchmarkGetSkillTemplate_WrongLevel benchmarks existing skillID but wrong level.
// Expected: ~10-20ns (first map hit, second map miss/nil).
func BenchmarkGetSkillTemplate_WrongLevel(b *testing.B) {
	b.ReportAllocs()
	skillID, _ := findExistingSkillID()
	if skillID == 0 {
		b.Skip("no skills loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = GetSkillTemplate(skillID, 9999)
	}
}

// BenchmarkGetSkillMaxLevel benchmarks max level lookup.
// Expected: ~10-30ns (map lookup + iteration).
func BenchmarkGetSkillMaxLevel(b *testing.B) {
	b.ReportAllocs()
	skillID, _ := findExistingSkillID()
	if skillID == 0 {
		b.Skip("no skills loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = GetSkillMaxLevel(skillID)
	}
}

// BenchmarkGetSkillMaxLevel_Miss benchmarks max level for non-existing skill.
// Expected: ~5-10ns (nil check + map miss).
func BenchmarkGetSkillMaxLevel_Miss(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		_ = GetSkillMaxLevel(999999)
	}
}

// BenchmarkGetSkillTemplate_Concurrent benchmarks concurrent skill lookups.
// Expected: measures contention on map[int32] reads (no mutex, but CPU cache).
func BenchmarkGetSkillTemplate_Concurrent(b *testing.B) {
	b.ReportAllocs()
	skillID, level := findExistingSkillID()
	if skillID == 0 {
		b.Skip("no skills loaded")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetSkillTemplate(skillID, level)
		}
	})
}
