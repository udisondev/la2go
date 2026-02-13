package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// testSummon -- helper, создающий Summon для тестов pet-пакетов.
func testSummon(t *testing.T, objectID, ownerID uint32, summonType model.SummonType) *model.Summon {
	t.Helper()
	tmpl := model.NewNpcTemplate(
		12077, "Wolf", "", 15,
		500, 200,
		100, 50, 80, 40,
		0,
		120, 300,
		60, 120,
		100, 10,
	)
	s := model.NewSummon(
		objectID, ownerID,
		summonType, 12077, tmpl,
		"Wolf", 15,
		500, 200,
		100, 50, 80, 40,
	)
	s.SetLocation(model.NewLocation(10000, 20000, -3000, 512))
	return s
}

// readInt32 -- helper для чтения int32 через packet.Reader.
func readInt32(t *testing.T, r *packet.Reader) int32 {
	t.Helper()
	v, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt: %v", err)
	}
	return v
}

// readInt16 -- helper для чтения int16 через packet.Reader.
func readInt16(t *testing.T, r *packet.Reader) int16 {
	t.Helper()
	v, err := r.ReadShort()
	if err != nil {
		t.Fatalf("ReadShort: %v", err)
	}
	return v
}

// readInt64 -- helper для чтения int64 через packet.Reader.
func readInt64(t *testing.T, r *packet.Reader) int64 {
	t.Helper()
	v, err := r.ReadLong()
	if err != nil {
		t.Fatalf("ReadLong: %v", err)
	}
	return v
}

// readFloat64 -- helper для чтения float64 через packet.Reader.
func readFloat64(t *testing.T, r *packet.Reader) float64 {
	t.Helper()
	v, err := r.ReadDouble()
	if err != nil {
		t.Fatalf("ReadDouble: %v", err)
	}
	return v
}

// readByte -- helper для чтения byte через packet.Reader.
func readByte(t *testing.T, r *packet.Reader) byte {
	t.Helper()
	v, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	return v
}

// readString -- helper для чтения строки через packet.Reader.
func readString(t *testing.T, r *packet.Reader) string {
	t.Helper()
	v, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	return v
}

// ---------- PetInfo (0xB1) ----------

