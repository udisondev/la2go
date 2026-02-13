// Package zone implements Lineage 2 world zones with geometric bounds,
// spatial indexing and zone-specific behavior (peace, pvp, damage, etc.).
package zone

import "github.com/udisondev/la2go/internal/model"

// Zone type string constants matching data.zoneDef.zoneType values.
const (
	TypePeace                    = "PeaceZone"
	TypeTown                     = "TownZone"
	TypeCastle                   = "CastleZone"
	TypeEffect                   = "EffectZone"
	TypeWater                    = "WaterZone"
	TypeDamage                   = "DamageZone"
	TypeSiege                    = "SiegeZone"
	TypeArena                    = "ArenaZone"
	TypeFishing                  = "FishingZone"
	TypeClanHall                 = "ClanHallZone"
	TypeNoStore                  = "NoStoreZone"
	TypeNoLanding                = "NoLandingZone"
	TypeNoSummonFriend           = "NoSummonFriendZone"
	TypeNoRestart                = "NoRestartZone"
	TypeJail                     = "JailZone"
	TypeMotherTree               = "MotherTreeZone"
	TypeSwamp                    = "SwampZone"
	TypeNoPvP                    = "NoPvPZone"
	TypeOlympiadStadium          = "OlympiadStadiumZone"
	TypeHQ                       = "HqZone"
	TypeRespawn                  = "RespawnZone"
	TypeScript                   = "ScriptZone"
	TypeBoss                     = "BossZone"
	TypeCondition                = "ConditionZone"
	TypeDerbyTrack               = "DerbyTrackZone"
	TypeSiegableHall             = "SiegableHallZone"
	TypeResidenceTeleport        = "ResidenceTeleportZone"
	TypeResidenceHallTeleport    = "ResidenceHallTeleportZone"
)

// Zone represents a game world zone with geometric bounds, character tracking,
// and entry/exit callbacks.
// Java reference: model/zone/ZoneType.java
type Zone interface {
	ID() int32
	Name() string
	ZoneType() string
	Contains(x, y, z int32) bool
	IsPeace() bool
	AllowsPvP() bool

	// Character tracking (Phase 11)
	// RevalidateInZone checks if creature is inside/outside and fires onEnter/onExit.
	RevalidateInZone(creature *model.Character)
	// RemoveCharacter forcibly removes creature from zone and fires onExit.
	RemoveCharacter(creature *model.Character)
	// GetCharactersInside returns all characters currently tracked in this zone.
	GetCharactersInside() []*model.Character
	// GetPlayersInside returns only Player characters tracked in this zone.
	GetPlayersInside() []*model.Player
}
