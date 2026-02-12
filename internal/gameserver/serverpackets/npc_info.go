package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeNpcInfo is the opcode for NpcInfo packet (S2C 0x16)
	OpcodeNpcInfo = 0x16
)

// NpcInfo packet (S2C 0x16) sends information about an NPC.
// Sent when NPC enters visibility range of player.
type NpcInfo struct {
	Npc *model.Npc
}

// NewNpcInfo creates NpcInfo packet from Npc model.
func NewNpcInfo(npc *model.Npc) NpcInfo {
	return NpcInfo{
		Npc: npc,
	}
}

// Write serializes NpcInfo packet to binary format.
// NpcInfo is simpler than CharInfo (no equipment, no clan, etc).
func (p *NpcInfo) Write() ([]byte, error) {
	w := packet.NewWriter(256)

	loc := p.Npc.Location()

	w.WriteByte(OpcodeNpcInfo)

	// Object ID
	w.WriteInt(int32(p.Npc.ObjectID()))

	// Template ID + ID offset (for client-side caching)
	// idOffset = templateID + 1000000 (L2J formula)
	w.WriteInt(p.Npc.TemplateID() + 1000000)

	// Attackable (1=attackable, 0=peaceful)
	// TODO Phase 5.1: read from NpcTemplate
	w.WriteInt(1) // default attackable

	// Position
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)
	w.WriteInt(int32(loc.Heading))

	// Stats
	w.WriteInt(0)                    // MAtkSpd (magical attack speed)
	w.WriteInt(p.Npc.AtkSpeed())     // PAtkSpd (physical attack speed)
	w.WriteInt(p.Npc.MoveSpeed())    // Run speed
	w.WriteInt(p.Npc.MoveSpeed() / 2) // Walk speed (half of run speed)

	// Float movement speed (same as run/walk but as float)
	w.WriteInt(p.Npc.MoveSpeed())    // fRunSpd
	w.WriteInt(p.Npc.MoveSpeed() / 2) // fWalkSpd
	w.WriteInt(p.Npc.MoveSpeed())    // fSwimRunSpd
	w.WriteInt(p.Npc.MoveSpeed() / 2) // fSwimWalkSpd
	w.WriteInt(0)                    // fFlyRunSpd (not used for ground NPCs)
	w.WriteInt(0)                    // fFlyWalkSpd

	// Movement multiplier (1.0 = normal speed)
	w.WriteDouble(1.0)

	// Attack speed multiplier (1.0 = normal)
	w.WriteDouble(1.0)

	// Collision radius and height
	// TODO Phase 5.1: read from NpcTemplate
	w.WriteDouble(8.0)  // collision radius
	w.WriteDouble(23.0) // collision height

	// Right hand weapon
	w.WriteInt(0) // right hand item (0=none)

	// Chest (for armor display)
	w.WriteInt(0) // chest item (0=none)

	// Left hand weapon/shield
	w.WriteInt(0) // left hand item (0=none)

	// Name
	w.WriteByte(1) // name exists
	w.WriteString(p.Npc.Name())

	// Title
	w.WriteByte(1) // title exists
	w.WriteString(p.Npc.Title())

	// Status
	w.WriteInt(0) // PvP flag (0=none)
	w.WriteInt(0) // Karma

	// Abnormal Effect (buffs/debuffs bitmask)
	w.WriteInt(0)

	// Clan ID (NPCs don't have clans)
	w.WriteInt(0)

	// Clan crest ID
	w.WriteInt(0)

	// Ally ID
	w.WriteInt(0)

	// Ally crest ID
	w.WriteInt(0)

	// Is Flying (wyverns, etc)
	w.WriteByte(0) // 0=ground, 1=flying

	// Team circle (for events/olympiad)
	w.WriteByte(0) // 0=none, 1=blue, 2=red

	// Collision radius (again, for client physics)
	w.WriteDouble(8.0)

	// Collision height
	w.WriteDouble(23.0)

	// Enchant effect (glowing aura)
	w.WriteInt(0) // 0=none

	// Is Flying (again, for client rendering)
	w.WriteInt(0) // 0=ground

	return w.Bytes(), nil
}
