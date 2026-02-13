package manor

import "testing"

func TestMode_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode Mode
		want string
	}{
		{ModeDisabled, "DISABLED"},
		{ModeModifiable, "MODIFIABLE"},
		{ModeMaintenance, "MAINTENANCE"},
		{ModeApproved, "APPROVED"},
		{Mode(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.want {
			t.Errorf("Mode(%d).String() = %q; want %q", tt.mode, got, tt.want)
		}
	}
}
