package data

// ItemReq describes an item requirement for learning a skill.
type ItemReq struct {
	ItemID int32
	Count  int32
}

// SkillLearn описывает навык в дереве скиллов класса.
// Определяет условия изучения скилла: уровень, стоимость SP, автоматическое получение.
type SkillLearn struct {
	SkillID      int32     // ID скилла в SkillTable
	SkillLevel   int32     // Level скилла
	MinLevel     int32     // Минимальный уровень игрока для изучения
	SpCost       int64     // Стоимость в SP
	AutoGet      bool      // Автоматически выдаётся при достижении уровня
	ClassID      int32     // ID класса (для фильтрации)
	LearnedByNpc bool      // Изучается у NPC (не autoGet)
	Items        []ItemReq // Необходимые предметы для изучения
}
