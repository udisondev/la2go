package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseRequestConfirmTargetItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		wantObjID int32
		wantErr  bool
	}{
		{"valid", makeInt32Bytes(42), 42, false},
		{"zero", makeInt32Bytes(0), 0, false},
		{"too short", []byte{1, 2}, 0, true},
		{"empty", nil, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestConfirmTargetItem(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseRequestConfirmTargetItem() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRequestConfirmTargetItem() error = %v", err)
			}
			if pkt.ObjectID != tt.wantObjID {
				t.Errorf("ObjectID = %d, want %d", pkt.ObjectID, tt.wantObjID)
			}
		})
	}
}

func TestParseRequestConfirmRefinerItem(t *testing.T) {
	t.Parallel()

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[:4], 10)
	binary.LittleEndian.PutUint32(data[4:8], 20)

	pkt, err := ParseRequestConfirmRefinerItem(data)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if pkt.TargetObjectID != 10 {
		t.Errorf("TargetObjectID = %d, want 10", pkt.TargetObjectID)
	}
	if pkt.RefinerObjectID != 20 {
		t.Errorf("RefinerObjectID = %d, want 20", pkt.RefinerObjectID)
	}
}

func TestParseRequestConfirmRefinerItem_TooShort(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestConfirmRefinerItem([]byte{1, 2, 3})
	if err == nil {
		t.Error("ParseRequestConfirmRefinerItem(short) = nil, want error")
	}
}

func TestParseRequestRefine(t *testing.T) {
	t.Parallel()

	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[:4], 100)
	binary.LittleEndian.PutUint32(data[4:8], 200)
	binary.LittleEndian.PutUint32(data[8:12], 300)
	binary.LittleEndian.PutUint64(data[12:20], 25)

	pkt, err := ParseRequestRefine(data)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if pkt.TargetObjectID != 100 {
		t.Errorf("TargetObjectID = %d, want 100", pkt.TargetObjectID)
	}
	if pkt.RefinerObjectID != 200 {
		t.Errorf("RefinerObjectID = %d, want 200", pkt.RefinerObjectID)
	}
	if pkt.GemStoneObjectID != 300 {
		t.Errorf("GemStoneObjectID = %d, want 300", pkt.GemStoneObjectID)
	}
	if pkt.GemStoneCount != 25 {
		t.Errorf("GemStoneCount = %d, want 25", pkt.GemStoneCount)
	}
}

func TestParseRequestRefine_TooShort(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestRefine([]byte{1, 2, 3, 4, 5})
	if err == nil {
		t.Error("ParseRequestRefine(short) = nil, want error")
	}
}

func TestParseRequestRefineCancel(t *testing.T) {
	t.Parallel()

	pkt, err := ParseRequestRefineCancel(makeInt32Bytes(777))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if pkt.ObjectID != 777 {
		t.Errorf("ObjectID = %d, want 777", pkt.ObjectID)
	}
}

func TestParseRequestConfirmCancelItem(t *testing.T) {
	t.Parallel()

	pkt, err := ParseRequestConfirmCancelItem(makeInt32Bytes(555))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if pkt.ObjectID != 555 {
		t.Errorf("ObjectID = %d, want 555", pkt.ObjectID)
	}
}

func TestParseRequestConfirmGemStone(t *testing.T) {
	t.Parallel()

	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[:4], 1)
	binary.LittleEndian.PutUint32(data[4:8], 2)
	binary.LittleEndian.PutUint32(data[8:12], 3)
	binary.LittleEndian.PutUint64(data[12:20], 50)

	pkt, err := ParseRequestConfirmGemStone(data)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if pkt.TargetObjectID != 1 {
		t.Errorf("TargetObjectID = %d, want 1", pkt.TargetObjectID)
	}
	if pkt.RefinerObjectID != 2 {
		t.Errorf("RefinerObjectID = %d, want 2", pkt.RefinerObjectID)
	}
	if pkt.GemStoneObjectID != 3 {
		t.Errorf("GemStoneObjectID = %d, want 3", pkt.GemStoneObjectID)
	}
	if pkt.GemStoneCount != 50 {
		t.Errorf("GemStoneCount = %d, want 50", pkt.GemStoneCount)
	}
}

func TestSubOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int16
		want int16
	}{
		{"ConfirmTargetItem", SubOpcodeRequestConfirmTargetItem, 0x29},
		{"ConfirmRefinerItem", SubOpcodeRequestConfirmRefinerItem, 0x2A},
		{"ConfirmGemStone", SubOpcodeRequestConfirmGemStone, 0x2B},
		{"Refine", SubOpcodeRequestRefine, 0x2C},
		{"ConfirmCancelItem", SubOpcodeRequestConfirmCancelItem, 0x2D},
		{"RefineCancel", SubOpcodeRequestRefineCancel, 0x2E},
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

func makeInt32Bytes(val int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(val))
	return b
}
