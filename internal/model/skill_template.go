package model

// SkillInfo описывает изученный скилл игрока (хранится в Player.skills).
// Содержит только ID и Level — шаблон хранится в data.SkillTable.
//
// Phase 5.9.2: Player Skills.
type SkillInfo struct {
	SkillID int32
	Level   int32
	Passive bool
}
