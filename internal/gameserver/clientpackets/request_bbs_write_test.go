package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestOpcodeRequestBBSwrite(t *testing.T) {
	if OpcodeRequestBBSwrite != 0x22 {
		t.Errorf("OpcodeRequestBBSwrite = 0x%02X; want 0x22", OpcodeRequestBBSwrite)
	}
}

func TestParseRequestBBSwrite(t *testing.T) {
	// Создаём пакет с 6 строками (url + 5 args)
	w := packet.NewWriter(256)
	w.WriteString("Topic")
	w.WriteString("arg1_value")
	w.WriteString("arg2_value")
	w.WriteString("arg3_value")
	w.WriteString("arg4_value")
	w.WriteString("arg5_value")

	pkt, err := ParseRequestBBSwrite(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestBBSwrite: %v", err)
	}

	if pkt.URL != "Topic" {
		t.Errorf("URL = %q; want %q", pkt.URL, "Topic")
	}
	if pkt.Args[0] != "arg1_value" {
		t.Errorf("Args[0] = %q; want %q", pkt.Args[0], "arg1_value")
	}
	if pkt.Args[4] != "arg5_value" {
		t.Errorf("Args[4] = %q; want %q", pkt.Args[4], "arg5_value")
	}
}

func TestParseRequestBBSwrite_EmptyArgs(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("Mail")
	for range 5 {
		w.WriteString("")
	}

	pkt, err := ParseRequestBBSwrite(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestBBSwrite: %v", err)
	}

	if pkt.URL != "Mail" {
		t.Errorf("URL = %q; want %q", pkt.URL, "Mail")
	}
	for i, arg := range pkt.Args {
		if arg != "" {
			t.Errorf("Args[%d] = %q; want empty", i, arg)
		}
	}
}
