package model

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/data"
)

// StatBonusProvider provides stat bonuses from active effects (buffs/debuffs).
// Interface to avoid import cycle between model ↔ skill packages.
// Phase 5.9.3: Effect Framework.
type StatBonusProvider interface {
	GetStatBonus(stat string) float64
}

// Player — игровой персонаж.
// Добавляет player-specific данные к Character.
type Player struct {
	*Character // embedded

	characterID int64
	accountID   int64
	level       int32
	raceID      int32
	classID     int32
	experience  int64
	sp          int64
	createdAt   time.Time
	lastLogin   time.Time

	playerMu sync.RWMutex // отдельный mutex для player data

	// Visibility cache (Phase 4.5 PR3)
	// Stores *VisibilityCache — updated by VisibilityManager every 100ms
	// atomic.Value allows lock-free concurrent reads
	visibilityCache atomic.Value // *VisibilityCache (defined in internal/world)

	// Movement tracking (Phase 5.1)
	// Tracks client and server positions separately for desync detection
	movement *PlayerMovement

	// Target tracking (Phase 5.2)
	// Currently selected target (player, NPC, or item)
	// Protected by playerMu for thread-safe access
	target *WorldObject

	// Inventory (Phase 5.5)
	// Player's inventory and equipped items (paperdoll)
	inventory *Inventory

	// Skills (Phase 5.9.2)
	// Learned skills: map[skillID]*SkillInfo
	// Protected by playerMu for thread-safe access
	skills map[int32]*SkillInfo

	// Effect manager (Phase 5.9.3)
	// Tracks active buffs/debuffs and provides stat bonuses
	// Interface to avoid import cycle between model ↔ skill packages
	effectManager StatBonusProvider

	// Last attack timestamp (UnixNano) for combat stance detection.
	// Atomic to avoid depending on combat package (no import cycle).
	lastAttackTime atomic.Int64

	// Party (Phase 7.3)
	// Current party membership. Nil if not in a party.
	// Protected by playerMu.
	party *Party

	// Pending party invite. Nil if no invite pending.
	// Protected by playerMu.
	pendingPartyInvite *PartyInvite

	// Private Store (Phase 8.1)
	// Tracks store mode (SELL/BUY/MANUFACTURE/PACKAGE_SELL)
	// and trade lists for sell/buy operations.
	privateStoreType PrivateStoreType // Protected by playerMu
	sellList         *TradeList       // Active sell list (nil if not selling)
	buyList          *TradeList       // Active buy list (nil if not buying)
	storeMessage     string           // Store title shown above player head

	// Manufacture Shop (Phase 54)
	// Items offered in manufacture (crafting) store.
	// Protected by playerMu.
	manufactureItems []*ManufactureItem

	// Recipe Book (Phase 15)
	// Learned recipes: set of recipeListIDs.
	// Protected by playerMu.
	dwarvenRecipes map[int32]struct{} // Dwarven craft recipes
	commonRecipes  map[int32]struct{} // Common craft recipes

	// Admin (Phase 17)
	// Access level from account. 0 = normal player, 1+ = GM, 100+ = full admin.
	// Negative = banned. Protected by playerMu.
	accessLevel      int32
	lastAdminMessage string // Last message from admin command system
	invisible        bool   // GM invisible mode
	invulnerable     bool   // GM invulnerable mode

	// Clan (Phase 18)
	// Clan membership. 0 = not in a clan.
	// Protected by playerMu.
	clanID             int32
	clanTitle          string     // Title assigned by clan leader
	pendingClanInvite  *ClanInvite // Pending clan invite (nil = none)

	// Henna System (Phase 13)
	// 3 slots for tattoos, each may be nil (empty).
	// Protected by playerMu.
	hennas    [MaxHennaSlots]*HennaSlot
	hennaStat hennaStats // cached stat bonuses

	// Pet/Summon (Phase 19)
	// Currently summoned creature. Nil if no active summon.
	// Protected by playerMu.
	summon *Summon

	// Duel System (Phase 20)
	// Active duel ID (0 = not in duel). Protected by playerMu.
	duelID int32
	// Pending duel request from another player. Protected by playerMu.
	pendingDuelRequest *DuelRequest

	// Karma & PK/PvP (Phase 32)
	// Karma: reputation penalty for killing innocent players.
	// PK = non-flagged kills, PvP = flagged kills.
	// Protected by playerMu.
	karma   int32
	pkKills int32

	// Cursed Weapon (Phase 32)
	// Item ID of the equipped cursed weapon (8190=Zariche, 8689=Akamanah).
	// 0 = no cursed weapon. Protected by playerMu.
	cursedWeaponEquippedID int32

	// Marriage System (Phase 33)
	// Partner ObjectID (0 = not engaged/married). Protected by playerMu.
	partnerID int32
	// Couple row ID in DB. Protected by playerMu.
	coupleID int32
	// True if the ceremony is complete. Protected by playerMu.
	married bool
	// Pending engage request state. Protected by playerMu.
	engageRequest bool
	engageFromID  int32 // ObjectID of the player who proposed

	// PvP State
	// Protected by playerMu.
	pvpKills int32
	pvpFlag  int32 // 0=not flagged, 1=flagged

	// Appearance
	// Protected by playerMu.
	title      string
	isFemale   bool
	hairStyle  int32
	hairColor  int32
	face       int32
	nameColor  int32 // BGR format, default 0xFFFFFF (white)
	titleColor int32 // BGR format, default 0xFFFF77 (light yellow)

	// Movement state. Protected by playerMu.
	running bool // true=running, false=walking (default true)
	sitting bool

	// Noble/Hero. Protected by playerMu.
	noble bool
	hero  bool

	// Fishing state. Protected by playerMu.
	fishing bool
	fishX   int32
	fishY   int32
	fishZ   int32

	// Clan display. Protected by playerMu.
	pledgeClass int32
	pledgeType  int32

	// Recommendations. Protected by playerMu.
	recomLeft int32
	recomHave int32

	// Abnormal visual effects bitmask. Protected by playerMu.
	abnormalVisualEffects int32

	// Team/Events. Protected by playerMu.
	teamID int32

	// Mount. Protected by playerMu.
	mountType  int32
	mountNpcID int32

	// Enchant System
	// ObjectID of the active enchant scroll (0 = no enchant in progress).
	// Protected by playerMu.
	activeEnchantItemID int32

	// P2P Trade System
	// Active trade session list. Nil if not in a trade.
	// Protected by playerMu.
	activeTradeList *P2PTradeList
	// Player who requested trade. Nil if no pending request.
	// Protected by playerMu.
	activeRequester *Player
	// Unix timestamp when trade request expires.
	// Protected by playerMu.
	requestExpireTime int64

	// Shortcuts (Phase 34)
	// Action bar shortcut bindings (F1-F12 x 10 pages).
	// Key = slot + page*12. Protected by playerMu.
	shortcuts map[int32]*Shortcut

	// Friend & Block Lists (Phase 35)
	// friendList stores ObjectIDs of friends (relation=0 in DB).
	// blockList stores ObjectIDs of blocked players (relation=1 in DB).
	// messageRefusal is "block all messages" toggle.
	// All protected by playerMu.
	friendList     map[int32]bool
	blockList      map[int32]bool
	messageRefusal bool

	// Auto SoulShot (Phase 36)
	// Item IDs of auto-enabled soulshots/spiritshots.
	// Protected by playerMu.
	autoSoulShots map[int32]bool

	// Macros (Phase 36)
	// Player-defined macros (up to 24).
	// Protected by playerMu.
	macros        map[int32]*Macro
	macroRevision int32

	// Item cooldowns (Phase 51: Item Handler System)
	// Key = itemID, Value = expiry time.
	// Protected by playerMu.
	itemCooldowns map[int32]time.Time

	// Olympiad state (Phase 51 stub)
	inOlympiad bool

	// Subclass System (Phase 14)
	subclassFields
}

