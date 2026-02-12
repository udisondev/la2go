package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeUserInfo is the opcode for UserInfo packet (S2C 0x32)
	OpcodeUserInfo = 0x32
)

// UserInfo packet (S2C 0x32) sends complete character information to the client.
// Sent when character enters world (after EnterWorld packet).
// This is the primary packet that makes the character visible in the game world.
type UserInfo struct {
	Player *model.Player
}

// NewUserInfo creates UserInfo packet from Player model.
func NewUserInfo(player *model.Player) UserInfo {
	return UserInfo{
		Player: player,
	}
}

// Write serializes UserInfo packet to binary format.
// Returns byte slice containing full packet data (opcode + payload).
//
// UserInfo is one of the most complex L2 packets (~100+ fields).
// Contains everything needed to display character: position, stats, appearance, equipment.
func (p *UserInfo) Write() ([]byte, error) {
	// Estimate buffer size: ~500 bytes for UserInfo
	w := packet.NewWriter(512)

	loc := p.Player.Location()

	w.WriteByte(OpcodeUserInfo)

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
	w.WriteInt(0) // Sex (TODO: add to DB schema in Phase 4.7)
	w.WriteInt(p.Player.ClassID())

	// Level & Stats
	w.WriteInt(p.Player.Level())
	w.WriteLong(p.Player.Experience())
	w.WriteInt(int32(p.Player.SP()))

	// Vitals (current values)
	w.WriteInt(p.Player.CurrentHP())
	w.WriteInt(p.Player.MaxHP())
	w.WriteInt(p.Player.CurrentMP())
	w.WriteInt(p.Player.MaxMP())
	w.WriteInt(p.Player.CurrentCP())
	w.WriteInt(p.Player.MaxCP())

	// Combat Stats (TODO: calculate from formulas in Phase 5.1)
	w.WriteInt(0) // P.Atk
	w.WriteInt(0) // Atk.Speed
	w.WriteInt(0) // P.Def
	w.WriteInt(0) // Evasion
	w.WriteInt(0) // Accuracy
	w.WriteInt(0) // Critical Rate
	w.WriteInt(0) // M.Atk
	w.WriteInt(0) // Cast Speed
	w.WriteInt(0) // M.Def
	w.WriteInt(0) // PvP Flag (0=white, 1=purple flagged)
	w.WriteInt(0) // Karma

	// Movement
	w.WriteInt(80)  // Run speed (default human run speed)
	w.WriteInt(40)  // Walk speed (default human walk speed)
	w.WriteInt(80)  // Swim run speed
	w.WriteInt(40)  // Swim walk speed
	w.WriteInt(80)  // Fly run speed (unused in Interlude)
	w.WriteInt(40)  // Fly walk speed (unused in Interlude)
	w.WriteInt(80)  // Fly run speed (unused in Interlude)
	w.WriteInt(40)  // Fly walk speed (unused in Interlude)
	w.WriteDouble(1.0) // Movement speed multiplier

	// Attack Speed Multiplier
	w.WriteDouble(1.0) // Attack speed multiplier

	// Collision Radius & Height (default human values)
	w.WriteDouble(8.0)  // Collision radius
	w.WriteDouble(23.0) // Collision height

	// Appearance (TODO: add to DB schema in Phase 4.7)
	w.WriteInt(0) // Hair style
	w.WriteInt(0) // Hair color
	w.WriteInt(0) // Face

	// Title
	w.WriteString("") // Title (empty by default)

	// Clan
	w.WriteInt(0) // Clan ID (TODO: Phase 4.8 clan system)
	w.WriteInt(0) // Clan crest ID
	w.WriteInt(0) // Ally ID
	w.WriteInt(0) // Ally crest ID

	// Status
	w.WriteInt(0) // Sitting (0=standing, 1=sitting)
	w.WriteInt(0) // Running (0=walking, 1=running)
	w.WriteInt(0) // In combat (0=peace, 1=combat)
	w.WriteInt(0) // AFK (0=active, 1=away)

	// Mount
	w.WriteInt(0) // Mount type (0=none, 1=strider, 2=wyvern)
	w.WriteInt(0) // Private store type (0=none, 1=sell, 2=buy)

	// Cubics
	w.WriteInt(0) // Cubic count (TODO: Phase 5.3 cubic system)

	// Party
	w.WriteByte(0) // Find party members (0=no, 1=yes)

	// Abnormal Effect (buffs/debuffs visual)
	w.WriteInt(0) // Abnormal effect bitmask (TODO: Phase 5.2 effect system)

	// Recommendations
	w.WriteByte(0) // Recommendations left
	w.WriteShort(0) // Recommendations received

	// Inventory
	w.WriteInt(0) // Inventory limit (TODO: Phase 4.7 inventory system)

	// Class ID (for title/transformation display)
	w.WriteInt(p.Player.ClassID())

	// Special Effects
	w.WriteInt(0) // Special effects bitmask
	w.WriteInt(0) // Max CP

	// Equipped items (17 paperdoll slots)
	// TODO Phase 4.7: load from items table
	// For now, send all zeros (no equipment)
	for range 17 {
		w.WriteInt(0) // Item ID (0 = empty slot)
	}

	// Hero
	w.WriteByte(0) // Hero aura (0=no, 1=yes)

	// Fish
	w.WriteByte(0) // Fishing (0=no, 1=yes)

	// Fish location
	w.WriteInt(0) // Fish X
	w.WriteInt(0) // Fish Y
	w.WriteInt(0) // Fish Z

	// Name color
	w.WriteInt(0xFFFFFF) // Name color (white by default)

	// Pledge Class (clan rank)
	w.WriteInt(0) // Pledge class

	// Title color
	w.WriteInt(0xFFFF77) // Title color (yellow by default)

	// Cursed Weapon
	w.WriteInt(0) // Cursed weapon equipped ID (0=none)

	return w.Bytes(), nil
}
