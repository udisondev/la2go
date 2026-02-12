package testutil

import (
	"testing"

	"github.com/udisondev/la2go/internal/world"
)

// ResetTestWorld resets the world singleton and registers t.Cleanup
// to reset it again after the test (isolation between tests).
func ResetTestWorld(t testing.TB) {
	t.Helper()
	world.Instance().Reset()
	t.Cleanup(func() {
		world.Instance().Reset()
	})
}
