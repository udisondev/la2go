package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodePlaySound is the opcode for PlaySound packet (S2C 0x98).
	OpcodePlaySound = 0x98
)

// Sound type constants for PlaySound packet.
const (
	SoundTypeNormal = 0 // UI sound
	SoundType3D     = 1 // Positional 3D sound
)

// PlaySound packet (S2C 0x98) plays a sound effect on the client.
//
// Packet structure:
//   - opcode (byte) — 0x98
//   - soundType (int32) — 0=normal, 1=3D positional
//   - soundFile (string) — sound file name (e.g. "ItemSound.quest_accept")
//   - creatureID (int32) — source object ID (0 for UI sounds)
//   - x, y, z (int32) — position for 3D sounds
//
// Phase 16: Quest System Framework.
type PlaySound struct {
	SoundType  int32
	SoundFile  string
	CreatureID int32
	X, Y, Z    int32
}

// NewPlaySound creates a UI PlaySound packet (non-positional).
func NewPlaySound(soundFile string) PlaySound {
	return PlaySound{
		SoundType: SoundTypeNormal,
		SoundFile: soundFile,
	}
}

// NewPlaySound3D creates a positional PlaySound packet.
func NewPlaySound3D(soundFile string, creatureID int32, x, y, z int32) PlaySound {
	return PlaySound{
		SoundType:  SoundType3D,
		SoundFile:  soundFile,
		CreatureID: creatureID,
		X:          x,
		Y:          y,
		Z:          z,
	}
}

// Write serializes PlaySound packet to binary format.
func (p PlaySound) Write() ([]byte, error) {
	// 1 opcode + 4 soundType + string + 4 creatureID + 3*4 xyz
	w := packet.NewWriter(1 + 4 + len(p.SoundFile)*2 + 2 + 4 + 12)

	w.WriteByte(OpcodePlaySound)
	w.WriteInt(p.SoundType)
	w.WriteString(p.SoundFile)
	w.WriteInt(p.CreatureID)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)

	return w.Bytes(), nil
}

// Common quest sound constants.
const (
	SoundQuestAccept    = "ItemSound.quest_accept"
	SoundQuestMiddle    = "ItemSound.quest_middle"
	SoundQuestFinish    = "ItemSound.quest_finish"
	SoundQuestGiveUp    = "ItemSound.quest_giveup"
	SoundQuestItemGet   = "ItemSound.quest_itemget"
	SoundQuestTutorial  = "ItemSound.quest_tutorial"
	SoundQuestFanfareM  = "ItemSound.quest_fanfare_middle"
	SoundQuestFanfare1  = "ItemSound.quest_fanfare_1"
	SoundQuestFanfare2  = "ItemSound.quest_fanfare_2"
)
