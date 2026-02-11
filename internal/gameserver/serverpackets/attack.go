package serverpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeAttack is the server packet opcode for attack notification.
// Broadcasts physical attack result to all visible players.
//
// Phase 5.3: Basic Combat System.
const OpcodeAttack = 0x05

// Hit flags (bitmask).
// Used in Attack packet to indicate special conditions (miss, crit, shield block, soulshot).
//
// Phase 5.3: Basic Combat System.
const (
	HitFlagUseSS  = 0x10 // Soulshot used (includes grade for visual effect)
	HitFlagCrit   = 0x20 // Critical hit
	HitFlagShield = 0x40 // Blocked by shield
	HitFlagMiss   = 0x80 // Miss (no damage)
)

// Hit represents a single hit in an Attack packet.
// Dual weapons can produce multiple hits in one attack.
//
// Phase 5.3: Basic Combat System (single hit only, dual weapons TODO Phase 5.4).
type Hit struct {
	TargetID uint32 // Target objectID
	Damage   int32  // Damage dealt (0 if miss)
	Flags    byte   // Bitmask: miss/crit/shield/soulshot flags
}

// Attack represents physical attack packet (S2C 0x05).
// Broadcasts attack result to all visible players (LOD optimization).
//
// Packet structure:
//   - opcode (byte) — 0x05
//   - attackerObjectID (int32)
//   - firstHit.targetID (int32)
//   - firstHit.damage (int32)
//   - firstHit.flags (byte)
//   - attacker.x (int32)
//   - attacker.y (int32)
//   - attacker.z (int32)
//   - additionalHitsCount (int16) — size - 1
//   - for each additional hit:
//     - hit.targetID (int32)
//     - hit.damage (int32)
//     - hit.flags (byte)
//   - target.x (int32) — last hit target location
//   - target.y (int32)
//   - target.z (int32)
//
// Phase 5.3: Basic Combat System.
// Java reference: Attack.java (opcode 0x05, writeImpl line 100-118).
type Attack struct {
	AttackerID  uint32           // Attacker objectID
	AttackerLoc *model.Location  // Attacker coordinates (for animation)
	TargetLoc   *model.Location  // Target coordinates (last hit target)
	UseSoulshot bool             // Soulshot used flag (MVP: always false)
	SSGrade     int32            // Soulshot grade (MVP: 0, unused)
	Hits        []Hit            // Hit list (1+ hits, dual weapons can have 2)
}

// NewAttack creates new Attack packet for physical attack (Player attacker).
// Initializes empty hit list (caller must add hits via AddHit).
//
// Returns value (not pointer) to avoid heap allocation.
//
// Phase 5.3: Basic Combat System (soulshot always false).
func NewAttack(attacker *model.Player, target *model.WorldObject) Attack {
	attackerLoc := attacker.Location()
	targetLoc := target.Location()

	return Attack{
		AttackerID:  attacker.ObjectID(),
		AttackerLoc: &attackerLoc,
		TargetLoc:   &targetLoc,
		UseSoulshot: false,
		SSGrade:     0,
		Hits:        make([]Hit, 0, 1), // Capacity 1 for single weapon (MVP)
	}
}

// NewNpcAttack creates new Attack packet for NPC physical attack.
// Phase 5.7: NPC Aggro & Basic AI.
func NewNpcAttack(attackerID uint32, attackerLoc model.Location, target *model.WorldObject) Attack {
	targetLoc := target.Location()

	return Attack{
		AttackerID:  attackerID,
		AttackerLoc: &attackerLoc,
		TargetLoc:   &targetLoc,
		UseSoulshot: false,
		SSGrade:     0,
		Hits:        make([]Hit, 0, 1),
	}
}

// AddHit adds hit result to Attack packet.
// Must be called at least once before Write().
//
// Parameters:
//   - targetID: target objectID
//   - damage: damage dealt (0 if miss)
//   - miss: true if attack missed
//   - crit: true if critical hit (×2 damage)
//
// Phase 5.3: Basic Combat System (shield block TODO Phase 5.4).
func (a *Attack) AddHit(targetID uint32, damage int32, miss, crit bool) {
	var flags byte

	if miss {
		flags |= HitFlagMiss
	}
	if crit {
		flags |= HitFlagCrit
	}

	a.Hits = append(a.Hits, Hit{
		TargetID: targetID,
		Damage:   damage,
		Flags:    flags,
	})
}

// Write serializes Attack packet to bytes.
// Returns error if no hits added (at least one hit required).
//
// Packet size: 40 + N*9 bytes (N = number of hits).
//
// Phase 5.3: Basic Combat System.
func (a *Attack) Write() ([]byte, error) {
	// Validate: at least one hit required
	if len(a.Hits) == 0 {
		return nil, fmt.Errorf("Attack packet must have at least one hit")
	}

	// Estimate size: 1 + 4 + (4+4+1) + (4+4+4) + 2 + N*(4+4+1) + (4+4+4)
	// = 1 + 4 + 9 + 12 + 2 + N*9 + 12 = 40 + N*9
	w := packet.NewWriter(40 + len(a.Hits)*9)

	w.WriteByte(OpcodeAttack)
	w.WriteInt(int32(a.AttackerID))

	// First hit (mandatory)
	firstHit := a.Hits[0]
	w.WriteInt(int32(firstHit.TargetID))
	w.WriteInt(firstHit.Damage)
	w.WriteByte(firstHit.Flags)

	// Attacker location (for animation origin)
	w.WriteInt(a.AttackerLoc.X)
	w.WriteInt(a.AttackerLoc.Y)
	w.WriteInt(a.AttackerLoc.Z)

	// Additional hits count (size - 1)
	// Dual weapons can produce 2 hits (TODO Phase 5.4)
	w.WriteShort(int16(len(a.Hits) - 1))

	// Additional hits (if any)
	for i := 1; i < len(a.Hits); i++ {
		hit := a.Hits[i]
		w.WriteInt(int32(hit.TargetID))
		w.WriteInt(hit.Damage)
		w.WriteByte(hit.Flags)
	}

	// Target location (last hit target, for animation destination)
	w.WriteInt(a.TargetLoc.X)
	w.WriteInt(a.TargetLoc.Y)
	w.WriteInt(a.TargetLoc.Z)

	return w.Bytes(), nil
}
