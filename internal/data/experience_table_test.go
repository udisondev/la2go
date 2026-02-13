package data

import "testing"

func TestGetExpForLevel(t *testing.T) {
	tests := []struct {
		level int32
		want  int64
	}{
		{0, 0},
		{1, 0},
		{2, 68},
		{5, 2884},
		{10, 48229},
		{20, 835854},
		{40, 15422851},
		{60, 126509030},
		{80, 4200000000},
		{81, 6300000000},      // overflow cap
		{100, 6300000000},     // clamped to 81
	}

	for _, tt := range tests {
		got := GetExpForLevel(tt.level)
		if got != tt.want {
			t.Errorf("GetExpForLevel(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestGetLevelForExp(t *testing.T) {
	tests := []struct {
		exp        int64
		startLevel int32
		want       int32
	}{
		{0, 1, 1},
		{67, 1, 1},              // just below level 2
		{68, 1, 2},              // exactly level 2
		{69, 1, 2},              // just above level 2
		{48229, 1, 10},          // exactly level 10
		{48230, 1, 10},          // just above level 10
		{71200, 1, 10},          // just below level 11
		{71201, 1, 11},          // exactly level 11
		{4200000000, 1, 80},     // exactly level 80
		{9999999999, 1, 80},     // way above — capped at 80
		{126509030, 50, 60},     // start from level 50, should find 60
		{126509030, 60, 60},     // start from exact level
	}

	for _, tt := range tests {
		got := GetLevelForExp(tt.exp, tt.startLevel)
		if got != tt.want {
			t.Errorf("GetLevelForExp(%d, %d) = %d, want %d", tt.exp, tt.startLevel, got, tt.want)
		}
	}
}

func TestExperienceTableMonotonic(t *testing.T) {
	for i := 1; i <= MaxPlayerLevel; i++ {
		if ExperienceTable[i] >= ExperienceTable[i+1] {
			t.Errorf("ExperienceTable[%d]=%d >= ExperienceTable[%d]=%d — must be strictly increasing",
				i, ExperienceTable[i], i+1, ExperienceTable[i+1])
		}
	}
}
