package bbs

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

const (
	// ContentSeparator разделяет ID и HTML-контент в ShowBoard пакете.
	// Java: "\u0008" (ASCII Backspace).
	ContentSeparator = "\x08"

	// MaxChunkSize — максимальный размер одной части HTML (байт).
	// Java: HtmlUtil.sendCBHtml — 4090 bytes per chunk.
	MaxChunkSize = 4090

	// MaxChunks — максимальное количество частей.
	MaxChunks = 3

	// MaxHTMLSize — максимальный общий размер HTML.
	MaxHTMLSize = MaxChunkSize * MaxChunks // 12270

	// DefaultCommand — bypass-команда по умолчанию при ALT+B.
	DefaultCommand = "_bbshome"
)

// NavigationButtons — 8 фиксированных bypass-кнопок верхней панели Community Board.
// Java: ShowBoard.java — hardcoded в пакете.
var NavigationButtons = [8]string{
	"bypass _bbshome",
	"bypass _bbsgetfav",
	"bypass _bbsloc",
	"bypass _bbsclan",
	"bypass _bbsmemo",
	"bypass _bbsmail",
	"bypass _bbsfriends",
	"bypass bbs_add_fav",
}

// Board handles a set of community board bypass commands.
type Board interface {
	// OnCommand processes a bypass command.
	// Returns HTML to send, or empty string if handled internally.
	OnCommand(cmd string, charID int64, charName string) string

	// Commands returns the list of bypass prefixes this board handles.
	Commands() []string
}

// WriteBoard extends Board with form-write support (RequestBBSwrite).
type WriteBoard interface {
	Board
	// OnWrite processes a BBS write request.
	OnWrite(charID int64, charName string, url string, args [5]string) string
}

// Handler dispatches community board commands to registered boards.
// Thread-safe via sync.RWMutex.
//
// Java reference: CommunityBoardHandler.java
type Handler struct {
	mu      sync.RWMutex
	boards  map[string]Board // prefix → Board
	enabled bool
}

// NewHandler creates a new community board handler.
func NewHandler() *Handler {
	h := &Handler{
		boards:  make(map[string]Board),
		enabled: true,
	}

	// Регистрация стандартных досок
	h.Register(&HomeBoard{})
	h.Register(&RegionBoard{})
	h.Register(&ClanBoard{})
	h.Register(&MemoBoard{})
	h.Register(&MailBoard{})
	h.Register(&FriendsBoard{})
	h.Register(&FavoriteBoard{})

	return h
}

// Register adds a board handler for its command prefixes.
func (h *Handler) Register(b Board) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, cmd := range b.Commands() {
		h.boards[cmd] = b
	}
}

// Enabled reports whether the community board is enabled.
func (h *Handler) Enabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

// SetEnabled enables or disables the community board.
func (h *Handler) SetEnabled(v bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = v
}

// HandleCommand processes a community board bypass command.
// Returns HTML content to send to the client (empty if board not found or disabled).
func (h *Handler) HandleCommand(cmd string, charID int64, charName string) string {
	if !h.Enabled() {
		return ""
	}

	b := h.findBoard(cmd)
	if b == nil {
		slog.Debug("community board: unknown command", "cmd", cmd)
		return ""
	}

	return b.OnCommand(cmd, charID, charName)
}

// HandleWrite processes a BBS write request (RequestBBSwrite).
// Returns HTML content to send (empty if not handled).
func (h *Handler) HandleWrite(charID int64, charName string, url string, args [5]string) string {
	if !h.Enabled() {
		return ""
	}

	// Маппинг URL → bypass-команда (Java: CommunityBoardHandler.handleWriteCommand)
	cmd := mapWriteURL(url)
	if cmd == "" {
		slog.Debug("community board: unknown write URL", "url", url)
		return ""
	}

	b := h.findBoard(cmd)
	if b == nil {
		return ""
	}

	wb, ok := b.(WriteBoard)
	if !ok {
		return ""
	}

	return wb.OnWrite(charID, charName, url, args)
}

// IsBoardCommand reports whether the command is a community board bypass.
func (h *Handler) IsBoardCommand(cmd string) bool {
	return h.findBoard(cmd) != nil
}

// findBoard returns the board that handles the given command.
func (h *Handler) findBoard(cmd string) Board {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Точное совпадение
	if b, ok := h.boards[cmd]; ok {
		return b
	}

	// Поиск по префиксу (например "_bbsloc;5" → "_bbsloc")
	for prefix, b := range h.boards {
		if strings.HasPrefix(cmd, prefix) {
			return b
		}
	}

	return nil
}

// SplitHTML разбивает HTML на 3 части для отправки через ShowBoard.
// Каждая часть получает ID: "101", "102", "103".
// Returns slice of (id, content) pairs.
//
// Java: HtmlUtil.sendCBHtml — 3 chunks × 4090 bytes.
func SplitHTML(html string) []HTMLChunk {
	if len(html) == 0 {
		return []HTMLChunk{
			{ID: "101", Content: ""},
			{ID: "102", Content: ""},
			{ID: "103", Content: ""},
		}
	}

	if len(html) > MaxHTMLSize {
		slog.Warn("community board: HTML too long, truncating",
			"size", len(html),
			"max", MaxHTMLSize)
		html = html[:MaxHTMLSize]
	}

	chunks := make([]HTMLChunk, MaxChunks)
	for i := range MaxChunks {
		id := fmt.Sprintf("10%d", i+1)
		start := i * MaxChunkSize
		if start >= len(html) {
			chunks[i] = HTMLChunk{ID: id, Content: ""}
			continue
		}
		end := min(start+MaxChunkSize, len(html))
		chunks[i] = HTMLChunk{ID: id, Content: html[start:end]}
	}

	return chunks
}

// HTMLChunk represents a piece of HTML content with its ShowBoard ID.
type HTMLChunk struct {
	ID      string // "101", "102", "103"
	Content string
}

// FormatContent formats content for the ShowBoard packet.
// Java: _content = id + "\u0008" + htmlCode
func FormatContent(id, html string) string {
	return id + ContentSeparator + html
}

// mapWriteURL maps RequestBBSwrite URL to a bypass command.
// Java: CommunityBoardHandler.handleWriteCommand switch.
func mapWriteURL(url string) string {
	switch url {
	case "Topic":
		return "_bbstop"
	case "Post":
		return "_bbspos"
	case "Region":
		return "_bbsloc"
	case "Notice":
		return "_bbsclan"
	case "Mail":
		return "_bbsmail"
	default:
		return ""
	}
}