// NewPlayer создаёт нового игрока с валидацией.
// Phase 4.15: Added objectID parameter to link Player with WorldObject.
// objectID must be unique across all world objects (players, NPCs, items).
// Use 0 for testing/mock objects only (not suitable for production).
func NewPlayer(objectID uint32, characterID, accountID int64, name string, level, raceID, classID int32) (*Player, error) {
	if name == "" || len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters, got %q", name)
	}
	if level < 1 || level > 80 {
		return nil, fmt.Errorf("level must be between 1 and 80, got %d", level)
	}

	// Default spawn location (будет из config/DB в Phase 4.3+)
	loc := NewLocation(0, 0, 0, 0)

	// Default stats (будет из PlayerTemplate в Phase 4.3+)
	// Для MVP используем simple hardcoded values
	maxHP := int32(1000 + level*50)  // Linear scaling для тестов
	maxMP := int32(500 + level*25)
	maxCP := int32(800 + level*40)

	p := &Player{
		Character:   NewCharacter(objectID, name, loc, level, maxHP, maxMP, maxCP),
		characterID: characterID,
		accountID:   accountID,
		level:       level, // FIXME: level duplicated in Character and Player
		raceID:      raceID,
		classID:     classID,
		experience:  0,
		createdAt:   time.Now(),
		movement:    NewPlayerMovement(loc.X, loc.Y, loc.Z), // Phase 5.1
		inventory:   NewInventory(characterID),              // Phase 5.5
		shortcuts:     make(map[int32]*Shortcut),              // Phase 34
		friendList:    make(map[int32]bool),                   // Phase 35
		blockList:     make(map[int32]bool),                   // Phase 35
		autoSoulShots: make(map[int32]bool),                   // Phase 36
		macros:        make(map[int32]*Macro),                 // Phase 36
	}

	// Defaults
	p.running = true
	p.nameColor = 0xFFFFFF  // White (BGR)
	p.titleColor = 0xFFFF77 // Light yellow (BGR)

	// Initialize visibility cache (Phase 4.5 PR3)
	p.visibilityCache.Store((*VisibilityCache)(nil))

	// Phase 5.6: Set WorldObject.Data reference for PvE combat
	p.WorldObject.Data = p

	// Phase 14: Subclass System
	p.initSubclassFields()

	return p, nil
}

// CharacterID возвращает DB ID персонажа (immutable).
func (p *Player) CharacterID() int64 {
	return p.characterID
}

// AccountID возвращает ID аккаунта (immutable).
func (p *Player) AccountID() int64 {
	return p.accountID
}

// Level возвращает уровень персонажа.
func (p *Player) Level() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.level
}

// SetLevel устанавливает уровень с валидацией.
func (p *Player) SetLevel(level int32) error {
	if level < 1 || level > 80 {
		return fmt.Errorf("level must be between 1 and 80, got %d", level)
	}

	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.level = level
	return nil
}

// RaceID возвращает ID расы.
func (p *Player) RaceID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.raceID
}

// ClassID возвращает ID класса.
func (p *Player) ClassID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.classID
}

// Experience возвращает текущий опыт.
func (p *Player) Experience() int64 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.experience
}

// AddExperience добавляет опыт (может быть отрицательным для penalty).
func (p *Player) AddExperience(exp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.experience += exp
	if p.experience < 0 {
		p.experience = 0
	}
}

// SetExperience устанавливает точное значение опыта.
func (p *Player) SetExperience(exp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if exp < 0 {
		exp = 0
	}
	p.experience = exp
}

// SP возвращает текущие skill points.
func (p *Player) SP() int64 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.sp
}

// AddSP добавляет skill points.
func (p *Player) AddSP(sp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.sp += sp
	if p.sp < 0 {
		p.sp = 0
	}
}

// SetSP устанавливает точное значение SP (для загрузки из DB).
func (p *Player) SetSP(sp int64) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if sp < 0 {
		sp = 0
	}
	p.sp = sp
}

// CreatedAt возвращает время создания персонажа.
func (p *Player) CreatedAt() time.Time {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.createdAt
}

// LastLogin возвращает время последнего входа.
func (p *Player) LastLogin() time.Time {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.lastLogin
}

