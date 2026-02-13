package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/game/sevensigns"
)

func TestSSQStatus_Write_Page1(t *testing.T) {
	t.Parallel()

	p := &SSQStatus{
		Page:          1,
		CurrentPeriod: sevensigns.PeriodCompetition,
		CurrentCycle:  3,
		PlayerCabal:   sevensigns.CabalDawn,
		PlayerSeal:    sevensigns.SealAvarice,
		PlayerStones:  150,
		PlayerAdena:   5000,
		DawnStoneScore: 300,
		DuskStoneScore: 200,
		DawnFestival:   50,
		DuskFestival:   30,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeSSQStatus {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSSQStatus)
	}
	if data[1] != 1 {
		t.Errorf("page = %d, want 1", data[1])
	}
	if data[2] != byte(sevensigns.PeriodCompetition) {
		t.Errorf("period = %d, want %d", data[2], sevensigns.PeriodCompetition)
	}

	// CurrentCycle at offset 3.
	cycle := int32(binary.LittleEndian.Uint32(data[3:7]))
	if cycle != 3 {
		t.Errorf("cycle = %d, want 3", cycle)
	}
}

func TestSSQStatus_Write_Page1_ZeroScores(t *testing.T) {
	t.Parallel()

	p := &SSQStatus{
		Page:          1,
		CurrentPeriod: sevensigns.PeriodRecruitment,
		CurrentCycle:  1,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeSSQStatus {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSSQStatus)
	}
	if data[1] != 1 {
		t.Errorf("page = %d, want 1", data[1])
	}
}

func TestSSQStatus_Write_Page2(t *testing.T) {
	t.Parallel()

	p := &SSQStatus{
		Page:          2,
		CurrentPeriod: sevensigns.PeriodCompetition,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeSSQStatus {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSSQStatus)
	}
	if data[1] != 2 {
		t.Errorf("page = %d, want 2", data[1])
	}

	// After header (3 bytes): short(1) + byte(5) = 5 bytes.
	// Then 5 festivals × (byte + int32 + int32 + byte + int32 + byte) = 5 × 15 = 75 bytes.
	minLen := 3 + 2 + 1 + 5*15
	if len(data) < minLen {
		t.Errorf("len(data) = %d, want >= %d", len(data), minLen)
	}
}

func TestSSQStatus_Write_Page3(t *testing.T) {
	t.Parallel()

	p := &SSQStatus{
		Page:          3,
		CurrentPeriod: sevensigns.PeriodSealValidation,
		SealOwners:    [4]sevensigns.Cabal{0, sevensigns.CabalDawn, sevensigns.CabalDusk, sevensigns.CabalNull},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeSSQStatus {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSSQStatus)
	}
	if data[1] != 3 {
		t.Errorf("page = %d, want 3", data[1])
	}

	// Header (3) + min%(1) + claim%(1) + sealCount(1) + 3 seals × (sealID(1) + owner(1) + duskPct(1) + dawnPct(1)) = 3+3+12 = 18.
	wantLen := 3 + 3 + 3*4
	if len(data) != wantLen {
		t.Errorf("len(data) = %d, want %d", len(data), wantLen)
	}

	// Check min/claim thresholds.
	if data[3] != 10 {
		t.Errorf("minRetain = %d, want 10", data[3])
	}
	if data[4] != 35 {
		t.Errorf("minClaim = %d, want 35", data[4])
	}
}

func TestSSQStatus_Write_Page4(t *testing.T) {
	t.Parallel()

	p := &SSQStatus{
		Page:          4,
		CurrentPeriod: sevensigns.PeriodCompetition,
		WinnerCabal:   sevensigns.CabalDawn,
		SealOwners:    [4]sevensigns.Cabal{0, sevensigns.CabalDawn, sevensigns.CabalNull, sevensigns.CabalDusk},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != OpcodeSSQStatus {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSSQStatus)
	}
	if data[1] != 4 {
		t.Errorf("page = %d, want 4", data[1])
	}

	// Header (3) + winner(1) + count(1) + 3 seals × (sealID(1) + owner(1) + msgID(2)) = 3+2+12 = 17.
	wantLen := 3 + 2 + 3*4
	if len(data) != wantLen {
		t.Errorf("len(data) = %d, want %d", len(data), wantLen)
	}

	// Winner.
	if data[3] != byte(sevensigns.CabalDawn) {
		t.Errorf("winner = %d, want %d", data[3], sevensigns.CabalDawn)
	}
}

func TestFestivalMaxScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tier int
		want int32
	}{
		{0, 60},
		{1, 70},
		{2, 100},
		{3, 120},
		{4, 150},
		{-1, 0},
		{5, 0},
	}
	for _, tt := range tests {
		got := festivalMaxScore(tt.tier)
		if got != tt.want {
			t.Errorf("festivalMaxScore(%d) = %d, want %d", tt.tier, got, tt.want)
		}
	}
}

func TestPeriodMsgIDs(t *testing.T) {
	t.Parallel()

	// Все периоды должны возвращать ненулевые ID.
	for _, p := range []sevensigns.Period{
		sevensigns.PeriodRecruitment,
		sevensigns.PeriodCompetition,
		sevensigns.PeriodResults,
		sevensigns.PeriodSealValidation,
	} {
		if periodDescMsgID(p) == 0 {
			t.Errorf("periodDescMsgID(%d) = 0", p)
		}
		if periodEndMsgID(p) == 0 {
			t.Errorf("periodEndMsgID(%d) = 0", p)
		}
	}
}
