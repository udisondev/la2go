package integration

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/model"
)

// noopPersister is a no-op implementation of gameserver.PlayerPersister for tests.
type noopPersister struct{}

func (n *noopPersister) SavePlayer(_ context.Context, _ *model.Player) error {
	return nil
}

func (n *noopPersister) LoadPlayerData(_ context.Context, _ int64) (*db.PlayerData, error) {
	return &db.PlayerData{}, nil
}

// noopCharRepo is a no-op implementation of gameserver.CharacterRepository for tests
// that don't need database access.
type noopCharRepo struct{}

func (n *noopCharRepo) LoadByAccountName(_ context.Context, _ string) ([]*model.Player, error) {
	return nil, nil
}

func (n *noopCharRepo) Create(_ context.Context, _ string, _ *model.Player) error {
	return nil
}

func (n *noopCharRepo) NameExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (n *noopCharRepo) CountByAccountName(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (n *noopCharRepo) MarkForDeletion(_ context.Context, _ int64, _ int64) error {
	return nil
}

func (n *noopCharRepo) RestoreCharacter(_ context.Context, _ int64) error {
	return nil
}

func (n *noopCharRepo) GetClanID(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

// testOIDCounter provides unique ObjectIDs for parallel tests.
var testOIDCounter atomic.Uint32

func init() {
	testOIDCounter.Store(200_000) // far from production ranges
}

// nextOID returns a globally unique ObjectID for tests.
func nextOID() uint32 {
	return testOIDCounter.Add(1)
}

// schemaCounter provides unique schema names for parallel suites.
var schemaCounter atomic.Uint32

// acquireSchema creates an isolated PostgreSQL schema and returns DSN with search_path.
// Schema is automatically dropped via t.Cleanup.
func acquireSchema(t testing.TB) string {
	t.Helper()
	ctx := context.Background()

	schemaName := fmt.Sprintf("test_%d", schemaCounter.Add(1))

	conn, err := pgx.Connect(ctx, sharedPGBaseDSN)
	if err != nil {
		t.Fatalf("connect to shared postgres: %v", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create schema %s: %v", schemaName, err)
	}

	t.Cleanup(func() {
		cleanCtx := context.Background()
		cleanConn, err := pgx.Connect(cleanCtx, sharedPGBaseDSN)
		if err != nil {
			t.Logf("cleanup: connect failed: %v", err)
			return
		}
		defer cleanConn.Close(cleanCtx)
		if _, err := cleanConn.Exec(cleanCtx, "DROP SCHEMA "+schemaName+" CASCADE"); err != nil {
			t.Logf("cleanup: drop schema %s: %v", schemaName, err)
		}
	})

	// Append search_path to DSN
	sep := "&"
	if !strings.Contains(sharedPGBaseDSN, "?") {
		sep = "?"
	}
	return sharedPGBaseDSN + sep + "search_path=" + schemaName
}
