package olympiad

import "testing"

func TestNewStadiums(t *testing.T) {
	stadiums := NewStadiums()

	if len(stadiums) != StadiumCount {
		t.Fatalf("NewStadiums() count = %d; want %d", len(stadiums), StadiumCount)
	}

	for i, s := range stadiums {
		if s == nil {
			t.Fatalf("stadium[%d] is nil", i)
		}
		if s.ID() != int32(i) {
			t.Errorf("stadium[%d].ID() = %d; want %d", i, s.ID(), i)
		}
		if s.InUse() {
			t.Errorf("stadium[%d].InUse() = true; want false", i)
		}
	}
}

func TestStadium_SetInUse(t *testing.T) {
	stadiums := NewStadiums()
	s := stadiums[0]

	s.SetInUse(true)
	if !s.InUse() {
		t.Error("InUse() = false after SetInUse(true)")
	}

	s.SetInUse(false)
	if s.InUse() {
		t.Error("InUse() = true after SetInUse(false)")
	}
}

func TestStadium_Location(t *testing.T) {
	stadiums := NewStadiums()

	// Проверить первый стадион
	loc := stadiums[0].Location()
	if loc.X != -20814 || loc.Y != -21189 || loc.Z != -3030 {
		t.Errorf("stadium[0].Location() = (%d,%d,%d); want (-20814,-21189,-3030)",
			loc.X, loc.Y, loc.Z)
	}

	// Проверить последний стадион
	last := stadiums[StadiumCount-1].Location()
	if last.X != -114413 || last.Y != -213241 || last.Z != -3331 {
		t.Errorf("stadium[21].Location() = (%d,%d,%d); want (-114413,-213241,-3331)",
			last.X, last.Y, last.Z)
	}
}

func TestCompetitionType_String(t *testing.T) {
	tests := []struct {
		ct   CompetitionType
		want string
	}{
		{CompClassed, "CLASSED"},
		{CompNonClassed, "NON_CLASSED"},
		{CompetitionType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.ct.String()
		if got != tt.want {
			t.Errorf("CompetitionType(%d).String() = %q; want %q", tt.ct, got, tt.want)
		}
	}
}

func TestStadiumConstants(t *testing.T) {
	if StadiumCount != 22 {
		t.Errorf("StadiumCount = %d; want 22", StadiumCount)
	}
	if NonClassedStadiumStart != 0 {
		t.Errorf("NonClassedStadiumStart = %d; want 0", NonClassedStadiumStart)
	}
	if NonClassedStadiumEnd != 10 {
		t.Errorf("NonClassedStadiumEnd = %d; want 10", NonClassedStadiumEnd)
	}
	if ClassedStadiumStart != 11 {
		t.Errorf("ClassedStadiumStart = %d; want 11", ClassedStadiumStart)
	}
	if ClassedStadiumEnd != 21 {
		t.Errorf("ClassedStadiumEnd = %d; want 21", ClassedStadiumEnd)
	}
}
