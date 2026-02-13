package instance

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func testTemplate(id int32) *Template {
	return &Template{
		ID:         id,
		Name:       "Test Instance",
		Duration:   30 * time.Minute,
		MaxPlayers: 9,
		MinLevel:   20,
		MaxLevel:   70,
		Cooldown:   1 * time.Hour,
	}
}

func TestManager_RegisterTemplate(t *testing.T) {
	m := NewManager()

	tmpl := testTemplate(1)
	if err := m.RegisterTemplate(tmpl); err != nil {
		t.Fatalf("RegisterTemplate() error = %v", err)
	}

	if m.TemplateCount() != 1 {
		t.Errorf("TemplateCount() = %d; want 1", m.TemplateCount())
	}

	got := m.Template(1)
	if got == nil {
		t.Fatal("Template(1) = nil; want non-nil")
	}
	if got.Name != "Test Instance" {
		t.Errorf("Template(1).Name = %q; want %q", got.Name, "Test Instance")
	}
}

func TestManager_RegisterTemplate_Invalid(t *testing.T) {
	m := NewManager()

	if err := m.RegisterTemplate(&Template{ID: 0, Name: "Test"}); err == nil {
		t.Error("RegisterTemplate() with ID=0 should fail")
	}
}

func TestManager_CreateInstance(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, err := m.CreateInstance(1, 1000)
	if err != nil {
		t.Fatalf("CreateInstance() error = %v", err)
	}

	if inst.TemplateID() != 1 {
		t.Errorf("TemplateID() = %d; want 1", inst.TemplateID())
	}
	if inst.OwnerID() != 1000 {
		t.Errorf("OwnerID() = %d; want 1000", inst.OwnerID())
	}
	if inst.State() != StateActive {
		t.Errorf("State() = %v; want ACTIVE", inst.State())
	}
	if m.InstanceCount() != 1 {
		t.Errorf("InstanceCount() = %d; want 1", m.InstanceCount())
	}
}

func TestManager_CreateInstance_TemplateNotFound(t *testing.T) {
	m := NewManager()

	_, err := m.CreateInstance(999, 1000)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("CreateInstance() error = %v; want ErrTemplateNotFound", err)
	}
}

func TestManager_EnterExitInstance(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, err := m.CreateInstance(1, 1000)
	if err != nil {
		t.Fatal(err)
	}

	// Вход игрока.
	if err := m.EnterInstance(inst.ID(), 1000, 100, 50); err != nil {
		t.Fatalf("EnterInstance() error = %v", err)
	}

	if !m.IsInInstance(1000) {
		t.Error("IsInInstance(1000) = false; want true")
	}
	if inst.PlayerCount() != 1 {
		t.Errorf("PlayerCount() = %d; want 1", inst.PlayerCount())
	}

	pi := m.GetPlayerInstance(1000)
	if pi == nil || pi.ID() != inst.ID() {
		t.Errorf("GetPlayerInstance(1000) = %v; want instance %d", pi, inst.ID())
	}

	// Выход игрока.
	exitInst, err := m.ExitInstance(1000, 100)
	if err != nil {
		t.Fatalf("ExitInstance() error = %v", err)
	}
	if exitInst.ID() != inst.ID() {
		t.Errorf("ExitInstance() instance ID = %d; want %d", exitInst.ID(), inst.ID())
	}
	if m.IsInInstance(1000) {
		t.Error("IsInInstance(1000) after exit = true; want false")
	}
}

func TestManager_EnterInstance_AlreadyInInstance(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)
	if err := m.EnterInstance(inst.ID(), 1000, 100, 50); err != nil {
		t.Fatal(err)
	}

	err := m.EnterInstance(inst.ID(), 1000, 100, 50)
	if !errors.Is(err, ErrAlreadyInInstance) {
		t.Errorf("EnterInstance() second time error = %v; want ErrAlreadyInInstance", err)
	}
}

func TestManager_EnterInstance_LevelRestriction(t *testing.T) {
	m := NewManager()
	tmpl := testTemplate(1)
	tmpl.MinLevel = 30
	tmpl.MaxLevel = 60
	if err := m.RegisterTemplate(tmpl); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)

	// Слишком низкий уровень.
	err := m.EnterInstance(inst.ID(), 2000, 200, 20)
	if !errors.Is(err, ErrLevelTooLow) {
		t.Errorf("EnterInstance(level=20) error = %v; want ErrLevelTooLow", err)
	}

	// Слишком высокий уровень.
	err = m.EnterInstance(inst.ID(), 2000, 200, 70)
	if !errors.Is(err, ErrLevelTooHigh) {
		t.Errorf("EnterInstance(level=70) error = %v; want ErrLevelTooHigh", err)
	}

	// Подходящий уровень.
	if err := m.EnterInstance(inst.ID(), 2000, 200, 50); err != nil {
		t.Errorf("EnterInstance(level=50) error = %v; want nil", err)
	}
}

