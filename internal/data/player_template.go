package data

// PlayerTemplate содержит базовые характеристики класса персонажа.
// Загружается из XML файлов (L2J Mobius templates).
//
// Phase 5.4: Character Templates & Stats System.
type PlayerTemplate struct {
	ClassID   uint8  // 0-118 (Interlude: 0-87)
	ClassName string // "Human Fighter", "Elf Mystic"
	ParentID  uint8  // Parent class ID (для иерархии)

	// Base attributes (level 1)
	BaseSTR uint8
	BaseCON uint8
	BaseDEX uint8
	BaseINT uint8
	BaseWIT uint8
	BaseMEN uint8

	// Combat stats (BASE, before modifiers)
	BasePAtk     int32 // Physical Attack (base)
	BaseMAtk     int32 // Magic Attack (base)
	BasePDef     int32 // Physical Defense (sum of all slot defs)
	BaseMDef     int32 // Magic Defense (sum of all slot defs)
	BasePAtkSpd  int32 // Physical Attack Speed (default: 300)
	BaseMAtkSpd  int32 // Magic Attack Speed (default: 333)
	BaseCritRate int32 // Critical Rate (default: 4)
	BaseMCritRate int32 // Magic Critical Rate (default: 5)
	BaseAtkRange int32 // Attack Range (default: 40)
	RandomDamage int32 // Random damage range

	// Slot-based defense (для пустых слотов)
	// При экипировке предмета — вычитаем базовую защиту слота
	SlotDef map[uint8]int32 // map[slotID]defense

	// Collision (for client)
	CollisionRadiusMale   float32
	CollisionHeightMale   float32
	CollisionRadiusFemale float32
	CollisionHeightFemale float32

	// Level progression (arrays [0..79] for levels 1-80)
	HPByLevel      []float32 // HP at each level
	MPByLevel      []float32 // MP at each level
	CPByLevel      []float32 // CP at each level
	HPRegenByLevel []float64 // HP regen at each level
	MPRegenByLevel []float64 // MP regen at each level
	CPRegenByLevel []float64 // CP regen at each level

	// Spawn points (для новых персонажей)
	CreationPoints []Location
}

// Location представляет координаты точки спавна.
type Location struct {
	X int32
	Y int32
	Z int32
}

// PlayerTemplates — глобальный registry всех шаблонов классов.
// map[classID]template
// Загружается через LoadPlayerTemplates() при старте сервера.
var PlayerTemplates map[uint8]*PlayerTemplate

// GetTemplate возвращает template по classID.
// Returns nil если класс не найден.
func GetTemplate(classID uint8) *PlayerTemplate {
	return PlayerTemplates[classID]
}

// GetHPMax возвращает максимальный HP для уровня (1-80).
// Returns 0 если level вне диапазона.
func (t *PlayerTemplate) GetHPMax(level int32) float32 {
	if level < 1 || level > 80 {
		return 0
	}
	return t.HPByLevel[level-1]
}

// GetMPMax возвращает максимальный MP для уровня (1-80).
func (t *PlayerTemplate) GetMPMax(level int32) float32 {
	if level < 1 || level > 80 {
		return 0
	}
	return t.MPByLevel[level-1]
}

// GetCPMax возвращает максимальный CP для уровня (1-80).
func (t *PlayerTemplate) GetCPMax(level int32) float32 {
	if level < 1 || level > 80 {
		return 0
	}
	return t.CPByLevel[level-1]
}

// GetSlotDef возвращает базовую защиту пустого слота.
// Returns 0 если слот не найден.
func (t *PlayerTemplate) GetSlotDef(slotID uint8) int32 {
	return t.SlotDef[slotID]
}
