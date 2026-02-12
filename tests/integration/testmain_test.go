package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/udisondev/la2go/internal/data"
)

// sharedPGBaseDSN is the base DSN for the shared PostgreSQL container.
// Each suite gets its own schema via acquireSchema().
var sharedPGBaseDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start shared PostgreSQL container (once for all suites)
	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("la2go_test"),
		postgres.WithUsername("la2go"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	sharedPGBaseDSN, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Load data templates
	data.InitStatBonuses()
	if err := data.LoadPlayerTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "LoadPlayerTemplates failed: %v\n", err)
		os.Exit(1)
	}
	if err := data.LoadNpcTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "LoadNpcTemplates failed: %v\n", err)
		os.Exit(1)
	}
	if err := data.LoadItemTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "LoadItemTemplates failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testcontainers.TerminateContainer(container); err != nil {
		fmt.Fprintf(os.Stderr, "failed to terminate postgres container: %v\n", err)
	}

	os.Exit(code)
}
