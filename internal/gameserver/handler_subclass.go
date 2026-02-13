package gameserver

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleNpcSubclass processes VillageMaster subclass bypass commands.
//
// Bypass format: "Subclass <action> [args...]"
//   - action 1: list available subclasses to add
//   - action 2: show modify subclass selection
//   - action 3: show change active class selection
//   - action 4 <classId> <classIndex>: add subclass
//   - action 5 <newClassId> <classIndex>: modify subclass
//   - action 6 <classIndex>: switch active class
//   - action 7 <classIndex>: show available classes for modify slot
//
// Phase 14: Subclass System.
// Java reference: VillageMaster.java
func (h *Handler) handleNpcSubclass(
	player *model.Player,
	npc *model.Npc,
	cmdArg string,
	buf []byte,
) (int, bool, error) {
	parts := strings.Fields(cmdArg)
	if len(parts) == 0 {
		return sendActionFailed(buf), true, nil
	}

	action, err := strconv.Atoi(parts[0])
	if err != nil {
		slog.Debug("invalid subclass action", "arg", cmdArg)
		return sendActionFailed(buf), true, nil
	}

	npcObjID := int32(npc.ObjectID())

	switch action {
	case 1:
		return h.subclassListAdd(player, npcObjID, buf)
	case 2:
		return h.subclassListModify(player, npcObjID, buf)
	case 3:
		return h.subclassListSwitch(player, npcObjID, buf)
	case 4:
		if len(parts) < 3 {
			return sendActionFailed(buf), true, nil
		}
		return h.subclassAdd(player, npcObjID, parts[1], parts[2], buf)
	case 5:
		if len(parts) < 3 {
			return sendActionFailed(buf), true, nil
		}
		return h.subclassModify(player, npcObjID, parts[1], parts[2], buf)
	case 6:
		if len(parts) < 2 {
			return sendActionFailed(buf), true, nil
		}
		return h.subclassSwitch(player, npcObjID, parts[1], buf)
	case 7:
		if len(parts) < 2 {
			return sendActionFailed(buf), true, nil
		}
		return h.subclassModifySlot(player, npcObjID, parts[1], buf)
	default:
		slog.Debug("unknown subclass action", "action", action)
		return sendActionFailed(buf), true, nil
	}
}

// subclassListAdd shows available subclasses for adding (action 1).
func (h *Handler) subclassListAdd(
	player *model.Player,
	npcObjID int32,
	buf []byte,
) (int, bool, error) {
	if !data.IsSubclassEligible(player.BaseClassID()) {
		return sendNpcHtml(npcObjID, subclassNotEligibleHTML, buf)
	}

	if player.SubClassCount() >= data.MaxSubclasses {
		return sendNpcHtml(npcObjID, subclassMaxReachedHTML, buf)
	}

	available := data.GetAvailableSubClasses(
		player.BaseClassID(),
		player.RaceID(),
		player.ExistingSubClassIDs(),
	)

	if len(available) == 0 {
		return sendNpcHtml(npcObjID, subclassNoneAvailableHTML, buf)
	}

	// Find first empty slot
	var slotIndex int32
	for i := int32(1); i <= int32(data.MaxSubclasses); i++ {
		if player.GetSubClass(i) == nil {
			slotIndex = i
			break
		}
	}

	htmlContent := buildSubclassListHTML("Add Subclass", available, 4, slotIndex)
	return sendNpcHtml(npcObjID, htmlContent, buf)
}

// subclassListModify shows existing subclasses available for modification (action 2).
func (h *Handler) subclassListModify(
	player *model.Player,
	npcObjID int32,
	buf []byte,
) (int, bool, error) {
	subs := player.SubClasses()
	if len(subs) == 0 {
		return sendNpcHtml(npcObjID, subclassNoSubsHTML, buf)
	}

	var sb strings.Builder
	sb.WriteString("<html><body>Modify Subclass<br><br>")
	sb.WriteString("Select the subclass slot to modify:<br>")

	for idx, sub := range subs {
		info := data.GetClassInfo(sub.ClassID)
		name := fmt.Sprintf("ClassID %d", sub.ClassID)
		if info != nil {
			name = info.Name
		}
		sb.WriteString(fmt.Sprintf(
			"<a action=\"bypass -h npc_%d_Subclass 7 %d\">%s (Lv.%d)</a><br>",
			npcObjID, idx, name, sub.Level,
		))
	}

	sb.WriteString("</body></html>")
	return sendNpcHtml(npcObjID, sb.String(), buf)
}

