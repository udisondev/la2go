package model

import "fmt"

const (
	MaxMacros        = 24  // max macros per player
	MaxMacroCommands = 12  // max commands per macro
	MaxMacroCmdLen   = 255 // total command string length
	MaxMacroDescLen  = 32  // max description length
)

// MacroCmdType defines the type of a macro command.
type MacroCmdType int8

const (
	MacroCmdSkill    MacroCmdType = 1 // Use skill
	MacroCmdAction   MacroCmdType = 3 // Social action
	MacroCmdShortcut MacroCmdType = 4 // Activate shortcut
)

// MacroCmd is a single command inside a macro.
type MacroCmd struct {
	Entry   int8
	Type    MacroCmdType
	D1      int32 // skill ID or page number
	D2      int8
	Command string
}

// Macro is a player-defined macro (sequence of commands).
type Macro struct {
	ID       int32
	Name     string
	Desc     string
	Acronym  string
	Icon     int8
	Commands []MacroCmd
}

// --- Auto SoulShot methods (Phase 36) ---

// AddAutoSoulShot enables auto-use for the given soulshot item ID.
func (p *Player) AddAutoSoulShot(itemID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.autoSoulShots[itemID] = true
}

// RemoveAutoSoulShot disables auto-use for the given soulshot item ID.
func (p *Player) RemoveAutoSoulShot(itemID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	delete(p.autoSoulShots, itemID)
}

// HasAutoSoulShot checks if the given soulshot item ID is auto-enabled.
func (p *Player) HasAutoSoulShot(itemID int32) bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.autoSoulShots[itemID]
}

// AutoSoulShots returns a copy of all auto-enabled soulshot item IDs.
func (p *Player) AutoSoulShots() []int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	result := make([]int32, 0, len(p.autoSoulShots))
	for id := range p.autoSoulShots {
		result = append(result, id)
	}
	return result
}

// SetAutoSoulShots replaces the auto-soulshot set (used on login restore).
func (p *Player) SetAutoSoulShots(items []int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.autoSoulShots = make(map[int32]bool, len(items))
	for _, id := range items {
		p.autoSoulShots[id] = true
	}
}

// --- Macro methods (Phase 36) ---

// RegisterMacro adds or updates a macro for this player.
func (p *Player) RegisterMacro(m *Macro) error {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if _, exists := p.macros[m.ID]; !exists && len(p.macros) >= MaxMacros {
		return fmt.Errorf("macro limit reached (%d)", MaxMacros)
	}

	p.macros[m.ID] = m
	p.macroRevision++
	return nil
}

// DeleteMacro removes a macro by ID.
func (p *Player) DeleteMacro(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	delete(p.macros, id)
	p.macroRevision++
}

// GetMacro returns a single macro by ID or nil.
func (p *Player) GetMacro(id int32) *Macro {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.macros[id]
}

// GetMacros returns a copy of all player macros.
func (p *Player) GetMacros() []*Macro {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	result := make([]*Macro, 0, len(p.macros))
	for _, m := range p.macros {
		result = append(result, m)
	}
	return result
}

// SetMacros replaces all macros (used on login restore).
func (p *Player) SetMacros(macros []*Macro) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.macros = make(map[int32]*Macro, len(macros))
	for _, m := range macros {
		p.macros[m.ID] = m
	}
}

// MacroRevision returns the current macro revision counter.
func (p *Player) MacroRevision() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.macroRevision
}

// MacroCount returns the number of macros.
func (p *Player) MacroCount() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return len(p.macros)
}