// UpdateLastLogin обновляет время последнего входа на текущее.
func (p *Player) UpdateLastLogin() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.lastLogin = time.Now()
}

// SetLastLogin устанавливает время последнего входа (для загрузки из DB).
func (p *Player) SetLastLogin(t time.Time) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.lastLogin = t
}

// SetCreatedAt устанавливает время создания (для загрузки из DB).
func (p *Player) SetCreatedAt(t time.Time) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.createdAt = t
}

// SetCharacterID устанавливает DB ID после создания в БД.
func (p *Player) SetCharacterID(id int64) {
	// characterID immutable, но setter нужен для repository.Create
	p.characterID = id
}

// SetRaceID устанавливает ID расы (для загрузки из DB).
func (p *Player) SetRaceID(raceID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.raceID = raceID
}

// SetClassID устанавливает ID класса (для изменения профессии).
func (p *Player) SetClassID(classID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.classID = classID
}

// GetVisibilityCache returns current visibility cache (may be nil if not initialized).
// Safe for concurrent reads (atomic.Value.Load).
// Phase 4.5 PR3: Visibility Cache optimization.
func (p *Player) GetVisibilityCache() *VisibilityCache {
	v := p.visibilityCache.Load()
	if v == nil {
		return nil
	}
	return v.(*VisibilityCache)
}

// SetVisibilityCache updates visibility cache atomically.
// Safe for concurrent writes (atomic.Value.Store is thread-safe).
// Phase 4.5 PR3: Called by VisibilityManager every 100ms.
func (p *Player) SetVisibilityCache(cache *VisibilityCache) {
	p.visibilityCache.Store(cache)
}

// InvalidateVisibilityCache clears visibility cache (sets to nil).
// Called when player moves to different region, teleports, or logs out.
// Phase 4.5 PR3: Forces fresh query on next visibility check.
func (p *Player) InvalidateVisibilityCache() {
	p.visibilityCache.Store((*VisibilityCache)(nil))
}

// Movement returns the movement tracking state.
// Thread-safe: PlayerMovement uses RWMutex internally.
// Phase 5.1: Movement validation and desync detection.
func (p *Player) Movement() *PlayerMovement {
	return p.movement
}

// CanLogout returns true if player can logout/restart safely.
// Checks multiple conditions: subclass lock, attack stance, enchanting, events, festivals.
//
// Reference: L2J_Mobius Player.canLogout() (8270-8313)
func (p *Player) CanLogout() bool {
	// Check subclass lock — if a subclass operation is in progress, deny logout.
	// TryLock returns false if mutex is already held (subclass change in progress).
	if !p.subclassMu.TryLock() {
		return false
	}
	p.subclassMu.Unlock()

	// Check attack stance (combat) — 15s cooldown after last attack.
	if p.HasAttackStance() {
		return false
	}

	// Cannot logout while enchanting
	if p.ActiveEnchantItemID() != 0 {
		return false
	}

	// Event registration system not yet implemented — will check IsRegisteredOnEvent here.
	// Festival participant check not yet implemented — will check IsFestivalParticipant here.

	return true
}

// HasAttackStance returns true if player is in combat (attacked or was attacked recently).
// Combat state persists for 15 seconds after last attack (COMBAT_TIME).
// Uses atomic lastAttackTime to avoid depending on combat package (no import cycle).
func (p *Player) HasAttackStance() bool {
	ts := p.lastAttackTime.Load()
	if ts == 0 {
		return false
	}
	return time.Since(time.Unix(0, ts)) < 15*time.Second
}

// MarkAttackStance records current time as last attack moment.
// Called by combat.AttackStanceManager when player enters combat.
func (p *Player) MarkAttackStance() {
	p.lastAttackTime.Store(time.Now().UnixNano())
}

// LastAttackTime returns the last attack timestamp as UnixNano (0 if never attacked).
func (p *Player) LastAttackTime() int64 {
	return p.lastAttackTime.Load()
}

// IsTrading returns true if player is in trade mode (private store, manufacture).
//
// Phase 8.1: Private Store System.
// Reference: L2J_Mobius Player.getPrivateStoreType()
func (p *Player) IsTrading() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.privateStoreType != StoreNone
}

// IsInStoreMode returns true if player has an active store (not manage mode).
func (p *Player) IsInStoreMode() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.privateStoreType.IsInStoreMode()
}

// PrivateStoreType returns current store type.
func (p *Player) PrivateStoreType() PrivateStoreType {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.privateStoreType
}

// SetPrivateStoreType sets the store type.
func (p *Player) SetPrivateStoreType(storeType PrivateStoreType) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.privateStoreType = storeType
}

// SellList returns the active sell trade list (may be nil).
func (p *Player) SellList() *TradeList {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.sellList
}

// SetSellList sets the sell trade list.
func (p *Player) SetSellList(list *TradeList) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.sellList = list
}

// BuyList returns the active buy trade list (may be nil).
func (p *Player) BuyList() *TradeList {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.buyList
}

// SetBuyList sets the buy trade list.
func (p *Player) SetBuyList(list *TradeList) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.buyList = list
}

// StoreMessage returns the store title shown above player's head.
func (p *Player) StoreMessage() string {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.storeMessage
}

// SetStoreMessage sets the store message (max 29 characters).
func (p *Player) SetStoreMessage(msg string) {
	if len(msg) > 29 {
		msg = msg[:29]
	}
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.storeMessage = msg
}

// ClosePrivateStore closes the player's private store and resets all store data.
func (p *Player) ClosePrivateStore() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.privateStoreType = StoreNone
	if p.sellList != nil {
		p.sellList.Clear()
		p.sellList = nil
	}
	if p.buyList != nil {
		p.buyList.Clear()
		p.buyList = nil
	}
	p.manufactureItems = nil
	p.storeMessage = ""
}

// ManufactureItems returns a copy of the manufacture items list.
// Thread-safe: acquires read lock.
//
// Phase 54: Recipe Shop (Manufacture) System.
func (p *Player) ManufactureItems() []*ManufactureItem {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.manufactureItems == nil {
		return nil
	}
	result := make([]*ManufactureItem, len(p.manufactureItems))
	copy(result, p.manufactureItems)
	return result
}