// subclassListSwitch shows existing subclasses for active class switching (action 3).
func (h *Handler) subclassListSwitch(
	player *model.Player,
	npcObjID int32,
	buf []byte,
) (int, bool, error) {
	subs := player.SubClasses()

	var sb strings.Builder
	sb.WriteString("<html><body>Change Active Class<br><br>")
	sb.WriteString("Select your class:<br>")

	// Base class option
	if player.IsSubClassActive() {
		baseInfo := data.GetClassInfo(player.BaseClassID())
		baseName := "Base Class"
		if baseInfo != nil {
			baseName = baseInfo.Name
		}
		sb.WriteString(fmt.Sprintf(
			"<a action=\"bypass -h npc_%d_Subclass 6 0\">%s (Base)</a><br>",
			npcObjID, baseName,
		))
	}

	// Subclass options
	activeIdx := player.ActiveClassIndex()
	for idx, sub := range subs {
		if idx == activeIdx {
			continue // skip currently active
		}
		info := data.GetClassInfo(sub.ClassID)
		name := fmt.Sprintf("ClassID %d", sub.ClassID)
		if info != nil {
			name = info.Name
		}
		sb.WriteString(fmt.Sprintf(
			"<a action=\"bypass -h npc_%d_Subclass 6 %d\">%s (Lv.%d)</a><br>",
			npcObjID, idx, name, sub.Level,
		))
	}

	sb.WriteString("</body></html>")
	return sendNpcHtml(npcObjID, sb.String(), buf)
}

// subclassAdd confirms adding a new subclass (action 4).
func (h *Handler) subclassAdd(
	player *model.Player,
	npcObjID int32,
	classIDStr, classIndexStr string,
	buf []byte,
) (int, bool, error) {
	classID, err := strconv.ParseInt(classIDStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}
	classIndex, err := strconv.ParseInt(classIndexStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}

	sub, addErr := player.AddSubClass(int32(classID), int32(classIndex))
	if addErr != nil {
		slog.Info("subclass add rejected",
			"character", player.Name(),
			"classID", classID,
			"classIndex", classIndex,
			"error", addErr)
		htmlContent := fmt.Sprintf(
			"<html><body>Subclass Error<br><br>%s</body></html>",
			addErr.Error(),
		)
		return sendNpcHtml(npcObjID, htmlContent, buf)
	}

	info := data.GetClassInfo(sub.ClassID)
	name := fmt.Sprintf("ClassID %d", sub.ClassID)
	if info != nil {
		name = info.Name
	}

	slog.Info("subclass added",
		"character", player.Name(),
		"classID", sub.ClassID,
		"className", name,
		"classIndex", sub.ClassIndex)

	htmlContent := fmt.Sprintf(
		"<html><body>Subclass Added<br><br>%s has been added as subclass #%d at level %d.</body></html>",
		name, sub.ClassIndex, sub.Level,
	)
	return sendNpcHtml(npcObjID, htmlContent, buf)
}

// subclassModifySlot shows available classes for a specific modify slot (action 7).
func (h *Handler) subclassModifySlot(
	player *model.Player,
	npcObjID int32,
	classIndexStr string,
	buf []byte,
) (int, bool, error) {
	classIndex, err := strconv.ParseInt(classIndexStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}

	sub := player.GetSubClass(int32(classIndex))
	if sub == nil {
		return sendNpcHtml(npcObjID, subclassNoSubsHTML, buf)
	}

	available := data.GetAvailableSubClasses(
		player.BaseClassID(),
		player.RaceID(),
		player.ExistingSubClassIDs(),
	)

	if len(available) == 0 {
		return sendNpcHtml(npcObjID, subclassNoneAvailableHTML, buf)
	}

	htmlContent := buildSubclassListHTML("Modify Subclass", available, 5, int32(classIndex))
	return sendNpcHtml(npcObjID, htmlContent, buf)
}

