package ai

import "sync/atomic"

// debugLoggingEnabled controls whether debug logging is enabled for AI subsystem.
// This is a package-level flag to avoid the overhead of checking log level on every call.
// Set via EnableDebugLogging() during initialization based on config.LogLevel.
var debugLoggingEnabled atomic.Bool

// EnableDebugLogging enables or disables debug logging for AI subsystem.
// Must be called during initialization (e.g., from main.go after parsing config).
func EnableDebugLogging(enabled bool) {
	debugLoggingEnabled.Store(enabled)
}

// IsDebugEnabled returns true if debug logging is enabled.
// Use this to guard expensive debug log calls:
//
//	if ai.IsDebugEnabled() {
//	    slog.Debug("expensive operation", "data", computeExpensiveData())
//	}
func IsDebugEnabled() bool {
	return debugLoggingEnabled.Load()
}