func TestPetInfo_Write(t *testing.T) {
	tests := []struct {
		name       string
		summonType model.SummonType
		objectID   uint32
		ownerID    uint32
		ownerName  string
		currentFed int32
		maxFed     int32
		exp        int64
		expMax     int64
		withFeed   bool
	}{
		{
			name:       "pet basic",
			summonType: model.SummonTypePet,
			objectID:   5000,
			ownerID:    1000,
			ownerName:  "TestOwner",
			withFeed:   false,
		},
		{
			name:       "servitor basic",
			summonType: model.SummonTypeServitor,
			objectID:   6000,
			ownerID:    2000,
			ownerName:  "Summoner",
			withFeed:   false,
		},
		{
			name:       "pet with feed",
			summonType: model.SummonTypePet,
			objectID:   7000,
			ownerID:    3000,
			ownerName:  "PetKeeper",
			currentFed: 50,
			maxFed:     100,
			exp:        12345,
			expMax:     99999,
			withFeed:   true,
		},
		{
			name:       "servitor with feed",
			summonType: model.SummonTypeServitor,
			objectID:   8000,
			ownerID:    4000,
			ownerName:  "WizardX",
			currentFed: 0,
			maxFed:     0,
			exp:        0,
			expMax:     0,
			withFeed:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summon := testSummon(t, tc.objectID, tc.ownerID, tc.summonType)

			var pkt PetInfo
			if tc.withFeed {
				pkt = NewPetInfoWithFeed(summon, tc.ownerName, tc.currentFed, tc.maxFed, tc.exp, tc.expMax)
			} else {
				pkt = NewPetInfo(summon, tc.ownerName)
			}

			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			r := packet.NewReader(data)

			// Opcode
			opcode := readByte(t, r)
			if opcode != OpcodePetInfo {
				t.Errorf("opcode: got 0x%02X, want 0x%02X", opcode, OpcodePetInfo)
			}

			// Summon type
			gotType := readInt32(t, r)
			if gotType != int32(tc.summonType) {
				t.Errorf("summonType: got %d, want %d", gotType, int32(tc.summonType))
			}

			// ObjectID
			gotObjectID := readInt32(t, r)
			if gotObjectID != int32(tc.objectID) {
				t.Errorf("objectID: got %d, want %d", gotObjectID, int32(tc.objectID))
			}

			// TemplateID + 1000000
			gotTemplateID := readInt32(t, r)
			wantTemplateID := int32(12077 + 1000000)
			if gotTemplateID != wantTemplateID {
				t.Errorf("templateID: got %d, want %d", gotTemplateID, wantTemplateID)
			}

			// isAutoAttackable
			gotAutoAtk := readInt32(t, r)
			if gotAutoAtk != 0 {
				t.Errorf("isAutoAttackable: got %d, want 0", gotAutoAtk)
			}

			// Position X, Y, Z, Heading
			gotX := readInt32(t, r)
			if gotX != 10000 {
				t.Errorf("X: got %d, want 10000", gotX)
			}
			gotY := readInt32(t, r)
			if gotY != 20000 {
				t.Errorf("Y: got %d, want 20000", gotY)
			}
			gotZ := readInt32(t, r)
			if gotZ != -3000 {
				t.Errorf("Z: got %d, want -3000", gotZ)
			}
			gotHeading := readInt32(t, r)
			if gotHeading != 512 {
				t.Errorf("Heading: got %d, want 512", gotHeading)
			}

			// padding
			_ = readInt32(t, r)

			// MAtkSpd (0), PAtkSpd, RunSpeed, WalkSpeed
			gotMAtkSpd := readInt32(t, r)
			if gotMAtkSpd != 0 {
				t.Errorf("MAtkSpd: got %d, want 0", gotMAtkSpd)
			}

			gotPAtkSpd := readInt32(t, r)
			wantAtkSpeed := summon.AtkSpeed()
			if gotPAtkSpd != wantAtkSpeed {
				t.Errorf("PAtkSpd: got %d, want %d", gotPAtkSpd, wantAtkSpeed)
			}

			gotRunSpeed := readInt32(t, r)
			wantMoveSpeed := summon.MoveSpeed()
			if gotRunSpeed != wantMoveSpeed {
				t.Errorf("RunSpeed: got %d, want %d", gotRunSpeed, wantMoveSpeed)
			}

			gotWalkSpeed := readInt32(t, r)
			wantWalkSpeed := wantMoveSpeed / 2
			if gotWalkSpeed != wantWalkSpeed {
				t.Errorf("WalkSpeed: got %d, want %d", gotWalkSpeed, wantWalkSpeed)
			}

			// Swim/Fly speeds (6 ints)
			gotSwimRun := readInt32(t, r)
			if gotSwimRun != wantMoveSpeed {
				t.Errorf("SwimRunSpeed: got %d, want %d", gotSwimRun, wantMoveSpeed)
			}
			gotSwimWalk := readInt32(t, r)
			if gotSwimWalk != wantWalkSpeed {
				t.Errorf("SwimWalkSpeed: got %d, want %d", gotSwimWalk, wantWalkSpeed)
			}
			// 4 fly speeds (all 0)
			for i := range 4 {
				flySpeed := readInt32(t, r)
				if flySpeed != 0 {
					t.Errorf("FlySpeed[%d]: got %d, want 0", i, flySpeed)
				}
			}

			// Movement and attack speed multipliers (double)
			gotMoveMult := readFloat64(t, r)
			if gotMoveMult != 1.0 {
				t.Errorf("moveMultiplier: got %f, want 1.0", gotMoveMult)
			}
			gotAtkMult := readFloat64(t, r)
			if gotAtkMult != 1.0 {
				t.Errorf("atkMultiplier: got %f, want 1.0", gotAtkMult)
			}

			// Collision radius and height (doubles)
			gotRadius := readFloat64(t, r)
			if gotRadius != 12.0 {
				t.Errorf("collisionRadius: got %f, want 12.0", gotRadius)
			}
			gotHeight := readFloat64(t, r)
			if gotHeight != 22.0 {
				t.Errorf("collisionHeight: got %f, want 22.0", gotHeight)
			}

			// Weapon/Armor (3 ints)
			for i := range 3 {
				val := readInt32(t, r)
				if val != 0 {
					t.Errorf("weapon/armor[%d]: got %d, want 0", i, val)
				}
			}

			// Owner objectID
			gotOwnerID := readInt32(t, r)
			if gotOwnerID != int32(tc.ownerID) {
				t.Errorf("ownerID: got %d, want %d", gotOwnerID, int32(tc.ownerID))
			}

			// Booleans (4 bytes)
			gotAutoAtkBool := readByte(t, r)
			if gotAutoAtkBool != 0 {
				t.Errorf("isAutoAttackable (bool): got %d, want 0", gotAutoAtkBool)
			}
			gotAtkNow := readByte(t, r)
			if gotAtkNow != 0 {
				t.Errorf("isAttackingNow: got %d, want 0", gotAtkNow)
			}
			gotAlikeDead := readByte(t, r)
			if gotAlikeDead != 0 {
				t.Errorf("isAlikeDead: got %d, want 0", gotAlikeDead)
			}
			gotShowName := readByte(t, r)
			if gotShowName != 1 {
				t.Errorf("showName: got %d, want 1", gotShowName)
			}

			// Name
			gotName := readString(t, r)
			if gotName != "Wolf" {
				t.Errorf("name: got %q, want %q", gotName, "Wolf")
			}

			// Title (owner name)
			gotTitle := readString(t, r)
			if gotTitle != tc.ownerName {
				t.Errorf("title: got %q, want %q", gotTitle, tc.ownerName)
			}

			// showSpawnAnimation
			gotSpawnAnim := readInt32(t, r)
			if gotSpawnAnim != 1 {
				t.Errorf("showSpawnAnimation: got %d, want 1", gotSpawnAnim)
			}

			// pvpFlag
			gotPvP := readInt32(t, r)
			if gotPvP != 0 {
				t.Errorf("pvpFlag: got %d, want 0", gotPvP)
			}

			// karma
			gotKarma := readInt32(t, r)
			if gotKarma != 0 {
				t.Errorf("karma: got %d, want 0", gotKarma)
			}

			// Feed
			gotFed := readInt32(t, r)
			gotMaxFed := readInt32(t, r)
			if tc.withFeed {
				if gotFed != tc.currentFed {
					t.Errorf("currentFed: got %d, want %d", gotFed, tc.currentFed)
				}
				if gotMaxFed != tc.maxFed {
					t.Errorf("maxFed: got %d, want %d", gotMaxFed, tc.maxFed)
				}
			} else {
				if gotFed != 0 {
					t.Errorf("currentFed (no feed): got %d, want 0", gotFed)
				}
				if gotMaxFed != 0 {
					t.Errorf("maxFed (no feed): got %d, want 0", gotMaxFed)
				}
			}

			// HP/MP
			gotCurHP := readInt32(t, r)
			if gotCurHP != summon.CurrentHP() {
				t.Errorf("currentHP: got %d, want %d", gotCurHP, summon.CurrentHP())
			}
			gotMaxHP := readInt32(t, r)
			if gotMaxHP != summon.MaxHP() {
				t.Errorf("maxHP: got %d, want %d", gotMaxHP, summon.MaxHP())
			}
			gotCurMP := readInt32(t, r)
			if gotCurMP != summon.CurrentMP() {
				t.Errorf("currentMP: got %d, want %d", gotCurMP, summon.CurrentMP())
			}
			gotMaxMP := readInt32(t, r)
			if gotMaxMP != summon.MaxMP() {
				t.Errorf("maxMP: got %d, want %d", gotMaxMP, summon.MaxMP())
			}

			// Level
			gotLevel := readInt32(t, r)
			if gotLevel != summon.Level() {
				t.Errorf("level: got %d, want %d", gotLevel, summon.Level())
			}

			// Exp, ExpMax, ExpNext
			gotExp := readInt64(t, r)
			gotExpMax := readInt64(t, r)
			_ = readInt64(t, r) // expNext (0)

			if tc.withFeed {
				if gotExp != tc.exp {
					t.Errorf("exp: got %d, want %d", gotExp, tc.exp)
				}
				if gotExpMax != tc.expMax {
					t.Errorf("expMax: got %d, want %d", gotExpMax, tc.expMax)
				}
			} else {
				if gotExp != 0 {
					t.Errorf("exp (no feed): got %d, want 0", gotExp)
				}
				if gotExpMax != 0 {
					t.Errorf("expMax (no feed): got %d, want 0", gotExpMax)
				}
			}

			// Weight (2 ints)
			_ = readInt32(t, r) // current weight
			_ = readInt32(t, r) // max weight

			// Combat stats
			gotPAtk := readInt32(t, r)
			if gotPAtk != summon.PAtk() {
				t.Errorf("pAtk: got %d, want %d", gotPAtk, summon.PAtk())
			}
			gotPDef := readInt32(t, r)
			if gotPDef != summon.PDef() {
				t.Errorf("pDef: got %d, want %d", gotPDef, summon.PDef())
			}
			gotMAtk := readInt32(t, r)
			if gotMAtk != summon.MAtk() {
				t.Errorf("mAtk: got %d, want %d", gotMAtk, summon.MAtk())
			}
			gotMDef := readInt32(t, r)
			if gotMDef != summon.MDef() {
				t.Errorf("mDef: got %d, want %d", gotMDef, summon.MDef())
			}

			// accuracy, evasion, critical (all 0)
			for _, statName := range []string{"accuracy", "evasion", "critical"} {
				val := readInt32(t, r)
				if val != 0 {
					t.Errorf("%s: got %d, want 0", statName, val)
				}
			}

			// speed
			gotSpeed := readInt32(t, r)
			if gotSpeed != summon.MoveSpeed() {
				t.Errorf("speed: got %d, want %d", gotSpeed, summon.MoveSpeed())
			}

			// PAtkSpd / cast speed
			gotPAtkSpd2 := readInt32(t, r)
			if gotPAtkSpd2 != summon.AtkSpeed() {
				t.Errorf("PAtkSpd (combat): got %d, want %d", gotPAtkSpd2, summon.AtkSpeed())
			}
			gotCastSpeed := readInt32(t, r)
			if gotCastSpeed != 333 {
				t.Errorf("MAtkSpd (cast): got %d, want 333", gotCastSpeed)
			}
		})
	}
}