// subclassModify confirms modifying an existing subclass (action 5).
func (h *Handler) subclassModify(
	player *model.Player,
	npcObjID int32,
	classIDStr, classIndexStr string,
	buf []byte,
) (int, bool, error) {
	classID, err := strconv.ParseInt(classIDStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}
	classIndex, err := strconv.ParseInt(classIndexStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}

	sub, modErr := player.ModifySubClass(int32(classIndex), int32(classID))
	if modErr != nil {
		slog.Info("subclass modify rejected",
			"character", player.Name(),
			"classID", classID,
			"classIndex", classIndex,
			"error", modErr)
		htmlContent := fmt.Sprintf(
			"<html><body>Subclass Error<br><br>%s</body></html>",
			modErr.Error(),
		)
		return sendNpcHtml(npcObjID, htmlContent, buf)
	}

	info := data.GetClassInfo(sub.ClassID)
	name := fmt.Sprintf("ClassID %d", sub.ClassID)
	if info != nil {
		name = info.Name
	}

	slog.Info("subclass modified",
		"character", player.Name(),
		"classID", sub.ClassID,
		"className", name,
		"classIndex", sub.ClassIndex)

	htmlContent := fmt.Sprintf(
		"<html><body>Subclass Modified<br><br>Slot #%d replaced with %s at level %d.</body></html>",
		sub.ClassIndex, name, sub.Level,
	)
	return sendNpcHtml(npcObjID, htmlContent, buf)
}

// subclassSwitch switches the player's active class (action 6).
func (h *Handler) subclassSwitch(
	player *model.Player,
	npcObjID int32,
	classIndexStr string,
	buf []byte,
) (int, bool, error) {
	classIndex, err := strconv.ParseInt(classIndexStr, 10, 32)
	if err != nil {
		return sendActionFailed(buf), true, nil
	}

	if switchErr := player.SetActiveClass(int32(classIndex)); switchErr != nil {
		slog.Info("subclass switch rejected",
			"character", player.Name(),
			"classIndex", classIndex,
			"error", switchErr)
		htmlContent := fmt.Sprintf(
			"<html><body>Class Switch Error<br><br>%s</body></html>",
			switchErr.Error(),
		)
		return sendNpcHtml(npcObjID, htmlContent, buf)
	}

	slog.Info("subclass switched",
		"character", player.Name(),
		"classIndex", classIndex,
		"classID", player.ClassID())

	// Broadcast CharInfo to nearby players (new appearance)
	charInfo := serverpackets.NewCharInfo(player)
	charInfoData, charInfoErr := charInfo.Write()
	if charInfoErr == nil {
		h.clientManager.BroadcastToVisibleNear(player, charInfoData, len(charInfoData))
	}

	// Send UserInfo to the player (refresh stats/class)
	userInfo := serverpackets.NewUserInfo(player)
	pktData, pktErr := userInfo.Write()
	if pktErr != nil {
		return 0, false, fmt.Errorf("serializing UserInfo after subclass switch: %w", pktErr)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// sendActionFailed writes ActionFailed packet to buf and returns bytes written.
func sendActionFailed(buf []byte) int {
	af := serverpackets.NewActionFailed()
	afData, _ := af.Write()
	return copy(buf, afData)
}

// sendNpcHtml writes NpcHtmlMessage to buf and returns handler result.
func sendNpcHtml(npcObjID int32, htmlContent string, buf []byte) (int, bool, error) {
	pkt := serverpackets.NewNpcHtmlMessage(npcObjID, htmlContent)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing NpcHtmlMessage: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// buildSubclassListHTML generates an HTML list of available subclasses
// with bypass links using the specified action code and class index.
func buildSubclassListHTML(title string, classIDs []int32, action int, classIndex int32) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<html><body>%s<br><br>", title))
	sb.WriteString("Available classes:<br>")

	for _, id := range classIDs {
		info := data.GetClassInfo(id)
		name := fmt.Sprintf("ClassID %d", id)
		if info != nil {
			name = info.Name
		}
		// bypass uses %%objectId%% — NPC object ID placeholder resolved by client
		sb.WriteString(fmt.Sprintf(
			"<a action=\"bypass -h npc_%%objectId%%_Subclass %d %d %d\">%s</a><br>",
			action, id, classIndex, name,
		))
	}

	sb.WriteString("</body></html>")
	return sb.String()
}

// HTML constants for subclass dialogs.
const (
	subclassNotEligibleHTML  = "<html><body>Subclass Error<br><br>You must reach 2nd class transfer before using the subclass system.</body></html>"
	subclassMaxReachedHTML   = "<html><body>Subclass Error<br><br>You already have the maximum number of subclasses.</body></html>"
	subclassNoneAvailableHTML = "<html><body>Subclass Error<br><br>There are no subclasses available for you.</body></html>"
	subclassNoSubsHTML       = "<html><body>Subclass Error<br><br>You have no subclasses to manage.</body></html>"
)