func TestManager_EnterInstance_InstanceFull(t *testing.T) {
	m := NewManager()
	tmpl := testTemplate(1)
	tmpl.MaxPlayers = 2
	if err := m.RegisterTemplate(tmpl); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)

	if err := m.EnterInstance(inst.ID(), 1000, 100, 50); err != nil {
		t.Fatal(err)
	}
	if err := m.EnterInstance(inst.ID(), 2000, 200, 50); err != nil {
		t.Fatal(err)
	}

	err := m.EnterInstance(inst.ID(), 3000, 300, 50)
	if !errors.Is(err, ErrInstanceFull) {
		t.Errorf("EnterInstance() when full error = %v; want ErrInstanceFull", err)
	}
}

func TestManager_EnterInstance_Destroyed(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)
	inst.SetState(StateDestroyed)

	err := m.EnterInstance(inst.ID(), 1000, 100, 50)
	if !errors.Is(err, ErrInstanceDestroyed) {
		t.Errorf("EnterInstance() on destroyed error = %v; want ErrInstanceDestroyed", err)
	}
}

func TestManager_EnterInstance_NotFound(t *testing.T) {
	m := NewManager()

	err := m.EnterInstance(999, 1000, 100, 50)
	if !errors.Is(err, ErrInstanceNotFound) {
		t.Errorf("EnterInstance(999) error = %v; want ErrInstanceNotFound", err)
	}
}

func TestManager_ExitInstance_NotInInstance(t *testing.T) {
	m := NewManager()

	_, err := m.ExitInstance(1000, 100)
	if !errors.Is(err, ErrNotInInstance) {
		t.Errorf("ExitInstance() when not inside error = %v; want ErrNotInInstance", err)
	}
}

func TestManager_DestroyInstance(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)
	if err := m.EnterInstance(inst.ID(), 1000, 100, 50); err != nil {
		t.Fatal(err)
	}

	// Уничтожение удаляет игроков из byPlayer.
	if err := m.DestroyInstance(inst.ID()); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}

	if inst.State() != StateDestroyed {
		t.Errorf("State() = %v; want DESTROYED", inst.State())
	}
	if m.InstanceCount() != 0 {
		t.Errorf("InstanceCount() = %d; want 0", m.InstanceCount())
	}
	if m.IsInInstance(1000) {
		t.Error("IsInInstance(1000) after destroy = true; want false")
	}
}

func TestManager_DestroyInstance_NotFound(t *testing.T) {
	m := NewManager()

	err := m.DestroyInstance(999)
	if !errors.Is(err, ErrInstanceNotFound) {
		t.Errorf("DestroyInstance(999) error = %v; want ErrInstanceNotFound", err)
	}
}

func TestManager_Cooldown(t *testing.T) {
	m := NewManager()

	// Нет cooldown.
	onCD, _ := m.IsOnCooldown(100, 1)
	if onCD {
		t.Error("IsOnCooldown() before set = true; want false")
	}

	// Устанавливаем.
	m.SetCooldown(100, 1, 1*time.Hour)

	onCD, expire := m.IsOnCooldown(100, 1)
	if !onCD {
		t.Error("IsOnCooldown() after set = false; want true")
	}
	if expire.IsZero() {
		t.Error("expire time is zero")
	}
	if time.Until(expire) < 59*time.Minute {
		t.Errorf("expire too soon: %v", expire)
	}

	// Другой персонаж — нет cooldown.
	onCD, _ = m.IsOnCooldown(200, 1)
	if onCD {
		t.Error("IsOnCooldown(200, 1) = true; want false")
	}

	// Другой шаблон — нет cooldown.
	onCD, _ = m.IsOnCooldown(100, 2)
	if onCD {
		t.Error("IsOnCooldown(100, 2) = true; want false")
	}

	// Очистка.
	m.ClearCooldown(100, 1)
	onCD, _ = m.IsOnCooldown(100, 1)
	if onCD {
		t.Error("IsOnCooldown() after clear = true; want false")
	}
}

func TestManager_CooldownOnExit(t *testing.T) {
	m := NewManager()
	tmpl := testTemplate(1)
	tmpl.Cooldown = 2 * time.Hour
	if err := m.RegisterTemplate(tmpl); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)
	if err := m.EnterInstance(inst.ID(), 1000, 100, 50); err != nil {
		t.Fatal(err)
	}

	if _, err := m.ExitInstance(1000, 100); err != nil {
		t.Fatal(err)
	}

	// После выхода должен быть cooldown.
	onCD, _ := m.IsOnCooldown(100, 1)
	if !onCD {
		t.Error("IsOnCooldown() after exit = false; want true")
	}

	// Повторный вход должен быть отклонён.
	inst2, _ := m.CreateInstance(1, 2000)
	err := m.EnterInstance(inst2.ID(), 1000, 100, 50)
	if !errors.Is(err, ErrOnCooldown) {
		t.Errorf("EnterInstance() on cooldown error = %v; want ErrOnCooldown", err)
	}
}