func TestPetInfo_ConstructorFields(t *testing.T) {
	summon := testSummon(t, 5000, 1000, model.SummonTypePet)

	t.Run("NewPetInfo defaults", func(t *testing.T) {
		pkt := NewPetInfo(summon, "Owner")

		if pkt.Summon != summon {
			t.Errorf("Summon: got %p, want %p", pkt.Summon, summon)
		}
		if pkt.OwnerName != "Owner" {
			t.Errorf("OwnerName: got %q, want %q", pkt.OwnerName, "Owner")
		}
		if pkt.CurrentFed != 0 {
			t.Errorf("CurrentFed: got %d, want 0", pkt.CurrentFed)
		}
		if pkt.MaxFed != 0 {
			t.Errorf("MaxFed: got %d, want 0", pkt.MaxFed)
		}
		if pkt.Exp != 0 {
			t.Errorf("Exp: got %d, want 0", pkt.Exp)
		}
		if pkt.ExpMax != 0 {
			t.Errorf("ExpMax: got %d, want 0", pkt.ExpMax)
		}
	})

	t.Run("NewPetInfoWithFeed", func(t *testing.T) {
		pkt := NewPetInfoWithFeed(summon, "PetOwner", 75, 100, 50000, 100000)

		if pkt.OwnerName != "PetOwner" {
			t.Errorf("OwnerName: got %q, want %q", pkt.OwnerName, "PetOwner")
		}
		if pkt.CurrentFed != 75 {
			t.Errorf("CurrentFed: got %d, want 75", pkt.CurrentFed)
		}
		if pkt.MaxFed != 100 {
			t.Errorf("MaxFed: got %d, want 100", pkt.MaxFed)
		}
		if pkt.Exp != 50000 {
			t.Errorf("Exp: got %d, want 50000", pkt.Exp)
		}
		if pkt.ExpMax != 100000 {
			t.Errorf("ExpMax: got %d, want 100000", pkt.ExpMax)
		}
	})
}

