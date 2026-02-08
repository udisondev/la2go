package clientpackets

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestParseAuthLogin(t *testing.T) {
	accountName := "testuser"

	// Encode account name as UTF-16LE null-terminated
	runes := []rune(accountName)
	utf16Encoded := utf16.Encode(runes)
	nameBytes := make([]byte, (len(utf16Encoded)+1)*2) // +1 for null terminator
	for i, r := range utf16Encoded {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], r)
	}
	// null terminator already zero

	// SessionKey (4×int32)
	playOkID1 := int32(0x11111111)
	playOkID2 := int32(0x22222222)
	loginOkID1 := int32(0x33333333)
	loginOkID2 := int32(0x44444444)

	sessionKeyBytes := make([]byte, 16)
	binary.LittleEndian.PutUint32(sessionKeyBytes[0:], uint32(playOkID1))
	binary.LittleEndian.PutUint32(sessionKeyBytes[4:], uint32(playOkID2))
	binary.LittleEndian.PutUint32(sessionKeyBytes[8:], uint32(loginOkID1))
	binary.LittleEndian.PutUint32(sessionKeyBytes[12:], uint32(loginOkID2))

	// 4 unknown int32 fields (all zeros)
	unknownBytes := make([]byte, 16)

	// Concatenate
	data := append(nameBytes, sessionKeyBytes...)
	data = append(data, unknownBytes...)

	pkt, err := ParseAuthLogin(data)
	if err != nil {
		t.Fatalf("ParseAuthLogin failed: %v", err)
	}

	if pkt.AccountName != accountName {
		t.Errorf("expected account name %q, got %q", accountName, pkt.AccountName)
	}

	if pkt.SessionKey.PlayOkID1 != playOkID1 {
		t.Errorf("expected PlayOkID1 0x%08X, got 0x%08X", playOkID1, pkt.SessionKey.PlayOkID1)
	}

	if pkt.SessionKey.PlayOkID2 != playOkID2 {
		t.Errorf("expected PlayOkID2 0x%08X, got 0x%08X", playOkID2, pkt.SessionKey.PlayOkID2)
	}

	if pkt.SessionKey.LoginOkID1 != loginOkID1 {
		t.Errorf("expected LoginOkID1 0x%08X, got 0x%08X", loginOkID1, pkt.SessionKey.LoginOkID1)
	}

	if pkt.SessionKey.LoginOkID2 != loginOkID2 {
		t.Errorf("expected LoginOkID2 0x%08X, got 0x%08X", loginOkID2, pkt.SessionKey.LoginOkID2)
	}
}

func TestParseAuthLogin_EmptyAccountName(t *testing.T) {
	// Empty account name (only null terminator)
	nameBytes := []byte{0x00, 0x00}

	// SessionKey
	sessionKeyBytes := make([]byte, 16)
	binary.LittleEndian.PutUint32(sessionKeyBytes[0:], 0x11111111)
	binary.LittleEndian.PutUint32(sessionKeyBytes[4:], 0x22222222)
	binary.LittleEndian.PutUint32(sessionKeyBytes[8:], 0x33333333)
	binary.LittleEndian.PutUint32(sessionKeyBytes[12:], 0x44444444)

	// Unknown fields
	unknownBytes := make([]byte, 16)

	data := append(nameBytes, sessionKeyBytes...)
	data = append(data, unknownBytes...)

	pkt, err := ParseAuthLogin(data)
	if err != nil {
		t.Fatalf("ParseAuthLogin failed: %v", err)
	}

	if pkt.AccountName != "" {
		t.Errorf("expected empty account name, got %q", pkt.AccountName)
	}
}

func TestParseAuthLogin_NotEnoughData(t *testing.T) {
	// Incomplete packet (only account name, no SessionKey)
	data := []byte{0x74, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00, 0x00, 0x00} // "test" + null

	_, err := ParseAuthLogin(data)
	if err == nil {
		t.Error("expected error when parsing incomplete AuthLogin packet")
	}
}
