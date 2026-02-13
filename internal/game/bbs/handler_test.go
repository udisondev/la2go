package bbs

import (
	"strings"
	"testing"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler()

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if !h.Enabled() {
		t.Error("should be enabled by default")
	}
}

func TestHandler_SetEnabled(t *testing.T) {
	h := NewHandler()

	h.SetEnabled(false)
	if h.Enabled() {
		t.Error("should be disabled")
	}

	h.SetEnabled(true)
	if !h.Enabled() {
		t.Error("should be enabled")
	}
}

func TestHandler_HandleCommand_Home(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbshome", 1, "Alice")
	if html == "" {
		t.Error("_bbshome should return HTML")
	}
	if !strings.Contains(html, "Alice") {
		t.Error("home page should contain player name")
	}
	if !strings.Contains(html, "Community Board") {
		t.Error("home page should contain title")
	}
}

func TestHandler_HandleCommand_Top(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbstop", 1, "Bob")
	if html == "" {
		t.Error("_bbstop should return HTML")
	}
	if !strings.Contains(html, "Bob") {
		t.Error("_bbstop should contain player name")
	}
}

func TestHandler_HandleCommand_Region(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsloc", 1, "Alice")
	if html == "" {
		t.Error("_bbsloc should return HTML")
	}
	if !strings.Contains(html, "Region") {
		t.Error("region page should contain Region title")
	}
	if !strings.Contains(html, "Gludio") {
		t.Error("region page should contain Gludio")
	}
}

func TestHandler_HandleCommand_RegionDetail(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsloc;3", 1, "Alice")
	if html == "" {
		t.Error("_bbsloc;3 should return HTML")
	}
	if !strings.Contains(html, "#3") {
		t.Error("region detail should contain region ID")
	}
}

func TestHandler_HandleCommand_Clan(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsclan", 1, "Alice")
	if html == "" {
		t.Error("_bbsclan should return HTML")
	}
	if !strings.Contains(html, "Clan") {
		t.Error("clan page should contain Clan title")
	}
}

func TestHandler_HandleCommand_Memo(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsmemo", 1, "Alice")
	if html == "" {
		t.Error("_bbsmemo should return HTML")
	}
	if !strings.Contains(html, "Memo") {
		t.Error("memo page should contain Memo title")
	}
}

func TestHandler_HandleCommand_Mail(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsmail", 1, "Alice")
	if html == "" {
		t.Error("_bbsmail should return HTML")
	}
	if !strings.Contains(html, "Mailbox") {
		t.Error("mail page should contain Mailbox title")
	}
}

func TestHandler_HandleCommand_Friends(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsfriends", 1, "Alice")
	if html == "" {
		t.Error("_bbsfriends should return HTML")
	}
	if !strings.Contains(html, "Friends") {
		t.Error("friends page should contain Friends title")
	}
}

func TestHandler_HandleCommand_Favorites(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsgetfav", 1, "Alice")
	if html == "" {
		t.Error("_bbsgetfav should return HTML")
	}
	if !strings.Contains(html, "Favorites") {
		t.Error("favorites page should contain Favorites title")
	}
}

func TestHandler_HandleCommand_AddFavorite(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("bbs_add_fav", 1, "Alice")
	if html == "" {
		t.Error("bbs_add_fav should return HTML")
	}
	if !strings.Contains(html, "Added") {
		t.Error("add favorite should contain confirmation")
	}
}

func TestHandler_HandleCommand_Unknown(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbsunknown", 1, "Alice")
	if html != "" {
		t.Errorf("unknown command should return empty; got %q", html)
	}
}

func TestHandler_HandleCommand_Disabled(t *testing.T) {
	h := NewHandler()
	h.SetEnabled(false)

	html := h.HandleCommand("_bbshome", 1, "Alice")
	if html != "" {
		t.Error("disabled handler should return empty")
	}
}