// ---------- PetStatusUpdate (0xB5) ----------

func TestPetStatusUpdate_Write(t *testing.T) {
	tests := []struct {
		name       string
		summonType model.SummonType
		objectID   uint32
		ownerID    uint32
		currentFed int32
		maxFed     int32
		exp        int64
		expMax     int64
		withFeed   bool
	}{
		{
			name:       "servitor no feed",
			summonType: model.SummonTypeServitor,
			objectID:   6000,
			ownerID:    2000,
			withFeed:   false,
		},
		{
			name:       "pet with feed",
			summonType: model.SummonTypePet,
			objectID:   5000,
			ownerID:    1000,
			currentFed: 80,
			maxFed:     100,
			exp:        55555,
			expMax:     99999,
			withFeed:   true,
		},
		{
			name:       "pet zero feed",
			summonType: model.SummonTypePet,
			objectID:   7000,
			ownerID:    3000,
			currentFed: 0,
			maxFed:     100,
			exp:        0,
			expMax:     1000,
			withFeed:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summon := testSummon(t, tc.objectID, tc.ownerID, tc.summonType)

			var pkt PetStatusUpdate
			if tc.withFeed {
				pkt = NewPetStatusUpdateWithFeed(summon, tc.currentFed, tc.maxFed, tc.exp, tc.expMax)
			} else {
				pkt = NewPetStatusUpdate(summon)
			}

			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			r := packet.NewReader(data)

			// Opcode
			opcode := readByte(t, r)
			if opcode != OpcodePetStatusUpdate {
				t.Errorf("opcode: got 0x%02X, want 0x%02X", opcode, OpcodePetStatusUpdate)
			}

			// Summon type
			gotType := readInt32(t, r)
			if gotType != int32(tc.summonType) {
				t.Errorf("summonType: got %d, want %d", gotType, int32(tc.summonType))
			}

			// ObjectID
			gotObjectID := readInt32(t, r)
			if gotObjectID != int32(tc.objectID) {
				t.Errorf("objectID: got %d, want %d", gotObjectID, int32(tc.objectID))
			}

			// Position
			gotX := readInt32(t, r)
			if gotX != 10000 {
				t.Errorf("X: got %d, want 10000", gotX)
			}
			gotY := readInt32(t, r)
			if gotY != 20000 {
				t.Errorf("Y: got %d, want 20000", gotY)
			}
			gotZ := readInt32(t, r)
			if gotZ != -3000 {
				t.Errorf("Z: got %d, want -3000", gotZ)
			}

			// Title (empty string)
			gotTitle := readString(t, r)
			if gotTitle != "" {
				t.Errorf("title: got %q, want %q", gotTitle, "")
			}

			// Feed
			gotFed := readInt32(t, r)
			gotMaxFed := readInt32(t, r)
			if tc.withFeed {
				if gotFed != tc.currentFed {
					t.Errorf("currentFed: got %d, want %d", gotFed, tc.currentFed)
				}
				if gotMaxFed != tc.maxFed {
					t.Errorf("maxFed: got %d, want %d", gotMaxFed, tc.maxFed)
				}
			} else {
				if gotFed != 0 {
					t.Errorf("currentFed (no feed): got %d, want 0", gotFed)
				}
				if gotMaxFed != 0 {
					t.Errorf("maxFed (no feed): got %d, want 0", gotMaxFed)
				}
			}

			// HP / MP
			gotCurHP := readInt32(t, r)
			if gotCurHP != summon.CurrentHP() {
				t.Errorf("currentHP: got %d, want %d", gotCurHP, summon.CurrentHP())
			}
			gotMaxHP := readInt32(t, r)
			if gotMaxHP != summon.MaxHP() {
				t.Errorf("maxHP: got %d, want %d", gotMaxHP, summon.MaxHP())
			}
			gotCurMP := readInt32(t, r)
			if gotCurMP != summon.CurrentMP() {
				t.Errorf("currentMP: got %d, want %d", gotCurMP, summon.CurrentMP())
			}
			gotMaxMP := readInt32(t, r)
			if gotMaxMP != summon.MaxMP() {
				t.Errorf("maxMP: got %d, want %d", gotMaxMP, summon.MaxMP())
			}

			// Level
			gotLevel := readInt32(t, r)
			if gotLevel != summon.Level() {
				t.Errorf("level: got %d, want %d", gotLevel, summon.Level())
			}

			// Experience
			gotExp := readInt64(t, r)
			gotExpMax := readInt64(t, r)
			if tc.withFeed {
				if gotExp != tc.exp {
					t.Errorf("exp: got %d, want %d", gotExp, tc.exp)
				}
				if gotExpMax != tc.expMax {
					t.Errorf("expMax: got %d, want %d", gotExpMax, tc.expMax)
				}
			} else {
				if gotExp != 0 {
					t.Errorf("exp (no feed): got %d, want 0", gotExp)
				}
				if gotExpMax != 0 {
					t.Errorf("expMax (no feed): got %d, want 0", gotExpMax)
				}
			}

			// reader должен быть в конце данных
			if r.Remaining() != 0 {
				t.Errorf("remaining bytes: got %d, want 0", r.Remaining())
			}
		})
	}
}

