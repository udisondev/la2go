package data

// TeleportGroup — exported view of a teleport group for use outside the data package.
// Phase 12: Teleporter System.
type TeleportGroup struct {
	Type      string             // "NORMAL", "NOBLES_TOKEN", "NOBLES_ADENA", "OTHER"
	Locations []TeleportLocation // Список точек телепортации
}

// TeleportLocation — exported view of a teleport location for use outside the data package.
// Phase 12: Teleporter System.
type TeleportLocation struct {
	Index    int    // Порядковый индекс в группе (0-based)
	Name     string // Отображаемое название ("Town of Gludio")
	X, Y, Z  int32
	FeeID    int32 // ItemID для оплаты (0 = Adena по умолчанию)
	FeeCount int32 // Стоимость
	CastleID int32 // ID замка для проверки осады (0 = нет)
}

// GetTeleportGroups returns all teleport groups for a given NPC.
// Returns nil if NPC has no teleport data.
// Phase 12: Teleporter System.
func GetTeleportGroups(npcID int32) []TeleportGroup {
	def := TeleporterTable[npcID]
	if def == nil {
		return nil
	}

	groups := make([]TeleportGroup, len(def.teleports))
	for i, g := range def.teleports {
		locs := make([]TeleportLocation, len(g.locations))
		for j, loc := range g.locations {
			locs[j] = TeleportLocation{
				Index:    j,
				Name:     loc.name,
				X:        loc.x,
				Y:        loc.y,
				Z:        loc.z,
				FeeID:    loc.feeId,
				FeeCount: loc.feeCount,
				CastleID: loc.castleId,
			}
		}
		groups[i] = TeleportGroup{
			Type:      g.teleType,
			Locations: locs,
		}
	}
	return groups
}

// GetTeleportGroupByType returns a specific teleport group for NPC by type.
// Returns nil if not found.
// Phase 12: Teleporter System.
func GetTeleportGroupByType(npcID int32, teleType string) *TeleportGroup {
	def := TeleporterTable[npcID]
	if def == nil {
		return nil
	}

	for _, g := range def.teleports {
		if g.teleType == teleType {
			locs := make([]TeleportLocation, len(g.locations))
			for j, loc := range g.locations {
				locs[j] = TeleportLocation{
					Index:    j,
					Name:     loc.name,
					X:        loc.x,
					Y:        loc.y,
					Z:        loc.z,
					FeeID:    loc.feeId,
					FeeCount: loc.feeCount,
					CastleID: loc.castleId,
				}
			}
			return &TeleportGroup{
				Type:      g.teleType,
				Locations: locs,
			}
		}
	}
	return nil
}

// GetTeleportLocation returns a specific location from a group by NPC ID, type, and index.
// Returns nil if not found.
// Phase 12: Teleporter System.
func GetTeleportLocation(npcID int32, teleType string, index int) *TeleportLocation {
	def := TeleporterTable[npcID]
	if def == nil {
		return nil
	}

	for _, g := range def.teleports {
		if g.teleType != teleType {
			continue
		}
		if index < 0 || index >= len(g.locations) {
			return nil
		}
		loc := g.locations[index]
		return &TeleportLocation{
			Index:    index,
			Name:     loc.name,
			X:        loc.x,
			Y:        loc.y,
			Z:        loc.z,
			FeeID:    loc.feeId,
			FeeCount: loc.feeCount,
			CastleID: loc.castleId,
		}
	}
	return nil
}

// HasTeleporter returns true if NPC has teleport data.
// Phase 12: Teleporter System.
func HasTeleporter(npcID int32) bool {
	return TeleporterTable[npcID] != nil
}
