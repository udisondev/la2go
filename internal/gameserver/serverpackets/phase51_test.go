package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// --- AskJoinAlly ---

func TestAskJoinAlly_Write(t *testing.T) {
	t.Parallel()

	pkt := &AskJoinAlly{
		RequestorObjectID: 42,
		AllyName:          "TestAlly",
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(data) < 5 {
		t.Fatalf("data too short: %d bytes", len(data))
	}
	if data[0] != OpcodeAskJoinAlly {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAskJoinAlly)
	}

	objID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objID != 42 {
		t.Errorf("RequestorObjectID = %d; want 42", objID)
	}
}

func TestAskJoinAlly_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	pkt := &AskJoinAlly{
		RequestorObjectID: 99999,
		AllyName:          "Alliance",
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Пропускаем opcode (1 байт), парсим данные через Reader
	r := packet.NewReader(data[1:])

	objID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt objectID: %v", err)
	}
	if objID != 99999 {
		t.Errorf("objectID = %d; want 99999", objID)
	}

	allyName, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString allyName: %v", err)
	}
	if allyName != "Alliance" {
		t.Errorf("allyName = %q; want %q", allyName, "Alliance")
	}
}

func TestAskJoinAlly_Write_EmptyName(t *testing.T) {
	t.Parallel()

	pkt := &AskJoinAlly{
		RequestorObjectID: 1,
		AllyName:          "",
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeAskJoinAlly {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAskJoinAlly)
	}
	// opcode(1) + int32(4) + null-terminator UTF-16LE(2) = 7 bytes minimum
	if len(data) < 7 {
		t.Errorf("data length = %d; want >= 7", len(data))
	}
}

// --- AllianceInfo ---

func TestAllianceInfo_Write(t *testing.T) {
	t.Parallel()

	pkt := &AllianceInfo{
		AllyName:       "TestAlly",
		TotalMembers:   50,
		OnlineMembers:  30,
		LeaderClanName: "LeaderClan",
		LeaderName:     "LeaderPlayer",
		Clans: []AllianceClanInfo{
			{
				ClanName:          "Clan1",
				ClanLevel:         5,
				ClanLeaderName:    "Leader1",
				ClanTotalMembers:  25,
				ClanOnlineMembers: 15,
			},
			{
				ClanName:          "Clan2",
				ClanLevel:         4,
				ClanLeaderName:    "Leader2",
				ClanTotalMembers:  25,
				ClanOnlineMembers: 15,
			},
		},
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(data) < 10 {
		t.Fatalf("data too short: %d bytes", len(data))
	}
	if data[0] != OpcodeAllianceInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAllianceInfo)
	}
}

func TestAllianceInfo_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	pkt := &AllianceInfo{
		AllyName:       "MyAlly",
		TotalMembers:   100,
		OnlineMembers:  42,
		LeaderClanName: "BossClan",
		LeaderName:     "BossMan",
		Clans: []AllianceClanInfo{
			{
				ClanName:          "First",
				ClanLevel:         8,
				ClanLeaderName:    "ClanBoss",
				ClanTotalMembers:  40,
				ClanOnlineMembers: 20,
			},
		},
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data[1:]) // пропускаем opcode

	allyName, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString allyName: %v", err)
	}
	if allyName != "MyAlly" {
		t.Errorf("allyName = %q; want %q", allyName, "MyAlly")
	}

	totalMembers, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt totalMembers: %v", err)
	}
	if totalMembers != 100 {
		t.Errorf("totalMembers = %d; want 100", totalMembers)
	}

	onlineMembers, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt onlineMembers: %v", err)
	}
	if onlineMembers != 42 {
		t.Errorf("onlineMembers = %d; want 42", onlineMembers)
	}

	leaderClan, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString leaderClan: %v", err)
	}
	if leaderClan != "BossClan" {
		t.Errorf("leaderClan = %q; want %q", leaderClan, "BossClan")
	}

	leaderName, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString leaderName: %v", err)
	}
	if leaderName != "BossMan" {
		t.Errorf("leaderName = %q; want %q", leaderName, "BossMan")
	}

	clanCount, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt clanCount: %v", err)
	}
	if clanCount != 1 {
		t.Fatalf("clanCount = %d; want 1", clanCount)
	}

	// Первый (и единственный) клан
	clanName, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString clanName: %v", err)
	}
	if clanName != "First" {
		t.Errorf("clanName = %q; want %q", clanName, "First")
	}

	reserved, err := r.ReadInt() // unknown/reserved field
	if err != nil {
		t.Fatalf("ReadInt reserved: %v", err)
	}
	if reserved != 0 {
		t.Errorf("reserved = %d; want 0", reserved)
	}

	clanLevel, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt clanLevel: %v", err)
	}
	if clanLevel != 8 {
		t.Errorf("clanLevel = %d; want 8", clanLevel)
	}

	clanLeaderName, err := r.ReadString()
	if err != nil {
		t.Fatalf("ReadString clanLeaderName: %v", err)
	}
	if clanLeaderName != "ClanBoss" {
		t.Errorf("clanLeaderName = %q; want %q", clanLeaderName, "ClanBoss")
	}

	clanTotal, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt clanTotalMembers: %v", err)
	}
	if clanTotal != 40 {
		t.Errorf("clanTotalMembers = %d; want 40", clanTotal)
	}

	clanOnline, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt clanOnlineMembers: %v", err)
	}
	if clanOnline != 20 {
		t.Errorf("clanOnlineMembers = %d; want 20", clanOnline)
	}
}