func TestPetStatusUpdate_ConstructorFields(t *testing.T) {
	summon := testSummon(t, 5000, 1000, model.SummonTypePet)

	t.Run("NewPetStatusUpdate defaults", func(t *testing.T) {
		pkt := NewPetStatusUpdate(summon)

		if pkt.Summon != summon {
			t.Errorf("Summon: got %p, want %p", pkt.Summon, summon)
		}
		if pkt.CurrentFed != 0 {
			t.Errorf("CurrentFed: got %d, want 0", pkt.CurrentFed)
		}
		if pkt.Exp != 0 {
			t.Errorf("Exp: got %d, want 0", pkt.Exp)
		}
	})

	t.Run("NewPetStatusUpdateWithFeed", func(t *testing.T) {
		pkt := NewPetStatusUpdateWithFeed(summon, 42, 100, 7777, 8888)

		if pkt.CurrentFed != 42 {
			t.Errorf("CurrentFed: got %d, want 42", pkt.CurrentFed)
		}
		if pkt.MaxFed != 100 {
			t.Errorf("MaxFed: got %d, want 100", pkt.MaxFed)
		}
		if pkt.Exp != 7777 {
			t.Errorf("Exp: got %d, want 7777", pkt.Exp)
		}
		if pkt.ExpMax != 8888 {
			t.Errorf("ExpMax: got %d, want 8888", pkt.ExpMax)
		}
	})
}

