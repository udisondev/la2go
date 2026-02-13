package clientpackets

import (
	"testing"
)

func TestParseRequestCursedWeaponList(t *testing.T) {
	t.Parallel()

	pkt, err := ParseRequestCursedWeaponList(nil)
	if err != nil {
		t.Fatalf("ParseRequestCursedWeaponList() error = %v", err)
	}
	if pkt == nil {
		t.Fatal("ParseRequestCursedWeaponList() returned nil")
	}
}

func TestParseRequestCursedWeaponLocation(t *testing.T) {
	t.Parallel()

	pkt, err := ParseRequestCursedWeaponLocation([]byte{})
	if err != nil {
		t.Fatalf("ParseRequestCursedWeaponLocation() error = %v", err)
	}
	if pkt == nil {
		t.Fatal("ParseRequestCursedWeaponLocation() returned nil")
	}
}

func TestCursedWeaponClientSubOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int16
		want int16
	}{
		{"RequestCursedWeaponList", SubOpcodeRequestCursedWeaponList, 0x22},
		{"RequestCursedWeaponLocation", SubOpcodeRequestCursedWeaponLocation, 0x23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = 0x%02X, want 0x%02X", tt.name, tt.got, tt.want)
			}
		})
	}
}
