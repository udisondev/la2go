package sevensigns

import "testing"

func TestNewManager(t *testing.T) {
	t.Parallel()

	m := NewManager()
	if m.CurrentPeriod() != PeriodRecruitment {
		t.Errorf("CurrentPeriod() = %d, want %d", m.CurrentPeriod(), PeriodRecruitment)
	}
	if m.CurrentCycle() != 1 {
		t.Errorf("CurrentCycle() = %d, want 1", m.CurrentCycle())
	}
	if m.PreviousWinner() != CabalNull {
		t.Errorf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalNull)
	}
}

func TestManager_JoinCabal(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Присоединение к Dawn.
	if !m.JoinCabal(1, CabalDawn) {
		t.Error("JoinCabal(1, Dawn) = false, want true")
	}
	if m.PlayerCabal(1) != CabalDawn {
		t.Errorf("PlayerCabal(1) = %d, want %d", m.PlayerCabal(1), CabalDawn)
	}

	// Повторное присоединение — отказ.
	if m.JoinCabal(1, CabalDusk) {
		t.Error("JoinCabal(1, Dusk) = true, want false (already joined)")
	}

	// Некорректный cabal.
	if m.JoinCabal(2, CabalNull) {
		t.Error("JoinCabal(2, Null) = true, want false")
	}

	// Несуществующий игрок.
	if m.PlayerCabal(999) != CabalNull {
		t.Errorf("PlayerCabal(999) = %d, want %d", m.PlayerCabal(999), CabalNull)
	}
}

func TestManager_JoinCabal_WrongPeriod(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.TransitionPeriod() // → Competition

	if m.JoinCabal(1, CabalDawn) {
		t.Error("JoinCabal during Competition = true, want false")
	}
}

func TestManager_ChooseSeal(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.JoinCabal(1, CabalDawn)

	if !m.ChooseSeal(1, SealAvarice) {
		t.Error("ChooseSeal(1, Avarice) = false, want true")
	}
	if m.PlayerSeal(1) != SealAvarice {
		t.Errorf("PlayerSeal(1) = %d, want %d", m.PlayerSeal(1), SealAvarice)
	}

	// Нет cabal — отказ.
	if m.ChooseSeal(999, SealGnosis) {
		t.Error("ChooseSeal(999, Gnosis) = true, want false (no cabal)")
	}

	// Некорректная печать.
	if m.ChooseSeal(1, SealNull) {
		t.Error("ChooseSeal(1, Null) = true, want false")
	}
	if m.ChooseSeal(1, Seal(99)) {
		t.Error("ChooseSeal(1, 99) = true, want false")
	}
}

func TestManager_ContributeStones(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)

	contrib := m.ContributeStones(1, 10, 5, 2)
	wantContrib := int64(10*3 + 5*5 + 2*10)
	if contrib != wantContrib {
		t.Errorf("ContributeStones() = %d, want %d", contrib, wantContrib)
	}

	pd := m.PlayerData(1)
	if pd == nil {
		t.Fatal("PlayerData(1) = nil")
	}
	if pd.BlueStones != 10 {
		t.Errorf("BlueStones = %d, want 10", pd.BlueStones)
	}
	if pd.GreenStones != 5 {
		t.Errorf("GreenStones = %d, want 5", pd.GreenStones)
	}
	if pd.RedStones != 2 {
		t.Errorf("RedStones = %d, want 2", pd.RedStones)
	}
	if pd.ContributionScore != wantContrib {
		t.Errorf("ContributionScore = %d, want %d", pd.ContributionScore, wantContrib)
	}

	// Глобальные score (нормализованные: round(75/75*500)+0 = 500, round(0/75*500)+0 = 0).
	if m.DawnScore() != 500 {
		t.Errorf("DawnScore() = %d, want 500 (normalized)", m.DawnScore())
	}
	if m.DuskScore() != 0 {
		t.Errorf("DuskScore() = %d, want 0", m.DuskScore())
	}

	// Без cabal.
	got := m.ContributeStones(999, 1, 1, 1)
	if got != 0 {
		t.Errorf("ContributeStones(no cabal) = %d, want 0", got)
	}
}