func TestPetStatusUpdate_DamagedSummon(t *testing.T) {
	summon := testSummon(t, 5000, 1000, model.SummonTypePet)
	summon.SetCurrentHP(250)
	summon.SetCurrentMP(100)

	pkt := NewPetStatusUpdateWithFeed(summon, 30, 100, 1000, 5000)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	// Пропускаем: opcode(1) + type(4) + objectID(4) + X(4) + Y(4) + Z(4)
	_ = readByte(t, r)
	for range 5 {
		_ = readInt32(t, r)
	}

	// Пропускаем title (empty string)
	_ = readString(t, r)

	// Пропускаем feed
	_ = readInt32(t, r)
	_ = readInt32(t, r)

	// HP
	gotCurHP := readInt32(t, r)
	if gotCurHP != 250 {
		t.Errorf("currentHP after damage: got %d, want 250", gotCurHP)
	}
	gotMaxHP := readInt32(t, r)
	if gotMaxHP != 500 {
		t.Errorf("maxHP: got %d, want 500", gotMaxHP)
	}

	// MP
	gotCurMP := readInt32(t, r)
	if gotCurMP != 100 {
		t.Errorf("currentMP after damage: got %d, want 100", gotCurMP)
	}
	gotMaxMP := readInt32(t, r)
	if gotMaxMP != 200 {
		t.Errorf("maxMP: got %d, want 200", gotMaxMP)
	}
}

// ---------- PetItemList (0xB2) ----------