// SetManufactureItems sets the manufacture items list.
// Thread-safe: acquires write lock.
//
// Phase 54: Recipe Shop (Manufacture) System.
func (p *Player) SetManufactureItems(items []*ManufactureItem) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.manufactureItems = items
}

// ClearManufactureItems clears the manufacture items list.
// Thread-safe: acquires write lock.
//
// Phase 54: Recipe Shop (Manufacture) System.
func (p *Player) ClearManufactureItems() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.manufactureItems = nil
}

// FindManufactureItem finds a manufacture item by recipeID.
// Returns nil if not found.
// Thread-safe: acquires read lock.
//
// Phase 54: Recipe Shop (Manufacture) System.
func (p *Player) FindManufactureItem(recipeID int32) *ManufactureItem {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	for _, mi := range p.manufactureItems {
		if mi.RecipeID == recipeID {
			return mi
		}
	}
	return nil
}

// Target returns the currently selected target.
// Returns nil if no target is selected.
// Thread-safe: acquires read lock.
//
// Phase 5.2: Target System.
func (p *Player) Target() *WorldObject {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.target
}

// SetTarget sets the currently selected target.
// Pass nil to clear the target.
// Thread-safe: acquires write lock.
//
// Phase 5.2: Target System.
func (p *Player) SetTarget(target *WorldObject) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.target = target
}

// ClearTarget clears the currently selected target.
// Convenience method equivalent to SetTarget(nil).
// Thread-safe: acquires write lock.
//
// Phase 5.2: Target System.
func (p *Player) ClearTarget() {
	p.SetTarget(nil)
}

// HasTarget returns true if player has a target selected.
// Thread-safe: acquires read lock.
//
// Phase 5.2: Target System.
func (p *Player) HasTarget() bool {
	return p.Target() != nil
}

// Inventory returns player's inventory (inventory + equipped items).
// Thread-safe: Inventory methods use internal mutex.
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) Inventory() *Inventory {
	return p.inventory
}

// GetEquippedWeapon returns equipped weapon (может быть nil).
// Convenience method для Inventory().GetPaperdollItem(PaperdollRHand).
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) GetEquippedWeapon() *Item {
	if p.inventory == nil {
		return nil
	}
	return p.inventory.GetPaperdollItem(PaperdollRHand)
}

// GetEquippedArmor returns equipped armor для указанного slot (может быть nil).
// Convenience method для Inventory().GetPaperdollItem(slot).
//
// Parameters:
//   - slot: paperdoll slot index (PaperdollChest, PaperdollLegs, etc.)
//
// Phase 5.5: Weapon & Equipment System.
func (p *Player) GetEquippedArmor(slot int32) *Item {
	if p.inventory == nil {
		return nil
	}
	return p.inventory.GetPaperdollItem(slot)
}

// GetBasePAtk returns base physical attack power from character template (no weapon).
// This is the "nude" pAtk WITHOUT weapon bonus, but WITH STR bonus and level modifier.
//
// Formula: basePAtk × STRBonus[STR] × levelMod
// where levelMod = (level + 89) / 100.0
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: FuncPAtkMod.java:45
func (p *Player) GetBasePAtk() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 100 // Fallback для неизвестного класса
	}

	basePAtk := float64(template.BasePAtk)
	strBonus := data.GetSTRBonus(p.GetSTR())
	levelMod := p.GetLevelMod()

	finalPAtk := basePAtk * strBonus * levelMod
	return int32(finalPAtk)
}

// GetPAtk returns final physical attack power WITH weapon bonus.
// Weapon pAtk is added BEFORE applying STR bonus and level modifier.
//
// Formula: (basePAtk + weaponPAtk) × STRBonus[STR] × levelMod
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: CreatureStat.java:539, FuncAdd + FuncPAtkMod
func (p *Player) GetPAtk() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 100 // Fallback
	}

	// Base pAtk from character template
	basePAtk := float64(template.BasePAtk)

	// Add weapon pAtk
	weaponPAtk := int32(0)
	if weapon := p.GetEquippedWeapon(); weapon != nil {
		weaponPAtk = weapon.Template().PAtk
	}

	// Apply STR bonus and level modifier
	strBonus := data.GetSTRBonus(p.GetSTR())
	levelMod := p.GetLevelMod()

	finalPAtk := (basePAtk + float64(weaponPAtk)) * strBonus * levelMod
	return int32(finalPAtk)
}

// GetPAtkSpd returns physical attack speed with DEX bonus.
// Formula: basePAtkSpd × DEXBonus[DEX]
//
// Java reference: CreatureStat.getPAtkSpd(), Formulas.calcPAtkSpd()
func (p *Player) GetPAtkSpd() float64 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 300.0 // Fallback
	}

	baseSpd := float64(template.BasePAtkSpd)
	dexBonus := data.GetDEXBonus(p.GetDEX())

	return baseSpd * dexBonus
}

// GetAttackRange returns physical attack range in game units.
// Weapon overrides template base range (fists=20 → sword=40 → bow=500).
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: CreatureStat.java:591-605
func (p *Player) GetAttackRange() int32 {
	// If weapon equipped, use weapon range
	if weapon := p.GetEquippedWeapon(); weapon != nil {
		return weapon.Template().AttackRange
	}

	// Fists: use template base range
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		return 20 // Fallback (typical fists range)
	}
	return template.BaseAtkRange
}

// GetAttackDelay returns delay between attacks (attack speed).
// MVP: simplified formula for unarmed player.
//
// Formula: 500000 / PAtkSpd (in milliseconds).
// Example: 300 PAtkSpd → 1666ms delay (~0.6 attacks/sec).
//
// Phase 5.3: Basic Combat System.
// Java reference: Creature.getAttackEndTime() (line 5419, 5433).
func (p *Player) GetAttackDelay() time.Duration {
	pAtkSpd := p.GetPAtkSpd()

	// Formula: 500000 / PAtkSpd (в миллисекундах)
	// Typical value: 300 PAtkSpd → 1666ms delay
	delayMs := int(500000 / pAtkSpd)

	return time.Duration(delayMs) * time.Millisecond
}

