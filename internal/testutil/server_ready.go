package testutil

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// WaitForTCPReady ждёт пока TCP сервер станет доступен (polling с timeout).
// Используется вместо time.Sleep для синхронизации в integration тестах.
//
// Пример:
//
//	go server.Serve(ctx, listener)
//	if err := testutil.WaitForTCPReady(addr, 5*time.Second); err != nil {
//	    t.Fatalf("server failed to start: %v", err)
//	}
func WaitForTCPReady(addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server at %s: %w", addr, ctx.Err())
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return nil
			}
			// Продолжаем polling если не удалось подключиться
		}
	}
}

// WaitForCondition polls condition every 10ms until it returns true or timeout.
// Returns error on timeout (does not call t.Fatal).
func WaitForCondition(condition func() bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("condition not met within %v: %w", timeout, ctx.Err())
		case <-ticker.C:
			if condition() {
				return nil
			}
		}
	}
}

// Eventually asserts that condition becomes true within timeout.
// Calls t.Fatal on timeout. Polling interval: 10ms.
func Eventually(t testing.TB, condition func() bool, timeout time.Duration, msgAndArgs ...any) {
	t.Helper()

	if err := WaitForCondition(condition, timeout); err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: %v", fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...), err)
		}
		t.Fatalf("eventually: %v", err)
	}
}

// WaitForCleanup ждёт пока cleanup condition будет выполнено (polling с timeout).
// Используется для явной проверки cleanup после disconnect в integration тестах.
func WaitForCleanup(t testing.TB, check func() bool, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("cleanup timeout: condition not met within %v", timeout)
		case <-ticker.C:
			if check() {
				return
			}
		}
	}
}
