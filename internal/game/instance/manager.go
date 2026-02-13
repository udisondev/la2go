package instance

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Manager manages all active instances and per-player cooldowns.
// Thread-safe for concurrent access.
type Manager struct {
	mu        sync.RWMutex
	instances map[int32]*Instance   // instanceID → Instance
	byPlayer  map[uint32]int32      // objectID → instanceID
	templates map[int32]*Template   // templateID → Template
	cooldowns map[cooldownKey]int64 // (charID, templateID) → unix timestamp (expire time)
	nextID    atomic.Int32
}

// cooldownKey identifies a per-character, per-template cooldown.
type cooldownKey struct {
	characterID int64
	templateID  int32
}

// NewManager creates a new instance manager.
func NewManager() *Manager {
	return &Manager{
		instances: make(map[int32]*Instance, 16),
		byPlayer:  make(map[uint32]int32, 64),
		templates: make(map[int32]*Template, 16),
		cooldowns: make(map[cooldownKey]int64, 128),
	}
}

// RegisterTemplate registers an instance template.
func (m *Manager) RegisterTemplate(tmpl *Template) error {
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("validate template %d: %w", tmpl.ID, err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[tmpl.ID] = tmpl
	return nil
}

// Template returns a registered template by ID.
func (m *Manager) Template(templateID int32) *Template {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.templates[templateID]
}

// TemplateCount returns the number of registered templates.
func (m *Manager) TemplateCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.templates)
}

// CreateInstance creates a new instance from a template.
// ownerID is the objectID of the player creating the instance.
func (m *Manager) CreateInstance(templateID int32, ownerID uint32) (*Instance, error) {
	m.mu.RLock()
	tmpl, ok := m.templates[templateID]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrTemplateNotFound
	}

	id := m.nextID.Add(1)
	inst := NewInstance(id, templateID, ownerID, tmpl.Duration)

	// Применяем настройки шаблона.
	if tmpl.Duration > 0 {
		// Таймер автоуничтожения по истечении времени жизни.
		time.AfterFunc(tmpl.Duration, func() {
			m.onInstanceExpired(id)
		})
	}

	inst.SetState(StateActive)

	m.mu.Lock()
	m.instances[id] = inst
	m.mu.Unlock()

	slog.Debug("instance created",
		"instanceID", id,
		"templateID", templateID,
		"owner", ownerID,
		"duration", tmpl.Duration)

	return inst, nil
}

// EnterInstance adds a player to an instance, enforcing all restrictions.
// characterID is used for cooldown tracking (persistent across sessions).
// objectID is the player's world object ID.
// level is the player's current level.
func (m *Manager) EnterInstance(instanceID int32, objectID uint32, characterID int64, level int32) error {
	m.mu.RLock()
	inst, ok := m.instances[instanceID]
	if !ok {
		m.mu.RUnlock()
		return ErrInstanceNotFound
	}

	// Уже в инстансе?
	if _, exists := m.byPlayer[objectID]; exists {
		m.mu.RUnlock()
		return ErrAlreadyInInstance
	}
	m.mu.RUnlock()

	// Проверка состояния.
	if inst.State() != StateActive {
		return ErrInstanceDestroyed
	}
	if inst.IsExpired() {
		return ErrInstanceExpired
	}

	// Проверка шаблона.
	tmpl := m.Template(inst.TemplateID())
	if tmpl != nil {
		if tmpl.MaxPlayers > 0 && int32(inst.PlayerCount()) >= tmpl.MaxPlayers {
			return ErrInstanceFull
		}
		if tmpl.MinLevel > 0 && level < tmpl.MinLevel {
			return ErrLevelTooLow
		}
		if tmpl.MaxLevel > 0 && level > tmpl.MaxLevel {
			return ErrLevelTooHigh
		}

		// Проверка cooldown.
		if tmpl.Cooldown > 0 {
			if onCD, _ := m.IsOnCooldown(characterID, inst.TemplateID()); onCD {
				return ErrOnCooldown
			}
		}
	}

	// Добавляем игрока в инстанс.
	if !inst.AddPlayer(objectID) {
		return ErrAlreadyInInstance
	}

	m.mu.Lock()
	m.byPlayer[objectID] = instanceID
	m.mu.Unlock()

	slog.Debug("player entered instance",
		"objectID", objectID,
		"instanceID", instanceID)

	return nil
}