// GetLevelMod returns level modifier для stat scaling.
// Formula: (level + 89) / 100.0
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: Creature.getLevelMod() (CreatureTemplate.java)
func (p *Player) GetLevelMod() float64 {
	return float64(p.Level()+89) / 100.0
}

// GetSTR returns current STR attribute (base + henna bonus).
// Phase 5.4: Character Templates & Stats System.
// Phase 13: Added henna stat bonus.
func (p *Player) GetSTR() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(40) // Default fallback (Human Fighter base STR)
	if template != nil {
		base = template.BaseSTR
	}
	return uint8(int32(base) + p.HennaStatSTR())
}

// GetINT returns current INT attribute (base + henna bonus).
func (p *Player) GetINT() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(21) // Default fallback
	if template != nil {
		base = template.BaseINT
	}
	return uint8(int32(base) + p.HennaStatINT())
}

// GetCON returns current CON attribute (base + henna bonus).
func (p *Player) GetCON() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(43) // Default fallback
	if template != nil {
		base = template.BaseCON
	}
	return uint8(int32(base) + p.HennaStatCON())
}

// GetMEN returns current MEN attribute (base + henna bonus).
func (p *Player) GetMEN() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(25) // Default fallback
	if template != nil {
		base = template.BaseMEN
	}
	return uint8(int32(base) + p.HennaStatMEN())
}

// GetDEX returns current DEX attribute (base + henna bonus).
func (p *Player) GetDEX() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(30) // Default fallback
	if template != nil {
		base = template.BaseDEX
	}
	return uint8(int32(base) + p.HennaStatDEX())
}

// GetWIT returns current WIT attribute (base + henna bonus).
func (p *Player) GetWIT() uint8 {
	template := data.GetTemplate(uint8(p.ClassID()))
	base := uint8(11) // Default fallback
	if template != nil {
		base = template.BaseWIT
	}
	return uint8(int32(base) + p.HennaStatWIT())
}

// GetBasePDef returns base physical defense (nude, no equipment).
// Equipment defense is handled by GetPDef() which subtracts slot defs and adds armor.
//
// Formula: basePDef × levelMod
// where levelMod = (level + 89) / 100.0
//
// Java reference: FuncPDefMod.java:45-87
func (p *Player) GetBasePDef() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		// Fallback to Character.GetBasePDef
		return p.Character.GetBasePDef()
	}

	basePDef := float64(template.BasePDef) // Nude defense (sum of all slots)
	levelMod := p.GetLevelMod()

	finalPDef := basePDef * levelMod
	return int32(finalPDef)
}

// GetPDef returns final physical defense WITH armor (slot subtraction!).
// Armor pDef uses slot-based subtraction: subtract base slot def for equipped slots,
// then add armor pDef, then apply level modifier.
//
// Formula: (basePDef - equippedSlotsDef + armorPDef) × levelMod
//
// Example (Human Fighter level 10):
//   - Nude: 80 × 0.99 = 79.2 ≈ 79
//   - Leather Shirt (chest pDef=43): (80 - 31 + 43) × 0.99 = 92 × 0.99 = 91.08 ≈ 91
//   - Full Plate (chest=100, legs=50): (80 - 31 - 18 + 100 + 50) × 0.99 = 181 × 0.99 = 179.19 ≈ 179
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: FuncPDefMod.java:44-88
func (p *Player) GetPDef() int32 {
	template := data.GetTemplate(uint8(p.ClassID()))
	if template == nil {
		// Fallback to GetBasePDef (nude)
		return p.GetBasePDef()
	}

	// Start with nude pDef (sum of all empty slot defs)
	basePDef := float64(template.BasePDef)

	// Armor slots that contribute to pDef
	slots := []int32{
		PaperdollChest, PaperdollLegs, PaperdollHead,
		PaperdollFeet, PaperdollGloves, PaperdollUnder, PaperdollCloak,
	}

	// Subtract equipped slot base defs
	for _, slot := range slots {
		if p.GetEquippedArmor(slot) != nil {
			// Slot occupied → subtract base slot def
			templateSlot := paperdollSlotToTemplateSlot(slot)
			if slotDef, exists := template.SlotDef[templateSlot]; exists {
				basePDef -= float64(slotDef)
			}
		}
	}

	// Add armor pDef
	armorPDef := float64(0)
	for _, slot := range slots {
		if armor := p.GetEquippedArmor(slot); armor != nil {
			armorPDef += float64(armor.Template().PDef)
		}
	}

	// Apply level modifier
	levelMod := p.GetLevelMod()
	finalPDef := (basePDef + armorPDef) * levelMod

	return int32(finalPDef)
}

// paperdollSlotToTemplateSlot converts paperdoll slot index to template SlotDef key.
// Used for armor slot subtraction in GetPDef().
//
// Phase 5.5: Weapon & Equipment System.
func paperdollSlotToTemplateSlot(paperdollSlot int32) uint8 {
	switch paperdollSlot {
	case PaperdollChest:
		return data.SlotChest
	case PaperdollLegs:
		return data.SlotLegs
	case PaperdollHead:
		return data.SlotHead
	case PaperdollFeet:
		return data.SlotFeet
	case PaperdollGloves:
		return data.SlotGloves
	case PaperdollUnder:
		return data.SlotUnderwear
	case PaperdollCloak:
		return data.SlotCloak
	default:
		return 0 // Unknown slot
	}
}

// DoAttack выполняет физическую атаку на target.
// NOTE: This is a no-op stub to satisfy signature requirements.
// Actual combat logic is in combat.CombatMgr.ExecuteAttack() (called from handler).
//
// Phase 5.3: Basic Combat System (stub to avoid import cycle).
func (p *Player) DoAttack(target *WorldObject) {
	// No-op: Combat logic delegated to combat.CombatMgr to avoid import cycle.
	// Handler calls combat.CombatMgr.ExecuteAttack() directly.
}

