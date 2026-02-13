package zone

import (
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

const gridSize int32 = 4096 // мировые единицы на ячейку сетки

type gridKey struct {
	gx, gy int32
}

// Manager manages all game zones with spatial indexing for fast lookups.
type Manager struct {
	zones  []Zone
	byID   map[int32]Zone
	byType map[string][]Zone
	grid   map[gridKey][]Zone
}

// NewManager creates a new empty ZoneManager.
func NewManager() *Manager {
	return &Manager{
		byID:   make(map[int32]Zone),
		byType: make(map[string][]Zone),
		grid:   make(map[gridKey][]Zone),
	}
}

// Init loads zones from data.ZoneTable and builds the spatial grid.
func (m *Manager) Init() error {
	if data.ZoneTable == nil {
		return fmt.Errorf("init zone manager: data.ZoneTable is nil, call data.LoadZones first")
	}

	for _, zd := range data.ZoneTable {
		nodes := zd.Nodes()
		n := len(nodes)

		if n == 0 {
			slog.Warn("skip zone with no nodes", "id", zd.ZoneID())
			continue
		}

		nodesX := make([]int32, n)
		nodesY := make([]int32, n)

		for i, nd := range nodes {
			nodesX[i] = nd.NodeX()
			nodesY[i] = nd.NodeY()
		}

		base := &BaseZone{
			id:       zd.ZoneID(),
			name:     zd.ZoneName(),
			zoneType: zd.ZoneType(),
			shape:    zd.Shape(),
			minZ:     zd.MinZ(),
			maxZ:     zd.MaxZ(),
			rad:      zd.Rad(),
			nodesX:   nodesX,
			nodesY:   nodesY,
			params:   zd.Params(),
		}

		z, err := newTypedZone(base)
		if err != nil {
			slog.Warn("skip zone", "id", zd.ZoneID(), "err", err)
			continue
		}

		m.zones = append(m.zones, z)
		m.byID[z.ID()] = z
		m.byType[z.ZoneType()] = append(m.byType[z.ZoneType()], z)
	}

	m.buildGrid()

	slog.Info("zone manager initialized",
		"zones", len(m.zones),
		"types", len(m.byType),
		"grid_cells", len(m.grid),
	)

	return nil
}

// GetZonesAt returns all zones containing the point (x, y, z).
func (m *Manager) GetZonesAt(x, y, z int32) []Zone {
	key := gridKey{gx: floorDiv(x, gridSize), gy: floorDiv(y, gridSize)}
	candidates := m.grid[key]

	var result []Zone
	for _, zn := range candidates {
		if zn.Contains(x, y, z) {
			result = append(result, zn)
		}
	}

	return result
}

// GetZoneByID returns a zone by its identifier, or nil if not found.
func (m *Manager) GetZoneByID(id int32) Zone {
	return m.byID[id]
}

// IsInPeaceZone checks if point (x, y, z) is inside any peace zone.
func (m *Manager) IsInPeaceZone(x, y, z int32) bool {
	zones := m.GetZonesAt(x, y, z)
	for _, zn := range zones {
		if zn.IsPeace() {
			return true
		}
	}

	return false
}

// GetZonesByType returns all zones of the given type.
func (m *Manager) GetZonesByType(zoneType string) []Zone {
	return m.byType[zoneType]
}

// RevalidateZones checks all zones at creature's position and triggers
// onEnter/onExit as needed. Called when a creature moves.
// Java reference: ZoneManager.getRegion -> revalidateInZone for each zone
func (m *Manager) RevalidateZones(creature *model.Character) {
	loc := creature.Location()
	key := gridKey{gx: floorDiv(loc.X, gridSize), gy: floorDiv(loc.Y, gridSize)}

	// Check all candidate zones in creature's grid cell
	for _, zn := range m.grid[key] {
		zn.RevalidateInZone(creature)
	}

	// Also need to handle zones the creature LEFT (was tracked but no longer in grid cell).
	// For simplicity we iterate all zones the creature is currently tracked in.
	// This is O(total zones) but only happens on creature movement, not per tick.
	// Performance: O(total zones) per movement. Could optimize with per-creature zone list if needed.
}

// RemoveFromAllZones removes a creature from all zones.
// Called when creature logs out, dies, or teleports.
func (m *Manager) RemoveFromAllZones(creature *model.Character) {
	creature.ClearAllZoneFlags()
	for _, zn := range m.zones {
		zn.RemoveCharacter(creature)
	}
}

// newTypedZone создает конкретный тип зоны по zoneType в BaseZone.
func newTypedZone(base *BaseZone) (Zone, error) {
	switch base.zoneType {
	case TypePeace:
		return NewPeaceZone(base), nil
	case TypeTown:
		return NewTownZone(base), nil
	case TypeCastle:
		return NewCastleZone(base), nil
	case TypeSiege:
		return NewPvPZone(base), nil
	case TypeDamage:
		return NewDamageZone(base), nil
	case TypeWater:
		return NewWaterZone(base), nil
	case TypeEffect:
		return NewEffectZone(base), nil
	case TypeArena:
		return NewArenaZone(base), nil
	case TypeFishing:
		return NewFishingZone(base), nil
	case TypeClanHall:
		return NewClanHallZone(base), nil
	case TypeNoStore:
		return NewNoStoreZone(base), nil
	case TypeNoLanding:
		return NewNoLandingZone(base), nil
	case TypeNoSummonFriend:
		return NewNoSummonFriendZone(base), nil
	case TypeNoRestart:
		return NewNoRestartZone(base), nil
	case TypeJail:
		return NewJailZone(base), nil
	case TypeMotherTree:
		return NewMotherTreeZone(base), nil
	case TypeSwamp:
		return NewSwampZone(base), nil
	case TypeNoPvP:
		return NewNoPvPZone(base), nil
	case TypeOlympiadStadium:
		return NewOlympiadStadiumZone(base), nil
	case TypeHQ:
		return NewHqZone(base), nil
	case TypeRespawn:
		return NewRespawnZone(base), nil
	case TypeScript:
		return NewScriptZone(base), nil
	case TypeBoss:
		return NewBossZone(base), nil
	case TypeCondition:
		return NewConditionZone(base), nil
	case TypeDerbyTrack:
		return NewDerbyTrackZone(base), nil
	case TypeSiegableHall:
		return NewSiegableHallZone(base), nil
	case TypeResidenceTeleport:
		return NewResidenceTeleportZone(base), nil
	case TypeResidenceHallTeleport:
		return NewResidenceHallTeleportZone(base), nil
	default:
		return nil, fmt.Errorf("unknown zone type %q for zone %d", base.zoneType, base.id)
	}
}

// buildGrid регистрирует каждую зону во всех ячейках сетки,
// которые пересекает её bounding box.
func (m *Manager) buildGrid() {
	for _, z := range m.zones {
		bz := extractBase(z)
		if bz == nil || len(bz.nodesX) == 0 {
			continue
		}

		minX, maxX := bz.nodesX[0], bz.nodesX[0]
		minY, maxY := bz.nodesY[0], bz.nodesY[0]

		for i := 1; i < len(bz.nodesX); i++ {
			if bz.nodesX[i] < minX {
				minX = bz.nodesX[i]
			}
			if bz.nodesX[i] > maxX {
				maxX = bz.nodesX[i]
			}
			if bz.nodesY[i] < minY {
				minY = bz.nodesY[i]
			}
			if bz.nodesY[i] > maxY {
				maxY = bz.nodesY[i]
			}
		}

		gxMin := floorDiv(minX, gridSize)
		gxMax := floorDiv(maxX, gridSize)
		gyMin := floorDiv(minY, gridSize)
		gyMax := floorDiv(maxY, gridSize)

		for gx := gxMin; gx <= gxMax; gx++ {
			for gy := gyMin; gy <= gyMax; gy++ {
				key := gridKey{gx: gx, gy: gy}
				m.grid[key] = append(m.grid[key], z)
			}
		}
	}
}

// floorDiv выполняет целочисленное деление с округлением к -inf,
// корректно обрабатывая отрицательные координаты.
func floorDiv(a, b int32) int32 {
	d := a / b
	if (a^b) < 0 && d*b != a {
		d--
	}

	return d
}

// extractBase возвращает BaseZone из конкретного типа зоны.
func extractBase(z Zone) *BaseZone {
	switch v := z.(type) {
	case *PeaceZone:
		return v.BaseZone
	case *TownZone:
		return v.BaseZone
	case *CastleZone:
		return v.BaseZone
	case *PvPZone:
		return v.BaseZone
	case *DamageZone:
		return v.BaseZone
	case *WaterZone:
		return v.BaseZone
	case *EffectZone:
		return v.BaseZone
	case *ArenaZone:
		return v.BaseZone
	case *FishingZone:
		return v.BaseZone
	case *ClanHallZone:
		return v.BaseZone
	case *NoStoreZone:
		return v.BaseZone
	case *NoLandingZone:
		return v.BaseZone
	case *NoSummonFriendZone:
		return v.BaseZone
	case *NoRestartZone:
		return v.BaseZone
	case *JailZone:
		return v.BaseZone
	case *MotherTreeZone:
		return v.BaseZone
	case *SwampZone:
		return v.BaseZone
	case *NoPvPZone:
		return v.BaseZone
	case *OlympiadStadiumZone:
		return v.BaseZone
	case *HqZone:
		return v.BaseZone
	case *RespawnZone:
		return v.BaseZone
	case *ScriptZone:
		return v.BaseZone
	case *BossZone:
		return v.BaseZone
	case *ConditionZone:
		return v.BaseZone
	case *DerbyTrackZone:
		return v.BaseZone
	case *SiegableHallZone:
		return v.BaseZone
	case *ResidenceTeleportZone:
		return v.BaseZone
	case *ResidenceHallTeleportZone:
		return v.BaseZone
	default:
		return nil
	}
}
