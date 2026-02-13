package sevensigns

import "testing"

func TestPlayerData_ContributionFromStones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		pd    PlayerData
		want  int64
	}{
		{"all zeros", PlayerData{}, 0},
		{"blue only", PlayerData{BlueStones: 10}, 30},
		{"green only", PlayerData{GreenStones: 10}, 50},
		{"red only", PlayerData{RedStones: 10}, 100},
		{"mixed", PlayerData{BlueStones: 5, GreenStones: 3, RedStones: 2}, 5*3 + 3*5 + 2*10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pd.ContributionFromStones()
			if got != tt.want {
				t.Errorf("ContributionFromStones() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStatus_SealOwner(t *testing.T) {
	t.Parallel()

	s := Status{
		AvariceOwner: CabalDawn,
		GnosisOwner:  CabalDusk,
		StrifeOwner:  CabalNull,
	}

	tests := []struct {
		seal Seal
		want Cabal
	}{
		{SealAvarice, CabalDawn},
		{SealGnosis, CabalDusk},
		{SealStrife, CabalNull},
		{SealNull, CabalNull},
	}
	for _, tt := range tests {
		got := s.SealOwner(tt.seal)
		if got != tt.want {
			t.Errorf("SealOwner(%d) = %d, want %d", tt.seal, got, tt.want)
		}
	}
}

func TestStatus_SetSealOwner(t *testing.T) {
	t.Parallel()

	var s Status
	s.SetSealOwner(SealAvarice, CabalDawn)
	s.SetSealOwner(SealGnosis, CabalDusk)
	s.SetSealOwner(SealStrife, CabalDawn)

	if s.AvariceOwner != CabalDawn {
		t.Errorf("AvariceOwner = %d, want %d", s.AvariceOwner, CabalDawn)
	}
	if s.GnosisOwner != CabalDusk {
		t.Errorf("GnosisOwner = %d, want %d", s.GnosisOwner, CabalDusk)
	}
	if s.StrifeOwner != CabalDawn {
		t.Errorf("StrifeOwner = %d, want %d", s.StrifeOwner, CabalDawn)
	}
}

func TestStatus_SealScore(t *testing.T) {
	t.Parallel()

	s := Status{
		AvariceDawnScore: 100,
		AvariceDuskScore: 200,
		GnosisDawnScore:  50,
		GnosisDuskScore:  75,
		StrifeDawnScore:  300,
		StrifeDuskScore:  150,
	}

	dawn, dusk := s.SealScore(SealAvarice)
	if dawn != 100 || dusk != 200 {
		t.Errorf("SealScore(Avarice) = (%d, %d), want (100, 200)", dawn, dusk)
	}

	dawn, dusk = s.SealScore(SealGnosis)
	if dawn != 50 || dusk != 75 {
		t.Errorf("SealScore(Gnosis) = (%d, %d), want (50, 75)", dawn, dusk)
	}

	dawn, dusk = s.SealScore(SealStrife)
	if dawn != 300 || dusk != 150 {
		t.Errorf("SealScore(Strife) = (%d, %d), want (300, 150)", dawn, dusk)
	}

	dawn, dusk = s.SealScore(SealNull)
	if dawn != 0 || dusk != 0 {
		t.Errorf("SealScore(Null) = (%d, %d), want (0, 0)", dawn, dusk)
	}
}

func TestCabalShortName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c    Cabal
		want string
	}{
		{CabalDawn, "dawn"},
		{CabalDusk, "dusk"},
		{CabalNull, ""},
	}
	for _, tt := range tests {
		got := CabalShortName(tt.c)
		if got != tt.want {
			t.Errorf("CabalShortName(%d) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestParseCabal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    string
		want Cabal
	}{
		{"dawn", CabalDawn},
		{"dusk", CabalDusk},
		{"", CabalNull},
		{"invalid", CabalNull},
	}
	for _, tt := range tests {
		got := ParseCabal(tt.s)
		if got != tt.want {
			t.Errorf("ParseCabal(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}