func TestAllianceInfo_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &AllianceInfo{
		AllyName:       "Solo",
		TotalMembers:   10,
		OnlineMembers:  5,
		LeaderClanName: "SoloClan",
		LeaderName:     "SoloLeader",
		Clans:          nil,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if data[0] != OpcodeAllianceInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAllianceInfo)
	}

	// Проверяем, что clanCount == 0
	r := packet.NewReader(data[1:])
	if _, err := r.ReadString(); err != nil { // allyName
		t.Fatalf("ReadString allyName: %v", err)
	}
	if _, err := r.ReadInt(); err != nil { // totalMembers
		t.Fatalf("ReadInt totalMembers: %v", err)
	}
	if _, err := r.ReadInt(); err != nil { // onlineMembers
		t.Fatalf("ReadInt onlineMembers: %v", err)
	}
	if _, err := r.ReadString(); err != nil { // leaderClanName
		t.Fatalf("ReadString leaderClanName: %v", err)
	}
	if _, err := r.ReadString(); err != nil { // leaderName
		t.Fatalf("ReadString leaderName: %v", err)
	}

	clanCount, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt clanCount: %v", err)
	}
	if clanCount != 0 {
		t.Errorf("clanCount = %d; want 0 for empty Clans", clanCount)
	}
}

// --- AllyCrest (round-trip) ---

func TestAllyCrest_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	pkt := &AllyCrest{
		CrestID: 42,
		Data:    []byte{0x01, 0x02, 0x03},
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// opcode(1) + crestID(4) + length(4) + data(3) = 12
	if len(data) != 12 {
		t.Fatalf("len(data) = %d; want 12", len(data))
	}
	if data[0] != OpcodeAllyCrest {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeAllyCrest)
	}

	r := packet.NewReader(data[1:])

	crestID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt crestID: %v", err)
	}
	if crestID != 42 {
		t.Errorf("crestID = %d; want 42", crestID)
	}

	length, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt length: %v", err)
	}
	if length != 3 {
		t.Errorf("length = %d; want 3", length)
	}

	crestData, err := r.ReadBytes(int(length))
	if err != nil {
		t.Fatalf("ReadBytes: %v", err)
	}
	if len(crestData) != 3 {
		t.Fatalf("len(crestData) = %d; want 3", len(crestData))
	}
	if crestData[0] != 0x01 || crestData[1] != 0x02 || crestData[2] != 0x03 {
		t.Errorf("crestData = %x; want [01 02 03]", crestData)
	}
}

func TestAllyCrest_Write_EmptyData(t *testing.T) {
	t.Parallel()

	pkt := &AllyCrest{
		CrestID: 0,
		Data:    nil,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// opcode(1) + crestID(4) + length(4) = 9 bytes (no data)
	if len(data) != 9 {
		t.Fatalf("len(data) = %d; want 9", len(data))
	}

	r := packet.NewReader(data[1:])

	crestID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt crestID: %v", err)
	}
	if crestID != 0 {
		t.Errorf("crestID = %d; want 0", crestID)
	}

	length, err := r.ReadInt()
	if err != nil {
		t.Fatalf("ReadInt length: %v", err)
	}
	if length != 0 {
		t.Errorf("length = %d; want 0", length)
	}
}

// --- Opcode Constants ---

func TestAllianceServerOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  byte
		want byte
	}{
		{"OpcodeAskJoinAlly", OpcodeAskJoinAlly, 0xA8},
		{"OpcodeAllianceInfo", OpcodeAllianceInfo, 0xB4},
		{"OpcodeAllyCrest", OpcodeAllyCrest, 0xAE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = 0x%02X; want 0x%02X", tt.name, tt.got, tt.want)
			}
		})
	}
}