func TestManager_CabalHighestScore(t *testing.T) {
	t.Parallel()

	m := NewManager()

	if m.CabalHighestScore() != CabalNull {
		t.Errorf("CabalHighestScore() = %d, want %d (tied at zero)", m.CabalHighestScore(), CabalNull)
	}

	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 10) // 100 pts

	m.JoinCabal(2, CabalDusk)
	m.ChooseSeal(2, SealGnosis)
	m.ContributeStones(2, 0, 0, 5) // 50 pts

	if m.CabalHighestScore() != CabalDawn {
		t.Errorf("CabalHighestScore() = %d, want %d", m.CabalHighestScore(), CabalDawn)
	}
}

func TestManager_TransitionPeriod(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Recruitment → Competition.
	p := m.TransitionPeriod()
	if p != PeriodCompetition {
		t.Errorf("TransitionPeriod() = %d, want %d", p, PeriodCompetition)
	}

	// Competition → Results.
	p = m.TransitionPeriod()
	if p != PeriodResults {
		t.Errorf("TransitionPeriod() = %d, want %d", p, PeriodResults)
	}

	// Results → SealValidation.
	p = m.TransitionPeriod()
	if p != PeriodSealValidation {
		t.Errorf("TransitionPeriod() = %d, want %d", p, PeriodSealValidation)
	}

	// SealValidation → Recruitment (new cycle).
	p = m.TransitionPeriod()
	if p != PeriodRecruitment {
		t.Errorf("TransitionPeriod() = %d, want %d", p, PeriodRecruitment)
	}
	if m.CurrentCycle() != 2 {
		t.Errorf("CurrentCycle() = %d, want 2", m.CurrentCycle())
	}
}

func TestManager_ComputeResults_DawnWins(t *testing.T) {
	t.Parallel()

	m := NewManager()

	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 20) // 200 pts

	m.JoinCabal(2, CabalDusk)
	m.ChooseSeal(2, SealGnosis)
	m.ContributeStones(2, 0, 0, 5) // 50 pts

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results (triggers computeResults)

	if m.PreviousWinner() != CabalDawn {
		t.Errorf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalDawn)
	}

	// Dawn победила и контрибьютила в Avarice → Dawn владеет Avarice.
	if m.SealOwner(SealAvarice) != CabalDawn {
		t.Errorf("SealOwner(Avarice) = %d, want %d", m.SealOwner(SealAvarice), CabalDawn)
	}

	// Dawn не контрибьютила в Gnosis → никто не владеет.
	if m.SealOwner(SealGnosis) != CabalNull {
		t.Errorf("SealOwner(Gnosis) = %d, want %d", m.SealOwner(SealGnosis), CabalNull)
	}
}

func TestManager_ComputeResults_Tied(t *testing.T) {
	t.Parallel()

	m := NewManager()
	// Никто не контрибьютил → ничья.

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	if m.PreviousWinner() != CabalNull {
		t.Errorf("PreviousWinner() = %d, want %d (tied)", m.PreviousWinner(), CabalNull)
	}

	for _, seal := range []Seal{SealAvarice, SealGnosis, SealStrife} {
		if m.SealOwner(seal) != CabalNull {
			t.Errorf("SealOwner(%d) = %d, want %d", seal, m.SealOwner(seal), CabalNull)
		}
	}
}

func TestManager_StartNewCycle_ResetsData(t *testing.T) {
	t.Parallel()

	m := NewManager()

	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 5, 5, 5)

	// Полный цикл.
	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results
	m.TransitionPeriod() // → SealValidation
	m.TransitionPeriod() // → Recruitment (new cycle)

	if m.CurrentCycle() != 2 {
		t.Errorf("CurrentCycle() = %d, want 2", m.CurrentCycle())
	}

	// Глобальные score обнулены.
	if m.DawnScore() != 0 {
		t.Errorf("DawnScore() = %d, want 0", m.DawnScore())
	}

	// Игроки обнулены.
	pd := m.PlayerData(1)
	if pd == nil {
		t.Fatal("PlayerData(1) = nil after new cycle")
	}
	if pd.Cabal != CabalNull {
		t.Errorf("Cabal = %d, want %d after new cycle", pd.Cabal, CabalNull)
	}
	if pd.ContributionScore != 0 {
		t.Errorf("ContributionScore = %d, want 0", pd.ContributionScore)
	}
}

