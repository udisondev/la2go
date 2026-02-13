package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/skill"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/html"
	"github.com/udisondev/la2go/internal/model"
)

// resolveSkillLearn looks up a SkillLearn entry from the appropriate skill tree.
func resolveSkillLearn(classID, skillID, skillLevel int32, skillType clientpackets.AcquireSkillType) *data.SkillLearn {
	switch skillType {
	case clientpackets.AcquireSkillTypeClass:
		return data.GetSkillLearn(classID, skillID, skillLevel)
	case clientpackets.AcquireSkillTypeFishing:
		return data.GetSpecialSkillLearn("fishingSkillTree", skillID, skillLevel)
	case clientpackets.AcquireSkillTypePledge:
		return data.GetSpecialSkillLearn("pledgeSkillTree", skillID, skillLevel)
	default:
		return nil
	}
}

// handleRequestAcquireSkillInfo processes 0x6B — skill info request before learning.
func (h *Handler) handleRequestAcquireSkillInfo(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAcquireSkillInfo(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAcquireSkillInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for acquire skill info")
	}

	if pkt.SkillID <= 0 || pkt.Level <= 0 {
		return 0, true, nil
	}

	// Find skill learn entry from appropriate tree
	sl := resolveSkillLearn(player.ClassID(), pkt.SkillID, pkt.Level, pkt.SkillType)
	if sl == nil {
		slog.Warn("acquire skill info: skill not in tree",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"level", pkt.Level,
			"type", pkt.SkillType)
		return 0, true, nil
	}

	// Build response
	resp := &serverpackets.AcquireSkillInfo{
		SkillID:   pkt.SkillID,
		Level:     pkt.Level,
		SpCost:    int32(sl.SpCost),
		SkillType: int32(pkt.SkillType),
	}

	// Add item requirements
	for _, item := range sl.Items {
		resp.Reqs = append(resp.Reqs, serverpackets.AcquireSkillReq{
			Type:   99, // item requirement
			ItemID: item.ItemID,
			Count:  int64(item.Count),
			Unk:    50, // Java constant
		})
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing AcquireSkillInfo: %w", err)
	}
	n := copy(buf, respData)

	slog.Debug("sent acquire skill info",
		"player", player.Name(),
		"skillID", pkt.SkillID,
		"level", pkt.Level,
		"spCost", sl.SpCost)

	return n, true, nil
}

// handleRequestAcquireSkill processes 0x6C — learn skill request.
func (h *Handler) handleRequestAcquireSkill(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAcquireSkill(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAcquireSkill: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for acquire skill")
	}

	if pkt.SkillID <= 0 || pkt.Level <= 0 {
		return 0, true, nil
	}

	// Find skill learn entry from appropriate tree
	sl := resolveSkillLearn(player.ClassID(), pkt.SkillID, pkt.Level, pkt.SkillType)
	if sl == nil {
		slog.Warn("acquire skill: skill not in tree",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"level", pkt.Level,
			"type", pkt.SkillType)
		return 0, true, nil
	}

	// Level check
	if player.Level() < sl.MinLevel {
		slog.Warn("acquire skill: level too low",
			"player", player.Name(),
			"playerLevel", player.Level(),
			"minLevel", sl.MinLevel)
		return 0, true, nil
	}

	// Skill level check: must have previous level learned (or 0 for new skill)
	currentLevel := player.GetSkillLevel(pkt.SkillID)
	if pkt.Level > 1 && currentLevel != pkt.Level-1 {
		slog.Warn("acquire skill: wrong current level",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"current", currentLevel,
			"requested", pkt.Level)
		return 0, true, nil
	}

	// SP check
	if player.SP() < sl.SpCost {
		slog.Warn("acquire skill: insufficient SP",
			"player", player.Name(),
			"have", player.SP(),
			"need", sl.SpCost)
		return 0, true, nil
	}

	// Item check + consume
	inv := player.Inventory()
	for _, item := range sl.Items {
		found := false
		for _, invItem := range inv.GetItems() {
			if invItem.ItemID() == item.ItemID && int64(invItem.Count()) >= int64(item.Count) {
				found = true
				break
			}
		}
		if !found {
			slog.Warn("acquire skill: missing item",
				"player", player.Name(),
				"itemID", item.ItemID,
				"count", item.Count)
			return 0, true, nil
		}
	}

	// Consume items
	for _, item := range sl.Items {
		inv.RemoveItemsByID(item.ItemID, int64(item.Count))
	}

	// Consume SP
	player.SetSP(player.SP() - sl.SpCost)

	// Learn skill
	isPassive := false
	if tmpl := data.GetSkillTemplate(pkt.SkillID, pkt.Level); tmpl != nil {
		isPassive = tmpl.IsPassive()
	}
	player.AddSkill(pkt.SkillID, pkt.Level, isPassive)

	slog.Info("player learned skill",
		"player", player.Name(),
		"skillID", pkt.SkillID,
		"level", pkt.Level,
		"spCost", sl.SpCost,
		"type", pkt.SkillType)

	// Send updated skill list
	skillList := serverpackets.NewSkillList(player.Skills())
	slData, err := skillList.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SkillList: %w", err)
	}
	n := copy(buf, slData)

	// Send updated acquire skill list (so player can learn next skill)
	acquireList := buildAcquireSkillListByType(player, pkt.SkillType)
	acquireData, err := acquireList.Write()
	if err != nil {
		slog.Error("serializing AcquireSkillList", "error", err)
	} else {
		n2 := copy(buf[n:], acquireData)
		n += n2
	}

	return n, true, nil
}

