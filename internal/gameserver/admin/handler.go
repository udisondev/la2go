package admin

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// Command is the interface for admin commands (//command).
// Each command registers one or more names and a required access level.
//
// Java reference: IAdminCommandHandler.java
type Command interface {
	// Handle executes the command. args includes command name at [0].
	Handle(player *model.Player, args []string) error
	// Names returns all registered command names (without // prefix).
	Names() []string
	// RequiredAccessLevel returns the minimum access level to use this command.
	RequiredAccessLevel() int32
}

// UserCommand is the interface for user commands (/command).
// Available to all players (no access level check).
//
// Java reference: IVoicedCommandHandler.java
type UserCommand interface {
	// Handle executes the user command. params is the rest of the message after command name.
	Handle(player *model.Player, params string) error
	// Names returns all registered command names (without / prefix).
	Names() []string
}

// Handler dispatches admin (//) and user (/) commands.
// Thread-safe: commands are registered once at startup, then read-only.
//
// Java reference: AdminCommandHandler.java, VoicedCommandHandler.java
type Handler struct {
	mu           sync.RWMutex
	adminCmds    map[string]Command     // name → Command (lowercase)
	userCmds     map[string]UserCommand // name → UserCommand (lowercase)
}

// NewHandler creates a new admin/user command handler.
func NewHandler() *Handler {
	return &Handler{
		adminCmds: make(map[string]Command, 32),
		userCmds:  make(map[string]UserCommand, 8),
	}
}

// RegisterAdmin registers an admin command.
// All command names are lowercased for case-insensitive lookup.
func (h *Handler) RegisterAdmin(cmd Command) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, name := range cmd.Names() {
		h.adminCmds[strings.ToLower(name)] = cmd
	}
}

// RegisterUser registers a user command.
func (h *Handler) RegisterUser(cmd UserCommand) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, name := range cmd.Names() {
		h.userCmds[strings.ToLower(name)] = cmd
	}
}

// HandleAdminCommand processes a message starting with //.
// Returns true if a command was found and executed.
// text is the full message WITHOUT the // prefix.
//
// Java reference: AdminCommandHandler.onCommand()
func (h *Handler) HandleAdminCommand(player *model.Player, text string) bool {
	if text == "" {
		return false
	}

	parts := strings.Fields(text)
	cmdName := strings.ToLower(parts[0])

	h.mu.RLock()
	cmd, ok := h.adminCmds[cmdName]
	h.mu.RUnlock()

	if !ok {
		sendMessage(player, "Unknown command: //"+cmdName)
		return false
	}

	// Access level check
	accessLevel := player.AccessLevel()
	al := GetAccessLevel(accessLevel)
	if al == nil || !al.CanUseAdminCommands {
		slog.Warn("unauthorized admin command attempt",
			"player", player.Name(),
			"command", cmdName,
			"accessLevel", accessLevel)
		return false
	}

	if accessLevel < cmd.RequiredAccessLevel() {
		sendMessage(player, fmt.Sprintf("Insufficient access level for //%s (need %d, have %d)",
			cmdName, cmd.RequiredAccessLevel(), accessLevel))
		slog.Warn("admin command access denied",
			"player", player.Name(),
			"command", cmdName,
			"required", cmd.RequiredAccessLevel(),
			"actual", accessLevel)
		return false
	}

	slog.Info("admin command",
		"player", player.Name(),
		"command", text)

	if err := cmd.Handle(player, parts); err != nil {
		sendMessage(player, fmt.Sprintf("Command error: %s", err))
		slog.Error("admin command failed",
			"player", player.Name(),
			"command", text,
			"error", err)
	}

	return true
}

// HandleUserCommand processes a message starting with /.
// Returns true if a command was found and executed.
// text is the full message WITHOUT the / prefix.
//
// Java reference: ChatGeneral — voiced command handling
func (h *Handler) HandleUserCommand(player *model.Player, text string) bool {
	if text == "" {
		return false
	}

	parts := strings.Fields(text)
	cmdName := strings.ToLower(parts[0])

	h.mu.RLock()
	cmd, ok := h.userCmds[cmdName]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	// Extract params (everything after command name)
	var params string
	if len(text) > len(parts[0]) {
		params = strings.TrimSpace(text[len(parts[0]):])
	}

	if err := cmd.Handle(player, params); err != nil {
		sendMessage(player, fmt.Sprintf("Command error: %s", err))
		slog.Error("user command failed",
			"player", player.Name(),
			"command", text,
			"error", err)
	}

	return true
}

// sendMessage is a placeholder for sending a message to player.
// In production this would send a CreatureSay or SystemMessage packet.
// The actual packet sending is handled by the gameserver handler that
// calls admin.Handler, since admin package doesn't have access to
// GameClient or packet serialization.
//
// Phase 17: We store the message on player using GM message mechanism.
func sendMessage(player *model.Player, msg string) {
	player.SetLastAdminMessage(msg)
}

// AdminCommandCount returns number of registered admin commands.
func (h *Handler) AdminCommandCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.adminCmds)
}

// UserCommandCount returns number of registered user commands.
func (h *Handler) UserCommandCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userCmds)
}
