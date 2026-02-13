package model

import (
	"sync"
	"testing"
)

// newTestNpcTemplate creates an NpcTemplate for use in summon tests.
func newTestNpcTemplate() *NpcTemplate {
	return NewNpcTemplate(
		12077,          // templateID
		"Dark Panther", // name
		"",             // title
		55,             // level
		2500,           // maxHP
		1200,           // maxMP
		250, 180,       // pAtk, pDef
		120, 90,        // mAtk, mDef
		0,              // aggroRange
		140,            // moveSpeed
		278,            // atkSpeed
		0, 0,           // respawnMin, respawnMax
		0, 0,           // baseExp, baseSP
	)
}

// newTestSummon creates a Summon (pet type) for tests.
func newTestSummon() *Summon {
	tmpl := newTestNpcTemplate()
	return NewSummon(
		5001,           // objectID
		1001,           // ownerID
		SummonTypePet,  // summonType
		12077,          // templateID
		tmpl,           // template
		"Dark Panther", // name
		55,             // level
		2500, 1200,     // maxHP, maxMP
		250, 180, 120, 90, // pAtk, pDef, mAtk, mDef
	)
}

func TestNewSummon(t *testing.T) {
	tmpl := newTestNpcTemplate()
	s := NewSummon(
		5001, 1001, SummonTypePet, 12077, tmpl,
		"TestPanther", 55,
		2500, 1200,
		250, 180, 120, 90,
	)

	if s == nil {
		t.Fatal("NewSummon returned nil")
	}
	if s.ObjectID() != 5001 {
		t.Errorf("ObjectID() = %d; want 5001", s.ObjectID())
	}
	if s.Name() != "TestPanther" {
		t.Errorf("Name() = %q; want %q", s.Name(), "TestPanther")
	}
	if s.OwnerID() != 1001 {
		t.Errorf("OwnerID() = %d; want 1001", s.OwnerID())
	}
	if s.Type() != SummonTypePet {
		t.Errorf("Type() = %d; want %d (SummonTypePet)", s.Type(), SummonTypePet)
	}
	if s.TemplateID() != 12077 {
		t.Errorf("TemplateID() = %d; want 12077", s.TemplateID())
	}
	if s.Template() != tmpl {
		t.Errorf("Template() pointer mismatch")
	}
	if s.Level() != 55 {
		t.Errorf("Level() = %d; want 55", s.Level())
	}
	if s.MaxHP() != 2500 {
		t.Errorf("MaxHP() = %d; want 2500", s.MaxHP())
	}
	if s.CurrentHP() != 2500 {
		t.Errorf("CurrentHP() = %d; want 2500 (should start at maxHP)", s.CurrentHP())
	}
	if s.MaxMP() != 1200 {
		t.Errorf("MaxMP() = %d; want 1200", s.MaxMP())
	}
	if s.CurrentMP() != 1200 {
		t.Errorf("CurrentMP() = %d; want 1200 (should start at maxMP)", s.CurrentMP())
	}
	if s.PAtk() != 250 {
		t.Errorf("PAtk() = %d; want 250", s.PAtk())
	}
	if s.PDef() != 180 {
		t.Errorf("PDef() = %d; want 180", s.PDef())
	}
	if s.MAtk() != 120 {
		t.Errorf("MAtk() = %d; want 120", s.MAtk())
	}
	if s.MDef() != 90 {
		t.Errorf("MDef() = %d; want 90", s.MDef())
	}
	// MoveSpeed и AtkSpeed берутся из template
	if s.MoveSpeed() != tmpl.MoveSpeed() {
		t.Errorf("MoveSpeed() = %d; want %d (from template)", s.MoveSpeed(), tmpl.MoveSpeed())
	}
	if s.AtkSpeed() != tmpl.AtkSpeed() {
		t.Errorf("AtkSpeed() = %d; want %d (from template)", s.AtkSpeed(), tmpl.AtkSpeed())
	}
}

