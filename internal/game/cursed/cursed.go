package cursed

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// Item IDs for cursed weapons.
const (
	ZaricheItemID  int32 = 8190
	AkamanahItemID int32 = 8689
)

// Skill IDs for cursed weapons.
const (
	ZaricheSkillID  int32 = 3603
	AkamanahSkillID int32 = 3629
	VoidBurstSkill  int32 = 3630
	VoidFlowSkill   int32 = 3631
)

// Default weapon parameters (from CursedWeapons.xml).
const (
	DefaultDropRate        int32 = 1      // 1 in 100000
	DropRateDenominator    int32 = 100000 // denominator for drop rate
	DefaultDuration        int64 = 300    // minutes
	DefaultDurationLost    int64 = 3      // minutes lost per kill
	DefaultDisappearChance int32 = 50     // % chance to vanish on death
	DefaultStageKills      int32 = 10     // kills per skill level
	MaxCursedKarma         int32 = 9999999
	SocialActionAnimation  int32 = 17
)

// Weapon states.
const (
	StateInactive  int32 = 0
	StateDropped   int32 = 1
	StateActivated int32 = 2
)

// Errors.
var (
	ErrNotCursedWeapon = errors.New("not a cursed weapon")
	ErrAlreadyActive   = errors.New("weapon already active")
	ErrNotActive       = errors.New("weapon not active")
	ErrPlayerOnMount   = errors.New("player is mounted")
)

// Weapon represents a single cursed weapon instance (Zariche or Akamanah).
type Weapon struct {
	mu sync.RWMutex

	// Config (immutable after init)
	itemID         int32
	name           string
	skillID        int32
	skillMaxLevel  int32
	dropRate       int32
	duration       int64 // minutes
	durationLost   int64 // minutes lost per kill
	disappearChance int32 // % chance to vanish on death
	stageKills     int32

	// State
	state   int32
	nbKills int32
	endTime int64 // Unix milliseconds

	// Player tracking
	playerID      int32 // ObjectID of owner
	player        *model.Player
	item          *model.Item
	playerKarma   int32
	playerPKKills int32

	// Location when dropped on ground
	dropX, dropY, dropZ int32

	// Timer
	stopCh chan struct{}
}

// NewWeapon creates a new cursed weapon with default configuration.
func NewWeapon(itemID int32, name string, skillID int32) *Weapon {
	return &Weapon{
		itemID:          itemID,
		name:            name,
		skillID:         skillID,
		skillMaxLevel:   10,
		dropRate:        DefaultDropRate,
		duration:        DefaultDuration,
		durationLost:    DefaultDurationLost,
		disappearChance: DefaultDisappearChance,
		stageKills:      DefaultStageKills,
		state:           StateInactive,
	}
}

// ItemID returns the weapon's item ID.
func (w *Weapon) ItemID() int32 {
	return w.itemID
}

// Name returns the weapon's display name.
func (w *Weapon) Name() string {
	return w.name
}

// SkillID returns the weapon's skill ID.
func (w *Weapon) SkillID() int32 {
	return w.skillID
}

// State returns the current state.
func (w *Weapon) State() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

// IsActive returns true if the weapon is dropped or activated.
func (w *Weapon) IsActive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state != StateInactive
}

// IsDropped returns true if the weapon is on the ground.
func (w *Weapon) IsDropped() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state == StateDropped
}

// IsActivated returns true if a player is carrying the weapon.
func (w *Weapon) IsActivated() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state == StateActivated
}

// Player returns the current owner.
func (w *Weapon) Player() *model.Player {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.player
}

// PlayerID returns the owner's object ID.
func (w *Weapon) PlayerID() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.playerID
}

// NBKills returns the current kill count.
func (w *Weapon) NBKills() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.nbKills
}

// EndTime returns the expiration unix timestamp in milliseconds.
func (w *Weapon) EndTime() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.endTime
}