func TestHandler_IsBoardCommand(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		cmd  string
		want bool
	}{
		{"_bbshome", true},
		{"_bbstop", true},
		{"_bbsloc", true},
		{"_bbsloc;5", true},
		{"_bbsclan", true},
		{"_bbsmemo", true},
		{"_bbsmail", true},
		{"_bbsfriends", true},
		{"_bbsgetfav", true},
		{"bbs_add_fav", true},
		{"_bbsunknown", false},
		{"npc_123_Shop", false},
		{"", false},
	}

	for _, tt := range tests {
		got := h.IsBoardCommand(tt.cmd)
		if got != tt.want {
			t.Errorf("IsBoardCommand(%q) = %v; want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestHandler_HandleWrite(t *testing.T) {
	h := NewHandler()

	// Не все доски поддерживают Write — проверяем что не паникует
	html := h.HandleWrite(1, "Alice", "Topic", [5]string{"a", "b", "c", "d", "e"})
	// HomeBoard не реализует WriteBoard — пустой ответ
	if html != "" {
		t.Errorf("HandleWrite(Topic) should return empty for non-WriteBoard; got %q", html)
	}
}

func TestHandler_HandleWrite_Disabled(t *testing.T) {
	h := NewHandler()
	h.SetEnabled(false)

	html := h.HandleWrite(1, "Alice", "Topic", [5]string{})
	if html != "" {
		t.Error("disabled handler should return empty")
	}
}

func TestHandler_HandleWrite_UnknownURL(t *testing.T) {
	h := NewHandler()

	html := h.HandleWrite(1, "Alice", "Unknown", [5]string{})
	if html != "" {
		t.Error("unknown URL should return empty")
	}
}

func TestHandler_Register_Custom(t *testing.T) {
	h := NewHandler()

	custom := &testBoard{cmds: []string{"_bbstest"}, html: "<html>test</html>"}
	h.Register(custom)

	html := h.HandleCommand("_bbstest", 1, "Alice")
	if html != "<html>test</html>" {
		t.Errorf("custom board returned %q; want <html>test</html>", html)
	}
}

func TestHandler_TopicPage(t *testing.T) {
	h := NewHandler()

	html := h.HandleCommand("_bbstop;home.html", 1, "Alice")
	if html == "" {
		t.Error("_bbstop;home.html should return HTML")
	}
	if !strings.Contains(html, "home.html") {
		t.Error("topic page should contain page name")
	}
}

func TestHandler_TopicPage_NoSuffix(t *testing.T) {
	h := NewHandler()

	// Без .html суффикса — должен вернуть home page
	html := h.HandleCommand("_bbstop;nope", 1, "Alice")
	if html == "" {
		t.Error("should return home page as fallback")
	}
	if !strings.Contains(html, "Community Board") {
		t.Error("should fallback to home")
	}
}

// --- SplitHTML tests ---

func TestSplitHTML_Empty(t *testing.T) {
	chunks := SplitHTML("")

	if len(chunks) != 3 {
		t.Fatalf("SplitHTML(\"\") chunks = %d; want 3", len(chunks))
	}
	for i, c := range chunks {
		if c.Content != "" {
			t.Errorf("chunk[%d] content should be empty", i)
		}
	}
	if chunks[0].ID != "101" || chunks[1].ID != "102" || chunks[2].ID != "103" {
		t.Error("chunk IDs should be 101, 102, 103")
	}
}

func TestSplitHTML_Short(t *testing.T) {
	html := "<html>short</html>"
	chunks := SplitHTML(html)

	if len(chunks) != 3 {
		t.Fatalf("chunks count = %d; want 3", len(chunks))
	}
	if chunks[0].Content != html {
		t.Errorf("chunk[0] = %q; want %q", chunks[0].Content, html)
	}
	if chunks[1].Content != "" {
		t.Error("chunk[1] should be empty")
	}
	if chunks[2].Content != "" {
		t.Error("chunk[2] should be empty")
	}
}

func TestSplitHTML_TwoParts(t *testing.T) {
	html := strings.Repeat("A", MaxChunkSize+100)
	chunks := SplitHTML(html)

	if len(chunks) != 3 {
		t.Fatalf("chunks count = %d; want 3", len(chunks))
	}
	if len(chunks[0].Content) != MaxChunkSize {
		t.Errorf("chunk[0] len = %d; want %d", len(chunks[0].Content), MaxChunkSize)
	}
	if len(chunks[1].Content) != 100 {
		t.Errorf("chunk[1] len = %d; want 100", len(chunks[1].Content))
	}
	if chunks[2].Content != "" {
		t.Error("chunk[2] should be empty")
	}
}

func TestSplitHTML_ThreeParts(t *testing.T) {
	html := strings.Repeat("B", MaxChunkSize*2+500)
	chunks := SplitHTML(html)

	if len(chunks[0].Content) != MaxChunkSize {
		t.Errorf("chunk[0] len = %d; want %d", len(chunks[0].Content), MaxChunkSize)
	}
	if len(chunks[1].Content) != MaxChunkSize {
		t.Errorf("chunk[1] len = %d; want %d", len(chunks[1].Content), MaxChunkSize)
	}
	if len(chunks[2].Content) != 500 {
		t.Errorf("chunk[2] len = %d; want 500", len(chunks[2].Content))
	}
}

func TestSplitHTML_Overflow(t *testing.T) {
	html := strings.Repeat("C", MaxHTMLSize+1000)
	chunks := SplitHTML(html)

	total := 0
	for _, c := range chunks {
		total += len(c.Content)
	}
	if total > MaxHTMLSize {
		t.Errorf("total content size = %d; should be <= %d", total, MaxHTMLSize)
	}
}

func TestFormatContent(t *testing.T) {
	content := FormatContent("101", "<html>hello</html>")
	if content != "101\x08<html>hello</html>" {
		t.Errorf("FormatContent() = %q; want %q", content, "101\x08<html>hello</html>")
	}
}

func TestMapWriteURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"Topic", "_bbstop"},
		{"Post", "_bbspos"},
		{"Region", "_bbsloc"},
		{"Notice", "_bbsclan"},
		{"Mail", "_bbsmail"},
		{"Unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := mapWriteURL(tt.url)
		if got != tt.want {
			t.Errorf("mapWriteURL(%q) = %q; want %q", tt.url, got, tt.want)
		}
	}
}

