package model

// ZoneID represents a zone flag type.
// Each creature maintains a bitfield of active zone flags via atomic.Uint32.
// Java reference: model/zone/ZoneId.java (22 enum values).
type ZoneID uint8

const (
	ZoneIDPVP            ZoneID = iota // 0: PvP zone (arena, siege during combat)
	ZoneIDPeace                        // 1: Peace zone (towns, safe areas)
	ZoneIDSiege                        // 2: Siege zone (castle siege area)
	ZoneIDMotherTree                   // 3: Mother Tree zone (elf village regen)
	ZoneIDClanHall                     // 4: Clan Hall zone
	ZoneIDNoLanding                    // 5: No wyvern landing zone
	ZoneIDWater                        // 6: Water zone (swimming)
	ZoneIDJail                         // 7: Jail zone
	ZoneIDMonsterTrack                 // 8: Monster track zone
	ZoneIDCastle                       // 9: Castle zone
	ZoneIDSwamp                        // 10: Swamp zone (slow movement)
	ZoneIDNoSummonFriend               // 11: No summon friend zone
	ZoneIDNoStore                      // 12: No private store zone
	ZoneIDNoPVP                        // 13: No PvP zone (explicit no-PvP)
	ZoneIDTown                         // 14: Town zone
	ZoneIDScript                       // 15: Script zone
	ZoneIDHQ                           // 16: Headquarters zone (siege)
	ZoneIDDangerArea                   // 17: Danger area zone
	ZoneIDAltered                      // 18: Altered zone
	ZoneIDNoBookmark                   // 19: No bookmark zone
	ZoneIDNoItemDrop                   // 20: No item drop zone
	ZoneIDNoRestart                    // 21: No restart zone

	ZoneIDCount // 22 — total zone flag count
)
