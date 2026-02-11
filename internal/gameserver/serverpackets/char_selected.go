package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeCharSelected is the opcode for CharSelected packet (S2C 0x15)
	OpcodeCharSelected = 0x15
)

// CharSelected confirms character selection after CharacterSelect packet.
// Sent from GameServer to client when user selects a character.
//
// Packet structure (266+ bytes):
//   - opcode: byte (0x15)
//   - name: String (UTF-16LE, variable length)
//   - objectID: int32
//   - title: String (UTF-16LE, variable length)
//   - sessionID: int32 (playOkID1 from SessionKey)
//   - clanID: int32
//   - unknown1: int32 (always 0)
//   - isFemale: int32 (0=male, 1=female)
//   - race: int32 (0=Human, 1=Elf, 2=DarkElf, 3=Orc, 4=Dwarf)
//   - classID: int32
//   - active: int32 (always 1)
//   - x, y, z: int32 (world coordinates)
//   - currentHP, currentMP: float64 (IEEE 754, LE)
//   - sp: int32 (skill points, cast from int64)
//   - exp: int64
//   - level: int32
//   - karma: int32
//   - pkKills: int32
//   - INT, STR, CON, MEN, DEX, WIT: int32 (base stats)
//   - reserved[30]: int32 (30 zero-filled fields)
//   - unknown2, unknown3: int32 (always 0)
//   - gameTime: int32 (minutes since midnight, 0..1439)
//   - unknown4: int32 (always 0)
//   - classIDDup: int32 (duplicate of classID)
//   - padding[12]: int32 (12 zero-filled fields)
//
// Reference: L2J_Mobius CharSelected.java:42-92
type CharSelected struct {
	Name      string
	ObjectID  int32
	Title     string
	SessionID int32 // playOkID1 from SessionKey
	ClanID    int32
	IsFemale  int32 // 0=male, 1=female (boolean as int32)
	Race      int32 // 0=Human, 1=Elf, 2=DarkElf, 3=Orc, 4=Dwarf
	ClassID   int32
	X         int32
	Y         int32
	Z         int32
	CurrentHP float64
	CurrentMP float64
	SP        int32 // Skill Points (cast from int64)
	Exp       int64
	Level     int32
	Karma     int32
	PkKills   int32
	INT       int32
	STR       int32
	CON       int32
	MEN       int32
	DEX       int32
	WIT       int32
	GameTime  int32 // Minutes since midnight (0..1439)
}

// NewCharSelected creates a CharSelected packet from Player.
// sessionID is playOkID1 from SessionKey (first int32 of 4-element key).
//
// TODO Phase 5.x: Many fields use temporary placeholders (isFemale, title, SP, karma, pkKills, base stats).
// Full character system will be implemented in Phase 5 (Character Templates, Stats, Appearance).
func NewCharSelected(player *model.Player, sessionID int32) *CharSelected {
	loc := player.Location()

	// TODO Phase 5.x: Add Player.IsFemale() method
	// For now, hardcode to male (0)
	isFemale := int32(0)

	// TODO Phase 5.x: Add Player.Title() method
	// For now, use empty title
	title := ""

	// TODO Phase 5.x: Implement game time manager
	// For now, use placeholder: 12:00 (720 minutes since midnight)
	gameTime := int32(720)

	// TODO Phase 5.x: Add SP, Karma, PkKills, base stats fields to Player
	// For now, use placeholders
	sp := int32(player.SP())
	karma := int32(0)
	pkKills := int32(0)

	// Default base stats (10 each for MVP)
	baseStats := int32(10)

	return &CharSelected{
		Name:      player.Name(),
		ObjectID:  int32(player.ObjectID()),
		Title:     title,
		SessionID: sessionID,
		ClanID:    0, // TODO: Implement clan system (Phase 5.x)
		IsFemale:  isFemale,
		Race:      player.RaceID(),
		ClassID:   player.ClassID(),
		X:         loc.X,
		Y:         loc.Y,
		Z:         loc.Z,
		CurrentHP: float64(player.CurrentHP()), // Cast int32 to float64
		CurrentMP: float64(player.CurrentMP()), // Cast int32 to float64
		SP:        sp,
		Exp:       player.Experience(), // Use existing Experience() method
		Level:     player.Level(),
		Karma:     karma,
		PkKills:   pkKills,
		INT:       baseStats,
		STR:       baseStats,
		CON:       baseStats,
		MEN:       baseStats,
		DEX:       baseStats,
		WIT:       baseStats,
		GameTime:  gameTime,
	}
}

// Write serializes the CharSelected packet to bytes.
func (p *CharSelected) Write() ([]byte, error) {
	w := packet.NewWriter(512) // Estimate: 266+ bytes

	// Opcode
	if err := w.WriteByte(OpcodeCharSelected); err != nil {
		return nil, err
	}

	// Character info
	w.WriteString(p.Name)
	w.WriteInt(p.ObjectID)
	w.WriteString(p.Title)
	w.WriteInt(p.SessionID)
	w.WriteInt(p.ClanID)
	w.WriteInt(0) // unknown1

	// Appearance
	w.WriteInt(p.IsFemale)
	w.WriteInt(p.Race)
	w.WriteInt(p.ClassID)
	w.WriteInt(1) // active (always 1)

	// Position
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)

	// Status
	w.WriteDouble(p.CurrentHP)
	w.WriteDouble(p.CurrentMP)
	w.WriteInt(p.SP)
	w.WriteLong(p.Exp)
	w.WriteInt(p.Level)
	w.WriteInt(p.Karma)
	w.WriteInt(p.PkKills)

	// Base stats
	w.WriteInt(p.INT)
	w.WriteInt(p.STR)
	w.WriteInt(p.CON)
	w.WriteInt(p.MEN)
	w.WriteInt(p.DEX)
	w.WriteInt(p.WIT)

	// Reserved fields (30 × int32 = 120 bytes)
	for range 30 {
		w.WriteInt(0)
	}

	// Additional fields
	w.WriteInt(0)         // unknown2
	w.WriteInt(0)         // unknown3
	w.WriteInt(p.GameTime) // minutes since midnight
	w.WriteInt(0)         // unknown4
	w.WriteInt(p.ClassID) // classID duplicate

	// Padding (12 × int32 = 48 bytes)
	for range 12 {
		w.WriteInt(0)
	}

	return w.Bytes(), nil
}
