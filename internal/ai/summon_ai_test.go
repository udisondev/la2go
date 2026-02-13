package ai

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// newTestSummon creates a test summon with sensible defaults.
// Template: Wolf (12077), level 15, 500 HP, 200 MP.
func newTestSummon(objectID, ownerID uint32) *model.Summon {
	tmpl := model.NewNpcTemplate(
		12077, "Wolf", "", 15,
		500, 200,
		100, 50, 80, 40,
		0, 120, 300,
		60, 120,
		100, 10,
	)
	return model.NewSummon(
		objectID, ownerID,
		model.SummonTypePet, 12077, tmpl,
		"Wolf", 15,
		500, 200,
		100, 50, 80, 40,
	)
}

// newTestObject creates a WorldObject at given location.
func newTestObject(objectID uint32, name string, x, y, z int32) *model.WorldObject {
	return model.NewWorldObject(objectID, name, model.NewLocation(x, y, z, 0))
}

func TestSummonAI_NewCreatesValidController(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	getObj := func(uint32) (*model.WorldObject, bool) { return nil, false }
	attackFn := func(*model.Summon, *model.WorldObject) {}

	ai := NewSummonAI(summon, getObj, attackFn)
	if ai == nil {
		t.Fatal("NewSummonAI() returned nil")
	}

	// Проверяем, что контроллер реализует интерфейс Controller
	var _ Controller = ai
}

func TestSummonAI_NpcReturnsNil(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)

	if got := ai.Npc(); got != nil {
		t.Errorf("Npc() = %v; want nil", got)
	}
}

func TestSummonAI_SummonReturnsSummon(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)

	if got := ai.Summon(); got != summon {
		t.Errorf("Summon() = %p; want %p", got, summon)
	}
}

func TestSummonAI_Start(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)

	ai.Start()

	if !summon.IsFollowing() {
		t.Errorf("after Start() IsFollowing() = false; want true")
	}
	if got := ai.CurrentIntention(); got != model.IntentionFollow {
		t.Errorf("after Start() CurrentIntention() = %v; want %v", got, model.IntentionFollow)
	}
}

func TestSummonAI_Stop(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)

	ai.Start()
	// Ставим таргет, чтобы убедиться, что Stop его очищает
	summon.SetTarget(9999)

	ai.Stop()

	if got := summon.Target(); got != 0 {
		t.Errorf("after Stop() Target() = %d; want 0", got)
	}
	if got := ai.CurrentIntention(); got != model.IntentionIdle {
		t.Errorf("after Stop() CurrentIntention() = %v; want %v", got, model.IntentionIdle)
	}
}

func TestSummonAI_SetIntentionCurrentIntention(t *testing.T) {
	tests := []struct {
		name      string
		intention model.Intention
	}{
		{"idle", model.IntentionIdle},
		{"attack", model.IntentionAttack},
		{"follow", model.IntentionFollow},
		{"active", model.IntentionActive},
		{"cast", model.IntentionCast},
		{"moveTo", model.IntentionMoveTo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summon := newTestSummon(5000, 1000)
			ai := NewSummonAI(summon, nil, nil)

			ai.SetIntention(tt.intention)
			if got := ai.CurrentIntention(); got != tt.intention {
				t.Errorf("CurrentIntention() = %v; want %v", got, tt.intention)
			}
		})
	}
}

func TestSummonAI_NotifyDamage(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*model.Summon, *SummonAI)
		attackerID      uint32
		damage          int32
		wantTarget      uint32
		wantIntention   model.Intention
		wantDescription string
	}{
		{
			name: "no_target_retaliates",
			setup: func(s *model.Summon, ai *SummonAI) {
				ai.Start()
			},
			attackerID:    2000,
			damage:        50,
			wantTarget:    2000,
			wantIntention: model.IntentionAttack,
		},
		{
			name: "has_target_no_change",
			setup: func(s *model.Summon, ai *SummonAI) {
				ai.Start()
				s.SetTarget(3000)
				ai.SetIntention(model.IntentionAttack)
			},
			attackerID:    2000,
			damage:        50,
			wantTarget:    3000,
			wantIntention: model.IntentionAttack,
		},
		{
			name: "dead_no_effect",
			setup: func(s *model.Summon, ai *SummonAI) {
				ai.Start()
				s.SetCurrentHP(0)
			},
			attackerID:    2000,
			damage:        50,
			wantTarget:    0,
			wantIntention: model.IntentionFollow,
		},
		{
			name: "not_running_no_effect",
			setup: func(s *model.Summon, ai *SummonAI) {
				// Не вызываем Start() — AI не запущен
			},
			attackerID:    2000,
			damage:        50,
			wantTarget:    0,
			wantIntention: model.IntentionIdle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summon := newTestSummon(5000, 1000)
			ai := NewSummonAI(summon, nil, nil)
			tt.setup(summon, ai)

			ai.NotifyDamage(tt.attackerID, tt.damage)

			if got := summon.Target(); got != tt.wantTarget {
				t.Errorf("Target() = %d; want %d", got, tt.wantTarget)
			}
			if got := ai.CurrentIntention(); got != tt.wantIntention {
				t.Errorf("CurrentIntention() = %v; want %v", got, tt.wantIntention)
			}
		})
	}
}

