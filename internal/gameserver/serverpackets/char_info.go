package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeCharInfo is the opcode for CharInfo packet (S2C 0x31)
	OpcodeCharInfo = 0x31
)

// CharInfo packet (S2C 0x31) sends information about another player character.
// Sent when player enters visibility range of another player.
// Similar to UserInfo but for OTHER players (not self).
type CharInfo struct {
	Player *model.Player
}

// NewCharInfo creates CharInfo packet from Player model.
func NewCharInfo(player *model.Player) *CharInfo {
	return &CharInfo{
		Player: player,
	}
}

// Write serializes CharInfo packet to binary format.
// CharInfo is similar to UserInfo but simplified (other players don't need full detail).
func (p *CharInfo) Write() ([]byte, error) {
	w := packet.NewWriter(512)

	loc := p.Player.Location()

	w.WriteByte(OpcodeCharInfo)

	// Position
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)
	w.WriteInt(int32(loc.Heading))

	// Identity
	w.WriteInt(int32(p.Player.CharacterID())) // Object ID
	w.WriteString(p.Player.Name())

	// Race & Class
	w.WriteInt(p.Player.RaceID())
	w.WriteInt(0) // Sex (TODO: add to DB schema)
	w.WriteInt(p.Player.ClassID())

	// Equipped items (17 paperdoll slots)
	// TODO Phase 4.9: load from items table
	for range 17 {
		w.WriteInt(0) // Item ID (0 = empty slot)
	}

	// Appearance (TODO: add to DB schema)
	w.WriteInt(0) // Hair style
	w.WriteInt(0) // Hair color
	w.WriteInt(0) // Face

	// Title
	w.WriteString("") // Title

	// Status
	w.WriteInt(0) // Running (0=walking, 1=running)
	w.WriteInt(0) // In combat (0=peace, 1=combat)
	w.WriteInt(0) // AFK (0=active, 1=away)

	// Mount
	w.WriteInt(0) // Mount type (0=none)

	// Clan
	w.WriteInt(0) // Clan ID
	w.WriteInt(0) // Clan crest ID
	w.WriteInt(0) // Ally ID
	w.WriteInt(0) // Ally crest ID

	// Sitting
	w.WriteInt(0) // Sitting (0=standing, 1=sitting)

	// PvP/Karma
	w.WriteInt(0) // PvP flag
	w.WriteInt(0) // Karma

	// Abnormal Effect
	w.WriteInt(0) // Abnormal effect bitmask (buffs/debuffs)

	// Clan Privileges
	w.WriteInt(0) // Clan privileges

	// Recommendations
	w.WriteShort(0) // Recommendations left
	w.WriteShort(0) // Recommendations received
	w.WriteShort(0) // Recommendations given

	// Mount NPC ID (for strider/wyvern display)
	w.WriteInt(0) // Mount NPC ID

	// Class ID (for transformation display)
	w.WriteInt(p.Player.ClassID())

	// Special Effects
	w.WriteInt(0) // Special effects bitmask

	// Team (for event/olympiad)
	w.WriteByte(0) // Team circle color (0=none, 1=blue, 2=red)

	// Hero
	w.WriteByte(0) // Hero aura

	// Fish
	w.WriteByte(0) // Fishing

	// Fish location
	w.WriteInt(0) // Fish X
	w.WriteInt(0) // Fish Y
	w.WriteInt(0) // Fish Z

	// Name color
	w.WriteInt(0xFFFFFF) // Name color (white)

	// Title color
	w.WriteInt(0xFFFF77) // Title color (yellow)

	return w.Bytes(), nil
}