// --- Skills (Phase 5.9.2) ---

// AddSkill adds or updates a skill in player's collection.
// Thread-safe: acquires write lock.
func (p *Player) AddSkill(skillID, level int32, passive bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.skills == nil {
		p.skills = make(map[int32]*SkillInfo)
	}
	p.skills[skillID] = &SkillInfo{
		SkillID: skillID,
		Level:   level,
		Passive: passive,
	}
}

// GetSkill returns SkillInfo for given skill ID.
// Returns nil if skill not learned.
// Thread-safe: acquires read lock.
func (p *Player) GetSkill(skillID int32) *SkillInfo {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.skills == nil {
		return nil
	}
	return p.skills[skillID]
}

// HasSkill returns true if player has learned the given skill.
// Thread-safe: acquires read lock.
func (p *Player) HasSkill(skillID int32) bool {
	return p.GetSkill(skillID) != nil
}

// Skills returns a copy of all learned skills.
// Thread-safe: acquires read lock.
func (p *Player) Skills() []*SkillInfo {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.skills == nil {
		return nil
	}

	result := make([]*SkillInfo, 0, len(p.skills))
	for _, s := range p.skills {
		result = append(result, s)
	}
	return result
}

// SkillCount returns the number of learned skills.
// Thread-safe: acquires read lock.
func (p *Player) SkillCount() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	return len(p.skills)
}

// RemoveSkill removes a skill from player's collection.
// Thread-safe: acquires write lock.
func (p *Player) RemoveSkill(skillID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	delete(p.skills, skillID)
}

// --- Effect Manager (Phase 5.9.3) ---

// SetEffectManager sets the effect manager for this player.
// Called during player initialization.
func (p *Player) SetEffectManager(em StatBonusProvider) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.effectManager = em
}

// EffectManager returns the effect manager.
func (p *Player) EffectManager() StatBonusProvider {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.effectManager
}

// --- Party System (Phase 7.3) ---

// PartyInvite tracks a pending party invite from another player.
type PartyInvite struct {
	FromObjectID uint32
	FromName     string
	LootRule     int32
}

// GetParty returns the player's current party (nil if not in a party).
func (p *Player) GetParty() *Party {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.party
}

// SetParty sets or clears the player's party membership.
func (p *Player) SetParty(party *Party) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.party = party
}

// IsInParty returns true if the player is in a party.
func (p *Player) IsInParty() bool {
	return p.GetParty() != nil
}

// PendingPartyInvite returns the pending party invite (nil if none).
func (p *Player) PendingPartyInvite() *PartyInvite {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pendingPartyInvite
}

// SetPendingPartyInvite sets a pending party invite.
func (p *Player) SetPendingPartyInvite(invite *PartyInvite) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pendingPartyInvite = invite
}

// ClearPendingPartyInvite clears the pending party invite.
func (p *Player) ClearPendingPartyInvite() {
	p.SetPendingPartyInvite(nil)
}

// --- Clan Invite (Phase 18) ---

// ClanInvite tracks a pending clan invite from a clan leader/officer.
type ClanInvite struct {
	ClanID     int32
	ClanName   string
	InviterID  uint32 // ObjectID of the player who sent the invite
	PledgeType int32  // Sub-pledge target (0=main, -1=academy, etc.)
}

// PendingClanInvite returns the pending clan invite (nil if none).
func (p *Player) PendingClanInvite() *ClanInvite {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pendingClanInvite
}

// SetPendingClanInvite sets a pending clan invite.
func (p *Player) SetPendingClanInvite(invite *ClanInvite) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pendingClanInvite = invite
}

// ClearPendingClanInvite clears the pending clan invite.
func (p *Player) ClearPendingClanInvite() {
	p.SetPendingClanInvite(nil)
}

// --- Recipe Book (Phase 15) ---

// LearnRecipe adds a recipe to the player's recipe book.
// Returns error if already learned.
func (p *Player) LearnRecipe(recipeID int32, isDwarven bool) error {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	book := p.commonRecipes
	if isDwarven {
		book = p.dwarvenRecipes
	}
	if book == nil {
		if isDwarven {
			p.dwarvenRecipes = make(map[int32]struct{})
			book = p.dwarvenRecipes
		} else {
			p.commonRecipes = make(map[int32]struct{})
			book = p.commonRecipes
		}
	}

	if _, exists := book[recipeID]; exists {
		return fmt.Errorf("recipe %d already learned", recipeID)
	}
	book[recipeID] = struct{}{}
	return nil
}

// ForgetRecipe removes a recipe from the player's recipe book.
// Returns error if not learned.
func (p *Player) ForgetRecipe(recipeID int32, isDwarven bool) error {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	book := p.commonRecipes
	if isDwarven {
		book = p.dwarvenRecipes
	}
	if book == nil {
		return fmt.Errorf("recipe %d not learned", recipeID)
	}
	if _, exists := book[recipeID]; !exists {
		return fmt.Errorf("recipe %d not learned", recipeID)
	}
	delete(book, recipeID)
	return nil
}

// HasRecipe checks if the player has learned a specific recipe.
func (p *Player) HasRecipe(recipeID int32) bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if p.dwarvenRecipes != nil {
		if _, ok := p.dwarvenRecipes[recipeID]; ok {
			return true
		}
	}
	if p.commonRecipes != nil {
		if _, ok := p.commonRecipes[recipeID]; ok {
			return true
		}
	}
	return false
}

// GetRecipeBook returns a list of recipe IDs for the given type.
func (p *Player) GetRecipeBook(isDwarven bool) []int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	book := p.commonRecipes
	if isDwarven {
		book = p.dwarvenRecipes
	}
	if book == nil {
		return nil
	}

	result := make([]int32, 0, len(book))
	for id := range book {
		result = append(result, id)
	}
	return result
}

// RecipeCount returns total number of learned recipes (both types).
func (p *Player) RecipeCount() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return len(p.dwarvenRecipes) + len(p.commonRecipes)
}

