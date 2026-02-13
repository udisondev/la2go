package admin

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// mockAdminCmd is a test admin command.
type mockAdminCmd struct {
	names       []string
	required    int32
	handleCalls int
	lastArgs    []string
}

func (c *mockAdminCmd) Names() []string           { return c.names }
func (c *mockAdminCmd) RequiredAccessLevel() int32 { return c.required }
func (c *mockAdminCmd) Handle(player *model.Player, args []string) error {
	c.handleCalls++
	c.lastArgs = args
	player.SetLastAdminMessage("admin ok: " + args[0])
	return nil
}

// mockUserCmd is a test user command.
type mockUserCmd struct {
	names       []string
	handleCalls int
	lastParams  string
}

func (c *mockUserCmd) Names() []string { return c.names }
func (c *mockUserCmd) Handle(player *model.Player, params string) error {
	c.handleCalls++
	c.lastParams = params
	player.SetLastAdminMessage("user ok")
	return nil
}

func newGMPlayer(t *testing.T, accessLevel int32) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(1, 100, 200, "TestGM", 80, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	p.SetAccessLevel(accessLevel)
	return p
}

func TestHandler_RegisterAndCount(t *testing.T) {
	h := NewHandler()
	if h.AdminCommandCount() != 0 {
		t.Errorf("AdminCommandCount = %d, want 0", h.AdminCommandCount())
	}
	if h.UserCommandCount() != 0 {
		t.Errorf("UserCommandCount = %d, want 0", h.UserCommandCount())
	}

	h.RegisterAdmin(&mockAdminCmd{names: []string{"test", "test2"}, required: 1})
	if h.AdminCommandCount() != 2 {
		t.Errorf("AdminCommandCount = %d, want 2 (two aliases)", h.AdminCommandCount())
	}

	h.RegisterUser(&mockUserCmd{names: []string{"cmd"}})
	if h.UserCommandCount() != 1 {
		t.Errorf("UserCommandCount = %d, want 1", h.UserCommandCount())
	}
}

func TestHandler_AdminCommand_Success(t *testing.T) {
	h := NewHandler()
	cmd := &mockAdminCmd{names: []string{"heal"}, required: 1}
	h.RegisterAdmin(cmd)

	player := newGMPlayer(t, 2) // Game Master, level 2

	ok := h.HandleAdminCommand(player, "heal 50")
	if !ok {
		t.Error("HandleAdminCommand returned false, want true")
	}
	if cmd.handleCalls != 1 {
		t.Errorf("Handle called %d times, want 1", cmd.handleCalls)
	}
	if len(cmd.lastArgs) != 2 || cmd.lastArgs[0] != "heal" || cmd.lastArgs[1] != "50" {
		t.Errorf("Handle args = %v, want [heal 50]", cmd.lastArgs)
	}
	if msg := player.ClearLastAdminMessage(); msg != "admin ok: heal" {
		t.Errorf("LastAdminMessage = %q, want %q", msg, "admin ok: heal")
	}
}

func TestHandler_AdminCommand_CaseInsensitive(t *testing.T) {
	h := NewHandler()
	cmd := &mockAdminCmd{names: []string{"teleport"}, required: 1}
	h.RegisterAdmin(cmd)

	player := newGMPlayer(t, 1)
	ok := h.HandleAdminCommand(player, "TELEPORT 0 0 0")
	if !ok {
		t.Error("HandleAdminCommand with uppercase should still find command")
	}
	if cmd.handleCalls != 1 {
		t.Errorf("Handle called %d times, want 1", cmd.handleCalls)
	}
}

func TestHandler_AdminCommand_UnknownCommand(t *testing.T) {
	h := NewHandler()
	player := newGMPlayer(t, 100)

	ok := h.HandleAdminCommand(player, "nosuchcmd")
	if ok {
		t.Error("HandleAdminCommand should return false for unknown command")
	}
	msg := player.ClearLastAdminMessage()
	if msg == "" {
		t.Error("Expected error message about unknown command")
	}
}

func TestHandler_AdminCommand_EmptyText(t *testing.T) {
	h := NewHandler()
	player := newGMPlayer(t, 100)

	ok := h.HandleAdminCommand(player, "")
	if ok {
		t.Error("HandleAdminCommand should return false for empty text")
	}
}

func TestHandler_AdminCommand_InsufficientAccess(t *testing.T) {
	h := NewHandler()
	cmd := &mockAdminCmd{names: []string{"spawn"}, required: 2}
	h.RegisterAdmin(cmd)

	player := newGMPlayer(t, 1) // Moderator, level 1 (needs 2)

	ok := h.HandleAdminCommand(player, "spawn 100")
	if ok {
		t.Error("HandleAdminCommand should return false when access level insufficient")
	}
	if cmd.handleCalls != 0 {
		t.Error("Handle should not be called when access denied")
	}
}

func TestHandler_AdminCommand_NormalPlayerDenied(t *testing.T) {
	h := NewHandler()
	cmd := &mockAdminCmd{names: []string{"kick"}, required: 1}
	h.RegisterAdmin(cmd)

	player := newGMPlayer(t, 0) // normal user

	ok := h.HandleAdminCommand(player, "kick someone")
	if ok {
		t.Error("HandleAdminCommand should return false for normal player")
	}
	if cmd.handleCalls != 0 {
		t.Error("Handle should not be called for normal player")
	}
}

func TestHandler_UserCommand_Success(t *testing.T) {
	h := NewHandler()
	cmd := &mockUserCmd{names: []string{"loc", "location"}}
	h.RegisterUser(cmd)

	player := newGMPlayer(t, 0) // normal user can use user commands

	ok := h.HandleUserCommand(player, "loc")
	if !ok {
		t.Error("HandleUserCommand returned false, want true")
	}
	if cmd.handleCalls != 1 {
		t.Errorf("Handle called %d times, want 1", cmd.handleCalls)
	}
}

func TestHandler_UserCommand_WithParams(t *testing.T) {
	h := NewHandler()
	cmd := &mockUserCmd{names: []string{"whisper"}}
	h.RegisterUser(cmd)

	player := newGMPlayer(t, 0)
	ok := h.HandleUserCommand(player, "whisper hello world")
	if !ok {
		t.Error("HandleUserCommand returned false, want true")
	}
	if cmd.lastParams != "hello world" {
		t.Errorf("params = %q, want %q", cmd.lastParams, "hello world")
	}
}

func TestHandler_UserCommand_Unknown(t *testing.T) {
	h := NewHandler()
	player := newGMPlayer(t, 0)

	ok := h.HandleUserCommand(player, "nosuchcmd")
	if ok {
		t.Error("HandleUserCommand should return false for unknown command")
	}
}

func TestHandler_UserCommand_EmptyText(t *testing.T) {
	h := NewHandler()
	player := newGMPlayer(t, 0)

	ok := h.HandleUserCommand(player, "")
	if ok {
		t.Error("HandleUserCommand should return false for empty text")
	}
}