func TestManager_AncientAdena(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.JoinCabal(1, CabalDawn)

	if !m.AddAncientAdena(1, 1000) {
		t.Error("AddAncientAdena = false, want true")
	}

	pd := m.PlayerData(1)
	if pd.AncientAdena != 1000 {
		t.Errorf("AncientAdena = %d, want 1000", pd.AncientAdena)
	}

	if !m.DeductAncientAdena(1, 500) {
		t.Error("DeductAncientAdena(500) = false, want true")
	}

	pd = m.PlayerData(1)
	if pd.AncientAdena != 500 {
		t.Errorf("AncientAdena = %d, want 500", pd.AncientAdena)
	}

	// Недостаточно.
	if m.DeductAncientAdena(1, 501) {
		t.Error("DeductAncientAdena(501) = true, want false (insufficient)")
	}

	// Неизвестный игрок.
	if m.AddAncientAdena(999, 100) {
		t.Error("AddAncientAdena(unknown) = true, want false")
	}
}

func TestManager_LoadStatus(t *testing.T) {
	t.Parallel()

	m := NewManager()
	s := Status{
		CurrentCycle:   5,
		ActivePeriod:   PeriodCompetition,
		PreviousWinner: CabalDusk,
		AvariceOwner:   CabalDawn,
	}
	m.LoadStatus(s)

	if m.CurrentCycle() != 5 {
		t.Errorf("CurrentCycle() = %d, want 5", m.CurrentCycle())
	}
	if m.CurrentPeriod() != PeriodCompetition {
		t.Errorf("CurrentPeriod() = %d, want %d", m.CurrentPeriod(), PeriodCompetition)
	}
	if m.PreviousWinner() != CabalDusk {
		t.Errorf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalDusk)
	}
	if m.SealOwner(SealAvarice) != CabalDawn {
		t.Errorf("SealOwner(Avarice) = %d, want %d", m.SealOwner(SealAvarice), CabalDawn)
	}
}

func TestManager_LoadPlayerData(t *testing.T) {
	t.Parallel()

	m := NewManager()
	pd := &PlayerData{
		CharID:            42,
		Cabal:             CabalDawn,
		Seal:              SealGnosis,
		BlueStones:        10,
		AncientAdena:      5000,
		ContributionScore: 30,
	}
	m.LoadPlayerData(pd)

	got := m.PlayerData(42)
	if got == nil {
		t.Fatal("PlayerData(42) = nil")
	}
	if got.Cabal != CabalDawn {
		t.Errorf("Cabal = %d, want %d", got.Cabal, CabalDawn)
	}
	if got.BlueStones != 10 {
		t.Errorf("BlueStones = %d, want 10", got.BlueStones)
	}
}

func TestManager_AllPlayerData(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.JoinCabal(1, CabalDawn)
	m.JoinCabal(2, CabalDusk)
	m.JoinCabal(3, CabalDawn)

	all := m.AllPlayerData()
	if len(all) != 3 {
		t.Errorf("AllPlayerData() len = %d, want 3", len(all))
	}
}

func TestManager_FestivalResults(t *testing.T) {
	t.Parallel()

	m := NewManager()

	m.SaveFestivalResult(&FestivalResult{
		FestivalID: 0,
		Cabal:      CabalDawn,
		Cycle:      1,
		Score:      100,
		Members:    "Player1,Player2",
	})
	m.SaveFestivalResult(&FestivalResult{
		FestivalID: 0,
		Cabal:      CabalDusk,
		Cycle:      1,
		Score:      200,
		Members:    "Player3,Player4",
	})

	results := m.FestivalResults(1)
	if len(results) != 2 {
		t.Errorf("FestivalResults(1) len = %d, want 2", len(results))
	}

	// Несуществующий цикл.
	results = m.FestivalResults(999)
	if results != nil {
		t.Errorf("FestivalResults(999) = %v, want nil", results)
	}
}

func TestManager_AddFestivalScore(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.AddFestivalScore(CabalDawn, 100)
	m.AddFestivalScore(CabalDusk, 50)

	if m.DawnScore() != 100 {
		t.Errorf("DawnScore() = %d, want 100", m.DawnScore())
	}
	if m.DuskScore() != 50 {
		t.Errorf("DuskScore() = %d, want 50", m.DuskScore())
	}
}

func TestManager_PeriodFlags(t *testing.T) {
	t.Parallel()

	m := NewManager()

	if !m.IsRecruitmentPeriod() {
		t.Error("IsRecruitmentPeriod() = false")
	}
	if m.IsCompetitionPeriod() {
		t.Error("IsCompetitionPeriod() = true")
	}

	m.TransitionPeriod()
	if !m.IsCompetitionPeriod() {
		t.Error("IsCompetitionPeriod() = false")
	}

	m.TransitionPeriod()
	if !m.IsCompResultsPeriod() {
		t.Error("IsCompResultsPeriod() = false")
	}

	m.TransitionPeriod()
	if !m.IsSealValidationPeriod() {
		t.Error("IsSealValidationPeriod() = false")
	}
}