// ExitInstance removes a player from their current instance.
// characterID is used for setting the reentry cooldown.
// Returns the instance the player was in, or ErrNotInInstance.
func (m *Manager) ExitInstance(objectID uint32, characterID int64) (*Instance, error) {
	m.mu.RLock()
	instanceID, ok := m.byPlayer[objectID]
	if !ok {
		m.mu.RUnlock()
		return nil, ErrNotInInstance
	}
	inst := m.instances[instanceID]
	m.mu.RUnlock()

	if inst == nil {
		m.mu.Lock()
		delete(m.byPlayer, objectID)
		m.mu.Unlock()
		return nil, ErrInstanceNotFound
	}

	removed, empty := inst.RemovePlayer(objectID)
	if !removed {
		return nil, ErrNotInInstance
	}

	m.mu.Lock()
	delete(m.byPlayer, objectID)
	m.mu.Unlock()

	// Устанавливаем cooldown при выходе.
	tmpl := m.Template(inst.TemplateID())
	if tmpl != nil && tmpl.Cooldown > 0 {
		m.SetCooldown(characterID, inst.TemplateID(), tmpl.Cooldown)
	}

	slog.Debug("player exited instance",
		"objectID", objectID,
		"instanceID", instanceID,
		"empty", empty)

	// Инстанс пустой — запускаем таймер уничтожения.
	if empty {
		m.scheduleEmptyDestroy(inst)
	}

	return inst, nil
}

// DestroyInstance forcefully destroys an instance.
// All players inside must be removed before calling this.
func (m *Manager) DestroyInstance(instanceID int32) error {
	m.mu.Lock()
	inst, ok := m.instances[instanceID]
	if !ok {
		m.mu.Unlock()
		return ErrInstanceNotFound
	}
	delete(m.instances, instanceID)

	// Удаляем всех оставшихся игроков из byPlayer.
	for objID, iid := range m.byPlayer {
		if iid == instanceID {
			delete(m.byPlayer, objID)
		}
	}
	m.mu.Unlock()

	inst.SetState(StateDestroyed)

	slog.Debug("instance destroyed",
		"instanceID", instanceID,
		"templateID", inst.TemplateID())

	return nil
}

// GetInstance returns an instance by ID, or nil if not found.
func (m *Manager) GetInstance(instanceID int32) *Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instances[instanceID]
}

// GetPlayerInstance returns the instance a player is currently in, or nil.
func (m *Manager) GetPlayerInstance(objectID uint32) *Instance {
	m.mu.RLock()
	instanceID, ok := m.byPlayer[objectID]
	if !ok {
		m.mu.RUnlock()
		return nil
	}
	inst := m.instances[instanceID]
	m.mu.RUnlock()
	return inst
}

// IsInInstance returns true if the player is in any instance.
func (m *Manager) IsInInstance(objectID uint32) bool {
	m.mu.RLock()
	_, ok := m.byPlayer[objectID]
	m.mu.RUnlock()
	return ok
}

// InstanceCount returns the number of active instances.
func (m *Manager) InstanceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.instances)
}

// IsOnCooldown checks if a character has a reentry cooldown for a template.
// Returns (true, expireTime) if on cooldown; (false, zero) if not.
func (m *Manager) IsOnCooldown(characterID int64, templateID int32) (bool, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := cooldownKey{characterID: characterID, templateID: templateID}
	expireNano, ok := m.cooldowns[key]
	if !ok {
		return false, time.Time{}
	}

	expire := time.Unix(0, expireNano)
	if time.Now().After(expire) {
		return false, time.Time{}
	}
	return true, expire
}

