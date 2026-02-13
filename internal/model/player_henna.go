package model

import (
	"fmt"

	"github.com/udisondev/la2go/internal/data"
)

// MaxHennaSlots — maximum henna slots per player.
const MaxHennaSlots = 3

// maxHennaStat — maximum stat bonus from hennas (+5 cap per stat).
const maxHennaStat = 5

// HennaSlot stores a henna equipped in a specific slot.
type HennaSlot struct {
	DyeID int32
}

// hennaStats кэширует суммарные модификаторы статов от всех хенн.
type hennaStats struct {
	str int32
	con int32
	dex int32
	inT int32 // "int" is reserved keyword
	men int32
	wit int32
}

// GetHenna returns the henna in given slot (1-3). Returns nil if slot is empty.
// Thread-safe: acquires read lock.
func (p *Player) GetHenna(slot int) *HennaSlot {
	if slot < 1 || slot > MaxHennaSlots {
		return nil
	}
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennas[slot-1]
}

// GetHennaList returns a copy of all equipped hennas (some may be nil).
// Thread-safe: acquires read lock.
func (p *Player) GetHennaList() [MaxHennaSlots]*HennaSlot {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennas
}

// GetHennaEmptySlots returns number of free henna slots.
// Thread-safe: acquires read lock.
func (p *Player) GetHennaEmptySlots() int {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	empty := 0
	for _, h := range p.hennas {
		if h == nil {
			empty++
		}
	}
	return empty
}

// HasHennas returns true if player has at least one henna equipped.
func (p *Player) HasHennas() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	for _, h := range p.hennas {
		if h != nil {
			return true
		}
	}
	return false
}

// AddHenna equips a henna into the first free slot.
// Returns the slot number (1-3) on success.
// Thread-safe: acquires write lock.
func (p *Player) AddHenna(dyeID int32) (int, error) {
	def := data.GetHennaDef(dyeID)
	if def == nil {
		return 0, fmt.Errorf("henna %d not found", dyeID)
	}

	if !def.IsAllowedClass(p.ClassID()) {
		return 0, fmt.Errorf("henna %d not allowed for class %d", dyeID, p.ClassID())
	}

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	// Find first free slot
	for i := range MaxHennaSlots {
		if p.hennas[i] == nil {
			p.hennas[i] = &HennaSlot{DyeID: dyeID}
			p.recalcHennaStatsLocked()
			return i + 1, nil
		}
	}

	return 0, fmt.Errorf("no free henna slots")
}

// RemoveHenna removes henna from slot (1-3).
// Returns the removed dye ID.
// Thread-safe: acquires write lock.
func (p *Player) RemoveHenna(slot int) (int32, error) {
	if slot < 1 || slot > MaxHennaSlots {
		return 0, fmt.Errorf("invalid henna slot %d (must be 1-%d)", slot, MaxHennaSlots)
	}

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	h := p.hennas[slot-1]
	if h == nil {
		return 0, fmt.Errorf("henna slot %d is empty", slot)
	}

	dyeID := h.DyeID
	p.hennas[slot-1] = nil
	p.recalcHennaStatsLocked()

	return dyeID, nil
}

// SetHenna sets henna directly into a slot (for DB restore).
// Slot is 1-3. Does NOT validate class restrictions.
// Thread-safe: acquires write lock.
func (p *Player) SetHenna(slot int, dyeID int32) error {
	if slot < 1 || slot > MaxHennaSlots {
		return fmt.Errorf("invalid henna slot %d", slot)
	}

	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.hennas[slot-1] = &HennaSlot{DyeID: dyeID}
	return nil
}

// RecalcHennaStats recalculates cached stat bonuses from all hennas.
// Thread-safe: acquires write lock.
func (p *Player) RecalcHennaStats() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.recalcHennaStatsLocked()
}

// recalcHennaStatsLocked recalculates without locking (caller must hold write lock).
// Java reference: Player.recalcHennaStats() — each stat capped at +5.
func (p *Player) recalcHennaStatsLocked() {
	p.hennaStat = hennaStats{}

	for _, h := range p.hennas {
		if h == nil {
			continue
		}
		def := data.GetHennaDef(h.DyeID)
		if def == nil {
			continue
		}

		p.hennaStat.str = addHennaStat(p.hennaStat.str, def.StatSTR())
		p.hennaStat.con = addHennaStat(p.hennaStat.con, def.StatCON())
		p.hennaStat.dex = addHennaStat(p.hennaStat.dex, def.StatDEX())
		p.hennaStat.inT = addHennaStat(p.hennaStat.inT, def.StatINT())
		p.hennaStat.men = addHennaStat(p.hennaStat.men, def.StatMEN())
		p.hennaStat.wit = addHennaStat(p.hennaStat.wit, def.StatWIT())
	}
}

// addHennaStat adds stat bonus with +5 cap.
// Java: _hennaSTR += ((_hennaSTR + h.getStatSTR()) > 5) ? 5 - _hennaSTR : h.getStatSTR()
func addHennaStat(current, bonus int32) int32 {
	if current+bonus > maxHennaStat {
		return maxHennaStat
	}
	return current + bonus
}

// HennaStatSTR returns henna STR bonus.
func (p *Player) HennaStatSTR() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.str
}

// HennaStatCON returns henna CON bonus.
func (p *Player) HennaStatCON() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.con
}

// HennaStatDEX returns henna DEX bonus.
func (p *Player) HennaStatDEX() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.dex
}

// HennaStatINT returns henna INT bonus.
func (p *Player) HennaStatINT() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.inT
}

// HennaStatMEN returns henna MEN bonus.
func (p *Player) HennaStatMEN() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.men
}

// HennaStatWIT returns henna WIT bonus.
func (p *Player) HennaStatWIT() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hennaStat.wit
}