func TestSummonAI_Tick_Follow(t *testing.T) {
	tests := []struct {
		name     string
		ownerX   int32
		ownerY   int32
		summonX  int32
		summonY  int32
		wantMove bool // ожидаем ли изменение позиции саммона
		wantTele bool // ожидаем ли телепорт (>2000 units)
	}{
		{
			name:     "teleport_far_from_owner",
			ownerX:   10000,
			ownerY:   10000,
			summonX:  0,
			summonY:  0,
			wantMove: true,
			wantTele: true,
		},
		{
			name:     "move_towards_owner",
			ownerX:   200,
			ownerY:   200,
			summonX:  0,
			summonY:  0,
			wantMove: true,
			wantTele: false,
		},
		{
			name:     "stay_close_to_owner",
			ownerX:   30,
			ownerY:   30,
			summonX:  0,
			summonY:  0,
			wantMove: false,
			wantTele: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerObj := newTestObject(1000, "Owner", tt.ownerX, tt.ownerY, 0)
			getObj := func(objectID uint32) (*model.WorldObject, bool) {
				if objectID == 1000 {
					return ownerObj, true
				}
				return nil, false
			}

			summon := newTestSummon(5000, 1000)
			summon.SetLocation(model.NewLocation(tt.summonX, tt.summonY, 0, 0))
			ai := NewSummonAI(summon, getObj, nil)
			ai.Start()

			ai.Tick()

			loc := summon.Location()
			moved := loc.X != tt.summonX || loc.Y != tt.summonY

			if moved != tt.wantMove {
				t.Errorf("position changed = %v; want %v (summon at %d,%d)",
					moved, tt.wantMove, loc.X, loc.Y)
			}

			if tt.wantTele && moved {
				// При телепорте саммон должен оказаться рядом с владельцем (+50, +50)
				ownerLoc := ownerObj.Location()
				if loc.X != ownerLoc.X+50 || loc.Y != ownerLoc.Y+50 {
					t.Errorf("teleport position = (%d, %d); want (%d, %d)",
						loc.X, loc.Y, ownerLoc.X+50, ownerLoc.Y+50)
				}
			}

			if !tt.wantTele && tt.wantMove && moved {
				// При обычном перемещении саммон должен оказаться рядом (+30, +30)
				ownerLoc := ownerObj.Location()
				if loc.X != ownerLoc.X+30 || loc.Y != ownerLoc.Y+30 {
					t.Errorf("move position = (%d, %d); want (%d, %d)",
						loc.X, loc.Y, ownerLoc.X+30, ownerLoc.Y+30)
				}
			}
		})
	}
}

func TestSummonAI_Tick_Attack_CallsAttackFunc(t *testing.T) {
	// Цель в радиусе атаки — attackFunc вызывается
	targetObj := newTestObject(2000, "Target", 50, 50, 0)
	getObj := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == 2000 {
			return targetObj, true
		}
		return nil, false
	}

	var attackCalled bool
	var attackedSummon *model.Summon
	var attackedTarget *model.WorldObject
	attackFn := func(s *model.Summon, target *model.WorldObject) {
		attackCalled = true
		attackedSummon = s
		attackedTarget = target
	}

	summon := newTestSummon(5000, 1000)
	summon.SetLocation(model.NewLocation(50, 50, 0, 0))
	ai := NewSummonAI(summon, getObj, attackFn)
	ai.Start()

	ai.OrderAttack(2000)
	ai.Tick()

	if !attackCalled {
		t.Error("attackFunc was not called; want called")
	}
	if attackedSummon != summon {
		t.Errorf("attackFunc summon = %p; want %p", attackedSummon, summon)
	}
	if attackedTarget != targetObj {
		t.Errorf("attackFunc target = %p; want %p", attackedTarget, targetObj)
	}
}