func TestManager_ClearExpiredCooldowns(t *testing.T) {
	m := NewManager()

	// Устанавливаем просроченный cooldown.
	m.mu.Lock()
	m.cooldowns[cooldownKey{characterID: 100, templateID: 1}] = time.Now().Add(-1 * time.Hour).UnixNano()
	m.cooldowns[cooldownKey{characterID: 200, templateID: 1}] = time.Now().Add(1 * time.Hour).UnixNano()
	m.mu.Unlock()

	removed := m.ClearExpiredCooldowns()
	if removed != 1 {
		t.Errorf("ClearExpiredCooldowns() = %d; want 1", removed)
	}

	onCD, _ := m.IsOnCooldown(100, 1)
	if onCD {
		t.Error("expired cooldown still active")
	}

	onCD, _ = m.IsOnCooldown(200, 1)
	if !onCD {
		t.Error("valid cooldown was removed")
	}
}

func TestManager_ExportLoadCooldowns(t *testing.T) {
	m := NewManager()
	m.SetCooldown(100, 1, 2*time.Hour)
	m.SetCooldown(200, 2, 4*time.Hour)

	entries := m.ExportCooldowns()
	if len(entries) != 2 {
		t.Fatalf("ExportCooldowns() len = %d; want 2", len(entries))
	}

	// Загружаем в новый менеджер.
	m2 := NewManager()
	m2.LoadCooldowns(entries)

	onCD, _ := m2.IsOnCooldown(100, 1)
	if !onCD {
		t.Error("IsOnCooldown(100, 1) after load = false; want true")
	}
	onCD, _ = m2.IsOnCooldown(200, 2)
	if !onCD {
		t.Error("IsOnCooldown(200, 2) after load = false; want true")
	}
}

func TestManager_LoadCooldowns_SkipsExpired(t *testing.T) {
	m := NewManager()

	entries := []CooldownEntry{
		{CharacterID: 100, TemplateID: 1, ExpireNano: time.Now().Add(-1 * time.Hour).UnixNano()},
		{CharacterID: 200, TemplateID: 2, ExpireNano: time.Now().Add(2 * time.Hour).UnixNano()},
	}

	m.LoadCooldowns(entries)

	onCD, _ := m.IsOnCooldown(100, 1)
	if onCD {
		t.Error("expired cooldown should not be loaded")
	}
	onCD, _ = m.IsOnCooldown(200, 2)
	if !onCD {
		t.Error("valid cooldown should be loaded")
	}
}

func TestManager_GetInstance(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)

	got := m.GetInstance(inst.ID())
	if got == nil {
		t.Fatal("GetInstance() = nil; want non-nil")
	}
	if got.ID() != inst.ID() {
		t.Errorf("GetInstance().ID() = %d; want %d", got.ID(), inst.ID())
	}

	if m.GetInstance(999) != nil {
		t.Error("GetInstance(999) should be nil")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	tmpl := testTemplate(1)
	tmpl.MaxPlayers = 0 // unlimited
	tmpl.Cooldown = 0   // no cooldown
	if err := m.RegisterTemplate(tmpl); err != nil {
		t.Fatal(err)
	}

	inst, _ := m.CreateInstance(1, 1000)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			objID := uint32(i + 1)
			charID := int64(i + 1)

			if err := m.EnterInstance(inst.ID(), objID, charID, 50); err != nil {
				return // concurrent enter may race, OK
			}
			_ = m.IsInInstance(objID)
			_ = m.GetPlayerInstance(objID)
			_, _ = m.ExitInstance(objID, charID)
		})
	}
	wg.Wait()

	// После всех выходов инстанс должен быть пуст.
	if inst.PlayerCount() != 0 {
		t.Errorf("PlayerCount() = %d; want 0 after all exits", inst.PlayerCount())
	}
}

func TestManager_MultipleInstances(t *testing.T) {
	m := NewManager()
	if err := m.RegisterTemplate(testTemplate(1)); err != nil {
		t.Fatal(err)
	}
	tmpl2 := testTemplate(2)
	tmpl2.Name = "Second Instance"
	if err := m.RegisterTemplate(tmpl2); err != nil {
		t.Fatal(err)
	}

	inst1, _ := m.CreateInstance(1, 1000)
	inst2, _ := m.CreateInstance(2, 2000)

	if m.InstanceCount() != 2 {
		t.Errorf("InstanceCount() = %d; want 2", m.InstanceCount())
	}

	// Игрок в первом инстансе.
	if err := m.EnterInstance(inst1.ID(), 1000, 100, 50); err != nil {
		t.Fatal(err)
	}

	// Тот же игрок не может войти во второй.
	err := m.EnterInstance(inst2.ID(), 1000, 100, 50)
	if !errors.Is(err, ErrAlreadyInInstance) {
		t.Errorf("EnterInstance() second instance error = %v; want ErrAlreadyInInstance", err)
	}

	// Другой игрок может.
	if err := m.EnterInstance(inst2.ID(), 2000, 200, 50); err != nil {
		t.Fatalf("EnterInstance() error = %v", err)
	}

	// Проверяем что они в разных инстансах.
	pi1 := m.GetPlayerInstance(1000)
	pi2 := m.GetPlayerInstance(2000)
	if pi1.ID() == pi2.ID() {
		t.Error("players should be in different instances")
	}
}