func TestManager_ComputeResults_ThresholdBelow35(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// 3 Dawn members, only 1 chooses Avarice = 33% < 35% threshold.
	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 10) // 100 pts

	m.JoinCabal(2, CabalDawn)
	m.ChooseSeal(2, SealGnosis)
	m.ContributeStones(2, 0, 0, 10)

	m.JoinCabal(3, CabalDawn)
	m.ChooseSeal(3, SealGnosis)
	m.ContributeStones(3, 0, 0, 10)

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	if m.PreviousWinner() != CabalDawn {
		t.Errorf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalDawn)
	}

	// Avarice: prevOwner=NULL, winner=Dawn, dawnPercent=33% < 35% → NULL.
	if m.SealOwner(SealAvarice) != CabalNull {
		t.Errorf("SealOwner(Avarice) = %d, want %d (33%% < 35%% threshold)", m.SealOwner(SealAvarice), CabalNull)
	}

	// Gnosis: prevOwner=NULL, winner=Dawn, dawnPercent=67% >= 35% → Dawn.
	if m.SealOwner(SealGnosis) != CabalDawn {
		t.Errorf("SealOwner(Gnosis) = %d, want %d (67%% >= 35%%)", m.SealOwner(SealGnosis), CabalDawn)
	}
}

func TestManager_ComputeResults_RetainAt10Percent(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Cycle 1: Dawn wins and captures Avarice.
	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 10)

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results (Dawn captures Avarice)

	if m.SealOwner(SealAvarice) != CabalDawn {
		t.Fatalf("Cycle 1: SealOwner(Avarice) = %d, want %d", m.SealOwner(SealAvarice), CabalDawn)
	}

	m.TransitionPeriod() // → SealValidation
	m.TransitionPeriod() // → Recruitment (new cycle)

	// Cycle 2: 10 Dawn members, only 1 picks Avarice = 10% → retain.
	for i := int64(10); i < 20; i++ {
		m.JoinCabal(i, CabalDawn)
		if i == 10 {
			m.ChooseSeal(i, SealAvarice)
		} else {
			m.ChooseSeal(i, SealGnosis)
		}
		m.ContributeStones(i, 0, 0, 1)
	}

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	// Avarice: prevOwner=Dawn, winner=Dawn, dawnPercent=10% >= 10% → retained.
	if m.SealOwner(SealAvarice) != CabalDawn {
		t.Errorf("SealOwner(Avarice) = %d, want %d (10%% retain threshold)", m.SealOwner(SealAvarice), CabalDawn)
	}
}

func TestManager_ComputeResults_LoseAt9Percent(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Pre-load: Dawn owns Avarice.
	m.LoadStatus(Status{
		CurrentCycle: 1,
		ActivePeriod: PeriodRecruitment,
		AvariceOwner: CabalDawn,
	})

	// 11 Dawn members, only 1 picks Avarice = 9% < 10% → lose.
	for i := int64(1); i <= 11; i++ {
		m.JoinCabal(i, CabalDawn)
		if i == 1 {
			m.ChooseSeal(i, SealAvarice)
		} else {
			m.ChooseSeal(i, SealGnosis)
		}
		m.ContributeStones(i, 0, 0, 1)
	}

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	// Avarice: prevOwner=Dawn, winner=Dawn, dawnPercent=9% < 10% → NULL.
	if m.SealOwner(SealAvarice) != CabalNull {
		t.Errorf("SealOwner(Avarice) = %d, want %d (9%% < 10%% threshold)", m.SealOwner(SealAvarice), CabalNull)
	}
}

