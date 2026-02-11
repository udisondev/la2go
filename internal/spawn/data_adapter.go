package spawn

import (
	"context"
	"fmt"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// DataNpcRepo implements NpcRepository using the data package (XML-generated Go literals).
type DataNpcRepo struct{}

// NewDataNpcRepo creates a DataNpcRepo adapter.
func NewDataNpcRepo() *DataNpcRepo {
	return &DataNpcRepo{}
}

// LoadTemplate loads NPC template from data package by ID.
func (r *DataNpcRepo) LoadTemplate(_ context.Context, templateID int32) (*model.NpcTemplate, error) {
	def := data.GetNpcDef(templateID)
	if def == nil {
		return nil, fmt.Errorf("NPC template %d not found in data", templateID)
	}

	return model.NewNpcTemplate(
		def.ID(),
		def.Name(),
		def.Title(),
		def.Level(),
		int32(def.HP()),
		int32(def.MP()),
		int32(def.PAtk()),
		int32(def.PDef()),
		int32(def.MAtk()),
		int32(def.MDef()),
		def.AggroRange(),
		def.RunSpeed(),
		def.AtkSpeed(),
		0, 0, // respawnMin/Max — now stored per-spawn, not per-template
		def.BaseExp(),
		def.BaseSP(),
	), nil
}

// DataSpawnRepo implements SpawnRepository using the data package (XML-generated Go literals).
type DataSpawnRepo struct{}

// NewDataSpawnRepo creates a DataSpawnRepo adapter.
func NewDataSpawnRepo() *DataSpawnRepo {
	return &DataSpawnRepo{}
}

// LoadAll loads all spawns from data package.
func (r *DataSpawnRepo) LoadAll(_ context.Context) ([]*model.Spawn, error) {
	spawns := make([]*model.Spawn, 0, len(data.SpawnList))

	for i := range data.SpawnList {
		sd := &data.SpawnList[i]

		count := sd.Count()
		if count <= 0 {
			count = 1
		}

		spawn := model.NewSpawn(
			int64(i+1), // sequential spawn ID
			sd.NpcID(),
			sd.X(), sd.Y(), sd.Z(),
			uint16(sd.Heading()),
			count,
			sd.RespawnDelay() > 0, // doRespawn if delay > 0
		)

		spawn.SetRespawnTimes(sd.RespawnDelay(), sd.RespawnRand())
		spawns = append(spawns, spawn)
	}

	return spawns, nil
}
