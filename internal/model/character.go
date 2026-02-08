package model

// Character — базовый класс для живых существ (Player, NPC).
// Добавляет HP, MP, CP, level к WorldObject.
type Character struct {
	*WorldObject // embedded

	level     int32
	currentHP int32
	maxHP     int32
	currentMP int32
	maxMP     int32
	currentCP int32
	maxCP     int32
}

// NewCharacter создаёт нового персонажа с указанными максимальными значениями.
// Текущие HP/MP/CP устанавливаются равными максимальным.
func NewCharacter(objectID uint32, name string, loc Location, level, maxHP, maxMP, maxCP int32) *Character {
	return &Character{
		WorldObject: NewWorldObject(objectID, name, loc),
		level:       level,
		currentHP:   maxHP,
		maxHP:       maxHP,
		currentMP:   maxMP,
		maxMP:       maxMP,
		currentCP:   maxCP,
		maxCP:       maxCP,
	}
}

// CurrentHP возвращает текущее HP.
func (c *Character) CurrentHP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentHP
}

// MaxHP возвращает максимальное HP.
func (c *Character) MaxHP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxHP
}

// SetCurrentHP устанавливает текущее HP с валидацией (clamp 0..maxHP).
func (c *Character) SetCurrentHP(hp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hp < 0 {
		hp = 0
	}
	if hp > c.maxHP {
		hp = c.maxHP
	}
	c.currentHP = hp
}

// SetMaxHP устанавливает максимальное HP и корректирует текущее если нужно.
func (c *Character) SetMaxHP(maxHP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxHP < 1 {
		maxHP = 1
	}

	c.maxHP = maxHP

	// Если текущее HP больше нового максимума — обрезаем
	if c.currentHP > c.maxHP {
		c.currentHP = c.maxHP
	}
}

// CurrentMP возвращает текущее MP.
func (c *Character) CurrentMP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMP
}

// MaxMP возвращает максимальное MP.
func (c *Character) MaxMP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxMP
}

// SetCurrentMP устанавливает текущее MP с валидацией (clamp 0..maxMP).
func (c *Character) SetCurrentMP(mp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if mp < 0 {
		mp = 0
	}
	if mp > c.maxMP {
		mp = c.maxMP
	}
	c.currentMP = mp
}

// SetMaxMP устанавливает максимальное MP и корректирует текущее если нужно.
func (c *Character) SetMaxMP(maxMP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxMP < 0 {
		maxMP = 0
	}

	c.maxMP = maxMP

	// Если текущее MP больше нового максимума — обрезаем
	if c.currentMP > c.maxMP {
		c.currentMP = c.maxMP
	}
}

// CurrentCP возвращает текущее CP.
func (c *Character) CurrentCP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentCP
}

// MaxCP возвращает максимальное CP.
func (c *Character) MaxCP() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxCP
}

// SetCurrentCP устанавливает текущее CP с валидацией (clamp 0..maxCP).
func (c *Character) SetCurrentCP(cp int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cp < 0 {
		cp = 0
	}
	if cp > c.maxCP {
		cp = c.maxCP
	}
	c.currentCP = cp
}

// SetMaxCP устанавливает максимальное CP и корректирует текущее если нужно.
func (c *Character) SetMaxCP(maxCP int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maxCP < 0 {
		maxCP = 0
	}

	c.maxCP = maxCP

	// Если текущее CP больше нового максимума — обрезаем
	if c.currentCP > c.maxCP {
		c.currentCP = c.maxCP
	}
}

// IsDead проверяет мёртв ли персонаж (HP <= 0).
func (c *Character) IsDead() bool {
	return c.CurrentHP() <= 0
}

// HPPercentage возвращает процент текущего HP (0.0 - 1.0).
func (c *Character) HPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxHP == 0 {
		return 0.0
	}
	return float64(c.currentHP) / float64(c.maxHP)
}

// MPPercentage возвращает процент текущего MP (0.0 - 1.0).
func (c *Character) MPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxMP == 0 {
		return 0.0
	}
	return float64(c.currentMP) / float64(c.maxMP)
}

// CPPercentage возвращает процент текущего CP (0.0 - 1.0).
func (c *Character) CPPercentage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxCP == 0 {
		return 0.0
	}
	return float64(c.currentCP) / float64(c.maxCP)
}

// Level возвращает уровень персонажа.
func (c *Character) Level() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.level
}

// SetLevel устанавливает уровень персонажа (clamp 1..100).
func (c *Character) SetLevel(level int32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}
	c.level = level
}