func TestManager_ComputeResults_OpponentCapture35(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Pre-load: Dawn owns Avarice.
	m.LoadStatus(Status{
		CurrentCycle: 1,
		ActivePeriod: PeriodRecruitment,
		AvariceOwner: CabalDawn,
	})

	// Dusk wins with 3 members, 2 chose Avarice = 67% >= 35% → Dusk captures.
	m.JoinCabal(1, CabalDusk)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 20)

	m.JoinCabal(2, CabalDusk)
	m.ChooseSeal(2, SealAvarice)
	m.ContributeStones(2, 0, 0, 20)

	m.JoinCabal(3, CabalDusk)
	m.ChooseSeal(3, SealGnosis)
	m.ContributeStones(3, 0, 0, 10)

	// Dawn has 1 member choosing Avarice (100% >= 10% → could retain, but Dusk wins overall).
	m.JoinCabal(4, CabalDawn)
	m.ChooseSeal(4, SealAvarice)
	m.ContributeStones(4, 0, 0, 1)

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	if m.PreviousWinner() != CabalDusk {
		t.Fatalf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalDusk)
	}

	// Avarice: prevOwner=Dawn, winner=Dusk, duskPercent=67% >= 35% → Dusk captures.
	if m.SealOwner(SealAvarice) != CabalDusk {
		t.Errorf("SealOwner(Avarice) = %d, want %d (opponent capture)", m.SealOwner(SealAvarice), CabalDusk)
	}
}

func TestManager_ComputeResults_OpponentFails35_DefenderRetains(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Pre-load: Dawn owns Avarice.
	m.LoadStatus(Status{
		CurrentCycle: 1,
		ActivePeriod: PeriodRecruitment,
		AvariceOwner: CabalDawn,
	})

	// Dusk wins overall, but only 1/4 = 25% chose Avarice < 35%.
	// Dawn has 1 member choosing Avarice (100% >= 10% → retain).
	m.JoinCabal(1, CabalDusk)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 20)

	m.JoinCabal(2, CabalDusk)
	m.ChooseSeal(2, SealGnosis)
	m.ContributeStones(2, 0, 0, 20)

	m.JoinCabal(3, CabalDusk)
	m.ChooseSeal(3, SealGnosis)
	m.ContributeStones(3, 0, 0, 20)

	m.JoinCabal(4, CabalDusk)
	m.ChooseSeal(4, SealGnosis)
	m.ContributeStones(4, 0, 0, 20)

	m.JoinCabal(5, CabalDawn)
	m.ChooseSeal(5, SealAvarice)
	m.ContributeStones(5, 0, 0, 1)

	m.TransitionPeriod() // → Competition
	m.TransitionPeriod() // → Results

	if m.PreviousWinner() != CabalDusk {
		t.Fatalf("PreviousWinner() = %d, want %d", m.PreviousWinner(), CabalDusk)
	}

	// Avarice: prevOwner=Dawn, winner=Dusk, duskPercent=25% < 35%,
	// dawnPercent=100% >= 10% → Dawn retains.
	if m.SealOwner(SealAvarice) != CabalDawn {
		t.Errorf("SealOwner(Avarice) = %d, want %d (defender retains)", m.SealOwner(SealAvarice), CabalDawn)
	}
}

func TestManager_NormalizedScoring(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Dawn: 400 stone + 100 festival, Dusk: 100 stone + 200 festival.
	m.JoinCabal(1, CabalDawn)
	m.ChooseSeal(1, SealAvarice)
	m.ContributeStones(1, 0, 0, 40) // 400 stone pts

	m.JoinCabal(2, CabalDusk)
	m.ChooseSeal(2, SealGnosis)
	m.ContributeStones(2, 0, 0, 10) // 100 stone pts

	m.AddFestivalScore(CabalDawn, 100)
	m.AddFestivalScore(CabalDusk, 200)

	// Dawn: round(400/500*500) + 100 = 400 + 100 = 500
	// Dusk: round(100/500*500) + 200 = 100 + 200 = 300
	if m.DawnScore() != 500 {
		t.Errorf("DawnScore() = %d, want 500", m.DawnScore())
	}
	if m.DuskScore() != 300 {
		t.Errorf("DuskScore() = %d, want 300", m.DuskScore())
	}
	if m.CabalHighestScore() != CabalDawn {
		t.Errorf("CabalHighestScore() = %d, want %d", m.CabalHighestScore(), CabalDawn)
	}
}

func TestManager_ContributeStones_DuskUpdatesGlobal(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.JoinCabal(1, CabalDusk)
	m.ChooseSeal(1, SealStrife)

	m.ContributeStones(1, 0, 0, 10) // 100 pts

	// Нормализованный score: round(100/100*500)+0 = 500.
	if m.DuskScore() != 500 {
		t.Errorf("DuskScore() = %d, want 500 (normalized)", m.DuskScore())
	}

	s := m.Status()
	_, dusk := s.SealScore(SealStrife)
	if dusk != 100 {
		t.Errorf("StrifeDuskScore = %d, want 100", dusk)
	}
}
