package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodePetInfo is the opcode for PetInfo packet (S2C 0xB1).
	OpcodePetInfo = 0xB1
)

// PetInfo sends full information about a pet/summon to the client.
// Phase 19: Pets/Summons System.
// Java reference: PetInfo.java
type PetInfo struct {
	Summon    *model.Summon
	OwnerName string

	// Pet-specific fields (set for pets, zero for servitors)
	CurrentFed int32
	MaxFed     int32
	Exp        int64
	ExpMax     int64
}

// NewPetInfo creates a PetInfo packet for a summon.
func NewPetInfo(summon *model.Summon, ownerName string) PetInfo {
	return PetInfo{
		Summon:    summon,
		OwnerName: ownerName,
	}
}

// NewPetInfoWithFeed creates PetInfo with pet-specific feed/exp data.
func NewPetInfoWithFeed(summon *model.Summon, ownerName string, currentFed, maxFed int32, exp, expMax int64) PetInfo {
	return PetInfo{
		Summon:     summon,
		OwnerName:  ownerName,
		CurrentFed: currentFed,
		MaxFed:     maxFed,
		Exp:        exp,
		ExpMax:     expMax,
	}
}

// Write serializes PetInfo packet to binary format.
func (p *PetInfo) Write() ([]byte, error) {
	w := packet.NewWriter(512)

	loc := p.Summon.Location()

	w.WriteByte(OpcodePetInfo)

	// Summon type (2=pet, 1=servitor)
	w.WriteInt(int32(p.Summon.Type()))

	// ObjectID
	w.WriteInt(int32(p.Summon.ObjectID()))

	// Template ID + offset (same as NpcInfo)
	w.WriteInt(p.Summon.TemplateID() + 1000000)

	// isAutoAttackable (0 for summons)
	w.WriteInt(0)

	// Position
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)
	w.WriteInt(int32(loc.Heading))

	// padding
	w.WriteInt(0)

	// Attack speed
	w.WriteInt(0)                        // MAtkSpd
	w.WriteInt(p.Summon.AtkSpeed())      // PAtkSpd
	w.WriteInt(p.Summon.MoveSpeed())     // Run speed
	w.WriteInt(p.Summon.MoveSpeed() / 2) // Walk speed

	// Swim/fly speeds (use run speed defaults)
	w.WriteInt(p.Summon.MoveSpeed())     // Swim run speed
	w.WriteInt(p.Summon.MoveSpeed() / 2) // Swim walk speed
	w.WriteInt(0)                        // Fly run speed
	w.WriteInt(0)                        // Fly walk speed
	w.WriteInt(0)                        // Fly run speed 2
	w.WriteInt(0)                        // Fly walk speed 2

	// Movement multiplier (1.0)
	w.WriteDouble(1.0)
	// Attack speed multiplier (1.0)
	w.WriteDouble(1.0)

	// Collision radius and height (NPC default values)
	w.WriteDouble(12.0) // collisionRadius
	w.WriteDouble(22.0) // collisionHeight

	// Weapon/Armor (0 for pets)
	w.WriteInt(0) // right hand weapon
	w.WriteInt(0) // body armor
	w.WriteInt(0) // left hand weapon

	// Owner objectID
	w.WriteInt(int32(p.Summon.OwnerID()))

	// Booleans
	w.WriteByte(0) // isAutoAttackable
	w.WriteByte(0) // isAttackingNow
	w.WriteByte(0) // isAlikeDead
	w.WriteByte(1) // showName

	// Name and title
	w.WriteString(p.Summon.Name())
	w.WriteString(p.OwnerName) // title = owner name

	// showSpawnAnimation (1 for new, 0 for existing)
	w.WriteInt(1)

	// pvpFlag
	w.WriteInt(0)

	// karma
	w.WriteInt(0)

	// Current feed / max feed
	w.WriteInt(p.CurrentFed)
	w.WriteInt(p.MaxFed)

	// HP/MP
	w.WriteInt(p.Summon.CurrentHP())
	w.WriteInt(p.Summon.MaxHP())
	w.WriteInt(p.Summon.CurrentMP())
	w.WriteInt(p.Summon.MaxMP())

	// Level + exp
	w.WriteInt(p.Summon.Level())
	w.WriteLong(p.Exp)    // current exp
	w.WriteLong(p.ExpMax) // exp for current level
	w.WriteLong(0)        // exp for next level

	// Weight (unused for basic summons)
	w.WriteInt(0) // current weight
	w.WriteInt(0) // max weight

	// Combat stats
	w.WriteInt(p.Summon.PAtk())
	w.WriteInt(p.Summon.PDef())
	w.WriteInt(p.Summon.MAtk())
	w.WriteInt(p.Summon.MDef())
	w.WriteInt(0) // accuracy
	w.WriteInt(0) // evasion
	w.WriteInt(0) // critical
	w.WriteInt(p.Summon.MoveSpeed()) // speed

	// PAtk speed / cast speed
	w.WriteInt(p.Summon.AtkSpeed())
	w.WriteInt(333) // MAtkSpd (default cast speed)

	// Abnormal visual effects
	w.WriteInt(0)

	// Mounted (0)
	w.WriteShort(0)

	// Swim flag (0)
	w.WriteByte(0)

	// Team color (0)
	w.WriteByte(0)

	// Soul shots / spirit shots
	w.WriteInt(0) // soulShotsPerHit
	w.WriteInt(0) // spiritShotsPerHit

	// Form
	w.WriteInt(0) // formID

	// Owner controlled flag
	w.WriteInt(0)

	// Secondary maxHP/maxMP (unused)
	w.WriteInt(0)
	w.WriteInt(0)

	// Abnormal effects count
	w.WriteInt(0)

	return w.Bytes(), nil
}
