package ai

import (
	"io"
	"log/slog"
	"testing"
)

// BenchmarkDebugLog_Disabled measures overhead when debug logging is DISABLED (production).
// Expected: ~1ns per check (atomic.Load only, slog.Debug not called).
func BenchmarkDebugLog_Disabled(b *testing.B) {
	// Configure slog with Info level (debug logs won't be emitted even if called)
	// Discard output to avoid benchmark noise
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Disable debug logging
	EnableDebugLogging(false)

	b.ResetTimer()
	for range b.N {
		if IsDebugEnabled() {
			slog.Debug("AI tick completed", "controllers", 100)
		}
	}
}

// BenchmarkDebugLog_Enabled measures overhead when debug logging is ENABLED (development).
// Expected: ~50ns per call (atomic.Load + slog.Debug call with formatting).
func BenchmarkDebugLog_Enabled(b *testing.B) {
	// Configure slog with Debug level
	// Discard output to measure only formatting overhead
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// Enable debug logging
	EnableDebugLogging(true)

	b.ResetTimer()
	for range b.N {
		if IsDebugEnabled() {
			slog.Debug("AI tick completed", "controllers", 100)
		}
	}
}

// BenchmarkDebugLog_Baseline_NoGuard measures baseline overhead WITHOUT guard (always calls slog.Debug).
// This simulates current (Phase 4.4) behavior where slog.Debug is called unconditionally.
func BenchmarkDebugLog_Baseline_NoGuard(b *testing.B) {
	// Configure slog with Info level (debug logs won't be emitted)
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	b.ResetTimer()
	for range b.N {
		slog.Debug("AI tick completed", "controllers", 100)
	}
}

// BenchmarkIsDebugEnabled measures raw performance of IsDebugEnabled() check.
func BenchmarkIsDebugEnabled(b *testing.B) {
	EnableDebugLogging(false)

	b.ResetTimer()
	for range b.N {
		_ = IsDebugEnabled()
	}
}