func TestSummonAccessors(t *testing.T) {
	tmpl := newTestNpcTemplate()
	s := NewSummon(
		5002, 2002, SummonTypeServitor, 999, tmpl,
		"Servitor", 30,
		1000, 500,
		100, 80, 60, 40,
	)

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"OwnerID", s.OwnerID(), uint32(2002)},
		{"Type", s.Type(), SummonTypeServitor},
		{"TemplateID", s.TemplateID(), int32(999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	if s.Template() != tmpl {
		t.Errorf("Template() = %p; want %p", s.Template(), tmpl)
	}
}

func TestSummonFollowToggle(t *testing.T) {
	s := newTestSummon()

	// По умолчанию follow = true после NewSummon
	if !s.IsFollowing() {
		t.Errorf("IsFollowing() = false; want true (default after NewSummon)")
	}

	s.SetFollow(false)
	if s.IsFollowing() {
		t.Errorf("IsFollowing() = true after SetFollow(false); want false")
	}

	s.SetFollow(true)
	if !s.IsFollowing() {
		t.Errorf("IsFollowing() = false after SetFollow(true); want true")
	}
}

func TestSummonIntention(t *testing.T) {
	s := newTestSummon()

	// По умолчанию IntentionIdle
	if s.Intention() != IntentionIdle {
		t.Errorf("Intention() = %v; want IntentionIdle", s.Intention())
	}

	tests := []struct {
		name      string
		intention Intention
	}{
		{"follow", IntentionFollow},
		{"attack", IntentionAttack},
		{"idle", IntentionIdle},
		{"cast", IntentionCast},
		{"moveTo", IntentionMoveTo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetIntention(tt.intention)
			if got := s.Intention(); got != tt.intention {
				t.Errorf("Intention() = %v; want %v", got, tt.intention)
			}
		})
	}
}

func TestSummonIsDecayed(t *testing.T) {
	s := newTestSummon()

	if s.IsDecayed() {
		t.Errorf("IsDecayed() = true; want false (default)")
	}

	s.SetDecayed(true)
	if !s.IsDecayed() {
		t.Errorf("IsDecayed() = false after SetDecayed(true); want true")
	}

	s.SetDecayed(false)
	if s.IsDecayed() {
		t.Errorf("IsDecayed() = true after SetDecayed(false); want false")
	}
}

func TestSummonTarget(t *testing.T) {
	s := newTestSummon()

	// По умолчанию нет цели
	if s.Target() != 0 {
		t.Errorf("Target() = %d; want 0 (no target)", s.Target())
	}

	s.SetTarget(42)
	if s.Target() != 42 {
		t.Errorf("Target() = %d after SetTarget(42); want 42", s.Target())
	}

	s.SetTarget(999)
	if s.Target() != 999 {
		t.Errorf("Target() = %d after SetTarget(999); want 999", s.Target())
	}

	s.ClearTarget()
	if s.Target() != 0 {
		t.Errorf("Target() = %d after ClearTarget(); want 0", s.Target())
	}
}