// AccessLevel returns the player's access level.
// 0 = normal player, 1+ = GM, negative = banned.
// Phase 17: Admin Commands.
func (p *Player) AccessLevel() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.accessLevel
}

// SetAccessLevel sets the player's access level.
// Phase 17: Admin Commands.
func (p *Player) SetAccessLevel(level int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.accessLevel = level
}

// IsGM returns true if player is a Game Master (accessLevel > 0).
// Phase 17: Admin Commands.
func (p *Player) IsGM() bool {
	return p.AccessLevel() > 0
}

// LastAdminMessage returns the last message from admin/user command system.
// Used to relay command feedback to the client.
// Phase 17: Admin Commands.
func (p *Player) LastAdminMessage() string {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.lastAdminMessage
}

// SetLastAdminMessage stores a message from admin/user command system.
// Phase 17: Admin Commands.
func (p *Player) SetLastAdminMessage(msg string) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.lastAdminMessage = msg
}

// ClearLastAdminMessage clears the last admin message after it has been sent.
// Phase 17: Admin Commands.
func (p *Player) ClearLastAdminMessage() string {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	msg := p.lastAdminMessage
	p.lastAdminMessage = ""
	return msg
}

// IsInvisible returns true if player is in GM invisible mode.
// Phase 17: Admin Commands.
func (p *Player) IsInvisible() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.invisible
}

// SetInvisible sets the GM invisible mode.
// Phase 17: Admin Commands.
func (p *Player) SetInvisible(invisible bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.invisible = invisible
}

// IsInvulnerable returns true if player is in GM invulnerable mode.
// Phase 17: Admin Commands.
func (p *Player) IsInvulnerable() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.invulnerable
}

// SetInvulnerable sets the GM invulnerable mode.
// Phase 17: Admin Commands.
func (p *Player) SetInvulnerable(invul bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.invulnerable = invul
}

// ClanID returns the player's clan ID (0 if not in a clan).
func (p *Player) ClanID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.clanID
}

// SetClanID sets the player's clan ID.
func (p *Player) SetClanID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.clanID = id
}

// ClanTitle returns the title assigned by the clan leader.
func (p *Player) ClanTitle() string {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.clanTitle
}

// SetClanTitle sets the title assigned by the clan leader.
func (p *Player) SetClanTitle(title string) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.clanTitle = title
}

// Summon returns the active summon (pet/servitor), or nil if none.
// Phase 19: Pets/Summons System.
func (p *Player) Summon() *Summon {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.summon
}

// SetSummon sets the active summon.
// Phase 19: Pets/Summons System.
func (p *Player) SetSummon(summon *Summon) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.summon = summon
}

// HasSummon returns true if player has an active summon.
// Phase 19: Pets/Summons System.
func (p *Player) HasSummon() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.summon != nil
}

// ClearSummon removes the active summon.
// Phase 19: Pets/Summons System.
func (p *Player) ClearSummon() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.summon = nil
}

// DuelRequest represents a pending duel invitation.
// Phase 20: Duel System.
type DuelRequest struct {
	RequestorID   uint32 // ObjectID of the challenger
	RequestorName string // Name of the challenger
	PartyDuel     bool   // true = party duel, false = 1v1
}

// DuelID returns the active duel ID (0 = not in duel).
// Phase 20: Duel System.
func (p *Player) DuelID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.duelID
}

// SetDuelID sets the active duel ID.
// Phase 20: Duel System.
func (p *Player) SetDuelID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.duelID = id
}

// IsInDuel returns true if the player is in an active duel.
// Phase 20: Duel System.
func (p *Player) IsInDuel() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.duelID != 0
}

// PendingDuelRequest returns the pending duel request, or nil if none.
// Phase 20: Duel System.
func (p *Player) PendingDuelRequest() *DuelRequest {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pendingDuelRequest
}

// SetPendingDuelRequest sets a pending duel invitation.
// Phase 20: Duel System.
func (p *Player) SetPendingDuelRequest(req *DuelRequest) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pendingDuelRequest = req
}

// ClearPendingDuelRequest clears the pending duel invitation.
// Phase 20: Duel System.
func (p *Player) ClearPendingDuelRequest() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pendingDuelRequest = nil
}

// Karma returns the player's karma value.
// Phase 32: Cursed Weapons.
func (p *Player) Karma() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.karma
}

// SetKarma sets the player's karma value.
// Phase 32: Cursed Weapons.
func (p *Player) SetKarma(karma int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.karma = karma
}

// PKKills returns the player's PK kill count.
// Phase 32: Cursed Weapons.
func (p *Player) PKKills() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pkKills
}

// SetPKKills sets the player's PK kill count.
// Phase 32: Cursed Weapons.
func (p *Player) SetPKKills(count int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pkKills = count
}

// CursedWeaponEquippedID returns the item ID of the equipped cursed weapon (0=none).
// Phase 32: Cursed Weapons.
func (p *Player) CursedWeaponEquippedID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.cursedWeaponEquippedID
}

// SetCursedWeaponEquippedID sets the equipped cursed weapon item ID.
// Phase 32: Cursed Weapons.
func (p *Player) SetCursedWeaponEquippedID(itemID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.cursedWeaponEquippedID = itemID
}

// IsCursedWeaponEquipped returns true if the player has a cursed weapon equipped.
// Phase 32: Cursed Weapons.
func (p *Player) IsCursedWeaponEquipped() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.cursedWeaponEquippedID != 0
}

// --- Marriage System (Phase 33) ---

// PartnerID returns the ObjectID of this player's partner (0 = not engaged).
func (p *Player) PartnerID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.partnerID
}

// SetPartnerID sets the partner ObjectID.
func (p *Player) SetPartnerID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.partnerID = id
}

// CoupleID returns the couple row ID in the database.
func (p *Player) CoupleID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.coupleID
}

// SetCoupleID sets the couple row ID.
func (p *Player) SetCoupleID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.coupleID = id
}

// IsMarried returns true if the player has completed a marriage ceremony.
func (p *Player) IsMarried() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.married
}

