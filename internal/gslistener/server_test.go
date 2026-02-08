package gslistener

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/login"
)

func TestNewServer(t *testing.T) {
	cfg := config.LoginServer{
		GSListenHost: "127.0.0.1",
		GSListenPort: 9013,
	}

	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)

	sessionManager := login.NewSessionManager()
	srv, err := NewServer(cfg, database, gsTable, sessionManager)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Verify RSA key pool (10 keys)
	assert.Len(t, srv.rsaKeyPairs, 10)
	for i, key := range srv.rsaKeyPairs {
		assert.NotNil(t, key, "RSA key %d should not be nil", i)
	}

	// Verify handler created
	assert.NotNil(t, srv.handler)

	// Verify pools created
	assert.NotNil(t, srv.sendPool)
	assert.NotNil(t, srv.readPool)
}

func TestServerRun(t *testing.T) {
	cfg := config.LoginServer{
		GSListenHost: "127.0.0.1",
		GSListenPort: 0, // random port
	}

	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)

	sessionManager := login.NewSessionManager()
	srv, err := NewServer(cfg, database, gsTable, sessionManager)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run server (will timeout after 1 second)
	err = srv.Run(ctx)

	// Should return context.DeadlineExceeded or nil (graceful shutdown)
	if err != nil {
		assert.Equal(t, context.DeadlineExceeded, err)
	}
}