// handleDlgAnswer processes 0xC5 — dialog confirmation answer.
func (h *Handler) handleDlgAnswer(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseDlgAnswer(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing DlgAnswer: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for DlgAnswer")
	}

	slog.Debug("dlg answer received",
		"player", player.Name(),
		"messageID", pkt.MessageID,
		"answer", pkt.Answer,
		"requesterID", pkt.RequesterID)

	// For now, log and acknowledge. Specific dialog handlers can be added later
	// as systems that use DlgAnswer are implemented (resurrection, teleport confirm, etc.)

	return 0, true, nil
}

// buildAcquireSkillList creates the AcquireSkillList packet for CLASS skills.
func buildAcquireSkillList(player *model.Player) *serverpackets.AcquireSkillList {
	return buildAcquireSkillListByType(player, clientpackets.AcquireSkillTypeClass)
}

// buildAcquireSkillListByType creates the AcquireSkillList packet for the given skill type.
func buildAcquireSkillListByType(player *model.Player, skillType clientpackets.AcquireSkillType) *serverpackets.AcquireSkillList {
	// Build known skill levels map
	knownSkills := make(map[int32]int32)
	for _, si := range player.Skills() {
		knownSkills[si.SkillID] = si.Level
	}

	var learnable []*data.SkillLearn
	switch skillType {
	case clientpackets.AcquireSkillTypeFishing:
		learnable = data.GetLearnableSpecialSkills("fishingSkillTree", player.Level(), knownSkills)
	case clientpackets.AcquireSkillTypePledge:
		learnable = data.GetLearnableSpecialSkills("pledgeSkillTree", player.Level(), knownSkills)
	default:
		learnable = data.GetLearnableSkills(player.ClassID(), player.Level(), knownSkills)
	}

	entries := make([]serverpackets.AcquireSkillEntry, 0, len(learnable))
	for _, sl := range learnable {
		entries = append(entries, serverpackets.AcquireSkillEntry{
			SkillID:      sl.SkillID,
			NextLevel:    sl.SkillLevel,
			MaxLevel:     sl.SkillLevel,
			SpCost:       int32(sl.SpCost),
			Requirements: 0,
		})
	}

	return &serverpackets.AcquireSkillList{
		SkillType: int32(skillType),
		Skills:    entries,
	}
}

