package data

import "log/slog"

// zoneDef — определение зоны (generated).
type zoneDef struct {
	name     string
	id       int32
	zoneType string // "TownZone","CastleZone","EffectZone","WaterZone","DamageZone","SiegeZone",...
	shape    string // "NPoly","Cuboid","Cylinder"
	minZ     int32
	maxZ     int32
	rad      int32  // radius for Cylinder shape
	nodes    []pointDef
	params   map[string]string // stat name→val pairs
	spawns   []zoneSpawnDef    // restart/banish points
}

type zoneSpawnDef struct {
	x, y, z   int32
	spawnType string // "","other","chaotic","banish"
}

// ZoneTable — все зоны по ID.
var ZoneTable map[int32]*zoneDef

// ZonesByType — зоны по типу.
var ZonesByType map[string][]*zoneDef

// LoadZones загружает зоны из Go-литералов.
func LoadZones() error {
	ZoneTable = make(map[int32]*zoneDef, len(zoneDefs))
	ZonesByType = make(map[string][]*zoneDef)

	for i := range zoneDefs {
		z := &zoneDefs[i]
		ZoneTable[z.id] = z
		ZonesByType[z.zoneType] = append(ZonesByType[z.zoneType], z)
	}

	slog.Info("loaded zones", "count", len(ZoneTable), "types", len(ZonesByType))
	return nil
}

// GetZone возвращает zoneDef по ID.
func GetZone(id int32) *zoneDef {
	if ZoneTable == nil {
		return nil
	}
	return ZoneTable[id]
}

// Accessor methods
func (z *zoneDef) ZoneName() string           { return z.name }
func (z *zoneDef) ZoneID() int32              { return z.id }
func (z *zoneDef) ZoneType() string           { return z.zoneType }
func (z *zoneDef) Shape() string              { return z.shape }
func (z *zoneDef) MinZ() int32                { return z.minZ }
func (z *zoneDef) MaxZ() int32                { return z.maxZ }
func (z *zoneDef) Nodes() []pointDef          { return z.nodes }
func (z *zoneDef) Params() map[string]string  { return z.params }
func (z *zoneDef) Rad() int32                 { return z.rad }
func (z *zoneDef) Spawns() []zoneSpawnDef     { return z.spawns }