// SetMarried sets the married flag.
func (p *Player) SetMarried(v bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.married = v
}

// IsEngageRequest returns true if this player has a pending engage proposal.
func (p *Player) IsEngageRequest() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.engageRequest
}

// SetEngageRequest sets or clears the pending engage request.
func (p *Player) SetEngageRequest(state bool, fromID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.engageRequest = state
	p.engageFromID = fromID
}

// EngageFromID returns the ObjectID of the player who proposed, or 0.
func (p *Player) EngageFromID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.engageFromID
}

// ClearMarriageState resets all marriage-related fields.
// Used on divorce or when the couple record is deleted.
func (p *Player) ClearMarriageState() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.partnerID = 0
	p.coupleID = 0
	p.married = false
	p.engageRequest = false
	p.engageFromID = 0
}

// ActiveEnchantItemID returns the ObjectID of the active enchant scroll.
// 0 means no enchant in progress.
func (p *Player) ActiveEnchantItemID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.activeEnchantItemID
}

// SetActiveEnchantItemID sets the active enchant scroll ObjectID.
// Pass 0 to clear the enchant state.
func (p *Player) SetActiveEnchantItemID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.activeEnchantItemID = id
}

// ActiveTradeList returns the current P2P trade session (nil if not trading).
func (p *Player) ActiveTradeList() *P2PTradeList {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.activeTradeList
}

// SetActiveTradeList sets the current P2P trade session.
func (p *Player) SetActiveTradeList(t *P2PTradeList) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.activeTradeList = t
}

// ActiveRequester returns the player who requested a trade (nil if none).
func (p *Player) ActiveRequester() *Player {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.activeRequester
}

// SetActiveRequester sets the player who requested a trade.
func (p *Player) SetActiveRequester(r *Player) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.activeRequester = r
}

// IsRequestExpired returns true if the pending trade request has expired.
func (p *Player) IsRequestExpired() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	if p.requestExpireTime == 0 {
		return true
	}
	return time.Now().Unix() > p.requestExpireTime
}

// OnTransactionRequest records a trade request from partner.
// Sets expire time to 10 seconds from now.
func (p *Player) OnTransactionRequest(partner *Player) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.activeRequester = partner
	p.requestExpireTime = time.Now().Unix() + 10 // 10s expire
}

// OnTransactionResponse clears the request expire time after responding.
func (p *Player) OnTransactionResponse() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.requestExpireTime = 0
}

// IsProcessingTransaction returns true if player is in an active P2P trade.
func (p *Player) IsProcessingTransaction() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.activeTradeList != nil
}

// CancelActiveTrade cancels the P2P trade for both this player and the partner.
// Clears trade state on both sides. Safe to call when no trade is active.
func (p *Player) CancelActiveTrade() {
	p.playerMu.Lock()
	tradeList := p.activeTradeList
	p.activeTradeList = nil
	p.activeRequester = nil
	p.requestExpireTime = 0
	p.playerMu.Unlock()

	if tradeList == nil {
		return
	}

	// Clear partner's trade state
	partner := tradeList.Partner()
	if partner != nil && partner != p {
		partner.playerMu.Lock()
		partner.activeTradeList = nil
		partner.activeRequester = nil
		partner.requestExpireTime = 0
		partner.playerMu.Unlock()
	}
}

// --- Shortcuts (Phase 34) ---

// RegisterShortcut registers or replaces a shortcut in the action bar.
// Thread-safe: acquires write lock.
func (p *Player) RegisterShortcut(sc *Shortcut) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.shortcuts == nil {
		p.shortcuts = make(map[int32]*Shortcut)
	}
	p.shortcuts[shortcutKey(sc.Slot, sc.Page)] = sc
}

// DeleteShortcut removes a shortcut from the action bar.
// Thread-safe: acquires write lock.
func (p *Player) DeleteShortcut(slot, page int8) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	delete(p.shortcuts, shortcutKey(slot, page))
}

// GetShortcuts returns a copy of all registered shortcuts.
// Thread-safe: acquires read lock.
func (p *Player) GetShortcuts() []*Shortcut {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if len(p.shortcuts) == 0 {
		return nil
	}

	result := make([]*Shortcut, 0, len(p.shortcuts))
	for _, sc := range p.shortcuts {
		result = append(result, sc)
	}
	return result
}

// SetShortcuts replaces all shortcuts with the given set (DB load).
// Thread-safe: acquires write lock.
func (p *Player) SetShortcuts(shortcuts []*Shortcut) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.shortcuts = make(map[int32]*Shortcut, len(shortcuts))
	for _, sc := range shortcuts {
		p.shortcuts[shortcutKey(sc.Slot, sc.Page)] = sc
	}
}

// GetSkillLevel returns the level of a learned skill (0 if not learned).
// Convenience method for shortcut registration.
// Thread-safe: acquires read lock.
func (p *Player) GetSkillLevel(skillID int32) int32 {
	si := p.GetSkill(skillID)
	if si == nil {
		return 0
	}
	return si.Level
}

// Auto SoulShot and Macro methods are in macro.go (Phase 36).

// --- Item Cooldown Methods (Phase 51: Item Handler System) ---

// IsItemOnCooldown returns true if the item is still on cooldown.
func (p *Player) IsItemOnCooldown(itemID int32) bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	if p.itemCooldowns == nil {
		return false
	}
	expiry, ok := p.itemCooldowns[itemID]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

// SetItemCooldown sets item use cooldown duration from now.
func (p *Player) SetItemCooldown(itemID int32, duration time.Duration) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	if p.itemCooldowns == nil {
		p.itemCooldowns = make(map[int32]time.Time)
	}
	p.itemCooldowns[itemID] = time.Now().Add(duration)
}

// --- Olympiad State (Phase 51 stub) ---

// IsInOlympiad returns true if the player is currently in an Olympiad match.
func (p *Player) IsInOlympiad() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.inOlympiad
}

// SetInOlympiad sets the Olympiad state flag.
func (p *Player) SetInOlympiad(inOly bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.inOlympiad = inOly
}