// PlayerKarma returns the saved player karma (before activation).
func (w *Weapon) PlayerKarma() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.playerKarma
}

// PlayerPKKills returns the saved player PK kills (before activation).
func (w *Weapon) PlayerPKKills() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.playerPKKills
}

// SkillLevel returns the current skill level based on kills.
func (w *Weapon) SkillLevel() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.skillLevel()
}

func (w *Weapon) skillLevel() int32 {
	level := 1 + w.nbKills/w.stageKills
	if level > w.skillMaxLevel {
		return w.skillMaxLevel
	}
	return level
}

// Location returns the weapon's drop location (only valid when dropped).
func (w *Weapon) Location() (x, y, z int32) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.dropX, w.dropY, w.dropZ
}

// RemainingMinutes returns the remaining time in minutes.
func (w *Weapon) RemainingMinutes() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	remaining := w.endTime - time.Now().UnixMilli()
	if remaining <= 0 {
		return 0
	}
	return remaining / 60000
}

// CheckDrop checks if a cursed weapon should drop from a killed monster.
// Returns true if the weapon should be dropped at the given location.
func (w *Weapon) CheckDrop(x, y, z int32) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state != StateInactive {
		return false
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(DropRateDenominator)))
	if err != nil {
		slog.Error("cursed weapon random check", "error", err)
		return false
	}
	if n.Int64() >= int64(w.dropRate) {
		return false
	}

	w.state = StateDropped
	w.dropX = x
	w.dropY = y
	w.dropZ = z
	w.endTime = time.Now().UnixMilli() + w.duration*60000

	w.startRemoveTask()

	slog.Info("cursed weapon dropped",
		"weapon", w.name,
		"x", x, "y", y, "z", z)

	return true
}

// Activate assigns the weapon to a player.
// Returns error if weapon can't be activated.
func (w *Weapon) Activate(player *model.Player) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state == StateActivated {
		return ErrAlreadyActive
	}

	// Сохраняем оригинальные значения
	w.playerKarma = player.Karma()
	w.playerPKKills = player.PKKills()

	// Устанавливаем cursed state
	player.SetKarma(MaxCursedKarma)
	player.SetPKKills(0)
	player.SetCursedWeaponEquippedID(w.itemID)

	w.player = player
	w.playerID = int32(player.ObjectID())
	w.state = StateActivated

	// Если оружие не было на земле (передача), устанавливаем endTime
	if w.endTime == 0 {
		w.endTime = time.Now().UnixMilli() + w.duration*60000
	}

	if w.stopCh == nil {
		w.startRemoveTask()
	}

	slog.Info("cursed weapon activated",
		"weapon", w.name,
		"player", player.Name(),
		"remaining_min", (w.endTime-time.Now().UnixMilli())/60000)

	return nil
}

// DropIt handles weapon drop when the owner dies.
// Returns true if weapon disappeared (50%), false if it dropped on ground.
func (w *Weapon) DropIt(killerX, killerY, killerZ int32) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state != StateActivated || w.player == nil {
		return true
	}

	// Восстанавливаем оригинальные значения
	w.player.SetKarma(w.playerKarma)
	w.player.SetPKKills(w.playerPKKills)
	w.player.SetCursedWeaponEquippedID(0)

	// Проверяем шанс исчезновения
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		slog.Error("cursed weapon disappear check", "error", err)
		w.reset()
		return true
	}

	if n.Int64() < int64(w.disappearChance) {
		// Оружие исчезает
		slog.Info("cursed weapon disappeared on death", "weapon", w.name)
		w.reset()
		return true
	}

	// Оружие дропается на землю
	w.dropX = killerX
	w.dropY = killerY
	w.dropZ = killerZ
	w.player = nil
	w.playerID = 0
	w.state = StateDropped

	slog.Info("cursed weapon dropped on death",
		"weapon", w.name,
		"x", killerX, "y", killerY, "z", killerZ)

	return false
}

