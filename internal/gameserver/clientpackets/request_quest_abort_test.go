package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseRequestQuestAbort(t *testing.T) {
	// Формируем пакет: questID = 303 (4 bytes LE)
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(303))

	pkt, err := ParseRequestQuestAbort(data)
	if err != nil {
		t.Fatalf("ParseRequestQuestAbort error: %v", err)
	}

	if pkt.QuestID != 303 {
		t.Errorf("QuestID = %d, want 303", pkt.QuestID)
	}
}

func TestParseRequestQuestAbort_EmptyData(t *testing.T) {
	_, err := ParseRequestQuestAbort(nil)
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}

func TestParseRequestQuestAbort_LargeID(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 0xFFFFFFFF)

	pkt, err := ParseRequestQuestAbort(data)
	if err != nil {
		t.Fatalf("ParseRequestQuestAbort error: %v", err)
	}

	// 0xFFFFFFFF as int32 = -1
	if pkt.QuestID != -1 {
		t.Errorf("QuestID = %d, want -1", pkt.QuestID)
	}
}
