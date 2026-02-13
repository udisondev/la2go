package serverpackets

import (
	"testing"
)

func TestExCursedWeaponList_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		weaponIDs []int32
		wantLen   int // minimum expected length
	}{
		{
			name:      "empty list",
			weaponIDs: nil,
			wantLen:   7, // 1 opcode + 2 subop + 4 count
		},
		{
			name:      "two weapons",
			weaponIDs: []int32{8190, 8689},
			wantLen:   15, // 1 + 2 + 4 + 2*4
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkt := ExCursedWeaponList{WeaponIDs: tt.weaponIDs}
			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if len(data) < tt.wantLen {
				t.Fatalf("Write() len = %d, want >= %d", len(data), tt.wantLen)
			}

			if data[0] != 0xFE {
				t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
			}

			subOp := int16(data[1]) | int16(data[2])<<8
			if subOp != SubOpcodeExCursedWeaponList {
				t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExCursedWeaponList)
			}
		})
	}
}

func TestExCursedWeaponLocation_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		weapons []CursedWeaponLocationInfo
		wantLen int
	}{
		{
			name:    "empty",
			weapons: nil,
			wantLen: 11, // 1 + 2 + 4 + 4 (zero case writes extra int)
		},
		{
			name: "one weapon dropped",
			weapons: []CursedWeaponLocationInfo{
				{ItemID: 8190, Activated: 0, X: 100, Y: 200, Z: 300},
			},
			wantLen: 27, // 1 + 2 + 4 + 1*20
		},
		{
			name: "two weapons",
			weapons: []CursedWeaponLocationInfo{
				{ItemID: 8190, Activated: 1, X: 100, Y: 200, Z: 300},
				{ItemID: 8689, Activated: 0, X: 400, Y: 500, Z: 600},
			},
			wantLen: 47, // 1 + 2 + 4 + 2*20
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkt := ExCursedWeaponLocation{Weapons: tt.weapons}
			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if len(data) < tt.wantLen {
				t.Fatalf("Write() len = %d, want >= %d", len(data), tt.wantLen)
			}

			if data[0] != 0xFE {
				t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
			}

			subOp := int16(data[1]) | int16(data[2])<<8
			if subOp != SubOpcodeExCursedWeaponLocation {
				t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExCursedWeaponLocation)
			}
		})
	}
}

func TestCursedWeaponSubOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int16
		want int16
	}{
		{"ExCursedWeaponList", SubOpcodeExCursedWeaponList, 0x45},
		{"ExCursedWeaponLocation", SubOpcodeExCursedWeaponLocation, 0x46},
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
