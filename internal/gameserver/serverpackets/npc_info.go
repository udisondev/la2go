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
	// Aggressive NPCs (aggroRange > 0) are attackable; peaceful NPCs (merchants, etc.) are not.
	var attackable int32
	if p.Npc.Template().AggroRange() > 0 {
		attackable = 1
	}
	w.WriteInt(attackable)

	// Position
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)
	w.WriteInt(int32(loc.Heading))

	// Stats — Java: AbstractNpcInfo.writeImpl()
	w.WriteInt(0)                    // unnamed zero (Java: writeInt(0))
	w.WriteInt(p.Npc.AtkSpeed())     // mAtkSpd (Java: writeInt(mAtkSpd)) — use pAtkSpd as approx
	w.WriteInt(p.Npc.AtkSpeed())     // pAtkSpd (Java: writeInt(pAtkSpd))
	w.WriteInt(p.Npc.MoveSpeed())    // runSpeed
	w.WriteInt(p.Npc.MoveSpeed() / 2) // walkSpeed (half of run speed)

	// Swimming/Flying speeds
	w.WriteInt(p.Npc.MoveSpeed())    // swimRunSpd
	w.WriteInt(p.Npc.MoveSpeed() / 2) // swimWalkSpd
	w.WriteInt(0)                    // flyRunSpd (not used for ground NPCs)
	w.WriteInt(0)                    // flyWalkSpd
	w.WriteInt(0)                    // flyRunSpd (duplicate)
	w.WriteInt(0)                    // flyWalkSpd (duplicate)

	// Movement multiplier (1.0 = normal speed)
	w.WriteDouble(1.0)

	// Attack speed multiplier (1.0 = normal)
	w.WriteDouble(1.0)

	// Collision radius and height (defaults; NpcTemplate collision fields deferred)
	w.WriteDouble(8.0)  // collision radius
	w.WriteDouble(23.0) // collision height

	// Right hand weapon
	w.WriteInt(0) // right hand item (0=none)

	// Chest (for armor display)
	w.WriteInt(0) // chest item (0=none)

	// Left hand weapon/shield
	w.WriteInt(0) // left hand item (0=none)

	// nameAboveChar (Java: writeByte(1))
	w.WriteByte(1)

	// Status bytes (Java: isRunning, isInCombat, isAlikeDead, isSummoned)
	w.WriteByte(1) // isRunning (1 = running)
	w.WriteByte(0) // isInCombat
	w.WriteByte(0) // isAlikeDead
	w.WriteByte(0) // isSummoned

	// Name
	w.WriteString(p.Npc.Name())

	// Title (Java does NOT write a "title exists" byte — writes string directly)
	w.WriteString(p.Npc.Title())

	// titleColor (Java: writeInt(0))
	w.WriteInt(0)

	// PvP/Karma
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