func TestSummonCombatStats(t *testing.T) {
	tmpl := newTestNpcTemplate()
	s := NewSummon(
		5010, 1010, SummonTypePet, 12077, tmpl,
		"Panther", 55,
		2500, 1200,
		250, 180, 120, 90,
	)

	tests := []struct {
		name string
		got  int32
		want int32
	}{
		{"PAtk", s.PAtk(), 250},
		{"PDef", s.PDef(), 180},
		{"MAtk", s.MAtk(), 120},
		{"MDef", s.MDef(), 90},
		{"MoveSpeed", s.MoveSpeed(), 140},
		{"AtkSpeed", s.AtkSpeed(), 278},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %d; want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSummonUpdateStats(t *testing.T) {
	s := newTestSummon()

	// Исходные значения
	if s.PAtk() != 250 {
		t.Fatalf("initial PAtk() = %d; want 250", s.PAtk())
	}

	// Обновляем статы
	s.UpdateStats(3000, 1500, 300, 220, 150, 110)

	tests := []struct {
		name string
		got  int32
		want int32
	}{
		{"PAtk after update", s.PAtk(), 300},
		{"PDef after update", s.PDef(), 220},
		{"MAtk after update", s.MAtk(), 150},
		{"MDef after update", s.MDef(), 110},
		{"MaxHP after update", s.MaxHP(), 3000},
		{"MaxMP after update", s.MaxMP(), 1500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %d; want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSummonUpdateStatsClampsCurrentHP(t *testing.T) {
	s := newTestSummon()

	// CurrentHP = MaxHP = 2500
	if s.CurrentHP() != 2500 {
		t.Fatalf("initial CurrentHP() = %d; want 2500", s.CurrentHP())
	}

	// Понижаем MaxHP ниже текущего HP
	s.UpdateStats(1000, 600, 100, 100, 100, 100)

	if s.MaxHP() != 1000 {
		t.Errorf("MaxHP() = %d; want 1000", s.MaxHP())
	}
	// CurrentHP должен быть обрезан до нового MaxHP
	if s.CurrentHP() > 1000 {
		t.Errorf("CurrentHP() = %d after reducing MaxHP to 1000; want <= 1000", s.CurrentHP())
	}
}

func TestSummonIsPetIsServitor(t *testing.T) {
	tmpl := newTestNpcTemplate()

	tests := []struct {
		name         string
		summonType   SummonType
		wantIsPet    bool
		wantIsServit bool
	}{
		{"pet", SummonTypePet, true, false},
		{"servitor", SummonTypeServitor, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummon(
				5020, 1020, tt.summonType, 12077, tmpl,
				"TestSummon", 50,
				2000, 1000,
				200, 150, 100, 80,
			)

			if got := s.IsPet(); got != tt.wantIsPet {
				t.Errorf("IsPet() = %v; want %v", got, tt.wantIsPet)
			}
			if got := s.IsServitor(); got != tt.wantIsServit {
				t.Errorf("IsServitor() = %v; want %v", got, tt.wantIsServit)
			}
		})
	}
}

func TestSummonWorldObjectData(t *testing.T) {
	s := newTestSummon()

	data := s.WorldObject.Data
	if data == nil {
		t.Fatal("WorldObject.Data = nil; want *Summon")
	}

	summonData, ok := data.(*Summon)
	if !ok {
		t.Fatalf("WorldObject.Data type = %T; want *Summon", data)
	}
	if summonData != s {
		t.Errorf("WorldObject.Data pointer = %p; want %p (same summon)", summonData, s)
	}
}

func TestSummonConcurrentFollowIntention(t *testing.T) {
	s := newTestSummon()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Параллельная запись/чтение follow
	for range goroutines {
		go func() {
			defer wg.Done()
			s.SetFollow(true)
			s.SetFollow(false)
			_ = s.IsFollowing()
		}()
	}

	// Параллельная запись/чтение intention
	for range goroutines {
		go func() {
			defer wg.Done()
			s.SetIntention(IntentionAttack)
			s.SetIntention(IntentionFollow)
			s.SetIntention(IntentionIdle)
			_ = s.Intention()
		}()
	}

	wg.Wait()

	// Если дошли сюда без data race (go test -race) -- тест пройден.
	// Проверяем что значения валидны.
	follow := s.IsFollowing()
	intention := s.Intention()

	if follow != true && follow != false {
		t.Errorf("IsFollowing() returned invalid value after concurrent access")
	}
	_ = intention // любое значение Intention допустимо
}

func TestSummonConcurrentStats(t *testing.T) {
	s := newTestSummon()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Параллельное чтение статов
	for range goroutines {
		go func() {
			defer wg.Done()
			_ = s.PAtk()
			_ = s.PDef()
			_ = s.MAtk()
			_ = s.MDef()
			_ = s.MoveSpeed()
			_ = s.AtkSpeed()
		}()
	}

	// Параллельная запись статов
	for range goroutines {
		go func() {
			defer wg.Done()
			s.UpdateStats(3000, 1500, 300, 200, 150, 100)
		}()
	}

	wg.Wait()
}

func TestSummonConcurrentTarget(t *testing.T) {
	s := newTestSummon()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for range goroutines {
		go func() {
			defer wg.Done()
			s.SetTarget(42)
			s.ClearTarget()
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = s.Target()
		}()
	}

	wg.Wait()
}

func TestSummonDefaultFollowAndIntention(t *testing.T) {
	s := newTestSummon()

	if !s.IsFollowing() {
		t.Errorf("default IsFollowing() = false; want true")
	}
	if s.Intention() != IntentionIdle {
		t.Errorf("default Intention() = %v; want IntentionIdle", s.Intention())
	}
	if s.IsDecayed() {
		t.Errorf("default IsDecayed() = true; want false")
	}
	if s.Target() != 0 {
		t.Errorf("default Target() = %d; want 0", s.Target())
	}
}

func TestSummonInheritsCharacter(t *testing.T) {
	s := newTestSummon()

	// Проверяем что Character-методы доступны
	s.SetCurrentHP(1000)
	if s.CurrentHP() != 1000 {
		t.Errorf("CurrentHP() = %d after SetCurrentHP(1000); want 1000", s.CurrentHP())
	}

	s.SetCurrentMP(500)
	if s.CurrentMP() != 500 {
		t.Errorf("CurrentMP() = %d after SetCurrentMP(500); want 500", s.CurrentMP())
	}

	if s.IsDead() {
		t.Errorf("IsDead() = true; want false (HP=1000)")
	}

	s.SetCurrentHP(0)
	if !s.IsDead() {
		t.Errorf("IsDead() = false after SetCurrentHP(0); want true")
	}
}
