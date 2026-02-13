package fishing

import "testing"

func TestCalcDamage_NoShot(t *testing.T) {
	t.Parallel()

	// rodDmg=20, expertise=5, skillPower=10, rodLevel=20, noShot
	// base = 35, grade = 2.0, shot = 1.0 → 70
	got := CalcDamage(20, 5, 10, 20, false)
	if got != 70 {
		t.Errorf("CalcDamage(20,5,10,20,false) = %d; want 70", got)
	}
}

func TestCalcDamage_WithShot(t *testing.T) {
	t.Parallel()

	// rodDmg=20, expertise=5, skillPower=10, rodLevel=20, shot
	// base = 35, grade = 2.0, shot = 2.0 → 140
	got := CalcDamage(20, 5, 10, 20, true)
	if got != 140 {
		t.Errorf("CalcDamage(20,5,10,20,true) = %d; want 140", got)
	}
}

func TestCalcDamage_HighLevelRod(t *testing.T) {
	t.Parallel()

	// rodDmg=39, expertise=10, skillPower=20, rodLevel=80
	// base = 69, grade = 8.0, shot = 1.0 → 552
	got := CalcDamage(39, 10, 20, 80, false)
	if got != 552 {
		t.Errorf("CalcDamage(39,10,20,80,false) = %d; want 552", got)
	}
}

func TestCalcPenalty_Applied(t *testing.T) {
	t.Parallel()

	// expertiseLevel=1, skillLevel=3 → 1 <= (3-2) → penalty = 100*0.05 = 5
	got := CalcPenalty(100, 1, 3)
	if got != 5 {
		t.Errorf("CalcPenalty(100,1,3) = %d; want 5", got)
	}
}

func TestCalcPenalty_NotApplied(t *testing.T) {
	t.Parallel()

	// expertiseLevel=5, skillLevel=5 → 5 > (5-2) → no penalty
	got := CalcPenalty(100, 5, 5)
	if got != 0 {
		t.Errorf("CalcPenalty(100,5,5) = %d; want 0", got)
	}
}

func TestCalcPenalty_BoundaryCase(t *testing.T) {
	t.Parallel()

	// expertiseLevel=3, skillLevel=5 → 3 <= (5-2) → penalty = 200*0.05 = 10
	got := CalcPenalty(200, 3, 5)
	if got != 10 {
		t.Errorf("CalcPenalty(200,3,5) = %d; want 10", got)
	}
}

func TestCalcDamage_ZeroValues(t *testing.T) {
	t.Parallel()

	got := CalcDamage(0, 0, 0, 0, false)
	if got != 0 {
		t.Errorf("CalcDamage(0,0,0,0,false) = %d; want 0", got)
	}
}
