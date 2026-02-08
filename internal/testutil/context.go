package testutil

import (
	"context"
	"testing"
	"time"
)

// ContextWithTimeout создаёт context с timeout и автоматически отменяет его при завершении теста.
func ContextWithTimeout(t testing.TB, duration time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	t.Cleanup(cancel)

	return ctx
}

// ContextWithDeadline создаёт context с deadline и автоматически отменяет его при завершении теста.
func ContextWithDeadline(t testing.TB, deadline time.Time) context.Context {
	t.Helper()

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	t.Cleanup(cancel)

	return ctx
}

// ContextWithCancel создаёт context с cancel и автоматически отменяет его при завершении теста.
func ContextWithCancel(t testing.TB) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return ctx, cancel
}
