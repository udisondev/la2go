package gameserver

import (
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/instance"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleInstanceBypass processes instance-related NPC bypass commands.
// Returns (n, handled, err):
//   - n: bytes written to buf
//   - handled: true if the command was recognized
//   - err: any error during processing
//
// Bypass commands:
//   - EnterInstance <templateID>  — create and enter a new instance
//   - LeaveInstance               — leave current instance
//
// Phase 26: Instance Zones.
func (h *Handler) handleInstanceBypass(client *GameClient, cmdName string, args []string, buf []byte) (int, bool, error) {
	if h.instanceManager == nil {
		return 0, false, nil
	}

	switch cmdName {
	case "EnterInstance":
		return h.bypassEnterInstance(client, args, buf)
	case "LeaveInstance":
		return h.bypassLeaveInstance(client, buf)
	default:
		return 0, false, nil
	}
}

// bypassEnterInstance creates a new instance (or joins existing) and teleports the player.
func (h *Handler) bypassEnterInstance(client *GameClient, args []string, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if len(args) == 0 {
		return 0, true, nil
	}

	var templateID int32
	if _, err := fmt.Sscan(args[0], &templateID); err != nil {
		slog.Debug("instance bypass: invalid template ID", "arg", args[0])
		return 0, true, nil
	}

	tmpl := h.instanceManager.Template(templateID)
	if tmpl == nil {
		slog.Debug("instance bypass: template not found", "templateID", templateID)
		return 0, true, nil
	}

	// Создаём инстанс.
	inst, err := h.instanceManager.CreateInstance(templateID, player.ObjectID())
	if err != nil {
		slog.Warn("create instance",
			"templateID", templateID,
			"player", player.Name(),
			"error", err)
		return 0, true, nil
	}

	// Входим в инстанс.
	if err := h.instanceManager.EnterInstance(inst.ID(), player.ObjectID(), player.CharacterID(), player.Level()); err != nil {
		slog.Warn("enter instance",
			"instanceID", inst.ID(),
			"player", player.Name(),
			"error", err)
		return 0, true, nil
	}

	// Телепортируем на точку входа.
	n := h.teleportToInstance(client, tmpl, buf)
	return n, true, nil
}

// bypassLeaveInstance removes the player from their current instance.
func (h *Handler) bypassLeaveInstance(client *GameClient, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	inst, err := h.instanceManager.ExitInstance(player.ObjectID(), player.CharacterID())
	if err != nil {
		if err == instance.ErrNotInInstance {
			return 0, true, nil
		}
		return 0, true, fmt.Errorf("exit instance: %w", err)
	}

	// Телепортируем на точку выхода.
	tmpl := h.instanceManager.Template(inst.TemplateID())
	if tmpl == nil {
		return 0, true, nil
	}

	n := h.teleportFromInstance(client, tmpl, buf)
	return n, true, nil
}

// teleportToInstance sends a TeleportToLocation packet to the instance spawn point.
func (h *Handler) teleportToInstance(client *GameClient, tmpl *instance.Template, buf []byte) int {
	player := client.ActivePlayer()
	if player == nil {
		return 0
	}

	// Обновляем позицию игрока.
	player.WorldObject.SetLocation(
		model.NewLocation(tmpl.SpawnX, tmpl.SpawnY, tmpl.SpawnZ, 0),
	)

	pkt := serverpackets.NewTeleportToLocation(
		int32(player.ObjectID()),
		tmpl.SpawnX,
		tmpl.SpawnY,
		tmpl.SpawnZ,
	)
	pktData, err := pkt.Write()
	if err != nil {
		slog.Error("write TeleportToLocation for instance", "error", err)
		return 0
	}
	return copy(buf, pktData)
}

// teleportFromInstance sends a TeleportToLocation packet to the instance exit point.
func (h *Handler) teleportFromInstance(client *GameClient, tmpl *instance.Template, buf []byte) int {
	player := client.ActivePlayer()
	if player == nil {
		return 0
	}

	player.WorldObject.SetLocation(
		model.NewLocation(tmpl.ExitX, tmpl.ExitY, tmpl.ExitZ, 0),
	)

	pkt := serverpackets.NewTeleportToLocation(
		int32(player.ObjectID()),
		tmpl.ExitX,
		tmpl.ExitY,
		tmpl.ExitZ,
	)
	pktData, err := pkt.Write()
	if err != nil {
		slog.Error("write TeleportToLocation for instance exit", "error", err)
		return 0
	}
	return copy(buf, pktData)
}