func TestSummonAI_Tick_Attack_TargetNotFound(t *testing.T) {
	// getObjectFunc не находит цель — саммон возвращается к follow
	getObj := func(uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	var attackCalled bool
	attackFn := func(*model.Summon, *model.WorldObject) {
		attackCalled = true
	}

	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, getObj, attackFn)
	ai.Start()
	summon.SetFollow(true) // Чтобы returnToFollow переключил на Follow
	summon.SetTarget(9999)
	ai.SetIntention(model.IntentionAttack)

	ai.Tick()

	if attackCalled {
		t.Error("attackFunc was called; want not called (target not found)")
	}
	if got := summon.Target(); got != 0 {
		t.Errorf("Target() = %d; want 0 (cleared after missing target)", got)
	}
	if got := ai.CurrentIntention(); got != model.IntentionFollow {
		t.Errorf("CurrentIntention() = %v; want %v", got, model.IntentionFollow)
	}
}

func TestSummonAI_Tick_Attack_TargetCleared(t *testing.T) {
	// target = 0 — саммон возвращается к follow
	getObj := func(uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, getObj, nil)
	ai.Start()
	summon.SetFollow(true)
	ai.SetIntention(model.IntentionAttack)
	// target = 0 (не устанавливаем)

	ai.Tick()

	if got := ai.CurrentIntention(); got != model.IntentionFollow {
		t.Errorf("CurrentIntention() = %v; want %v", got, model.IntentionFollow)
	}
}

func TestSummonAI_Tick_Idle_FollowingMode(t *testing.T) {
	// Когда intention=Idle, но follow=true — переключается на Follow
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	ai.Start()

	summon.SetFollow(true)
	ai.SetIntention(model.IntentionIdle)

	ai.Tick()

	if got := ai.CurrentIntention(); got != model.IntentionFollow {
		t.Errorf("CurrentIntention() = %v; want %v (idle + following should switch to follow)",
			got, model.IntentionFollow)
	}
}

func TestSummonAI_OrderAttack(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	ai.Start()

	ai.OrderAttack(2000)

	if got := summon.Target(); got != 2000 {
		t.Errorf("Target() = %d; want 2000", got)
	}
	if got := ai.CurrentIntention(); got != model.IntentionAttack {
		t.Errorf("CurrentIntention() = %v; want %v", got, model.IntentionAttack)
	}
	if summon.IsFollowing() {
		t.Error("IsFollowing() = true; want false (attack disables follow)")
	}
}

func TestSummonAI_OrderFollow(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	ai.Start()

	// Сначала ставим атаку, потом отдаём приказ следовать
	summon.SetTarget(2000)
	ai.SetIntention(model.IntentionAttack)

	ai.OrderFollow()

	if got := summon.Target(); got != 0 {
		t.Errorf("Target() = %d; want 0 (cleared on OrderFollow)", got)
	}
	if !summon.IsFollowing() {
		t.Error("IsFollowing() = false; want true")
	}
	if got := ai.CurrentIntention(); got != model.IntentionFollow {
		t.Errorf("CurrentIntention() = %v; want %v", got, model.IntentionFollow)
	}
}

func TestSummonAI_OrderStop(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	ai.Start()

	// Сначала ставим атаку, потом приказ стоять
	summon.SetTarget(2000)
	ai.SetIntention(model.IntentionAttack)

	ai.OrderStop()

	if got := summon.Target(); got != 0 {
		t.Errorf("Target() = %d; want 0 (cleared on OrderStop)", got)
	}
	if summon.IsFollowing() {
		t.Error("IsFollowing() = true; want false")
	}
	if got := ai.CurrentIntention(); got != model.IntentionIdle {
		t.Errorf("CurrentIntention() = %v; want %v", got, model.IntentionIdle)
	}
}

func TestSummonAI_OrderAttack_WhenDead(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	ai.Start()

	summon.SetCurrentHP(0)
	ai.OrderAttack(2000)

	// Мёртвый саммон не должен принимать приказы на атаку
	if got := summon.Target(); got != 0 {
		t.Errorf("Target() = %d; want 0 (dead summon ignores OrderAttack)", got)
	}
}

func TestSummonAI_OrderAttack_WhenNotRunning(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	// Не вызываем Start()

	ai.OrderAttack(2000)

	if got := summon.Target(); got != 0 {
		t.Errorf("Target() = %d; want 0 (not running, ignores OrderAttack)", got)
	}
}

