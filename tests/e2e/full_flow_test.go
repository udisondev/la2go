package e2e

import (
	"os"
	"testing"
)

// TestFullLoginFlow тестирует полный end-to-end flow:
// Client → LoginServer → gslistener → GameServer
// Requires running PostgreSQL, LoginServer, and GameServer infrastructure.
func TestFullLoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	dbAddr := os.Getenv("DB_ADDR")
	if dbAddr == "" {
		t.Skip("DB_ADDR not set, skipping e2e tests")
	}

	// Full E2E flow requires running LoginServer + GameServer instances with DB.
	// GameServer is implemented (Phase 4+), but E2E test harness (multi-process orchestration) not built yet.
	t.Skip("E2E test harness not implemented: requires multi-process orchestration for LS + GS")
}