// handleNpcSkillList handles "SkillList" NPC bypass — sends AcquireSkillList.
func (h *Handler) handleNpcSkillList(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	pkt := buildAcquireSkillList(player)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing AcquireSkillList: %w", err)
	}

	if len(pkt.Skills) == 0 {
		slog.Debug("no learnable skills",
			"player", player.Name(),
			"classID", player.ClassID(),
			"level", player.Level())
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// defaultRespawnLoc is Talking Island Village.
var defaultRespawnLoc = model.NewLocation(-84318, 244579, -3730, 0)

// handleRequestRestartPoint processes 0x6D — respawn after death.
func (h *Handler) handleRequestRestartPoint(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestRestartPoint(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestRestartPoint: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for restart point")
	}

	if !player.IsDead() {
		slog.Warn("restart point: player is not dead", "player", player.Name())
		return 0, true, nil
	}

	// Determine respawn location
	var loc model.Location
	switch pkt.PointType {
	case clientpackets.RestartPointClanHall:
		loc = h.getClanHallRespawn(player)
	case clientpackets.RestartPointCastle:
		loc = h.getCastleRespawn(player)
	case clientpackets.RestartPointSiegeHQ:
		loc = h.getSiegeHQRespawn(player)
	case clientpackets.RestartPointFixed:
		// Fixed resurrection handled separately (scroll)
		loc = defaultRespawnLoc
	default:
		// Village (0) — default
		loc = defaultRespawnLoc
	}

	// Revive: restore HP/MP/CP
	player.SetCurrentHP(player.MaxHP())
	player.SetCurrentMP(player.MaxMP())
	player.SetCurrentCP(player.MaxCP())
	player.ResetDeathOnce()

	// Teleport to respawn location
	player.SetLocation(loc)

	// Send Revive packet
	revivePkt := &serverpackets.Revive{ObjectID: int32(player.ObjectID())}
	reviveData, err := revivePkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing Revive: %w", err)
	}
	n := copy(buf, reviveData)

	// Send TeleportToLocation
	telePkt := serverpackets.NewTeleportToLocation(int32(player.ObjectID()), loc.X, loc.Y, loc.Z)
	teleData, err := telePkt.Write()
	if err != nil {
		slog.Error("serializing TeleportToLocation", "error", err)
	} else {
		n2 := copy(buf[n:], teleData)
		n += n2
	}

	// Broadcast CharInfo to visible players
	charInfo := serverpackets.NewCharInfo(player)
	charData, err := charInfo.Write()
	if err != nil {
		slog.Error("serializing CharInfo for respawn broadcast", "error", err)
	} else {
		h.clientManager.BroadcastFromPosition(loc.X, loc.Y, charData, len(charData))
	}

	slog.Info("player respawned",
		"player", player.Name(),
		"pointType", pkt.PointType,
		"x", loc.X, "y", loc.Y, "z", loc.Z)

	return n, true, nil
}

// getClanHallRespawn returns the respawn location for a player's clan hall.
// Falls back to defaultRespawnLoc if the player has no clan, the clan has no hall,
// or the hall zone has no spawn points.
func (h *Handler) getClanHallRespawn(player *model.Player) model.Location {
	if h.hallTable == nil {
		return defaultRespawnLoc
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return defaultRespawnLoc
	}

	ch := h.hallTable.HallByOwner(clanID)
	if ch == nil {
		return defaultRespawnLoc
	}

	loc, ok := findZoneSpawn("ClanHallZone", "clanHallId", ch.ID())
	if !ok {
		slog.Warn("clan hall zone spawn not found",
			"hallID", ch.ID(), "hall", ch.Name(), "clanID", clanID)
		return defaultRespawnLoc
	}

	return loc
}

// getCastleRespawn returns the respawn location for a player's castle.
// Falls back to defaultRespawnLoc if the player's clan does not own a castle
// or the castle zone has no spawn points.
func (h *Handler) getCastleRespawn(player *model.Player) model.Location {
	if h.siegeManager == nil {
		return defaultRespawnLoc
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return defaultRespawnLoc
	}

	// Поиск замка, принадлежащего клану игрока.
	var castleID int32
	for _, c := range h.siegeManager.Castles() {
		if c.OwnerClanID() == clanID {
			castleID = c.ID()
			break
		}
	}

	if castleID == 0 {
		return defaultRespawnLoc
	}

	loc, ok := findZoneSpawn("CastleZone", "castleId", castleID)
	if !ok {
		slog.Warn("castle zone spawn not found",
			"castleID", castleID, "clanID", clanID)
		return defaultRespawnLoc
	}

	return loc
}

