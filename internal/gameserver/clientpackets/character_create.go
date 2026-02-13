package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharacterCreate is the opcode for CharacterCreate (C2S 0x0B).
// Java reference: ClientPackets.CHARACTER_CREATE(0x0B).
const OpcodeCharacterCreate = 0x0B

// CharacterCreate represents the client packet for creating a new character.
//
// Packet structure (body after opcode):
//   - name       (string)
//   - race       (int32) — read but ignored (derived from classId)
//   - sex        (int32) — 0=male, 1=female
//   - classId    (int32) — base class ID
//   - INT        (int32) — ignored (from template)
//   - STR        (int32) — ignored
//   - CON        (int32) — ignored
//   - MEN        (int32) — ignored
//   - DEX        (int32) — ignored
//   - WIT        (int32) — ignored
//   - hairStyle  (int32)
//   - hairColor  (int32)
//   - face       (int32)
type CharacterCreate struct {
	Name      string
	Race      int32
	IsFemale  bool
	ClassID   int32
	HairStyle int32
	HairColor int32
	Face      int32
}

// ParseCharacterCreate parses CharacterCreate packet.
func ParseCharacterCreate(data []byte) (*CharacterCreate, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	race, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading race: %w", err)
	}

	sex, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading sex: %w", err)
	}

	classID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading classId: %w", err)
	}

	// Skip 6 stat values (INT, STR, CON, MEN, DEX, WIT) — server ignores them
	for i := range 6 {
		if _, err := r.ReadInt(); err != nil {
			return nil, fmt.Errorf("reading stat[%d]: %w", i, err)
		}
	}

	hairStyle, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading hairStyle: %w", err)
	}

	hairColor, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading hairColor: %w", err)
	}

	face, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading face: %w", err)
	}

	return &CharacterCreate{
		Name:      name,
		Race:      race,
		IsFemale:  sex != 0,
		ClassID:   classID,
		HairStyle: hairStyle,
		HairColor: hairColor,
		Face:      face,
	}, nil
}
