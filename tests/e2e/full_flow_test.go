package e2e

import (
	"os"
	"testing"
)

// TestFullLoginFlow тестирует полный end-to-end flow:
// Client → LoginServer → gslistener → GameServer
// TODO: Требует реализации GameServer для полного flow.
func TestFullLoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	dbAddr := os.Getenv("DB_ADDR")
	if dbAddr == "" {
		t.Skip("DB_ADDR not set, skipping e2e tests")
	}

	// TODO: Реализовать полный flow когда будет Phase 4+ (GameServer)
	t.Skip("E2E tests require GameServer implementation (Phase 4+)")
}