// getSiegeHQRespawn returns the respawn location at a siege headquarters flag.
// In Interlude, siege HQ is an NPC flag placed by the attacking clan.
// Since the flag spawn system is not yet implemented, this falls back to
// defaultRespawnLoc with a log message.
func (h *Handler) getSiegeHQRespawn(player *model.Player) model.Location {
	if h.siegeManager == nil {
		return defaultRespawnLoc
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return defaultRespawnLoc
	}

	// Проверяем, участвует ли клан в активной осаде как атакующий.
	castleID, registered := h.siegeManager.IsClanRegistered(clanID)
	if !registered {
		return defaultRespawnLoc
	}

	castle := h.siegeManager.Castle(castleID)
	if castle == nil {
		return defaultRespawnLoc
	}

	siege := castle.Siege()
	if siege == nil || !siege.IsInProgress() || !siege.IsAttacker(clanID) {
		return defaultRespawnLoc
	}

	// Флаговая система (siege HQ NPC) ещё не реализована,
	// используем дефолтный респаун.
	slog.Debug("siege HQ respawn: flag system not implemented, using default",
		"player", player.Name(), "castleID", castleID)
	return defaultRespawnLoc
}

// findZoneSpawn searches data.ZonesByType for a zone matching the given type and
// residence parameter, returning the first "normal" spawn point (empty spawnType).
func findZoneSpawn(zoneType, paramKey string, residenceID int32) (model.Location, bool) {
	zones := data.ZonesByType[zoneType]
	idStr := strconv.FormatInt(int64(residenceID), 10)

	for _, zd := range zones {
		if zd.Params()[paramKey] != idStr {
			continue
		}

		for _, sp := range zd.Spawns() {
			if sp.SpawnType() != "" {
				continue
			}
			return model.NewLocation(sp.SpawnX(), sp.SpawnY(), sp.SpawnZ(), 0), true
		}
	}

	return model.Location{}, false
}

// handleRequestLinkHtml processes 0x20 — load linked HTML page from NPC dialog.
// Java reference: RequestLinkHtml.java — validates link, checks interaction distance,
// loads HTML file and sends NpcHtmlMessage with NPC's objectID.
func (h *Handler) handleRequestLinkHtml(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestLinkHtml(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestLinkHtml: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for link html")
	}

	if pkt.Link == "" {
		slog.Warn("link html: empty link", "player", player.Name())
		return 0, true, nil
	}

	// Path traversal protection
	if strings.Contains(pkt.Link, "..") {
		slog.Warn("link html: path traversal attempt",
			"player", player.Name(),
			"link", pkt.Link)
		return 0, true, nil
	}

	if h.dialogManager == nil {
		slog.Debug("link html: no dialog manager", "player", player.Name())
		return 0, true, nil
	}

	// Determine NPC objectID from player's current target.
	// Java checks validateHtmlAction + interaction distance;
	// here we use the target NPC's objectID for the response packet.
	var npcObjectID int32
	if target := player.Target(); target != nil {
		npcObjectID = int32(target.ObjectID())
	}

	// Build dialog data with NPC objectID for template substitution
	dialogData := html.DialogData{
		"objectId": fmt.Sprintf("%d", npcObjectID),
	}

	content, err := h.dialogManager.ExecuteLink(pkt.Link, dialogData)
	if err != nil {
		slog.Warn("link html: load error",
			"player", player.Name(),
			"link", pkt.Link,
			"error", err)
		return 0, true, nil
	}

	msg := serverpackets.NewNpcHtmlMessage(npcObjectID, content)
	msgData, err := msg.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing NpcHtmlMessage: %w", err)
	}

	n := copy(buf, msgData)
	return n, true, nil
}

// handleRequestSkillCoolTime processes 0x9D — client requests skill cooldown sync.
func (h *Handler) handleRequestSkillCoolTime(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for skill cool time")
	}

	pkt := &serverpackets.SkillCoolTime{}

	if skill.CastMgr != nil {
		entries := skill.CastMgr.GetAllCooldowns(player.ObjectID())
		cds := make([]serverpackets.SkillCoolDown, 0, len(entries))
		for _, e := range entries {
			skillLevel := player.GetSkillLevel(e.SkillID)
			if skillLevel <= 0 {
				skillLevel = 1
			}
			remainSec := e.RemainMs / 1000
			if remainSec < 1 {
				remainSec = 1
			}
			reuseSec := e.ReuseMs / 1000
			if reuseSec < 1 {
				reuseSec = remainSec
			}
			cds = append(cds, serverpackets.SkillCoolDown{
				SkillID:       e.SkillID,
				SkillLevel:    skillLevel,
				ReuseTime:     reuseSec,
				RemainingTime: remainSec,
			})
		}
		pkt.CoolDowns = cds
	}

	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SkillCoolTime: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}