// IncreaseKills increments the kill counter and reduces duration.
func (w *Weapon) IncreaseKills() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.nbKills++

	if w.player != nil {
		w.player.SetPKKills(w.nbKills)
	}

	// Уменьшаем время жизни
	w.endTime -= w.durationLost * 60000

	slog.Debug("cursed weapon kill",
		"weapon", w.name,
		"kills", w.nbKills,
		"skillLevel", w.skillLevel(),
		"remaining_min", (w.endTime-time.Now().UnixMilli())/60000)
}

// EndOfLife forcefully removes the weapon. Called when timer expires.
func (w *Weapon) EndOfLife() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.player != nil {
		w.player.SetKarma(w.playerKarma)
		w.player.SetPKKills(w.playerPKKills)
		w.player.SetCursedWeaponEquippedID(0)
	}

	slog.Info("cursed weapon expired", "weapon", w.name)
	w.reset()
}

// SetPlayer updates the player reference (for login/logout).
func (w *Weapon) SetPlayer(player *model.Player) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.player = player
}

// Restore восстанавливает состояние из БД.
func (w *Weapon) Restore(charID, playerKarma, playerPKKills, nbKills int32, endTime int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.playerID = charID
	w.playerKarma = playerKarma
	w.playerPKKills = playerPKKills
	w.nbKills = nbKills
	w.endTime = endTime

	if endTime > 0 {
		w.state = StateActivated
		w.startRemoveTask()
	}
}

// SaveData returns data needed for DB persistence.
func (w *Weapon) SaveData() (itemID, charID, playerKarma, playerPKKills, nbKills int32, endTime int64) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.itemID, w.playerID, w.playerKarma, w.playerPKKills, w.nbKills, w.endTime
}

// Stop stops the remove task.
func (w *Weapon) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancelTask()
}

func (w *Weapon) reset() {
	w.cancelTask()
	w.state = StateInactive
	w.player = nil
	w.playerID = 0
	w.item = nil
	w.nbKills = 0
	w.endTime = 0
	w.playerKarma = 0
	w.playerPKKills = 0
	w.dropX = 0
	w.dropY = 0
	w.dropZ = 0
}

func (w *Weapon) startRemoveTask() {
	w.cancelTask()

	w.stopCh = make(chan struct{})
	interval := time.Duration(w.durationLost*12) * time.Second // 36 сек при durationLost=3

	go func(stopCh chan struct{}) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if time.Now().UnixMilli() >= w.EndTime() {
					w.EndOfLife()
					return
				}
			case <-stopCh:
				return
			}
		}
	}(w.stopCh)
}

func (w *Weapon) cancelTask() {
	if w.stopCh != nil {
		close(w.stopCh)
		w.stopCh = nil
	}
}

// IsCursedWeapon returns true if the given item ID is a cursed weapon.
func IsCursedWeapon(itemID int32) bool {
	return itemID == ZaricheItemID || itemID == AkamanahItemID
}

// Manager manages all cursed weapons in the game world.
type Manager struct {
	mu      sync.RWMutex
	weapons map[int32]*Weapon // itemID → Weapon
}

// NewManager creates a new cursed weapons manager with Zariche and Akamanah.
func NewManager() *Manager {
	m := &Manager{
		weapons: make(map[int32]*Weapon, 2),
	}

	m.weapons[ZaricheItemID] = NewWeapon(ZaricheItemID, "Demonic Sword Zariche", ZaricheSkillID)
	m.weapons[AkamanahItemID] = NewWeapon(AkamanahItemID, "Blood Sword Akamanah", AkamanahSkillID)

	return m
}

// Weapon returns the cursed weapon by item ID.
func (m *Manager) Weapon(itemID int32) *Weapon {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.weapons[itemID]
}