func TestPetItemList_Write(t *testing.T) {
	tests := []struct {
		name      string
		items     []*model.Item
		wantCount int16
	}{
		{
			name:      "empty inventory",
			items:     nil,
			wantCount: 0,
		},
		{
			name: "single item",
			items: func() []*model.Item {
				tmpl := &model.ItemTemplate{
					ItemID: 57,
					Name:   "Adena",
					Type:   model.ItemTypeEtcItem,
				}
				item, err := model.NewItem(1000, 57, 1, 100, tmpl)
				if err != nil {
					t.Fatalf("NewItem: %v", err)
				}
				return []*model.Item{item}
			}(),
			wantCount: 1,
		},
		{
			name: "multiple items",
			items: func() []*model.Item {
				adena := &model.ItemTemplate{ItemID: 57, Name: "Adena", Type: model.ItemTypeEtcItem}
				sword := &model.ItemTemplate{ItemID: 1, Name: "Short Sword", Type: model.ItemTypeWeapon}
				armor := &model.ItemTemplate{ItemID: 100, Name: "Leather Shirt", Type: model.ItemTypeArmor, BodyPart: model.ArmorSlotChest}

				i1, err := model.NewItem(1000, 57, 1, 5000, adena)
				if err != nil {
					t.Fatalf("NewItem adena: %v", err)
				}
				i2, err := model.NewItem(1001, 1, 1, 1, sword)
				if err != nil {
					t.Fatalf("NewItem sword: %v", err)
				}
				i3, err := model.NewItem(1002, 100, 1, 1, armor)
				if err != nil {
					t.Fatalf("NewItem armor: %v", err)
				}
				return []*model.Item{i1, i2, i3}
			}(),
			wantCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkt := NewPetItemList(tc.items)
			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			r := packet.NewReader(data)

			// Opcode
			opcode := readByte(t, r)
			if opcode != OpcodePetItemList {
				t.Errorf("opcode: got 0x%02X, want 0x%02X", opcode, OpcodePetItemList)
			}

			// Item count (short)
			gotCount := readInt16(t, r)
			if gotCount != tc.wantCount {
				t.Errorf("itemCount: got %d, want %d", gotCount, tc.wantCount)
			}

			// Проверяем каждый item
			for i, item := range tc.items {
				// type1 (short)
				_ = readInt16(t, r)

				// objectID (int)
				gotObjID := readInt32(t, r)
				if gotObjID != int32(item.ObjectID()) {
					t.Errorf("item[%d].objectID: got %d, want %d", i, gotObjID, int32(item.ObjectID()))
				}

				// itemID (int)
				gotItemID := readInt32(t, r)
				if gotItemID != item.ItemID() {
					t.Errorf("item[%d].itemID: got %d, want %d", i, gotItemID, item.ItemID())
				}

				// count (int)
				gotItemCount := readInt32(t, r)
				if gotItemCount != item.Count() {
					t.Errorf("item[%d].count: got %d, want %d", i, gotItemCount, item.Count())
				}

				// type2 (short)
				_ = readInt16(t, r)

				// customType1 (short)
				_ = readInt16(t, r)

				// equipped (short)
				_ = readInt16(t, r)

				// bodyPart (int)
				_ = readInt32(t, r)

				// enchant (short)
				_ = readInt16(t, r)

				// customType2 (short)
				_ = readInt16(t, r)

				// augmentation (int)
				_ = readInt32(t, r)

				// mana (int)
				_ = readInt32(t, r)
			}

			if r.Remaining() != 0 {
				t.Errorf("remaining bytes: got %d, want 0", r.Remaining())
			}
		})
	}
}

func TestPetItemList_ItemSize(t *testing.T) {
	// Каждый item занимает ровно 36 байт в writeItem
	tmpl := &model.ItemTemplate{ItemID: 57, Name: "Adena", Type: model.ItemTypeEtcItem}
	item, err := model.NewItem(1000, 57, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}

	pkt := NewPetItemList([]*model.Item{item})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// opcode(1) + count(2) + 1 item(36) = 39
	wantLen := 1 + 2 + 36
	if len(data) != wantLen {
		t.Errorf("packet size for 1 item: got %d, want %d", len(data), wantLen)
	}
}

