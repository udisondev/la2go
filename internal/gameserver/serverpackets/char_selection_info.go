package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeCharSelectionInfo is the opcode for CharSelectionInfo packet (S2C 0x13)
	OpcodeCharSelectionInfo = 0x13
)

// CharacterInfo holds display information for one character in the selection screen.
// Contains basic stats, appearance, and equipment visible in character list.
type CharacterInfo struct {
	// Basic identity
	Name      string
	ObjectID  int32
	ClanID    int32
	Sex       int32 // 0=male, 1=female
	Race      int32 // 0=human, 1=elf, 2=dark elf, 3=orc, 4=dwarf
	ClassID   int32
	BaseClass int32

	// Stats
	Level int32
	Exp   int64
	SP    int32
	Karma int32

	// Vitals
	CurrentHP float64
	CurrentMP float64
	MaxHP     float64
	MaxMP     float64

	// Appearance
	Face      int32
	HairStyle int32
	HairColor int32

	// Equipment (paperdoll slots)
	// Each slot has ObjectID (instance) and ItemID (template)
	PaperdollObjectIDs [17]int32 // Equipped item instance IDs
	PaperdollItemIDs   [17]int32 // Equipped item template IDs

	// Status
	DeleteTimer    int32 // Seconds until deletion (0=normal, -1=banned, >0=pending deletion)
	EnchantEffect  byte  // Visual enchant glow (max 127)
	AugmentationID int32 // Augmentation visual effect

	// Metadata
	LastAccess int64 // Timestamp for determining default selected character
}

// CharSelectionInfo packet (S2C 0x13) sends list of characters for account.
// Sent after successful AuthLogin, allows player to select character to enter world.
type CharSelectionInfo struct {
	LoginName  string
	SessionID  int32
	Characters []CharacterInfo
	ActiveID   int32 // Index of pre-selected character (-1 = auto-select by lastAccess)
}

// NewCharSelectionInfo creates CharSelectionInfo packet from list of characters.
// If activeID is -1, automatically selects character with most recent access.
func NewCharSelectionInfo(loginName string, sessionID int32, characters []CharacterInfo) *CharSelectionInfo {
	return &CharSelectionInfo{
		LoginName:  loginName,
		SessionID:  sessionID,
		Characters: characters,
		ActiveID:   -1, // Auto-select by last access
	}
}

// NewCharSelectionInfoFromPlayers creates CharSelectionInfo from model.Player slice.
// Converts Player models to CharacterInfo display format.
// NOTE: Uses default values for fields not yet in database schema (sex, face, hair, karma, SP).
// TODO Phase 4.7: Add migration for missing character appearance/stats fields.
func NewCharSelectionInfoFromPlayers(loginName string, sessionID int32, players []*model.Player) *CharSelectionInfo {
	charInfos := make([]CharacterInfo, len(players))
	for i, player := range players {
		// Get timestamp for last access (use lastLogin if available, otherwise createdAt)
		lastAccess := player.CreatedAt().Unix()
		if !player.LastLogin().IsZero() {
			lastAccess = player.LastLogin().Unix()
		}

		charInfos[i] = CharacterInfo{
			Name:      player.Name(),
			ObjectID:  int32(player.CharacterID()),
			ClanID:    0, // TODO Phase 4.8: load from clan system
			Sex:       0, // TODO Phase 4.7: add to DB schema (0=male default)
			Race:      player.RaceID(),
			ClassID:   player.ClassID(),
			BaseClass: player.ClassID(), // TODO Phase 4.7: add base_class to DB
			Level:     player.Level(),
			Exp:       player.Experience(),
			SP:        int32(player.SP()),
			Karma:     0,                // TODO Phase 4.7: add karma to DB schema
			CurrentHP: float64(player.CurrentHP()),
			CurrentMP: float64(player.CurrentMP()),
			MaxHP:     float64(player.MaxHP()),
			MaxMP:     float64(player.MaxMP()),
			Face:      0, // TODO Phase 4.7: add face to DB schema (default face)
			HairStyle: 0, // TODO Phase 4.7: add hair_style to DB
			HairColor: 0, // TODO Phase 4.7: add hair_color to DB
			// PaperdollObjectIDs and PaperdollItemIDs initialized to zero
			// TODO Phase 4.7: load from items table (paperdoll slots)
			DeleteTimer:    0, // TODO Phase 4.8: character deletion system
			EnchantEffect:  0, // TODO Phase 4.9: calculate from equipment enchants
			AugmentationID: 0, // TODO Phase 5.2: augmentation system
			LastAccess:     lastAccess,
		}
	}

	return NewCharSelectionInfo(loginName, sessionID, charInfos)
}

// Write serializes CharSelectionInfo packet to binary format.
// Returns byte slice containing full packet data (opcode + payload).
func (p *CharSelectionInfo) Write() ([]byte, error) {
	// Calculate required buffer size
	// Base: 1 (opcode) + 4 (character count)
	// Per character: ~300 bytes (estimate for all fields)
	bufSize := 5 + len(p.Characters)*300
	w := packet.NewWriter(bufSize)

	w.WriteByte(OpcodeCharSelectionInfo)
	w.WriteInt(int32(len(p.Characters)))

	// Auto-select character with most recent access if activeID not set
	activeID := p.ActiveID
	if activeID == -1 {
		var lastAccess int64
		for i, char := range p.Characters {
			if char.LastAccess > lastAccess {
				lastAccess = char.LastAccess
				activeID = int32(i)
			}
		}
	}

	// Write each character
	for i, char := range p.Characters {
		w.WriteString(char.Name)                // Character name
		w.WriteInt(char.ObjectID)               // Character ID
		w.WriteString(p.LoginName)              // Account name
		w.WriteInt(p.SessionID)                 // Session ID
		w.WriteInt(char.ClanID)                 // Clan ID
		w.WriteInt(0)                           // Builder level (always 0)
		w.WriteInt(char.Sex)                    // Sex
		w.WriteInt(char.Race)                   // Race
		w.WriteInt(char.BaseClass)              // Base class ID
		w.WriteInt(1)                           // Game server name (always 1)
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteDouble(char.CurrentHP)           // Current HP
		w.WriteDouble(char.CurrentMP)           // Current MP
		w.WriteInt(char.SP)                     // SP
		w.WriteLong(char.Exp)                   // Experience
		w.WriteInt(char.Level)                  // Level
		w.WriteInt(char.Karma)                  // Karma
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown
		w.WriteInt(0)                           // Unknown

		// Paperdoll ObjectIDs (17 slots)
		for _, objID := range char.PaperdollObjectIDs {
			w.WriteInt(objID)
		}

		// Paperdoll ItemIDs (17 slots)
		for _, itemID := range char.PaperdollItemIDs {
			w.WriteInt(itemID)
		}

		w.WriteInt(char.HairStyle)              // Hair style
		w.WriteInt(char.HairColor)              // Hair color
		w.WriteInt(char.Face)                   // Face
		w.WriteDouble(char.MaxHP)               // Max HP
		w.WriteDouble(char.MaxMP)               // Max MP
		w.WriteInt(char.DeleteTimer)            // Delete timer
		w.WriteInt(char.ClassID)                // Class ID
		w.WriteInt(boolToInt(i == int(activeID))) // Is selected
		w.WriteByte(char.EnchantEffect)         // Enchant effect
		w.WriteInt(char.AugmentationID)         // Augmentation ID
	}

	return w.Bytes(), nil
}

// boolToInt converts bool to int (1 for true, 0 for false)
func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