func TestNavigationButtons(t *testing.T) {
	if len(NavigationButtons) != 8 {
		t.Fatalf("NavigationButtons count = %d; want 8", len(NavigationButtons))
	}

	if NavigationButtons[0] != "bypass _bbshome" {
		t.Errorf("NavigationButtons[0] = %q; want %q", NavigationButtons[0], "bypass _bbshome")
	}
	if NavigationButtons[7] != "bypass bbs_add_fav" {
		t.Errorf("NavigationButtons[7] = %q; want %q", NavigationButtons[7], "bypass bbs_add_fav")
	}
}

func TestConstants(t *testing.T) {
	if MaxChunkSize != 4090 {
		t.Errorf("MaxChunkSize = %d; want 4090", MaxChunkSize)
	}
	if MaxChunks != 3 {
		t.Errorf("MaxChunks = %d; want 3", MaxChunks)
	}
	if MaxHTMLSize != 12270 {
		t.Errorf("MaxHTMLSize = %d; want 12270", MaxHTMLSize)
	}
	if DefaultCommand != "_bbshome" {
		t.Errorf("DefaultCommand = %q; want %q", DefaultCommand, "_bbshome")
	}
	if ContentSeparator != "\x08" {
		t.Errorf("ContentSeparator = %q; want \\x08", ContentSeparator)
	}
}

// --- test helpers ---

type testBoard struct {
	cmds []string
	html string
}

func (b *testBoard) Commands() []string {
	return b.cmds
}

func (b *testBoard) OnCommand(_ string, _ int64, _ string) string {
	return b.html
}