// Weapons returns all cursed weapons.
func (m *Manager) Weapons() []*Weapon {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Weapon, 0, len(m.weapons))
	for _, w := range m.weapons {
		result = append(result, w)
	}
	return result
}

// IsCursed returns true if the given item ID is a cursed weapon.
func (m *Manager) IsCursed(itemID int32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.weapons[itemID]
	return ok
}

// CursedWeaponIDs returns all registered cursed weapon item IDs.
func (m *Manager) CursedWeaponIDs() []int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]int32, 0, len(m.weapons))
	for id := range m.weapons {
		ids = append(ids, id)
	}
	return ids
}

// CheckDrop checks if any cursed weapon should drop from a killed monster.
// Returns the weapon that dropped, or nil.
func (m *Manager) CheckDrop(x, y, z int32) *Weapon {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, w := range m.weapons {
		if w.CheckDrop(x, y, z) {
			return w
		}
	}
	return nil
}

// Activate activates a cursed weapon for a player.
func (m *Manager) Activate(player *model.Player, itemID int32) error {
	w := m.Weapon(itemID)
	if w == nil {
		return ErrNotCursedWeapon
	}

	// Если игрок уже имеет другое cursed weapon — отказ
	if player.IsCursedWeaponEquipped() {
		existingID := player.CursedWeaponEquippedID()
		if existingID != itemID {
			// Добавляем +10 киллов к текущему оружию (1 стадию)
			if existing := m.Weapon(existingID); existing != nil {
				for range DefaultStageKills {
					existing.IncreaseKills()
				}
			}
			return fmt.Errorf("player already has cursed weapon %d", existingID)
		}
	}

	return w.Activate(player)
}

// IncreaseKills increases kill count for the weapon the player is carrying.
func (m *Manager) IncreaseKills(itemID int32) {
	w := m.Weapon(itemID)
	if w == nil {
		return
	}
	w.IncreaseKills()
}

// Drop handles weapon drop when owner dies.
func (m *Manager) Drop(itemID int32, killerX, killerY, killerZ int32) bool {
	w := m.Weapon(itemID)
	if w == nil {
		return true
	}
	return w.DropIt(killerX, killerY, killerZ)
}

// CheckPlayer checks if a player owns a cursed weapon on login.
// Returns the weapon if found, nil otherwise.
func (m *Manager) CheckPlayer(player *model.Player) *Weapon {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objectID := int32(player.ObjectID())
	for _, w := range m.weapons {
		if w.PlayerID() == objectID && w.IsActivated() {
			w.SetPlayer(player)
			return w
		}
	}
	return nil
}

// CheckOwnsWeapon returns the item ID of the cursed weapon owned by the player.
// Returns 0 if player doesn't own any.
func (m *Manager) CheckOwnsWeapon(playerObjectID int32) int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, w := range m.weapons {
		if w.PlayerID() == playerObjectID && w.IsActivated() {
			return w.ItemID()
		}
	}
	return 0
}

// LocationInfo returns location information for active cursed weapons.
func (m *Manager) LocationInfo() []WeaponLocationInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var infos []WeaponLocationInfo
	for _, w := range m.weapons {
		if !w.IsActive() {
			continue
		}

		info := WeaponLocationInfo{
			ItemID: w.ItemID(),
		}

		if w.IsActivated() {
			info.Activated = 1
			if p := w.Player(); p != nil {
				loc := p.Location()
				info.X = loc.X
				info.Y = loc.Y
				info.Z = loc.Z
			}
		} else {
			info.Activated = 0
			info.X, info.Y, info.Z = w.Location()
		}

		infos = append(infos, info)
	}
	return infos
}

// Stop stops all timers.
func (m *Manager) Stop() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.weapons {
		w.Stop()
	}
}

// WeaponLocationInfo holds location data for a cursed weapon.
type WeaponLocationInfo struct {
	ItemID    int32
	Activated int32 // 0=dropped, 1=equipped by player
	X, Y, Z  int32
}
