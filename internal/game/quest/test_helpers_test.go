package quest

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// createTestPlayer creates a minimal player for testing.
func createTestPlayer(t *testing.T, charID int64) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(uint32(charID), charID, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("creating test player: %v", err)
	}
	return p
}
