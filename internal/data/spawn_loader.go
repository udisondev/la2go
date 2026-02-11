package data

import "log/slog"

// SpawnList — все spawn определения из XML.
var SpawnList []spawnDef

// LoadSpawns загружает спавны из Go-литералов.
func LoadSpawns() error {
	SpawnList = spawnDefs
	slog.Info("loaded spawns", "count", len(SpawnList))
	return nil
}

// SpawnDef accessor methods
func (s *spawnDef) NpcID() int32        { return s.npcID }
func (s *spawnDef) Count() int32        { return s.count }
func (s *spawnDef) RespawnDelay() int32 { return s.respawnDelay }
func (s *spawnDef) RespawnRand() int32  { return s.respawnRand }
func (s *spawnDef) ChaseRange() int32   { return s.chaseRange }
func (s *spawnDef) X() int32            { return s.x }
func (s *spawnDef) Y() int32            { return s.y }
func (s *spawnDef) Z() int32            { return s.z }
func (s *spawnDef) Heading() int32      { return s.heading }
func (s *spawnDef) Territory() *territoryDef { return s.territory }
func (s *spawnDef) HasTerritory() bool  { return s.territory != nil }

func (t *territoryDef) MinZ() int32     { return t.minZ }
func (t *territoryDef) MaxZ() int32     { return t.maxZ }
func (t *territoryDef) Nodes() []pointDef { return t.nodes }

func (p *pointDef) X() int32 { return p.x }
func (p *pointDef) Y() int32 { return p.y }
