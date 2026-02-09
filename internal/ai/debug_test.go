package ai

import "testing"

func TestEnableDebugLogging(t *testing.T) {
	// Reset state
	EnableDebugLogging(false)

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"enable", true, true},
		{"disable", false, false},
		{"enable again", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnableDebugLogging(tt.enabled)
			if got := IsDebugEnabled(); got != tt.expected {
				t.Errorf("IsDebugEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsDebugEnabled_Concurrent(t *testing.T) {
	EnableDebugLogging(true)

	// Verify concurrent reads are safe
	done := make(chan bool)
	for range 100 {
		go func() {
			for range 1000 {
				_ = IsDebugEnabled()
			}
			done <- true
		}()
	}

	for range 100 {
		<-done
	}
}