// SetCooldown sets a reentry cooldown for a character on a template.
func (m *Manager) SetCooldown(characterID int64, templateID int32, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := cooldownKey{characterID: characterID, templateID: templateID}
	m.cooldowns[key] = time.Now().Add(duration).UnixNano()
}

// ClearCooldown removes a cooldown for a character on a template.
func (m *Manager) ClearCooldown(characterID int64, templateID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cooldowns, cooldownKey{characterID: characterID, templateID: templateID})
}

// ClearExpiredCooldowns removes all expired cooldowns.
func (m *Manager) ClearExpiredCooldowns() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UnixNano()
	removed := 0
	for key, expireNano := range m.cooldowns {
		if now > expireNano {
			delete(m.cooldowns, key)
			removed++
		}
	}
	return removed
}

// LoadCooldowns loads cooldowns from persistent storage.
// Called on server startup to restore reentry cooldowns.
func (m *Manager) LoadCooldowns(entries []CooldownEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UnixNano()
	loaded := 0
	for _, e := range entries {
		if e.ExpireNano > now {
			key := cooldownKey{characterID: e.CharacterID, templateID: e.TemplateID}
			m.cooldowns[key] = e.ExpireNano
			loaded++
		}
	}

	if loaded > 0 {
		slog.Info("loaded instance cooldowns", "count", loaded)
	}
}

// ExportCooldowns returns all active cooldowns for DB persistence.
func (m *Manager) ExportCooldowns() []CooldownEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UnixNano()
	entries := make([]CooldownEntry, 0, len(m.cooldowns))
	for key, expireNano := range m.cooldowns {
		if expireNano > now {
			entries = append(entries, CooldownEntry{
				CharacterID: key.characterID,
				TemplateID:  key.templateID,
				ExpireNano:  expireNano,
			})
		}
	}
	return entries
}

// CooldownEntry represents a persistent cooldown record.
type CooldownEntry struct {
	CharacterID int64
	TemplateID  int32
	ExpireNano  int64 // time.UnixNano
}

// onInstanceExpired handles instance lifetime expiration.
// Goroutine завершается после вызова DestroyInstance.
func (m *Manager) onInstanceExpired(instanceID int32) {
	inst := m.GetInstance(instanceID)
	if inst == nil {
		return
	}
	if inst.State() == StateDestroyed {
		return
	}

	slog.Info("instance expired",
		"instanceID", instanceID,
		"templateID", inst.TemplateID())

	inst.SetState(StateDestroying)

	if err := m.DestroyInstance(instanceID); err != nil {
		slog.Error("destroy expired instance",
			"instanceID", instanceID,
			"error", err)
	}
}

// scheduleEmptyDestroy starts a timer to destroy an empty instance.
// If a player enters before the timer fires, the timer is cancelled.
// Goroutine завершается когда таймер срабатывает или отменяется.
func (m *Manager) scheduleEmptyDestroy(inst *Instance) {
	delay := inst.EmptyDelay()
	if delay <= 0 {
		// Мгновенное уничтожение.
		if err := m.DestroyInstance(inst.ID()); err != nil {
			slog.Error("destroy empty instance",
				"instanceID", inst.ID(),
				"error", err)
		}
		return
	}

	timer := time.AfterFunc(delay, func() {
		if inst.PlayerCount() > 0 {
			return // Кто-то зашёл — не уничтожаем.
		}
		if inst.State() == StateDestroyed {
			return
		}
		inst.SetState(StateDestroying)
		if err := m.DestroyInstance(inst.ID()); err != nil {
			slog.Error("destroy empty instance after timeout",
				"instanceID", inst.ID(),
				"error", err)
		}
	})

	inst.SetEmptyTimer(timer)
}