func TestSummonAI_OrderFollow_WhenNotRunning(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	// Не вызываем Start()
	summon.SetFollow(false)

	ai.OrderFollow()

	// Приказ следовать не должен работать, если AI не запущен
	if summon.IsFollowing() {
		t.Error("IsFollowing() = true; want false (not running, ignores OrderFollow)")
	}
}

func TestSummonAI_OrderStop_WhenNotRunning(t *testing.T) {
	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, nil, nil)
	// Не вызываем Start()
	summon.SetFollow(true)

	ai.OrderStop()

	// Приказ стоять не должен работать, если AI не запущен
	// follow остаётся true (начальное значение из NewSummon)
	if !summon.IsFollowing() {
		t.Error("IsFollowing() = false; want true (not running, ignores OrderStop)")
	}
}

func TestSummonAI_Tick_WhenDead(t *testing.T) {
	ownerObj := newTestObject(1000, "Owner", 5000, 5000, 0)
	getObj := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == 1000 {
			return ownerObj, true
		}
		return nil, false
	}

	summon := newTestSummon(5000, 1000)
	summon.SetLocation(model.NewLocation(0, 0, 0, 0))
	ai := NewSummonAI(summon, getObj, nil)
	ai.Start()

	summon.SetCurrentHP(0)

	ai.Tick()

	// Мёртвый саммон не должен двигаться
	loc := summon.Location()
	if loc.X != 0 || loc.Y != 0 {
		t.Errorf("dead summon moved to (%d, %d); want (0, 0)", loc.X, loc.Y)
	}
}

func TestSummonAI_Tick_WhenNotRunning(t *testing.T) {
	ownerObj := newTestObject(1000, "Owner", 5000, 5000, 0)
	getObj := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == 1000 {
			return ownerObj, true
		}
		return nil, false
	}

	summon := newTestSummon(5000, 1000)
	summon.SetLocation(model.NewLocation(0, 0, 0, 0))
	ai := NewSummonAI(summon, getObj, nil)
	// Не вызываем Start()
	ai.SetIntention(model.IntentionFollow)

	ai.Tick()

	loc := summon.Location()
	if loc.X != 0 || loc.Y != 0 {
		t.Errorf("not running summon moved to (%d, %d); want (0, 0)", loc.X, loc.Y)
	}
}

func TestSummonAI_Tick_Attack_OutOfRange(t *testing.T) {
	// Цель далеко — attackFunc не вызывается (саммон двигается, но не атакует)
	targetObj := newTestObject(2000, "FarTarget", 500, 500, 0)
	getObj := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == 2000 {
			return targetObj, true
		}
		return nil, false
	}

	var attackCalled bool
	attackFn := func(*model.Summon, *model.WorldObject) {
		attackCalled = true
	}

	summon := newTestSummon(5000, 1000)
	summon.SetLocation(model.NewLocation(0, 0, 0, 0))
	ai := NewSummonAI(summon, getObj, attackFn)
	ai.Start()
	ai.OrderAttack(2000)

	ai.Tick()

	if attackCalled {
		t.Error("attackFunc was called; want not called (target out of range)")
	}
	// Intention остаётся Attack — саммон пытается приблизиться
	if got := ai.CurrentIntention(); got != model.IntentionAttack {
		t.Errorf("CurrentIntention() = %v; want %v (still chasing)", got, model.IntentionAttack)
	}
}

func TestSummonAI_Tick_Follow_OwnerNotFound(t *testing.T) {
	// Владелец не найден через getObjectFunc — саммон не двигается
	getObj := func(uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	summon := newTestSummon(5000, 1000)
	summon.SetLocation(model.NewLocation(0, 0, 0, 0))
	ai := NewSummonAI(summon, getObj, nil)
	ai.Start()

	ai.Tick()

	loc := summon.Location()
	if loc.X != 0 || loc.Y != 0 {
		t.Errorf("summon moved to (%d, %d) without owner; want (0, 0)", loc.X, loc.Y)
	}
}

func TestSummonAI_ReturnToFollow_WhenNotFollowing(t *testing.T) {
	// target=0, follow=false — returnToFollow ставит Idle, а не Follow
	getObj := func(uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	summon := newTestSummon(5000, 1000)
	ai := NewSummonAI(summon, getObj, nil)
	ai.Start()

	summon.SetFollow(false)
	summon.ClearTarget()
	ai.SetIntention(model.IntentionAttack)

	ai.Tick()

	if got := ai.CurrentIntention(); got != model.IntentionIdle {
		t.Errorf("CurrentIntention() = %v; want %v (not following, should go idle)",
			got, model.IntentionIdle)
	}
}
