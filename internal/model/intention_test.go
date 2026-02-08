package model

import "testing"

func TestIntentionString(t *testing.T) {
	tests := []struct {
		intention Intention
		want      string
	}{
		{IntentionIdle, "IDLE"},
		{IntentionActive, "ACTIVE"},
		{IntentionAttack, "ATTACK"},
		{IntentionCast, "CAST"},
		{IntentionMoveTo, "MOVE_TO"},
		{Intention(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.intention.String(); got != tt.want {
				t.Errorf("Intention.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