// ---------- PetDelete (0xB6) ----------

func TestPetDelete_Write(t *testing.T) {
	tests := []struct {
		name       string
		summonType int32
		objectID   uint32
	}{
		{
			name:       "pet delete",
			summonType: int32(model.SummonTypePet),
			objectID:   5000,
		},
		{
			name:       "servitor delete",
			summonType: int32(model.SummonTypeServitor),
			objectID:   6000,
		},
		{
			name:       "max object ID",
			summonType: 2,
			objectID:   0xFFFFFFFF,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkt := NewPetDelete(tc.summonType, tc.objectID)

			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			// opcode(1) + summonType(4) + objectID(4) = 9 bytes
			wantLen := 9
			if len(data) != wantLen {
				t.Fatalf("packet size: got %d, want %d", len(data), wantLen)
			}

			r := packet.NewReader(data)

			// Opcode
			opcode := readByte(t, r)
			if opcode != OpcodePetDelete {
				t.Errorf("opcode: got 0x%02X, want 0x%02X", opcode, OpcodePetDelete)
			}

			// Summon type
			gotType := readInt32(t, r)
			if gotType != tc.summonType {
				t.Errorf("summonType: got %d, want %d", gotType, tc.summonType)
			}

			// ObjectID
			gotObjectID := readInt32(t, r)
			if gotObjectID != int32(tc.objectID) {
				t.Errorf("objectID: got %d, want %d", gotObjectID, int32(tc.objectID))
			}

			if r.Remaining() != 0 {
				t.Errorf("remaining bytes: got %d, want 0", r.Remaining())
			}
		})
	}
}

func TestPetDelete_ConstructorFields(t *testing.T) {
	pkt := NewPetDelete(2, 12345)

	if pkt.SummonType != 2 {
		t.Errorf("SummonType: got %d, want 2", pkt.SummonType)
	}
	if pkt.ObjectID != 12345 {
		t.Errorf("ObjectID: got %d, want 12345", pkt.ObjectID)
	}
}

// ---------- PetStatusShow (0xB0) ----------

func TestPetStatusShow_Write(t *testing.T) {
	tests := []struct {
		name       string
		summonType int32
	}{
		{
			name:       "pet status show",
			summonType: int32(model.SummonTypePet),
		},
		{
			name:       "servitor status show",
			summonType: int32(model.SummonTypeServitor),
		},
		{
			name:       "zero type",
			summonType: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkt := NewPetStatusShow(tc.summonType)

			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			// opcode(1) + summonType(4) = 5 bytes
			wantLen := 5
			if len(data) != wantLen {
				t.Fatalf("packet size: got %d, want %d", len(data), wantLen)
			}

			r := packet.NewReader(data)

			// Opcode
			opcode := readByte(t, r)
			if opcode != OpcodePetStatusShow {
				t.Errorf("opcode: got 0x%02X, want 0x%02X", opcode, OpcodePetStatusShow)
			}

			// Summon type
			gotType := readInt32(t, r)
			if gotType != tc.summonType {
				t.Errorf("summonType: got %d, want %d", gotType, tc.summonType)
			}

			if r.Remaining() != 0 {
				t.Errorf("remaining bytes: got %d, want 0", r.Remaining())
			}
		})
	}
}

func TestPetStatusShow_ConstructorFields(t *testing.T) {
	pkt := NewPetStatusShow(2)
	if pkt.SummonType != 2 {
		t.Errorf("SummonType: got %d, want 2", pkt.SummonType)
	}
}

// ---------- Opcode constants ----------

func TestPetOpcodes(t *testing.T) {
	tests := []struct {
		name   string
		got    byte
		want   byte
	}{
		{"PetStatusShow", OpcodePetStatusShow, 0xB0},
		{"PetInfo", OpcodePetInfo, 0xB1},
		{"PetItemList", OpcodePetItemList, 0xB2},
		{"PetStatusUpdate", OpcodePetStatusUpdate, 0xB5},
		{"PetDelete", OpcodePetDelete, 0xB6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("opcode %s: got 0x%02X, want 0x%02X", tc.name, tc.got, tc.want)
			}
		})
	}
}
