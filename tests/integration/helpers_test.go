package integration

import (
	"context"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/model"
)

// noopPersister is a no-op implementation of gameserver.PlayerPersister for tests.
type noopPersister struct{}

func (n *noopPersister) SavePlayer(_ context.Context, _ *model.Player) error {
	return nil
}

func (n *noopPersister) LoadPlayerData(_ context.Context, _ int64) ([]db.ItemRow, []*model.SkillInfo, error) {
	return nil, nil, nil
}
