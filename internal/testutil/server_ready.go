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

// WaitForCleanup ждёт пока cleanup condition будет выполнено (polling с timeout).
// Используется для явной проверки cleanup после disconnect в integration тестах.
//
// Пример:
//
//	client.Close()
//	testutil.WaitForCleanup(t, func() bool {
//	    // Проверяем что сервер готов принимать новые подключения
//	    return canConnectTo(addr)
//	}, 5*time.Second)
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
