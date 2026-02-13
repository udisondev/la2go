package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/game/augment"
	"github.com/udisondev/la2go/internal/game/enchant"
	"github.com/udisondev/la2go/internal/game/bbs"
	"github.com/udisondev/la2go/internal/game/cursed"
	"github.com/udisondev/la2go/internal/constants"
	skilldata "github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/game/craft"
	"github.com/udisondev/la2go/internal/game/crest"
	"github.com/udisondev/la2go/internal/game/duel"
	"github.com/udisondev/la2go/internal/game/geo"
	"github.com/udisondev/la2go/internal/game/hall"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/game/instance"
	"github.com/udisondev/la2go/internal/game/olympiad"
	"github.com/udisondev/la2go/internal/game/party"
	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/game/sevensigns"
	"github.com/udisondev/la2go/internal/game/siege"
	"github.com/udisondev/la2go/internal/game/skill"
	"github.com/udisondev/la2go/internal/game/zone"
	"github.com/udisondev/la2go/internal/gameserver/admin"
	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/html"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/protocol"
	"github.com/udisondev/la2go/internal/world"
)

// Handler processes game client packets.
type Handler struct {
	sessionManager *login.SessionManager
	clientManager  *ClientManager      // Phase 4.5 PR4: register clients after auth
	charRepo       CharacterRepository // Phase 4.6: load characters for CharSelectionInfo
	persister      PlayerPersister     // Phase 6.0: DB persistence
	partyManager   *party.Manager      // Phase 7.3: party system
	zoneManager    *zone.Manager       // Phase 7.2: zone logic
	geoEngine      *geo.Engine         // Phase 7.1: GeoEngine pathfinding & LOS
	dialogManager   *html.DialogManager // Phase 11: NPC dialog system
	adminHandler    *admin.Handler      // Phase 17: Admin/User commands
	craftController *craft.Controller   // Phase 15: Recipe/Craft system
	questManager    *quest.Manager      // Phase 16: Quest system
	clanTable       *clan.Table         // Phase 18: Clan system
	siegeManager    *siege.Manager          // Phase 21: Siege system
	duelManager     *duel.Manager          // Phase 20: Duel system
	hallTable       *hall.Table            // Phase 22: Clan Halls
	sevenSignsMgr   *sevensigns.Manager    // Phase 25: Seven Signs
	instanceManager *instance.Manager      // Phase 26: Instance Zones
	olympiadMgr     *olympiad.Olympiad     // Phase 24: Olympiad + Hero System
	augmentService  *augment.Service       // Phase 28: Augmentation System
	bbsHandler      *bbs.Handler           // Phase 30: Community Board
	offlineSvc      OfflineTradeService    // Phase 31: Offline Trade
	cursedMgr       *cursed.Manager        // Phase 32: Cursed Weapons
	manorMgr        *manor.Manager         // Phase 49: Manor seed/crop handlers
	friendStore    FriendStore             // Phase 35: Friend/Block DB persistence
	crestTbl       *crest.Table           // Phase 52: Crest System (initialized internally)
}

// CharacterRepository defines interface for loading/creating characters in database.
// Used for dependency injection to keep handler testable.
type CharacterRepository interface {
	LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error)
	Create(ctx context.Context, accountName string, p *model.Player) error
	NameExists(ctx context.Context, name string) (bool, error)
	CountByAccountName(ctx context.Context, accountName string) (int, error)
	MarkForDeletion(ctx context.Context, characterID int64, deleteTimerMs int64) error
	RestoreCharacter(ctx context.Context, characterID int64) error
	GetClanID(ctx context.Context, characterID int64) (int64, error)
}

// PlayerPersister defines interface for saving/loading player data.
// Phase 6.0: DB Persistence.
// Phase 13: LoadPlayerData returns *db.PlayerData (items, skills, recipes, hennas).
type PlayerPersister interface {
	SavePlayer(ctx context.Context, player *model.Player) error
	LoadPlayerData(ctx context.Context, charID int64) (*db.PlayerData, error)
}

// FriendStore defines interface for immediate friend/block DB operations.
// Phase 35: Friend/Block persistence.
type FriendStore interface {
	InsertFriend(ctx context.Context, charID int64, friendID int32) error
	DeleteFriend(ctx context.Context, charID int64, friendID int32) error
	InsertBlock(ctx context.Context, charID int64, blockedID int32) error
	DeleteBlock(ctx context.Context, charID int64, blockedID int32) error
}

// NewHandler creates a new packet handler for game clients.
func NewHandler(
	sessionManager *login.SessionManager,
	clientManager *ClientManager,
	charRepo CharacterRepository,
	persister PlayerPersister,
	partyMgr *party.Manager,
	zoneMgr *zone.Manager,
	geoEng *geo.Engine,
	dialogMgr *html.DialogManager,
	adminHandler *admin.Handler,
	craftCtrl *craft.Controller,
	questMgr *quest.Manager,
	clanTbl *clan.Table,
	siegeMgr *siege.Manager,
	duelMgr *duel.Manager,
	hallTbl *hall.Table,
	ssMgr *sevensigns.Manager,
	instanceMgr *instance.Manager,
	olympiadMgr *olympiad.Olympiad,
	augmentSvc *augment.Service,
	bbsHandler *bbs.Handler,
	offlineSvc OfflineTradeService,
	cursedMgr *cursed.Manager,
	manorMgr *manor.Manager,
	friendStore FriendStore,
) *Handler {
	return &Handler{
		sessionManager:  sessionManager,
		clientManager:   clientManager,
		charRepo:        charRepo,
		persister:       persister,
		partyManager:    partyMgr,
		zoneManager:     zoneMgr,
		geoEngine:       geoEng,
		dialogManager:   dialogMgr,
		adminHandler:    adminHandler,
		craftController: craftCtrl,
		questManager:    questMgr,
		clanTable:       clanTbl,
		siegeManager:    siegeMgr,
		duelManager:     duelMgr,
		hallTable:       hallTbl,
		sevenSignsMgr:   ssMgr,
		instanceManager: instanceMgr,
		olympiadMgr:     olympiadMgr,
		augmentService:  augmentSvc,
		bbsHandler:      bbsHandler,
		offlineSvc:      offlineSvc,
		cursedMgr:       cursedMgr,
		manorMgr:        manorMgr,
		friendStore:     friendStore,
		crestTbl:        crest.NewTable(),
	}
}

// HandlePacket dispatches a decrypted packet to the appropriate handler.
// Writes response into buf. Returns: n — bytes written to buf (0 = nothing to send),
// ok — true if connection stays open (false = close after sending).
func (h *Handler) HandlePacket(
	ctx context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	if len(data) == 0 {
		return 0, false, fmt.Errorf("empty packet data")
	}

	opcode := data[0]
	body := data[1:]
	state := client.State()

	switch state {
	case ClientStateConnected:
		switch opcode {
		case clientpackets.OpcodeProtocolVersion:
			return handleProtocolVersion(client, body)
		default:
			slog.Warn("invalid opcode for state CONNECTED",
				"opcode", fmt.Sprintf("0x%02X", opcode),
				"client", client.IP())
			return 0, false, nil
		}

	case ClientStateAuthenticated, ClientStateEntering, ClientStateInGame:
		switch opcode {
		case clientpackets.OpcodeAuthLogin:
			return h.handleAuthLogin(ctx, client, body, buf)
		case clientpackets.OpcodeNewCharacter:
			return h.handleNewCharacter(ctx, client, body, buf)
		case clientpackets.OpcodeCharacterCreate:
			return h.handleCharacterCreate(ctx, client, body, buf)
		case clientpackets.OpcodeCharacterDelete:
			return h.handleCharacterDelete(ctx, client, body, buf)
		case clientpackets.OpcodeCharacterRestore:
			return h.handleCharacterRestore(ctx, client, body, buf)
		case clientpackets.OpcodeCharacterSelect:
			return h.handleCharacterSelect(ctx, client, body, buf)
		case clientpackets.OpcodeEnterWorld:
			return h.handleEnterWorld(ctx, client, body, buf)
		case clientpackets.OpcodeMoveToLocation:
			return h.handleMoveToLocation(ctx, client, body, buf)
		case clientpackets.OpcodeCannotMoveAnymore:
			return h.handleCannotMoveAnymore(ctx, client, body, buf)
		case clientpackets.OpcodeValidatePosition:
			return h.handleValidatePosition(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAction:
			return h.handleRequestAction(ctx, client, body, buf)
		case clientpackets.OpcodeAttackRequest:
			return h.handleAttackRequest(ctx, client, body, buf)
		case clientpackets.OpcodeRequestDropItem:
			return h.handleRequestDropItem(ctx, client, body, buf)
		case clientpackets.OpcodeUseItem:
			return h.handleUseItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSocialAction:
			return h.handleRequestSocialAction(ctx, client, body, buf)
		case clientpackets.OpcodeRequestItemList:
			return h.handleRequestItemList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestUnEquipItem:
			return h.handleRequestUnEquipItem(ctx, client, body, buf)
		case clientpackets.OpcodeChangeMoveType2:
			return h.handleChangeMoveType2(ctx, client, body, buf)
		case clientpackets.OpcodeChangeWaitType2:
			return h.handleChangeWaitType2(ctx, client, body, buf)
		case clientpackets.OpcodeAppearing:
			return h.handleAppearing(ctx, client, body, buf)
		case clientpackets.OpcodeRequestTargetCanceld:
			return h.handleRequestTargetCanceld(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSkillList:
			return h.handleRequestSkillList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestDestroyItem:
			return h.handleRequestDestroyItem(ctx, client, body, buf)
		case clientpackets.OpcodeStartRotating:
			return h.handleStartRotating(ctx, client, body, buf)
		case clientpackets.OpcodeFinishRotating:
			return h.handleFinishRotating(ctx, client, body, buf)
		case clientpackets.OpcodeLogout:
			return h.handleLogout(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRestart:
			return h.handleRequestRestart(ctx, client, body, buf)
		case clientpackets.OpcodeRequestMagicSkillUse:
			return h.handleRequestMagicSkillUse(ctx, client, body, buf)
		case clientpackets.OpcodeSay2:
			return h.handleSay2(ctx, client, body, buf)
		case clientpackets.OpcodeRequestLinkHtml:
			return h.handleRequestLinkHtml(ctx, client, body, buf)
		case clientpackets.OpcodeRequestBypassToServer:
			return h.handleRequestBypassToServer(ctx, client, body, buf)
		case clientpackets.OpcodeRequestBuyItem:
			return h.handleRequestBuyItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSellItem:
			return h.handleRequestSellItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestJoinParty:
			return h.handleRequestJoinParty(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAnswerJoinParty:
			return h.handleAnswerJoinParty(ctx, client, body, buf)
		case clientpackets.OpcodeRequestWithdrawalParty:
			return h.handleWithdrawalParty(ctx, client, body, buf)
		case clientpackets.OpcodeRequestOustPartyMember:
			return h.handleOustPartyMember(ctx, client, body, buf)
		case clientpackets.OpcodeSendWareHouseDepositList:
			return h.handleWarehouseDeposit(ctx, client, body, buf)
		case clientpackets.OpcodeSendWareHouseWithDrawList:
			return h.handleWarehouseWithdraw(ctx, client, body, buf)
		case clientpackets.OpcodeMultiSellChoose:
			return h.handleMultiSellChoose(ctx, client, body, buf)
		// Private Store packets (Phase 8.1)
		case clientpackets.OpcodeRequestPrivateStoreManageSell:
			return h.handleRequestPrivateStoreManageSell(ctx, client, body, buf)
		case clientpackets.OpcodeSetPrivateStoreListSell:
			return h.handleSetPrivateStoreListSell(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPrivateStoreQuitSell:
			return h.handleRequestPrivateStoreQuitSell(ctx, client, body, buf)
		case clientpackets.OpcodeSetPrivateStoreMsgSell:
			return h.handleSetPrivateStoreMsgSell(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPrivateStoreBuy:
			return h.handleRequestPrivateStoreBuy(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPrivateStoreManageBuy:
			return h.handleRequestPrivateStoreManageBuy(ctx, client, body, buf)
		case clientpackets.OpcodeSetPrivateStoreListBuy:
			return h.handleSetPrivateStoreListBuy(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPrivateStoreQuitBuy:
			return h.handleRequestPrivateStoreQuitBuy(ctx, client, body, buf)
		case clientpackets.OpcodeSetPrivateStoreMsgBuy:
			return h.handleSetPrivateStoreMsgBuy(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPrivateStoreSell:
			return h.handleRequestPrivateStoreSell(ctx, client, body, buf)
		// Recipe/Craft packets (Phase 15)
		case clientpackets.OpcodeRequestRecipeBookOpen:
			return h.handleRequestRecipeBookOpen(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeItemMakeInfo:
			return h.handleRequestRecipeItemMakeInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeItemMakeSelf:
			return h.handleRequestRecipeItemMakeSelf(ctx, client, body, buf)
		// Henna packets (Phase 13)
		case clientpackets.OpcodeRequestHennaItemList:
			return h.handleRequestHennaItemList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestHennaItemInfo:
			return h.handleRequestHennaItemInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestHennaEquip:
			return h.handleRequestHennaEquip(ctx, client, body, buf)
		case clientpackets.OpcodeRequestHennaRemoveList:
			return h.handleRequestHennaRemoveList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestHennaItemRemoveInfo:
			return h.handleRequestHennaItemRemoveInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestHennaRemove:
			return h.handleRequestHennaRemove(ctx, client, body, buf)
		// Quest packets (Phase 16)
		case clientpackets.OpcodeRequestQuestList:
			return h.handleRequestQuestList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestQuestAbort:
			return h.handleRequestQuestAbort(ctx, client, body, buf)
		// Clan packets (Phase 18)
		case clientpackets.OpcodeRequestJoinPledge:
			return h.handleRequestJoinPledge(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAnswerJoinPledge:
			return h.handleRequestAnswerJoinPledge(ctx, client, body, buf)
		case clientpackets.OpcodeRequestWithdrawalPledge:
			return h.handleRequestWithdrawalPledge(ctx, client, body, buf)
		case clientpackets.OpcodeRequestOustPledgeMember:
			return h.handleRequestOustPledgeMember(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPledgeInfo:
			return h.handleRequestPledgeInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPledgeCrest:
			return h.handleRequestPledgeCrest(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPledgeMemberList:
			return h.handleRequestPledgeMemberList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPledgePower:
			return h.handleRequestPledgePower(ctx, client, body, buf)
		// Clan War (Phase 50 stubs)
		case clientpackets.OpcodeRequestStartPledgeWar:
			return h.handleRequestStartPledgeWar(ctx, client, body, buf)
		case clientpackets.OpcodeRequestReplyStartPledgeWar:
			return h.handleRequestReplyStartPledgeWar(ctx, client, body, buf)
		case clientpackets.OpcodeRequestStopPledgeWar:
			return h.handleRequestStopPledgeWar(ctx, client, body, buf)
		case clientpackets.OpcodeRequestReplyStopPledgeWar:
			return h.handleRequestReplyStopPledgeWar(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSurrenderPledgeWar:
			return h.handleRequestSurrenderPledgeWar(ctx, client, body, buf)
		case clientpackets.OpcodeRequestReplySurrenderPledgeWar:
			return h.handleRequestReplySurrenderPledgeWar(ctx, client, body, buf)
		// Clan Title (Phase 50 stub)
		case clientpackets.OpcodeRequestGiveNickName:
			return h.handleRequestGiveNickName(ctx, client, body, buf)
		// Siege packets (Phase 21)
		case clientpackets.OpcodeRequestSiegeInfo:
			return h.handleRequestSiegeInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSiegeAttackerList:
			return h.handleRequestSiegeAttackerList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSiegeDefenderList:
			return h.handleRequestSiegeDefenderList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestJoinSiege:
			return h.handleRequestJoinSiege(ctx, client, body, buf)
		case clientpackets.OpcodeRequestConfirmSiegeWaitingList:
			return h.handleRequestConfirmSiegeWaitingList(ctx, client, body, buf)
		// Seven Signs packets (Phase 25)
		case clientpackets.OpcodeRequestSSQStatus:
			return h.handleRequestSSQStatus(ctx, client, body, buf)
		// Pet/Summon packets (Phase 19)
		case clientpackets.OpcodeRequestActionUse:
			return h.handleRequestActionUse(ctx, client, body, buf)
		case clientpackets.OpcodeRequestChangePetName:
			return h.handleRequestChangePetName(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPetUseItem:
			return h.handleRequestPetUseItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestGiveItemToPet:
			return h.handleRequestGiveItemToPet(ctx, client, body, buf)
		case clientpackets.OpcodeRequestGetItemFromPet:
			return h.handleRequestGetItemFromPet(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPetGetItem:
			return h.handleRequestPetGetItem(ctx, client, body, buf)
		// Alliance (Phase 50 stubs)
		case clientpackets.OpcodeRequestJoinAlly:
			return h.handleRequestJoinAlly(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAnswerJoinAlly:
			return h.handleRequestAnswerJoinAlly(ctx, client, body, buf)
		case clientpackets.OpcodeAllyLeave:
			return h.handleAllyLeave(ctx, client, body, buf)
		case clientpackets.OpcodeAllyDismiss:
			return h.handleAllyDismiss(ctx, client, body, buf)
		case clientpackets.OpcodeRequestDismissAlly:
			return h.handleRequestDismissAlly(ctx, client, body, buf)
		// Trade packets
		case clientpackets.OpcodeTradeRequest:
			return h.handleTradeRequest(ctx, client, body, buf)
		case clientpackets.OpcodeAnswerTradeRequest:
			return h.handleAnswerTradeRequest(ctx, client, body, buf)
		case clientpackets.OpcodeAddTradeItem:
			return h.handleAddTradeItem(ctx, client, body, buf)
		case clientpackets.OpcodeTradeDone:
			return h.handleTradeDone(ctx, client, body, buf)

		// Enchant packets
		case clientpackets.OpcodeRequestEnchantItem:
			return h.handleRequestEnchantItem(ctx, client, body, buf)

		// Community Board packets (Phase 30)
		case clientpackets.OpcodeRequestShowBoard:
			return h.handleRequestShowBoard(ctx, client, body, buf)
		case clientpackets.OpcodeRequestBBSwrite:
			return h.handleRequestBBSwrite(ctx, client, body, buf)

		// Friend packets (Phase 35)
		case clientpackets.OpcodeRequestFriendInvite:
			return h.handleRequestFriendInvite(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAnswerFriendInvite:
			return h.handleRequestAnswerFriendInvite(ctx, client, body, buf)
		case clientpackets.OpcodeRequestFriendList:
			return h.handleRequestFriendList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestFriendDel:
			return h.handleRequestFriendDel(ctx, client, body, buf)

		// Block packets (Phase 35)
		case clientpackets.OpcodeRequestBlock:
			return h.handleRequestBlock(ctx, client, body, buf)

		// Shortcut packets (Phase 34)
		case clientpackets.OpcodeRequestShortCutReg:
			return h.handleRequestShortCutReg(ctx, client, body, buf)
		case clientpackets.OpcodeRequestShortCutDel:
			return h.handleRequestShortCutDel(ctx, client, body, buf)

		// Macro packets (Phase 36)
		case clientpackets.OpcodeRequestMakeMacro:
			return h.handleRequestMakeMacro(ctx, client, body, buf)
		case clientpackets.OpcodeRequestDeleteMacro:
			return h.handleRequestDeleteMacro(ctx, client, body, buf)

		// Friend PM (Phase 36)
		case clientpackets.OpcodeRequestSendFriendMsg:
			return h.handleRequestSendFriendMsg(ctx, client, body, buf)

		// Crystallize (Phase 36)
		case clientpackets.OpcodeRequestCrystallizeItem:
			return h.handleRequestCrystallizeItem(ctx, client, body, buf)

		// Restart Point — respawn after death (Phase 37)
		case clientpackets.OpcodeRequestRestartPoint:
			return h.handleRequestRestartPoint(ctx, client, body, buf)

		// Skill Learning (Phase 38)
		case clientpackets.OpcodeRequestAcquireSkillInfo:
			return h.handleRequestAcquireSkillInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestAcquireSkill:
			return h.handleRequestAcquireSkill(ctx, client, body, buf)

		// Dialog Answer (Phase 38)
		case clientpackets.OpcodeDlgAnswer:
			return h.handleDlgAnswer(ctx, client, body, buf)

		// Skill Cool Time sync (Phase 49)
		case clientpackets.OpcodeRequestSkillCoolTime:
			return h.handleRequestSkillCoolTime(ctx, client, body, buf)

		// Manor: Buy Seeds (Phase 49)
		case clientpackets.OpcodeRequestBuySeed:
			return h.handleRequestBuySeed(ctx, client, body, buf)

		// Alliance Crest request (Phase 50)
		case clientpackets.OpcodeRequestAllyCrest:
			return h.handleRequestAllyCrest(ctx, client, body, buf)

		// Crest upload stubs (Phase 50)
		case clientpackets.OpcodeRequestSetPledgeCrest:
			return h.handleRequestSetPledgeCrest(ctx, client, body, buf)
		case clientpackets.OpcodeRequestSetAllyCrest:
			return h.handleRequestSetAllyCrest(ctx, client, body, buf)

		// GM list (Phase 50)
		case clientpackets.OpcodeRequestGmList:
			return h.handleRequestGmList(ctx, client, body, buf)

		// Alliance info (Phase 50)
		case clientpackets.OpcodeRequestAllyInfo:
			return h.handleRequestAllyInfo(ctx, client, body, buf)

		// GM //command bypass (Phase 50)
		case clientpackets.OpcodeSendBypassBuildCmd:
			return h.handleSendBypassBuildCmd(ctx, client, body, buf)

		// Recipe Shop (Phase 50 stubs)
		case clientpackets.OpcodeRequestRecipeShopMessageSet:
			return h.handleRequestRecipeShopMessageSet(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeShopListSet:
			return h.handleRequestRecipeShopListSet(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeShopManageQuit:
			return h.handleRequestRecipeShopManageQuit(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeShopMakeInfo:
			return h.handleRequestRecipeShopMakeInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeShopMakeItem:
			return h.handleRequestRecipeShopMakeItem(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeShopManagePrev:
			return h.handleRequestRecipeShopManagePrev(ctx, client, body, buf)

		// Miscellaneous (Phase 50)
		case clientpackets.OpcodeRequestRecordInfo:
			return h.handleRequestRecordInfo(ctx, client, body, buf)
		case clientpackets.OpcodeRequestShowMiniMap:
			return h.handleRequestShowMiniMap(ctx, client, body, buf)
		case clientpackets.OpcodeObserverReturn:
			return h.handleObserverReturn(ctx, client, body, buf)
		case clientpackets.OpcodeRequestRecipeBookDestroy:
			return h.handleRequestRecipeBookDestroy(ctx, client, body, buf)
		case clientpackets.OpcodeRequestEvaluate:
			return h.handleRequestEvaluate(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPartyMatchConfig:
			return h.handleRequestPartyMatchConfig(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPartyMatchList:
			return h.handleRequestPartyMatchList(ctx, client, body, buf)
		case clientpackets.OpcodeRequestPartyMatchDetail:
			return h.handleRequestPartyMatchDetail(ctx, client, body, buf)

		// Extended client packets (0xD0 + sub-opcode)
		case clientpackets.OpcodeRequestExPacket:
			return h.handleExtendedPacket(ctx, client, body, buf)
		default:
			slog.Warn("unknown packet opcode",
				"opcode", fmt.Sprintf("0x%02X", opcode),
				"state", state,
				"client", client.IP())
			return 0, true, nil
		}

	default:
		return 0, false, fmt.Errorf("invalid state: %v", state)
	}
}

// handleProtocolVersion processes the ProtocolVersion packet (opcode 0x0E).
func handleProtocolVersion(client *GameClient, data []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseProtocolVersion(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ProtocolVersion: %w", err)
	}

	if !pkt.IsValid() {
		slog.Warn("invalid protocol version",
			"expected", 0x0106,
			"got", pkt.ProtocolRevision,
			"client", client.IP())
		return 0, false, fmt.Errorf("invalid protocol revision: 0x%04X", pkt.ProtocolRevision)
	}

	slog.Debug("protocol version validated", "client", client.IP())

	// Protocol version is valid, wait for AuthLogin
	// No response packet
	return 0, true, nil
}

// handleAuthLogin processes the AuthLogin packet (opcode 0x08).
func (h *Handler) handleAuthLogin(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAuthLogin(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AuthLogin: %w", err)
	}

	// Validate SessionKey with SessionManager (shared with LoginServer)
	// showLicence=false because GameServer doesn't care about license state
	if !h.sessionManager.Validate(pkt.AccountName, pkt.SessionKey, false) {
		slog.Warn("session key validation failed",
			"account", pkt.AccountName,
			"client", client.IP())

		// Send AuthLoginFail packet before closing connection
		failPkt := serverpackets.NewAuthLoginFail(serverpackets.AuthFailReasonAccessDenied)
		failData, writeErr := failPkt.Write()
		if writeErr != nil {
			slog.Error("failed to serialize AuthLoginFail", "error", writeErr)
			return 0, false, fmt.Errorf("invalid session key for account %s", pkt.AccountName)
		}
		n := copy(buf, failData)
		return n, false, fmt.Errorf("invalid session key for account %s", pkt.AccountName)
	}

	// SessionKey is valid, set client state
	client.SetAccountName(pkt.AccountName)
	client.SetSessionKey(&pkt.SessionKey)
	client.SetState(ClientStateAuthenticated)

	// Register client in ClientManager (Phase 4.5 PR4)
	h.clientManager.Register(pkt.AccountName, client)

	slog.Info("client authenticated",
		"account", pkt.AccountName,
		"client", client.IP())

	// Load characters for this account (Phase 4.6)
	// Phase 4.18: Use cached loader to eliminate redundant DB queries
	players, err := client.GetCharacters(pkt.AccountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", pkt.AccountName, err)
	}

	// Create and send CharSelectionInfo packet
	// SessionID is derived from SessionKey (use PlayOkID1)
	sessionID := pkt.SessionKey.PlayOkID1
	charSelInfo := serverpackets.NewCharSelectionInfoFromPlayers(pkt.AccountName, sessionID, players)

	packetData, err := charSelInfo.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharSelectionInfo: %w", err)
	}

	// Copy packet data to response buffer
	n := copy(buf, packetData)
	if n != len(packetData) {
		return 0, false, fmt.Errorf("buffer too small: need %d bytes, have %d", len(packetData), len(buf))
	}

	slog.Debug("sent CharSelectionInfo",
		"account", pkt.AccountName,
		"character_count", len(players),
		"packet_size", n)

	return n, true, nil
}

// handleNewCharacter processes the NewCharacter packet (opcode 0x0E in AUTHENTICATED state).
// Client sends this when user clicks "Create" on character selection screen.
// Response: CharTemplates S2C (0x17) with 9 base class templates.
func (h *Handler) handleNewCharacter(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseNewCharacter(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing NewCharacter: %w", err)
	}

	charTmpl := &serverpackets.CharTemplates{}
	tmplData, err := charTmpl.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharTemplates: %w", err)
	}

	n := copy(buf, tmplData)

	slog.Debug("sent character templates", "account", client.AccountName())

	return n, true, nil
}

// handleCharacterCreate processes the CharacterCreate packet (opcode 0x0B).
// Client sends this when user creates a new character.
// Response: CharCreateOk (0x19) or CharCreateFail (0x1A).
func (h *Handler) handleCharacterCreate(ctx context.Context, client *GameClient, data []byte, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCharacterCreate(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CharacterCreate: %w", err)
	}

	accountName := client.AccountName()

	// Validate name length (1-16 chars)
	nameLen := len([]rune(pkt.Name))
	if nameLen < 1 || nameLen > 16 {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonIncorrectName)
	}

	// Validate name: alphanumeric only
	for _, r := range pkt.Name {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonIncorrectName)
		}
	}

	// Validate class is a base class (occupation level 0)
	if skilldata.ClassLevel(pkt.ClassID) != 0 {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonNotAllowed)
	}

	// Get player template for this class
	tmpl := skilldata.GetTemplate(uint8(pkt.ClassID))
	if tmpl == nil {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonNotAllowed)
	}

	// Validate appearance
	if pkt.Face < 0 || pkt.Face > 2 {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}
	maxHairStyle := int32(4)
	if pkt.IsFemale {
		maxHairStyle = 6
	}
	if pkt.HairStyle < 0 || pkt.HairStyle > maxHairStyle {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}
	if pkt.HairColor < 0 || pkt.HairColor > 3 {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}

	// Check max characters per account (7)
	count, err := h.charRepo.CountByAccountName(ctx, accountName)
	if err != nil {
		slog.Error("counting characters", "account", accountName, "error", err)
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}
	if count >= 7 {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonTooMany)
	}

	// Check name uniqueness
	exists, err := h.charRepo.NameExists(ctx, pkt.Name)
	if err != nil {
		slog.Error("checking name existence", "name", pkt.Name, "error", err)
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}
	if exists {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonNameExists)
	}

	// Derive raceID from classID
	classInfo := skilldata.GetClassInfo(pkt.ClassID)
	if classInfo == nil {
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonNotAllowed)
	}
	raceID := classInfo.Race

	// Get spawn location from template
	var spawnX, spawnY, spawnZ int32
	if len(tmpl.CreationPoints) > 0 {
		sp := tmpl.CreationPoints[0]
		spawnX, spawnY, spawnZ = sp.X, sp.Y, sp.Z
	}

	// Create the player model
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, 0, 0, pkt.Name, 1, raceID, pkt.ClassID)
	if err != nil {
		slog.Error("creating player model", "name", pkt.Name, "error", err)
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}

	// Set appearance
	player.SetIsFemale(pkt.IsFemale)
	player.SetFace(pkt.Face)
	player.SetHairStyle(pkt.HairStyle)
	player.SetHairColor(pkt.HairColor)

	// Set spawn location
	loc := model.NewLocation(spawnX, spawnY, spawnZ, 0)
	player.SetLocation(loc)

	// Set stats from template (level 1)
	hp := int32(tmpl.GetHPMax(1))
	mp := int32(tmpl.GetMPMax(1))
	cp := int32(tmpl.GetCPMax(1))
	player.SetMaxHP(hp)
	player.SetMaxMP(mp)
	player.SetMaxCP(cp)
	player.SetCurrentHP(hp)
	player.SetCurrentMP(mp)
	player.SetCurrentCP(cp)

	// Save to DB
	if err := h.charRepo.Create(ctx, accountName, player); err != nil {
		slog.Error("saving new character to DB", "name", pkt.Name, "account", accountName, "error", err)
		return h.sendCharCreateFail(buf, serverpackets.CharCreateReasonFailed)
	}

	// Invalidate character cache
	client.ClearCharacterCache()

	slog.Info("character created",
		"name", pkt.Name,
		"account", accountName,
		"classID", pkt.ClassID,
		"raceID", raceID,
		"characterID", player.CharacterID())

	// Send CharCreateOk
	okPkt := &serverpackets.CharCreateOk{}
	pktData, err := okPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharCreateOk: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// sendCharCreateFail sends a CharCreateFail packet with the given reason.
func (h *Handler) sendCharCreateFail(buf []byte, reason int32) (int, bool, error) {
	pkt := serverpackets.NewCharCreateFail(reason)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharCreateFail: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleCharacterDelete processes the CharacterDelete packet (opcode 0x0C).
// Sets a 7-day delete timer on the character; clan leaders/members are blocked.
func (h *Handler) handleCharacterDelete(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCharacterDelete(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CharacterDelete: %w", err)
	}

	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for delete: %w", err)
	}

	if int(pkt.CharSlot) < 0 || int(pkt.CharSlot) >= len(players) {
		return h.sendCharDeleteFail(buf, serverpackets.CharDeleteReasonFailed)
	}

	player := players[pkt.CharSlot]

	// Check clan membership
	clanID, err := h.charRepo.GetClanID(ctx, player.CharacterID())
	if err != nil {
		slog.Error("checking clan for delete", "characterID", player.CharacterID(), "error", err)
		return h.sendCharDeleteFail(buf, serverpackets.CharDeleteReasonFailed)
	}

	if clanID != 0 {
		// Check if clan leader
		if h.clanTable != nil {
			c := h.clanTable.Clan(int32(clanID))
			if c != nil && c.LeaderID() == player.CharacterID() {
				return h.sendCharDeleteFail(buf, serverpackets.CharDeleteReasonClanLeader)
			}
		}
		return h.sendCharDeleteFail(buf, serverpackets.CharDeleteReasonClanMember)
	}

	// Set 7-day delete timer (7 * 86400 * 1000 ms)
	deleteTime := time.Now().UnixMilli() + 7*86400*1000
	if err := h.charRepo.MarkForDeletion(ctx, player.CharacterID(), deleteTime); err != nil {
		slog.Error("marking character for deletion", "characterID", player.CharacterID(), "error", err)
		return h.sendCharDeleteFail(buf, serverpackets.CharDeleteReasonFailed)
	}

	// Clear character cache
	client.ClearCharacterCache()

	slog.Info("character marked for deletion",
		"name", player.Name(),
		"account", accountName,
		"characterID", player.CharacterID())

	// Send CharDeleteOk
	okPkt := &serverpackets.CharDeleteOk{}
	okData, err := okPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharDeleteOk: %w", err)
	}
	n := copy(buf, okData)

	// Send updated character list
	updatedPlayers, err := h.charRepo.LoadByAccountName(ctx, accountName)
	if err != nil {
		slog.Error("loading characters after delete", "account", accountName, "error", err)
		return n, true, nil
	}

	sessionKey := client.SessionKey()
	sessionID := int32(0)
	if sessionKey != nil {
		sessionID = sessionKey.PlayOkID1
	}
	charList := serverpackets.NewCharSelectionInfoFromPlayers(accountName, sessionID, updatedPlayers)
	charData, err := charList.Write()
	if err != nil {
		slog.Error("serializing CharSelectionInfo after delete", "error", err)
		return n, true, nil
	}
	n2 := copy(buf[n:], charData)

	return n + n2, true, nil
}

// sendCharDeleteFail sends a CharDeleteFail packet with the given reason.
func (h *Handler) sendCharDeleteFail(buf []byte, reason int32) (int, bool, error) {
	pkt := serverpackets.NewCharDeleteFail(reason)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharDeleteFail: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleCharacterRestore processes the CharacterRestore packet (opcode 0x62).
// Cancels pending deletion by clearing the delete timer.
func (h *Handler) handleCharacterRestore(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCharacterRestore(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CharacterRestore: %w", err)
	}

	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for restore: %w", err)
	}

	if int(pkt.CharSlot) < 0 || int(pkt.CharSlot) >= len(players) {
		return 0, false, fmt.Errorf("invalid character slot for restore: %d", pkt.CharSlot)
	}

	player := players[pkt.CharSlot]

	if err := h.charRepo.RestoreCharacter(ctx, player.CharacterID()); err != nil {
		slog.Error("restoring character", "characterID", player.CharacterID(), "error", err)
		return 0, false, fmt.Errorf("restoring character: %w", err)
	}

	// Clear character cache
	client.ClearCharacterCache()

	slog.Info("character restored",
		"name", player.Name(),
		"account", accountName,
		"characterID", player.CharacterID())

	// Send updated character list
	updatedPlayers, err := h.charRepo.LoadByAccountName(ctx, accountName)
	if err != nil {
		return 0, false, fmt.Errorf("loading characters after restore: %w", err)
	}

	sessionKey := client.SessionKey()
	sessionID := int32(0)
	if sessionKey != nil {
		sessionID = sessionKey.PlayOkID1
	}
	charList := serverpackets.NewCharSelectionInfoFromPlayers(accountName, sessionID, updatedPlayers)
	charData, err := charList.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharSelectionInfo after restore: %w", err)
	}
	n := copy(buf, charData)
	return n, true, nil
}

// handleCharacterSelect processes the CharacterSelect packet (opcode 0x0D).
// Client sends this when user selects a character from the character list.
// Response: CharSelected packet with character data.
func (h *Handler) handleCharacterSelect(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCharacterSelect(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CharacterSelect: %w", err)
	}

	// Validate character slot (0-7)
	if pkt.CharSlot < 0 || pkt.CharSlot > 7 {
		slog.Warn("invalid character slot",
			"slot", pkt.CharSlot,
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("invalid character slot: %d", pkt.CharSlot)
	}

	// Load characters for this account
	// Phase 4.18: Use cached loader (2nd call — cache hit expected)
	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", accountName, err)
	}

	// Validate slot index
	if int(pkt.CharSlot) >= len(players) {
		slog.Warn("character slot out of range",
			"slot", pkt.CharSlot,
			"character_count", len(players),
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("character slot %d out of range (have %d characters)", pkt.CharSlot, len(players))
	}

	// Get selected character
	player := players[pkt.CharSlot]

	// Get PlayOkID1 from SessionKey for CharSelected packet
	sessionKey := client.SessionKey()
	if sessionKey == nil {
		slog.Error("no session key for authenticated client",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("missing session key")
	}

	// Store selected character slot
	client.SetSelectedCharacter(pkt.CharSlot)

	// Send CharSelected packet (Phase 4.17.1)
	charSelected := serverpackets.NewCharSelected(player, sessionKey.PlayOkID1)
	charSelectedData, err := charSelected.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CharSelected: %w", err)
	}

	n := copy(buf, charSelectedData)
	if n != len(charSelectedData) {
		return 0, false, fmt.Errorf("buffer too small for CharSelected")
	}

	// Transition to ENTERING state (Phase 4.17.2)
	client.SetState(ClientStateEntering)

	slog.Info("character selected",
		"account", client.AccountName(),
		"character", player.Name(),
		"slot", pkt.CharSlot,
		"level", player.Level(),
		"client", client.IP())

	return n, true, nil
}

// handleEnterWorld processes the EnterWorld packet (opcode 0x03).
// Client sends this after CharacterSelect to spawn in the world.
func (h *Handler) handleEnterWorld(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseEnterWorld(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing EnterWorld: %w", err)
	}

	// Verify character was selected
	charSlot := client.SelectedCharacter()
	if charSlot < 0 {
		slog.Warn("EnterWorld without character selection",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("no character selected")
	}

	// Load characters for this account
	// Phase 4.18: Use cached loader (3rd call — cache hit expected)
	accountName := client.AccountName()
	players, err := client.GetCharacters(accountName, func(name string) ([]*model.Player, error) {
		return h.charRepo.LoadByAccountName(ctx, name)
	})
	if err != nil {
		return 0, false, fmt.Errorf("loading characters for account %s: %w", accountName, err)
	}

	// Validate slot index
	if int(charSlot) >= len(players) {
		slog.Warn("character slot out of range",
			"slot", charSlot,
			"character_count", len(players),
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("character slot %d out of range (have %d characters)", charSlot, len(players))
	}

	// Get selected character
	player := players[charSlot]

	// Cache player in GameClient (Phase 4.8 part 2)
	client.SetActivePlayer(player)

	// Phase 5.9.5: Apply auto-get skills for current level
	autoSkills := skilldata.GetAutoGetSkills(player.ClassID(), player.Level())
	for _, sl := range autoSkills {
		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(sl.SkillID, sl.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(sl.SkillID, sl.SkillLevel, isPassive)
	}

	// Phase 6.0: Load items, skills, recipes and hennas from DB
	playerData, err := h.persister.LoadPlayerData(ctx, player.CharacterID())
	if err != nil {
		slog.Error("load player data",
			"characterID", player.CharacterID(),
			"err", err)
		// Continue without — not fatal
	}

	// Restore skills from DB (override auto-get with saved levels)
	if playerData == nil {
		playerData = &db.PlayerData{}
	}
	for _, si := range playerData.Skills {
		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(si.SkillID, si.Level); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(si.SkillID, si.Level, isPassive)
	}

	// Restore items to inventory
	for _, row := range playerData.Items {
		template := db.ItemDefToTemplate(row.ItemTypeID)
		if template == nil {
			slog.Warn("item template not found, skipping",
				"itemTypeID", row.ItemTypeID,
				"characterID", player.CharacterID())
			continue
		}
		objectID := world.IDGenerator().NextItemID()
		item, itemErr := model.NewItem(objectID, row.ItemTypeID, player.CharacterID(), row.Count, template)
		if itemErr != nil {
			slog.Error("restore item failed",
				"itemTypeID", row.ItemTypeID,
				"error", itemErr)
			continue
		}
		if row.Enchant > 0 {
			if enchErr := item.SetEnchant(row.Enchant); enchErr != nil {
				slog.Error("set enchant failed",
					"itemTypeID", row.ItemTypeID,
					"error", enchErr)
			}
		}
		if addErr := player.Inventory().AddItem(item); addErr != nil {
			slog.Error("add item to inventory failed",
				"itemTypeID", row.ItemTypeID,
				"error", addErr)
			continue
		}
		if model.ItemLocation(row.Location) == model.ItemLocationPaperdoll && row.SlotID >= 0 {
			if _, equipErr := player.Inventory().EquipItem(item, row.SlotID); equipErr != nil {
				slog.Error("equip item failed",
					"itemTypeID", row.ItemTypeID,
					"slot", row.SlotID,
					"error", equipErr)
			}
		}
	}

	// Phase 15: Restore recipes from DB
	for _, rr := range playerData.Recipes {
		if err := player.LearnRecipe(rr.RecipeID, rr.IsDwarven); err != nil {
			slog.Warn("restore recipe failed",
				"recipeID", rr.RecipeID,
				"isDwarven", rr.IsDwarven,
				"characterID", player.CharacterID(),
				"error", err)
		}
	}

	// Phase 13: Restore hennas from DB
	for _, hr := range playerData.Hennas {
		if err := player.SetHenna(int(hr.Slot), hr.DyeID); err != nil {
			slog.Warn("restore henna failed",
				"slot", hr.Slot,
				"dyeID", hr.DyeID,
				"characterID", player.CharacterID(),
				"error", err)
		}
	}
	if len(playerData.Hennas) > 0 {
		player.RecalcHennaStats()
	}

	// Phase 14: Restore subclasses from DB
	for _, sr := range playerData.SubClasses {
		player.RestoreSubClass(&model.SubClass{
			ClassID:    sr.ClassID,
			ClassIndex: sr.ClassIndex,
			Level:      sr.Level,
			Exp:        sr.Exp,
			SP:         sr.SP,
		})
	}

	// Phase 35: Restore friends and blocks from DB
	if len(playerData.Friends) > 0 {
		friendIDs := make([]int32, 0, len(playerData.Friends))
		blockIDs := make([]int32, 0)
		for _, fr := range playerData.Friends {
			if fr.Relation == 0 {
				friendIDs = append(friendIDs, fr.FriendID)
			} else {
				blockIDs = append(blockIDs, fr.FriendID)
			}
		}
		if len(friendIDs) > 0 {
			player.SetFriendList(friendIDs)
		}
		if len(blockIDs) > 0 {
			player.SetBlockList(blockIDs)
		}
	}

	// Register player in World Grid (Phase 4.9)
	if err := world.Instance().AddObject(player.WorldObject); err != nil {
		return 0, false, fmt.Errorf("adding player to world: %w", err)
	}

	// Update client state
	client.SetState(ClientStateInGame)

	slog.Info("player entering world",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// Send multiple packets after EnterWorld (Phase 4.7)
	// Order is important: UserInfo must be first, then StatusUpdate, then others
	var totalBytes int

	// 1. UserInfo (spawns character in world)
	userInfo := serverpackets.NewUserInfo(player)
	userInfoData, err := userInfo.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing UserInfo: %w", err)
	}
	n := copy(buf[totalBytes:], userInfoData)
	if n != len(userInfoData) {
		return 0, false, fmt.Errorf("buffer too small for UserInfo")
	}
	totalBytes += n

	// 2. StatusUpdate (HP/MP/CP bars)
	statusUpdate := serverpackets.NewStatusUpdate(player)
	statusData, err := statusUpdate.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing StatusUpdate: %w", err)
	}
	n = copy(buf[totalBytes:], statusData)
	if n != len(statusData) {
		return 0, false, fmt.Errorf("buffer too small for StatusUpdate")
	}
	totalBytes += n

	// 3. InventoryItemList (Phase 6.0: send real items)
	invList := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invData, err := invList.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}
	n = copy(buf[totalBytes:], invData)
	if n != len(invData) {
		return 0, false, fmt.Errorf("buffer too small for InventoryItemList")
	}
	totalBytes += n

	// 4. ShortCutInit (sends registered shortcuts or empty list)
	shortcuts := serverpackets.NewShortCutInit(player.GetShortcuts())
	shortcutData, err := shortcuts.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ShortCutInit: %w", err)
	}
	n = copy(buf[totalBytes:], shortcutData)
	if n != len(shortcutData) {
		return 0, false, fmt.Errorf("buffer too small for ShortCutInit")
	}
	totalBytes += n

	// 5. FriendList (Phase 35: send friends with online status)
	friendInfos := make([]serverpackets.FriendInfo, 0, len(player.FriendList()))
	for _, friendID := range player.FriendList() {
		friendClient := h.clientManager.GetClientByObjectID(uint32(friendID))
		isOnline := friendClient != nil && friendClient.ActivePlayer() != nil
		name := ""
		if isOnline {
			name = friendClient.ActivePlayer().Name()
		}
		friendInfos = append(friendInfos, serverpackets.FriendInfo{
			ObjectID: friendID,
			Name:     name,
			IsOnline: isOnline,
		})
	}
	friendPkt := serverpackets.NewFriendListPacket(friendInfos)
	friendData, err := friendPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing FriendList: %w", err)
	}
	n = copy(buf[totalBytes:], friendData)
	if n != len(friendData) {
		return 0, false, fmt.Errorf("buffer too small for FriendList")
	}
	totalBytes += n

	// 6. SkillList (player's learned skills)
	skills := serverpackets.NewSkillList(player.Skills())
	skillData, err := skills.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SkillList: %w", err)
	}
	n = copy(buf[totalBytes:], skillData)
	if n != len(skillData) {
		return 0, false, fmt.Errorf("buffer too small for SkillList")
	}
	totalBytes += n

	// 6. QuestList (Phase 16: real quest data)
	questEntries := h.buildQuestListEntries(player)
	quests := serverpackets.NewQuestListWithEntries(questEntries)
	questData, err := quests.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing QuestList: %w", err)
	}
	n = copy(buf[totalBytes:], questData)
	if n != len(questData) {
		return 0, false, fmt.Errorf("buffer too small for QuestList")
	}
	totalBytes += n

	// 8. ExAutoSoulShot (Phase 36: restore auto-soulshot toggles)
	for _, shotID := range player.AutoSoulShots() {
		shotPkt := &serverpackets.ExAutoSoulShot{ItemID: shotID, Type: 1}
		shotData, shotErr := shotPkt.Write()
		if shotErr != nil {
			slog.Error("serializing ExAutoSoulShot", "error", shotErr)
			continue
		}
		n = copy(buf[totalBytes:], shotData)
		totalBytes += n
	}

	// 9. SendMacroList (Phase 36: restore macros)
	macros := player.GetMacros()
	for _, macro := range macros {
		macroPkt := &serverpackets.SendMacroList{
			Revision: player.MacroRevision(),
			Count:    int8(len(macros)),
			Macro:    macro,
		}
		macroData, macroErr := macroPkt.Write()
		if macroErr != nil {
			slog.Error("serializing SendMacroList", "error", macroErr)
			continue
		}
		n = copy(buf[totalBytes:], macroData)
		totalBytes += n
	}

	// 10. SkillCoolTime (Phase 38: sync skill cooldowns at login)
	// Currently sends empty list (cooldowns not persisted yet).
	coolTimePkt := &serverpackets.SkillCoolTime{}
	coolTimeData, err := coolTimePkt.Write()
	if err != nil {
		slog.Error("serializing SkillCoolTime", "error", err)
	} else {
		n = copy(buf[totalBytes:], coolTimeData)
		totalBytes += n
	}

	slog.Debug("sent spawn packets",
		"character", player.Name(),
		"total_bytes", totalBytes,
		"packets", "UserInfo+StatusUpdate+Inventory+Shortcuts+FriendList+Skills+Quests+AutoShots+Macros+CoolTime")

	// Broadcast CharInfo to visible players (Phase 4.8 part 2)
	// This makes the spawned player visible to others
	charInfo := serverpackets.NewCharInfo(player)
	charInfoData, err := charInfo.Write()
	if err != nil {
		slog.Error("failed to serialize CharInfo",
			"character", player.Name(),
			"error", err)
		// Continue даже если broadcast failed (player still spawns)
	} else {
		// Broadcast to all visible players
		visibleCount := h.clientManager.BroadcastToVisible(player, charInfoData, len(charInfoData))
		if visibleCount > 0 {
			slog.Debug("broadcasted CharInfo",
				"character", player.Name(),
				"visible_players", visibleCount)
		}
	}

	// Send CharInfo + NpcInfo TO client for all visible objects (Phase 4.9 Part 2 + Phase 4.10)
	// This makes other players and NPCs visible to the spawned player
	if err := h.sendVisibleObjectsInfo(client, player); err != nil {
		slog.Error("failed to send info for visible objects",
			"character", player.Name(),
			"error", err)
		// Continue даже если некоторые packets failed
	}

	// NpcInfo + ItemOnGround already sent by sendVisibleObjectsInfo above.

	return totalBytes, true, nil
}

// handleMoveToLocation processes the MoveToLocation packet (opcode 0x01).
// Client sends this when player clicks on ground to move.
func (h *Handler) handleMoveToLocation(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseMoveToLocation(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing MoveToLocation: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		slog.Warn("MoveToLocation before entering world",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, true, nil // Ignore silently
	}

	// Get cached player (Phase 4.18 Opt 3)
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("MoveToLocation without active player",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("no active player for account %s", client.AccountName())
	}

	// Phase 5.1: Validate movement (distance, Z-bounds)
	if err := ValidateMoveToLocation(player, pkt.TargetX, pkt.TargetY, pkt.TargetZ); err != nil {
		slog.Warn("movement validation failed",
			"character", player.Name(),
			"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
			"to", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ),
			"error", err)

		// Send ValidateLocation (force client to use server position)
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil // Continue даже если failed
		}
		n := copy(buf, validateData)

		// Broadcast StopMove to visible players (Phase 5.1)
		stopMove := serverpackets.NewStopMove(player)
		stopData, err := stopMove.Write()
		if err != nil {
			slog.Error("failed to serialize StopMove",
				"character", player.Name(),
				"error", err)
		} else {
			// Phase 5.1: Use BroadcastToVisibleNear (LOD optimization, -90% packets)
			h.clientManager.BroadcastToVisibleNear(player, stopData, len(stopData))
		}

		return n, true, nil // Connection stays open
	}

	// Phase 7.1: Geodata validation — check for walls/obstacles
	targetX, targetY, targetZ := pkt.TargetX, pkt.TargetY, pkt.TargetZ
	currentLoc := player.Location()

	geoResult := ValidateMoveWithGeo(h.geoEngine, currentLoc.X, currentLoc.Y, currentLoc.Z,
		targetX, targetY, targetZ)

	if geoResult.Blocked && geoResult.Path != nil && len(geoResult.Path) > 1 {
		// Direct path blocked but A* found alternative — use first waypoint
		targetX = geoResult.Path[1].X
		targetY = geoResult.Path[1].Y
		targetZ = geoResult.Path[1].Z

		slog.Debug("movement rerouted via geodata",
			"character", player.Name(),
			"original", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ),
			"waypoint", fmt.Sprintf("(%d,%d,%d)", targetX, targetY, targetZ))
	} else if geoResult.Blocked && geoResult.Path == nil {
		// No path found — stop player at current position
		slog.Debug("movement blocked by geodata (no path)",
			"character", player.Name(),
			"target", fmt.Sprintf("(%d,%d,%d)", pkt.TargetX, pkt.TargetY, pkt.TargetZ))

		stopMove := serverpackets.NewStopMove(player)
		stopData, err := stopMove.Write()
		if err != nil {
			slog.Error("failed to serialize StopMove",
				"character", player.Name(),
				"error", err)
		} else {
			h.clientManager.BroadcastToVisibleNear(player, stopData, len(stopData))
		}
		return 0, true, nil
	} else {
		// Direct movement OK — correct Z from geodata
		targetZ = geoResult.CorrectedZ
	}

	// Update player location (validated)
	newLoc := model.NewLocation(targetX, targetY, targetZ, player.Location().Heading)
	player.SetLocation(newLoc)

	// Phase 5.1: Track last server-validated position
	player.Movement().SetLastServerPosition(targetX, targetY, targetZ)

	slog.Debug("player moving",
		"character", player.Name(),
		"from", fmt.Sprintf("(%d,%d,%d)", pkt.OriginX, pkt.OriginY, pkt.OriginZ),
		"to", fmt.Sprintf("(%d,%d,%d)", targetX, targetY, targetZ))

	// Broadcast movement to visible players
	movePkt := serverpackets.NewCharMoveToLocation(player, targetX, targetY, targetZ)
	moveData, err := movePkt.Write()
	if err != nil {
		slog.Error("failed to serialize CharMoveToLocation",
			"character", player.Name(),
			"error", err)
		// Continue даже если broadcast failed
	} else {
		// Phase 5.1: Use BroadcastToVisibleNear (LOD optimization, -90% packets)
		visibleCount := h.clientManager.BroadcastToVisibleNear(player, moveData, len(moveData))
		if visibleCount > 0 {
			slog.Debug("broadcasted movement",
				"character", player.Name(),
				"visible_players", visibleCount)
		}
	}

	// No response packet to client (movement is client-predicted)
	return 0, true, nil
}

// handleCannotMoveAnymore processes CannotMoveAnymore packet (C2S opcode 0x36).
// Client sends this when it hits a wall or reaches destination and movement is blocked.
// Server updates player position to the blocked location and corrects Z via geodata.
//
// Java reference: CannotMoveAnymore.java → player.getAI().notifyAction(ARRIVED_BLOCKED, loc)
func (h *Handler) handleCannotMoveAnymore(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseCannotMoveAnymore(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing CannotMoveAnymore: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Correct Z via geodata if available
	correctedZ := pkt.Z
	if h.geoEngine != nil && h.geoEngine.IsLoaded() {
		correctedZ = h.geoEngine.GetHeight(pkt.X, pkt.Y, pkt.Z)
	}

	// Update player position to the blocked location
	newLoc := model.NewLocation(pkt.X, pkt.Y, correctedZ, uint16(pkt.Heading))
	player.SetLocation(newLoc)

	slog.Debug("player movement blocked",
		"character", player.Name(),
		"pos", fmt.Sprintf("(%d,%d,%d)", pkt.X, pkt.Y, correctedZ),
		"heading", pkt.Heading)

	// Broadcast StopMove to visible players
	stopMove := serverpackets.NewStopMove(player)
	stopData, err := stopMove.Write()
	if err != nil {
		slog.Error("failed to serialize StopMove",
			"character", player.Name(),
			"error", err)
	} else {
		h.clientManager.BroadcastToVisibleNear(player, stopData, len(stopData))
	}

	return 0, true, nil
}

// sendVisibleObjectsInfo sends CharInfo + NpcInfo + ItemOnGround packets TO client for all visible objects.
// Phase 4.19: Parallel encryption implementation — encrypts packets in parallel.
// Phase 7.0: Sends via client.Send() (writePump batches and writes).
//
// Uses ForEachVisibleObjectCached for efficient visibility queries.
// Handles Players (CharInfo), NPCs (NpcInfo), and Items (ItemOnGround).
//
// Thread-safety: Encryption is safe after authentication (firstPacket=false).
const maxConcurrent = 20

func (h *Handler) sendVisibleObjectsInfo(client *GameClient, player *model.Player) error {
	// Thread-safe packet collection
	var (
		mu                               sync.Mutex
		lastErr                          error
		wg                               sync.WaitGroup
		playerCount, npcCount, itemCount int
		encryptedPackets                 = make([][]byte, 0, 450)
	)

	// Semaphore to limit concurrent goroutines (avoid goroutine explosion)
	semaphore := make(chan struct{}, maxConcurrent)

	writePool := h.clientManager.writePool

	world.ForEachVisibleObjectCached(player, func(obj *model.WorldObject) bool {
		objectID := obj.ObjectID()

		// Skip self
		if constants.IsPlayerObjectID(objectID) {
			otherClient := h.clientManager.GetClientByObjectID(objectID)
			if otherClient != nil {
				if otherPlayer := otherClient.ActivePlayer(); otherPlayer != nil {
					if otherPlayer.CharacterID() == player.CharacterID() {
						return true // Don't send CharInfo for self
					}
				}
			}
		}

		semaphore <- struct{}{} // Acquire
		wg.Go(func() {
			defer func() { <-semaphore }() // Release

			// Serialize packet based on object type
			var payloadData []byte
			var packetType string
			var err error

			if constants.IsPlayerObjectID(obj.ObjectID()) {
				// This is a Player — send CharInfo
				otherClient := h.clientManager.GetClientByObjectID(obj.ObjectID())
				if otherClient == nil {
					return // Player offline, skip
				}

				otherPlayer := otherClient.ActivePlayer()
				if otherPlayer == nil {
					return // Player not in game yet, skip
				}

				charInfoPkt := serverpackets.NewCharInfo(otherPlayer)
				payloadData, err = charInfoPkt.Write()
				packetType = "CharInfo"

				mu.Lock()
				playerCount++
				mu.Unlock()

			} else if constants.IsNpcObjectID(obj.ObjectID()) {
				// This is an NPC — send NpcInfo
				npc, ok := world.Instance().GetNpc(obj.ObjectID())
				if !ok {
					return // NPC not found or despawned, skip
				}

				npcInfoPkt := serverpackets.NewNpcInfo(npc)
				payloadData, err = npcInfoPkt.Write()
				packetType = "NpcInfo"

				mu.Lock()
				npcCount++
				mu.Unlock()

			} else if constants.IsItemObjectID(obj.ObjectID()) {
				// This is a dropped item — send ItemOnGround
				droppedItem, ok := world.Instance().GetItem(obj.ObjectID())
				if !ok {
					return // Item not found or picked up, skip
				}

				itemOnGroundPkt := serverpackets.NewItemOnGround(droppedItem)
				payloadData, err = itemOnGroundPkt.Write()
				packetType = "ItemOnGround"

				mu.Lock()
				itemCount++
				mu.Unlock()

			} else {
				return // Unknown object type, skip
			}

			if err != nil {
				slog.Error("failed to serialize packet",
					"packet_type", packetType,
					"object_id", obj.ObjectID(),
					"error", err)
				mu.Lock()
				if lastErr == nil {
					lastErr = err
				}
				mu.Unlock()
				return
			}

			// Encrypt into pool buffer (zero-alloc in steady state)
			var encPkt []byte
			if writePool != nil {
				encPkt, err = writePool.EncryptToPooled(client.Encryption(), payloadData, len(payloadData))
			} else {
				// Fallback: allocate buffer (for tests without writePool)
				buf := make([]byte, constants.PacketHeaderSize+len(payloadData)+constants.PacketBufferPadding)
				copy(buf[constants.PacketHeaderSize:], payloadData)
				var encSize int
				encSize, err = protocol.EncryptInPlace(client.Encryption(), buf, len(payloadData))
				if err == nil {
					encPkt = buf[:encSize]
				}
			}

			if err != nil {
				slog.Error("failed to encrypt packet",
					"packet_type", packetType,
					"object_id", obj.ObjectID(),
					"error", err)
				mu.Lock()
				if lastErr == nil {
					lastErr = err
				}
				mu.Unlock()
				return
			}

			// Add encrypted packet to collection (mutex-protected)
			mu.Lock()
			encryptedPackets = append(encryptedPackets, encPkt)
			mu.Unlock()
		})

		return true // Continue iteration
	})

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors during packet creation/encryption
	if lastErr != nil {
		return fmt.Errorf("creating visible objects info packets: %w", lastErr)
	}

	// Send all packets via write queue (writePump will batch via drain loop)
	if len(encryptedPackets) > 0 {
		for _, pkt := range encryptedPackets {
			if err := client.Send(pkt); err != nil {
				return fmt.Errorf("queueing visible object packet: %w", err)
			}
		}

		slog.Debug("sent info for visible objects",
			"character", player.Name(),
			"visible_players", playerCount,
			"visible_npcs", npcCount,
			"visible_items", itemCount,
			"total_packets", len(encryptedPackets))
	}

	return nil
}

// handleLogout processes the Logout packet (opcode 0x09).
// Client sends this when user clicks Exit button.
//
// Phase 4.17.5: MVP implementation with basic logout flow.
// Phase 31: Offline trade mode support added.
//
// Reference: L2J_Mobius Logout.java (53-107)
func (h *Handler) handleLogout(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseLogout(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing Logout: %w", err)
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("Logout without active player",
			"account", client.AccountName(),
			"client", client.IP())
		// Close connection даже если player nil
		client.MarkForDisconnection()
		return 0, true, nil
	}

	// Check if can logout (15s combat delay)
	if !player.CanLogout() {
		slog.Info("logout denied (cannot logout)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		// Send SystemMessage "YOU_CANNOT_EXIT_WHILE_IN_COMBAT"
		totalBytes := 0
		sysMsgPkt := serverpackets.NewSystemMessage(serverpackets.SysMsgYouCannotExitWhileInCombat)
		sysMsgData, err := sysMsgPkt.Write()
		if err != nil {
			slog.Error("failed to serialize combat deny SystemMessage", "error", err)
		} else {
			n := copy(buf[totalBytes:], sysMsgData)
			totalBytes += n
		}

		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf[totalBytes:], afData)
		totalBytes += n

		return totalBytes, true, nil
	}

	slog.Info("player logging out",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// Phase 4.17.6: Boss zone removal and Olympiad unregister deferred until boss zone tracking is added.

	// Phase 4.17.6: Instance cleanup deferred until instance system tracks player locations.

	// Phase 31: Offline trade mode — keep player in world as detached trader
	if h.offlineSvc != nil && h.offlineSvc.Enabled() && player.IsInStoreMode() {
		if err := h.offlineSvc.EnteredOfflineMode(ctx, player, player.ObjectID(), client.AccountName()); err != nil {
			slog.Warn("offline trade mode failed on logout, proceeding with normal logout",
				"character", player.Name(),
				"error", err)
		} else {
			// Успешно: player остаётся в мире, TCP закрывается
			client.Detach()
			client.MarkForDisconnection()

			slog.Info("player entered offline trade on logout",
				"character", player.Name(),
				"objectID", player.ObjectID())

			// Отправляем LeaveWorld клиенту, но НЕ удаляем из мира
			leaveWorld := serverpackets.NewLeaveWorld()
			leaveWorldData, err := leaveWorld.Write()
			if err != nil {
				return 0, true, nil
			}
			n := copy(buf, leaveWorldData)
			return n, true, nil
		}
	}

	// Phase 8.1: Close private store on logout (if not entering offline trade)
	if player.IsTrading() {
		player.ClosePrivateStore()
	}

	// Phase 6.0: Save player to DB (location, inventory, skills)
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("failed to save player on logout",
			"character", player.Name(),
			"error", err)
	}

	// Remove from world (Phase 4.17.5)
	world.Instance().RemoveObject(player.ObjectID())

	// Clear active player from client
	client.SetActivePlayer(nil)

	// Send LeaveWorld packet (Phase 4.17.3)
	leaveWorld := serverpackets.NewLeaveWorld()
	leaveWorldData, err := leaveWorld.Write()
	if err != nil {
		slog.Error("failed to serialize LeaveWorld",
			"character", player.Name(),
			"error", err)
		// Continue with disconnect даже если packet failed
		client.MarkForDisconnection()
		return 0, true, nil
	}

	n := copy(buf, leaveWorldData)
	if n != len(leaveWorldData) {
		slog.Error("buffer too small for LeaveWorld",
			"character", player.Name(),
			"size", len(leaveWorldData),
			"buffer_size", len(buf))
		// Continue with disconnect
		client.MarkForDisconnection()
		return 0, true, nil
	}

	// Mark client for disconnection (server.go will close TCP after sending LeaveWorld)
	client.MarkForDisconnection()

	slog.Info("player logged out successfully",
		"account", client.AccountName(),
		"character", player.Name())

	return n, true, nil
}

// handleRequestRestart processes the RequestRestart packet (opcode 0x46).
// Client sends this when user clicks "Restart" to return to character selection screen.
// Unlike Logout, RequestRestart does NOT close TCP connection — client returns to char selection.
//
// Phase 4.17.6: MVP implementation with basic restart flow.
// Enchant, class change, and festival checks deferred until those systems are wired.
//
// Reference: L2J_Mobius RequestRestart.java (60-173)
func (h *Handler) handleRequestRestart(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseRequestRestart(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestRestart: %w", err)
	}

	// Verify client is in game
	if client.State() != ClientStateInGame {
		slog.Warn("RequestRestart from non-ingame state",
			"account", client.AccountName(),
			"state", client.State(),
			"client", client.IP())

		// Send denial
		restartResp := serverpackets.NewRestartResponse(false)
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		slog.Warn("RequestRestart without active player",
			"account", client.AccountName(),
			"client", client.IP())

		// Send denial
		restartResp := serverpackets.NewRestartResponse(false)
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Active enchant check deferred: requires enchant session tracking.

	// Class change check deferred: requires class change session tracking.

	// Check if in trade/store mode
	if player.IsTrading() {
		slog.Info("restart denied (trading)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		restartResp := serverpackets.NewRestartResponse(false)
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Check if can logout (includes attack stance check)
	if !player.CanLogout() {
		slog.Info("restart denied (cannot logout)",
			"account", client.AccountName(),
			"character", player.Name(),
			"client", client.IP())

		restartResp := serverpackets.NewRestartResponse(false)
		respData, err := restartResp.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
		}
		copy(buf, respData)
		return len(respData), true, nil
	}

	// Festival participant check deferred: requires Seven Signs Festival event wiring.

	slog.Info("player restarting to character selection",
		"account", client.AccountName(),
		"character", player.Name(),
		"level", player.Level(),
		"client", client.IP())

	// Phase 4.17.7: Boss zone removal and Olympiad unregister deferred until boss zone tracking is added.

	// Phase 4.17.7: Instance cleanup deferred until instance system tracks player locations.

	// Phase 8.1: Close private store on restart
	if player.IsTrading() {
		player.ClosePrivateStore()
	}

	// Phase 6.0: Save player to DB (location, inventory, skills)
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("failed to save player on restart",
			"character", player.Name(),
			"error", err)
	}

	// Remove from world (Phase 4.17.6)
	world.Instance().RemoveObject(player.ObjectID())

	// Clear active player from client
	client.SetActivePlayer(nil)
	client.SetSelectedCharacter(-1)

	// Transition to AUTHENTICATED state (Phase 4.17.6)
	// This allows client to access CharacterSelect, CharacterCreate, CharacterDelete packets
	client.SetState(ClientStateAuthenticated)

	slog.Info("player returned to character selection",
		"account", client.AccountName(),
		"character", player.Name())

	// Send response packets
	var totalBytes int

	// 1. RestartResponse(true) — confirms restart success
	restartResp := serverpackets.NewRestartResponse(true)
	respData, err := restartResp.Write()
	if err != nil {
		slog.Error("failed to serialize RestartResponse",
			"character", player.Name(),
			"error", err)
		return 0, false, fmt.Errorf("serializing RestartResponse: %w", err)
	}
	n := copy(buf[totalBytes:], respData)
	if n != len(respData) {
		return 0, false, fmt.Errorf("buffer too small for RestartResponse")
	}
	totalBytes += n

	// 2. CharSelectionInfo — sends list of characters for account
	// Get SessionKey PlayOkID1 for CharSelectionInfo
	sessionKey := client.SessionKey()
	if sessionKey == nil {
		slog.Error("no session key for authenticated client",
			"account", client.AccountName(),
			"client", client.IP())
		return 0, false, fmt.Errorf("missing session key")
	}

	// Load characters for this account (Phase 4.17.6)
	players, err := h.charRepo.LoadByAccountName(ctx, client.AccountName())
	if err != nil {
		slog.Error("failed to load characters for restart",
			"account", client.AccountName(),
			"error", err)
		return 0, false, fmt.Errorf("loading characters: %w", err)
	}

	charList := serverpackets.NewCharSelectionInfoFromPlayers(client.AccountName(), sessionKey.PlayOkID1, players)
	charListData, err := charList.Write()
	if err != nil {
		slog.Error("failed to serialize CharSelectionInfo",
			"account", client.AccountName(),
			"error", err)
		return 0, false, fmt.Errorf("serializing CharSelectionInfo: %w", err)
	}
	n = copy(buf[totalBytes:], charListData)
	if n != len(charListData) {
		return 0, false, fmt.Errorf("buffer too small for CharSelectionInfo")
	}
	totalBytes += n

	slog.Info("restart completed successfully",
		"account", client.AccountName(),
		"total_bytes", totalBytes)

	return totalBytes, true, nil
}

// handleValidatePosition processes ValidatePosition packet (opcode 0x48).
// Client sends this periodically (~200ms) to report current position.
// Server validates and corrects if desynced.
//
// Phase 5.1: Movement validation — desync detection and correction.
//
// Reference: L2J_Mobius ValidatePosition.java
func (h *Handler) handleValidatePosition(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseValidatePosition(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ValidatePosition: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil // Ignore silently
	}

	// Z-bounds check (prevent flying/underground exploits)
	// Reference: L2J_Mobius ValidatePosition.java:76-82
	if pkt.Z < MinZCoordinate || pkt.Z > MaxZCoordinate {
		slog.Warn("abnormal Z coordinate from client",
			"character", player.Name(),
			"z", pkt.Z,
			"allowed_range", fmt.Sprintf("[%d..%d]", MinZCoordinate, MaxZCoordinate))

		// Teleport player to last server-validated position
		lastX, lastY, lastZ := player.Movement().LastServerPosition()
		player.SetLocation(model.NewLocation(lastX, lastY, lastZ, player.Location().Heading))

		// Send ValidateLocation to force correction
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil
		}

		n := copy(buf, validateData)
		return n, true, nil
	}

	// Update client-reported position
	player.Movement().SetClientPosition(pkt.X, pkt.Y, pkt.Z, pkt.Heading)

	// Check desync between client and server positions
	needsCorrection, diffSq := ValidatePositionDesync(player, pkt.X, pkt.Y, pkt.Z)
	if needsCorrection {
		slog.Info("position desync detected",
			"character", player.Name(),
			"diff_squared", diffSq,
			"client", fmt.Sprintf("(%d,%d,%d)", pkt.X, pkt.Y, pkt.Z),
			"server", fmt.Sprintf("(%d,%d,%d)", player.Location().X, player.Location().Y, player.Location().Z))

		// Send ValidateLocation to correct client
		validateLoc := serverpackets.NewValidateLocation(player)
		validateData, err := validateLoc.Write()
		if err != nil {
			slog.Error("failed to serialize ValidateLocation",
				"character", player.Name(),
				"error", err)
			return 0, true, nil
		}

		n := copy(buf, validateData)
		return n, true, nil
	}

	// Position synchronized, no response needed
	return 0, true, nil
}

// handleRequestAction processes RequestAction packet (opcode 0x04).
// Client sends this when player clicks on an object (target selection or attack intent).
//
// Phase 5.2: Target System.
//
// Reference: L2J_Mobius RequestActionUse.java
func (h *Handler) handleRequestAction(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAction(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAction: %w", err)
	}

	// Verify character is in game
	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	// Get active player
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil // Ignore silently
	}

	worldInst := world.Instance()

	// Phase 35: DroppedItem pickup — clicking on a dropped item picks it up
	// In L2 Interlude, item pickup is done through Action (0x04), NOT a separate opcode.
	if droppedItem, ok := worldInst.GetItem(uint32(pkt.ObjectID)); ok {
		return h.pickupDroppedItem(player, droppedItem, pkt.ObjectID, buf)
	}

	// Validate target selection
	target, err := ValidateTargetSelection(player, uint32(pkt.ObjectID), worldInst)
	if err != nil {
		slog.Debug("target selection failed",
			"character", player.Name(),
			"targetID", pkt.ObjectID,
			"error", err)
		// Silent failure — client will not change target
		return 0, true, nil
	}

	// Set target
	player.SetTarget(target)

	slog.Debug("target selected",
		"character", player.Name(),
		"targetID", target.ObjectID(),
		"targetName", target.Name(),
		"attackIntent", pkt.IsAttackIntent())

	// Prepare response buffer
	totalBytes := 0

	// 1. Send MyTargetSelected (highlight target + show HP bar)
	myTargetSel := serverpackets.NewMyTargetSelected(target.ObjectID())
	targetSelData, err := myTargetSel.Write()
	if err != nil {
		slog.Error("failed to serialize MyTargetSelected",
			"character", player.Name(),
			"error", err)
		return 0, true, nil
	}
	n := copy(buf[totalBytes:], targetSelData)
	totalBytes += n

	// 2. Send StatusUpdate (HP/MP/CP values for target)
	// Check if target is a Character (has HP/MP/CP)
	if character := getCharacterFromObject(target, worldInst, h.clientManager); character != nil {
		statusUpdate := serverpackets.NewStatusUpdateForTarget(character)
		statusData, err := statusUpdate.Write()
		if err != nil {
			slog.Error("failed to serialize StatusUpdate",
				"character", player.Name(),
				"error", err)
		} else {
			n = copy(buf[totalBytes:], statusData)
			totalBytes += n
		}
	}

	// Phase 11: NPC Dialogues — show chat window on simple click for talkable NPC
	if pkt.ActionType == clientpackets.ActionSimpleClick {
		if npc, ok := worldInst.GetNpc(uint32(pkt.ObjectID)); ok {
			npcDef := skilldata.GetNpcDef(npc.TemplateID())
			if npcDef != nil && isNpcTalkable(npcDef.NpcType()) {
				htmlContent := h.buildNpcDialog(npc, player)
				htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), htmlContent)
				htmlData, err := htmlMsg.Write()
				if err != nil {
					slog.Error("failed to serialize NpcHtmlMessage",
						"character", player.Name(),
						"npcID", npc.TemplateID(),
						"error", err)
				} else {
					n = copy(buf[totalBytes:], htmlData)
					totalBytes += n
				}
			}
		}
	}

	return totalBytes, true, nil
}

// pickupDroppedItem handles picking up a dropped item from the world.
// Called from handleRequestAction when the clicked object is a DroppedItem.
func (h *Handler) pickupDroppedItem(player *model.Player, droppedItem *model.DroppedItem, objectID int32, buf []byte) (int, bool, error) {
	// Validate pickup range (200 units max)
	const maxPickupRangeSquared = 200 * 200

	playerLoc := player.Location()
	itemLoc := droppedItem.Location()

	dx := int64(playerLoc.X - itemLoc.X)
	dy := int64(playerLoc.Y - itemLoc.Y)
	distSq := dx*dx + dy*dy

	if distSq > maxPickupRangeSquared {
		slog.Debug("pickup failed: out of range",
			"character", player.Name(),
			"objectID", objectID,
			"distance_sq", distSq)
		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, true, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Get Item from DroppedItem
	item := droppedItem.Item()
	if item == nil {
		slog.Error("pickup failed: DroppedItem has nil item",
			"character", player.Name(),
			"objectID", objectID)
		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Add item to player's inventory
	if err := player.Inventory().AddItem(item); err != nil {
		slog.Error("pickup failed: cannot add to inventory",
			"character", player.Name(),
			"objectID", objectID,
			"itemID", item.ItemID(),
			"error", err)
		actionFailed := serverpackets.NewActionFailed()
		failedData, _ := actionFailed.Write()
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Remove DroppedItem from world
	worldInst := world.Instance()
	worldInst.RemoveObject(uint32(objectID))

	// Broadcast DeleteObject to visible players
	deleteObj := serverpackets.NewDeleteObject(objectID)
	deleteData, err := deleteObj.Write()
	if err != nil {
		slog.Error("failed to serialize DeleteObject for pickup",
			"objectID", objectID,
			"error", err)
	} else {
		h.clientManager.BroadcastToVisible(player, deleteData, len(deleteData))
	}

	slog.Info("item picked up",
		"character", player.Name(),
		"itemID", item.ItemID(),
		"count", item.Count(),
		"objectID", objectID)

	// Send InventoryUpdate to client
	invUpdate := serverpackets.NewInventoryUpdate(serverpackets.InvUpdateEntry{
		ChangeType: serverpackets.InvUpdateAdd,
		Item:       item,
	})
	invData, err := invUpdate.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryUpdate for pickup",
			"character", player.Name(),
			"error", err)
		return 0, true, nil
	}

	n := copy(buf, invData)
	return n, true, nil
}

// getCharacterFromObject attempts to extract Character from WorldObject.
// Returns nil if object is not a Character (e.g., dropped item).
// Uses clientManager to resolve player objectIDs to Character.
func getCharacterFromObject(obj *model.WorldObject, worldInst *world.World, cm *ClientManager) *model.Character {
	objectID := obj.ObjectID()

	// Check if it's an NPC
	if npc, ok := worldInst.GetNpc(objectID); ok {
		return npc.Character
	}

	// Check if it's a Player — look up via ClientManager
	if constants.IsPlayerObjectID(objectID) && cm != nil {
		if otherClient := cm.GetClientByObjectID(objectID); otherClient != nil {
			if otherPlayer := otherClient.ActivePlayer(); otherPlayer != nil {
				return otherPlayer.Character
			}
		}
		return nil
	}

	return nil
}

// handleAttackRequest processes AttackRequest packet (opcode 0x0A).
// Client sends this when player clicks on enemy to initiate auto-attack.
//
// Workflow:
//  1. Validate target exists in world
//  2. Validate attack (range, dead, etc)
//  3. Start auto-attack via player.DoAttack(target)
//
// Phase 5.3: Basic Combat System.
// Java reference: AttackRequest.java (runImpl, line 53-129).
func (h *Handler) handleAttackRequest(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAttackRequest(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AttackRequest: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil // Ignore silently
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Attack speed throttle — prevent attack speed exploits.
	// Minimum interval = 500000 / pAtkSpd ms (Java: Creature.doAttack).
	pAtkSpd := player.GetPAtkSpd()
	if pAtkSpd < 1 {
		pAtkSpd = 1
	}
	minIntervalMs := int64(500000.0 / pAtkSpd)
	if minIntervalMs < 100 {
		minIntervalMs = 100 // absolute minimum 100ms
	}
	lastAtk := player.LastAttackTime()
	if lastAtk > 0 {
		elapsed := time.Since(time.Unix(0, lastAtk)).Milliseconds()
		if elapsed < minIntervalMs {
			return 0, true, nil // silently ignore — too fast
		}
	}

	// Get target from world
	worldInst := world.Instance()
	target, exists := worldInst.GetObject(pkt.ObjectID)
	if !exists {
		// Target not found — send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Validate attack (range, dead, etc)
	if err := combat.ValidateAttack(player, target); err != nil {
		slog.Warn("attack validation failed",
			"character", player.Name(),
			"target", target.ObjectID(),
			"error", err)

		// Send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	// Phase 5.6: PvP + PvE combat (Player vs Player/NPC)
	// ExecuteAttack handles type assertion internally
	if combat.CombatMgr != nil {
		combat.CombatMgr.ExecuteAttack(player, target)
	}

	// No response to client (Attack packet sent via broadcast)
	return 0, true, nil
}

// NOTE: handleRequestPickup was removed — item pickup is now handled in
// handleRequestAction via pickupDroppedItem(). In L2 Interlude, there is no
// separate RequestPickup opcode; pickup is done through the Action packet (0x04)
// when the target object is a DroppedItem.

// handleRequestMagicSkillUse processes RequestMagicSkillUse packet (opcode 0x2F).
// Client sends this when player uses a skill from the skill bar.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: RequestMagicSkillUse.java
func (h *Handler) handleRequestMagicSkillUse(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestMagicSkillUse(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestMagicSkillUse: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if skill.CastMgr == nil {
		slog.Warn("CastManager not initialized, ignoring skill use")
		return 0, true, nil
	}

	if err := skill.CastMgr.UseMagic(player, pkt.SkillID, pkt.CtrlPressed, pkt.ShiftPressed); err != nil {
		slog.Debug("skill use failed",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"error", err)

		// Send ActionFailed
		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	return 0, true, nil
}

// handleSay2 processes the Say2 packet (opcode 0x38).
// Client sends this when player types a chat message.
//
// Phase 5.11: Chat System.
// Channels supported: GENERAL (radius), SHOUT (all), WHISPER (1 player), TRADE (all).
// Java reference: Say2.java, CreatureSay.java.
func (h *Handler) handleSay2(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSay2(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing Say2: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	chatType := ChatType(pkt.ChatType)

	// Validate chat type
	if !chatType.IsValid() {
		slog.Warn("invalid chat type",
			"character", player.Name(),
			"chatType", pkt.ChatType,
			"client", client.IP())
		return 0, false, nil // disconnect
	}

	// Validate empty message
	if len(pkt.Text) == 0 {
		slog.Warn("empty chat message",
			"character", player.Name(),
			"chatType", pkt.ChatType,
			"client", client.IP())
		return 0, false, nil // disconnect
	}

	// Phase 17: Intercept admin commands (//) and user commands (/)
	// Admin commands start with "//", user commands with "/"
	// Must check before message length validation (GM commands can be longer)
	if h.adminHandler != nil && chatType == ChatGeneral {
		if strings.HasPrefix(pkt.Text, "//") {
			// Admin command
			cmdText := pkt.Text[2:]
			h.adminHandler.HandleAdminCommand(player, cmdText)
			return h.sendAdminResponse(client, player, buf)
		}
		if strings.HasPrefix(pkt.Text, "/") && !strings.HasPrefix(pkt.Text, "//") {
			// User command
			cmdText := pkt.Text[1:]
			if h.adminHandler.HandleUserCommand(player, cmdText) {
				return h.sendAdminResponse(client, player, buf)
			}
			// If user command not found, fall through to normal chat
		}
	}

	// Validate message length (max 105 chars for non-GM)
	if len([]rune(pkt.Text)) > MaxMessageLength && !player.IsGM() {
		slog.Info("chat message too long",
			"character", player.Name(),
			"length", len([]rune(pkt.Text)),
			"max", MaxMessageLength)

		// Send system message: exceeded chat text limit
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgYouHaveExceededTheChatTextLimit)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	// Route by chat type
	switch chatType {
	case ChatGeneral:
		return h.handleChatGeneral(client, player, pkt.Text, buf)
	case ChatShout:
		return h.handleChatShout(player, pkt.Text, buf)
	case ChatWhisper:
		return h.handleChatWhisper(client, player, pkt.Text, pkt.Target, buf)
	case ChatTrade:
		return h.handleChatTrade(player, pkt.Text, buf)
	default:
		slog.Warn("unsupported chat type",
			"character", player.Name(),
			"chatType", pkt.ChatType)
		return 0, true, nil
	}
}

// sendAdminResponse reads the pending admin message from the player
// and sends it back to the client as a CreatureSay packet.
// If the message starts with "ANNOUNCE:", it broadcasts to all players.
//
// Phase 17: Admin Commands.
func (h *Handler) sendAdminResponse(_ *GameClient, player *model.Player, buf []byte) (int, bool, error) {
	msg := player.ClearLastAdminMessage()
	if msg == "" {
		return 0, true, nil
	}

	// Announce: broadcast to all players via ChatAnnounce channel
	if strings.HasPrefix(msg, "ANNOUNCE:") {
		text := strings.TrimPrefix(msg, "ANNOUNCE:")
		say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatAnnounce), player.Name(), text)
		sayData, err := say.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing CreatureSay ANNOUNCE: %w", err)
		}
		h.clientManager.BroadcastToAll(sayData, len(sayData))
		n := copy(buf, sayData)
		return n, true, nil
	}

	// Normal admin response: send to command issuer only via PETITION_GM channel (type 7)
	say := serverpackets.NewCreatureSay(0, int32(ChatPetitionGM), "System", msg)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing admin response: %w", err)
	}
	n := copy(buf, sayData)
	return n, true, nil
}

// handleChatGeneral broadcasts a GENERAL message to nearby visible players.
// Radius is LODNear (~1250 units, same region).
func (h *Handler) handleChatGeneral(client *GameClient, player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatGeneral), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay GENERAL: %w", err)
	}

	// Send to sender
	n := copy(buf, sayData)

	// Broadcast to nearby visible players
	h.clientManager.BroadcastToVisibleNear(player, sayData, len(sayData))

	return n, true, nil
}

// handleChatShout broadcasts a SHOUT message to all connected players.
func (h *Handler) handleChatShout(player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatShout), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay SHOUT: %w", err)
	}

	// Send to sender (included in BroadcastToAll but also return in response buffer)
	n := copy(buf, sayData)

	// Broadcast to all players
	h.clientManager.BroadcastToAll(sayData, len(sayData))

	return n, true, nil
}

// handleChatWhisper sends a WHISPER message to a specific player by name.
func (h *Handler) handleChatWhisper(senderClient *GameClient, sender *model.Player, text, targetName string, buf []byte) (int, bool, error) {
	if targetName == "" {
		return 0, true, nil
	}

	targetClient := h.clientManager.FindClientByPlayerName(targetName)
	if targetClient == nil {
		// Target not found — send system message
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgTargetIsNotFound).AddString(targetName)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgTargetIsNotFound).AddString(targetName)
		sysMsgData, err := sysMsg.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing SystemMessage: %w", err)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	// Send message to target
	sayToTarget := serverpackets.NewCreatureSay(int32(sender.ObjectID()), int32(ChatWhisper), sender.Name(), text)
	sayToTargetData, err := sayToTarget.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay WHISPER: %w", err)
	}

	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), sayToTargetData, len(sayToTargetData)); err != nil {
		slog.Warn("failed to send whisper to target",
			"sender", sender.Name(),
			"target", targetName,
			"error", err)
	}

	// Echo to sender: "-> targetName: text"
	sayToSender := serverpackets.NewCreatureSay(int32(sender.ObjectID()), int32(ChatWhisper), sender.Name(), "->"+targetPlayer.Name()+": "+text)
	sayToSenderData, err := sayToSender.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay WHISPER echo: %w", err)
	}

	n := copy(buf, sayToSenderData)
	return n, true, nil
}

// handleChatTrade broadcasts a TRADE message to all connected players.
func (h *Handler) handleChatTrade(player *model.Player, text string, buf []byte) (int, bool, error) {
	say := serverpackets.NewCreatureSay(int32(player.ObjectID()), int32(ChatTrade), player.Name(), text)
	sayData, err := say.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing CreatureSay TRADE: %w", err)
	}

	// Send to sender
	n := copy(buf, sayData)

	// Broadcast to all players
	h.clientManager.BroadcastToAll(sayData, len(sayData))

	return n, true, nil
}

// --- Phase 8: NPC Interaction ---

// NPC interaction distance limit (game units).
const maxNpcInteractionDistance = 150

// maxNpcInteractionDistanceSquared is squared for performance (avoid sqrt).
const maxNpcInteractionDistanceSquared = maxNpcInteractionDistance * maxNpcInteractionDistance

// isNpcTalkable returns true for NPC types that can show dialog.
// Phase 8.2: NPC Dialogues.
func isNpcTalkable(npcType string) bool {
	switch npcType {
	case "folk", "merchant", "guard", "teleporter", "warehouse":
		return true
	default:
		return false
	}
}

// buildNpcDialog builds HTML dialog for NPC using the DialogManager.
// Resolves template by NPC type/ID with fallback chain.
//
// Phase 11: NPC Dialog System.
func (h *Handler) buildNpcDialog(npc *model.Npc, player *model.Player) string {
	templateID := npc.TemplateID()
	npcDef := skilldata.GetNpcDef(templateID)

	npcType := ""
	if npcDef != nil {
		npcType = npcDef.NpcType()
	}

	data := h.buildDialogData(npc, npcDef, player)

	if h.dialogManager == nil {
		return h.buildNpcDefaultHtmlFallback(npc)
	}

	result, err := h.dialogManager.GetNpcDialog(npcType, templateID, data)
	if err != nil {
		slog.Warn("failed to get NPC dialog", "npcID", templateID, "error", err)
		return h.dialogManager.FallbackHTML(data)
	}
	return result
}

// buildDialogData creates a DialogData map with standard NPC dialog variables.
func (h *Handler) buildDialogData(npc *model.Npc, npcDef interface{ Name() string }, player *model.Player) html.DialogData {
	npcName := "NPC"
	if npcDef != nil {
		npcName = npcDef.Name()
	}

	return html.DialogData{
		"objectId": strconv.FormatUint(uint64(npc.ObjectID()), 10),
		"npcname":  npcName,
		"npc_name": npcName,
		"name":     player.Name(),
		"player":   player.Name(),
	}
}

// buildNpcDefaultHtmlFallback is a hardcoded fallback when DialogManager is nil.
// Used in tests where DialogManager is not wired.
func (h *Handler) buildNpcDefaultHtmlFallback(npc *model.Npc) string {
	templateID := npc.TemplateID()
	npcDef := skilldata.GetNpcDef(templateID)
	if npcDef == nil {
		return "<html><body>I have nothing to say.</body></html>"
	}

	var sb strings.Builder
	sb.WriteString("<html><body>")
	sb.WriteString(npcDef.Name())
	sb.WriteString(":<br>")

	npcType := npcDef.NpcType()

	if buylists := skilldata.GetBuylistsByNpc(templateID); len(buylists) > 0 {
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(strconv.FormatUint(uint64(npc.ObjectID()), 10))
		sb.WriteString("_Shop\">Shop</a><br>")
	}

	if npcType == "merchant" {
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(strconv.FormatUint(uint64(npc.ObjectID()), 10))
		sb.WriteString("_Sell\">Sell</a><br>")
	}

	if npcType == "teleporter" && skilldata.HasTeleporter(templateID) {
		objIDStr := strconv.FormatUint(uint64(npc.ObjectID()), 10)
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(objIDStr)
		sb.WriteString("_Teleport NORMAL\">Teleport</a><br>")
	}

	if npcType == "warehouse" {
		objIDStr := strconv.FormatUint(uint64(npc.ObjectID()), 10)
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(objIDStr)
		sb.WriteString("_DepositP\">Deposit Items</a><br>")
		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(objIDStr)
		sb.WriteString("_WithdrawP\">Withdraw Items</a><br>")
	}

	sb.WriteString("</body></html>")
	return sb.String()
}

// handleNpcChat handles "Chat N" bypass — shows dialog page N.
//
// Phase 11: NPC Dialog System.
func (h *Handler) handleNpcChat(player *model.Player, npc *model.Npc, arg string, buf []byte) (int, bool, error) {
	pageNum, _ := strconv.Atoi(arg)

	npcDef := skilldata.GetNpcDef(npc.TemplateID())
	npcType := ""
	npcName := "NPC"
	if npcDef != nil {
		npcType = npcDef.NpcType()
		npcName = npcDef.Name()
	}

	data := h.buildDialogData(npc, npcDef, player)

	var content string
	if h.dialogManager != nil {
		var err error
		content, err = h.dialogManager.GetDialogPage(npcType, npc.TemplateID(), pageNum, data)
		if err != nil {
			content = h.dialogManager.FallbackHTML(data)
		}
	} else {
		content = "<html><body>" + npcName + ":<br>I have nothing more to say.<br></body></html>"
	}

	htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), content)
	msgData, err := htmlMsg.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing NpcHtmlMessage: %w", err)
	}

	n := copy(buf, msgData)
	return n, true, nil
}

// handleNpcTeleport handles "Teleport" NPC bypass command.
// Two sub-commands:
//   - "Teleport <type>" — show teleport location list (HTML)
//   - "Teleport <type> <index>" — execute teleportation
//
// Types: NORMAL, NOBLES_TOKEN, NOBLES_ADENA
// Java reference: Teleporter.java, TeleportHolder.java
//
// Phase 12: Teleporter System.
func (h *Handler) handleNpcTeleport(player *model.Player, npc *model.Npc, arg string, buf []byte) (int, bool, error) {
	templateID := npc.TemplateID()

	parts := strings.Fields(arg)
	if len(parts) == 0 {
		// Без аргументов — показать NORMAL список
		return h.showTeleportList(player, npc, templateID, "NORMAL", buf)
	}

	teleType := parts[0]
	if len(parts) == 1 {
		// "Teleport NORMAL" — показать список
		return h.showTeleportList(player, npc, templateID, teleType, buf)
	}

	// "Teleport NORMAL 3" — выполнить телепортацию
	locIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		slog.Warn("invalid teleport location index",
			"arg", arg,
			"character", player.Name())
		return 0, true, nil
	}

	return h.doTeleport(player, npc, templateID, teleType, locIndex, buf)
}

// showTeleportList generates HTML with teleport destinations and sends NpcHtmlMessage.
//
// Phase 12: Teleporter System.
func (h *Handler) showTeleportList(player *model.Player, npc *model.Npc, npcID int32, teleType string, buf []byte) (int, bool, error) {
	group := skilldata.GetTeleportGroupByType(npcID, teleType)
	if group == nil {
		slog.Debug("no teleport group for NPC",
			"npcID", npcID,
			"type", teleType)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	objIDStr := strconv.FormatUint(uint64(npc.ObjectID()), 10)

	var sb strings.Builder
	sb.WriteString("<html><body>Select a destination:<br><br>")

	for _, loc := range group.Locations {
		displayName := loc.Name
		if displayName == "" {
			continue // пропускаем unnamed OTHER-type locations
		}

		// Показать стоимость
		if loc.FeeCount > 0 {
			feeLabel := "Adena"
			if loc.FeeID == 6651 {
				feeLabel = "Noblesse Token"
			} else if loc.FeeID > 0 {
				feeLabel = fmt.Sprintf("Item #%d", loc.FeeID)
			}
			displayName += fmt.Sprintf(" - %d %s", loc.FeeCount, feeLabel)
		}

		sb.WriteString("<a action=\"bypass -h npc_")
		sb.WriteString(objIDStr)
		sb.WriteString("_Teleport ")
		sb.WriteString(teleType)
		sb.WriteString(" ")
		sb.WriteString(strconv.Itoa(loc.Index))
		sb.WriteString("\">")
		sb.WriteString(displayName)
		sb.WriteString("</a><br1>")
	}

	sb.WriteString("</body></html>")

	htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), sb.String())
	msgData, err := htmlMsg.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing teleport list: %w", err)
	}

	slog.Debug("sent teleport list",
		"character", player.Name(),
		"npcID", npcID,
		"type", teleType,
		"locations", len(group.Locations))

	n := copy(buf, msgData)
	return n, true, nil
}

// doTeleport executes the actual teleportation after fee validation.
//
// Phase 12: Teleporter System.
func (h *Handler) doTeleport(player *model.Player, npc *model.Npc, npcID int32, teleType string, locIndex int, buf []byte) (int, bool, error) {
	loc := skilldata.GetTeleportLocation(npcID, teleType, locIndex)
	if loc == nil {
		slog.Warn("unknown teleport location",
			"npcID", npcID,
			"type", teleType,
			"index", locIndex,
			"character", player.Name())
		return 0, true, nil
	}

	// Проверка оплаты
	feeID := loc.FeeID
	feeCount := int64(loc.FeeCount)
	if feeCount > 0 {
		if feeID == 0 {
			// Adena (default)
			adena := player.Inventory().GetAdena()
			if adena < feeCount {
				slog.Debug("not enough adena for teleport",
					"character", player.Name(),
					"have", adena,
					"need", feeCount)
				// Отправляем системное сообщение "Not enough Adena" через HTML
				htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()),
					"<html><body>You do not have enough Adena.</body></html>")
				msgData, _ := htmlMsg.Write()
				n := copy(buf, msgData)
				return n, true, nil
			}
			if err := player.Inventory().RemoveAdena(int32(feeCount)); err != nil {
				slog.Error("remove adena for teleport",
					"error", err,
					"character", player.Name())
				return 0, true, nil
			}
		} else {
			// Оплата специальным предметом (напр. Noblesse Token 6651)
			itemCount := player.Inventory().CountItemsByID(feeID)
			if itemCount < feeCount {
				slog.Debug("not enough fee items for teleport",
					"character", player.Name(),
					"feeID", feeID,
					"have", itemCount,
					"need", feeCount)
				htmlMsg := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()),
					"<html><body>You do not have enough items to pay the fee.</body></html>")
				msgData, _ := htmlMsg.Write()
				n := copy(buf, msgData)
				return n, true, nil
			}
			removed := player.Inventory().RemoveItemsByID(feeID, feeCount)
			if removed < feeCount {
				slog.Error("partial fee item removal",
					"feeID", feeID,
					"removed", removed,
					"needed", feeCount,
					"character", player.Name())
				return 0, true, nil
			}
		}
	}

	// Телепортация — обновить координаты и отправить пакет
	newLoc := model.NewLocation(loc.X, loc.Y, loc.Z+5, 0)
	player.WorldObject.SetLocation(newLoc)

	telePkt := serverpackets.NewTeleportToLocation(int32(player.ObjectID()), loc.X, loc.Y, loc.Z+5)
	pktData, err := telePkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing TeleportToLocation: %w", err)
	}

	slog.Info("player teleported",
		"character", player.Name(),
		"destination", loc.Name,
		"x", loc.X, "y", loc.Y, "z", loc.Z,
		"fee", feeCount,
		"npcID", npcID)

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestBypassToServer processes RequestBypassToServer packet (opcode 0x21).
// Client sends this when player clicks a link in NPC HTML dialog.
//
// Bypass routing:
//   - "npc_%objectId%_Shop" → send BuyList
//   - "npc_%objectId%_Sell" → send SellList
//   - "npc_%objectId%_Teleport" → teleport list or execute teleport
//   - "_bbshome", "_bbsgetfav" → Community Board (Phase 11+)
//
// Phase 8.2: NPC Dialogues.
func (h *Handler) handleRequestBypassToServer(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBypassToServer(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestBypassToServer: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	bypass := pkt.Bypass
	slog.Debug("bypass received", "character", player.Name(), "bypass", bypass)

	// Route NPC bypass commands: "npc_<objectID>_<command>"
	if strings.HasPrefix(bypass, "npc_") {
		return h.handleNpcBypass(player, bypass, buf)
	}

	// Community Board bypass (_bbs*, bbs_add_fav)
	if strings.HasPrefix(bypass, "_bbs") || strings.HasPrefix(bypass, "bbs_") {
		return h.handleBBSBypass(client, player, bypass, buf)
	}

	// Phase 26: Instance zone bypass commands ("EnterInstance <id>" / "LeaveInstance").
	if strings.HasPrefix(bypass, "EnterInstance") || bypass == "LeaveInstance" {
		parts := strings.Fields(bypass)
		cmdName := parts[0]
		args := parts[1:]
		n, handled, err := h.handleInstanceBypass(client, cmdName, args, buf)
		if handled {
			return n, true, err
		}
	}

	slog.Warn("unknown bypass command", "bypass", bypass, "character", player.Name())
	return 0, true, nil
}

// handleNpcBypass routes NPC-specific bypass commands.
// Format: "npc_<objectID>_<command>"
//
// Phase 8.2/8.3: NPC Dialogues + Shops.
func (h *Handler) handleNpcBypass(player *model.Player, bypass string, buf []byte) (int, bool, error) {
	// Parse: "npc_<objectID>_<command>"
	parts := strings.SplitN(bypass, "_", 3)
	if len(parts) < 3 {
		slog.Warn("malformed npc bypass", "bypass", bypass)
		return 0, true, nil
	}

	npcObjectID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		slog.Warn("invalid npc objectID in bypass", "bypass", bypass, "error", err)
		return 0, true, nil
	}

	command := parts[2]
	worldInst := world.Instance()

	// Validate NPC exists and is within interaction distance
	npc, ok := worldInst.GetNpc(uint32(npcObjectID))
	if !ok {
		slog.Warn("bypass target NPC not found", "objectID", npcObjectID)
		return 0, true, nil
	}

	playerLoc := player.Location()
	npcLoc := npc.Location()
	distSq := playerLoc.DistanceSquared(npcLoc)
	if distSq > maxNpcInteractionDistanceSquared {
		slog.Debug("NPC too far for bypass interaction",
			"character", player.Name(),
			"npcID", npc.TemplateID(),
			"distSq", distSq)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Commands may have arguments after space: "Multisell 123"
	cmdParts := strings.SplitN(command, " ", 2)
	cmdName := cmdParts[0]
	var cmdArg string
	if len(cmdParts) > 1 {
		cmdArg = cmdParts[1]
	}

	switch cmdName {
	case "Shop":
		return h.handleNpcShop(player, npc, buf)
	case "Sell":
		return h.handleNpcSell(player, npc, buf)
	case "DepositP":
		return h.handleNpcWarehouseDeposit(player, npc, buf)
	case "WithdrawP":
		return h.handleNpcWarehouseWithdraw(player, npc, buf)
	case "DepositC":
		return h.handleNpcClanWarehouseDeposit(player, npc, buf)
	case "WithdrawC":
		return h.handleNpcClanWarehouseWithdraw(player, npc, buf)
	case "Multisell":
		return h.handleNpcMultisell(player, npc, cmdArg, buf)
	case "Chat":
		return h.handleNpcChat(player, npc, cmdArg, buf)
	case "Teleport":
		return h.handleNpcTeleport(player, npc, cmdArg, buf)
	case "Quest":
		return h.handleNpcQuestBypass(player, npc, cmdArg, buf)
	case "Subclass":
		return h.handleNpcSubclass(player, npc, cmdArg, buf)
	case "SSQJoin", "SSQSeal", "SSQContribute", "SSQStatus":
		return h.handleSevenSignsBypass(player, npc, cmdName, cmdArg, buf)
	case "Augment", "AugmentCancel":
		return h.handleAugmentBypass(cmdName, buf)
	case "SkillList":
		return h.handleNpcSkillList(player, npc, buf)
	case "create_ally":
		return h.handleNpcCreateAlly(player, cmdArg, buf)
	default:
		slog.Debug("unhandled NPC bypass command",
			"command", command,
			"npcID", npc.TemplateID(),
			"character", player.Name())
		return 0, true, nil
	}
}

// handleNpcShop sends BuyList packet for NPC's shop.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleNpcShop(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	templateID := npc.TemplateID()

	buylistIDs := skilldata.GetBuylistsByNpc(templateID)
	if len(buylistIDs) == 0 {
		slog.Debug("NPC has no buylists", "npcID", templateID)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Use the first buylist for this NPC
	listID := buylistIDs[0]

	// Build products list
	products := buildBuyListProducts(listID)

	playerAdena := player.Inventory().GetAdena()
	buyListPkt := serverpackets.NewBuyList(playerAdena, listID, products)

	pktData, err := buyListPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing BuyList: %w", err)
	}

	slog.Debug("sent BuyList",
		"character", player.Name(),
		"npcID", templateID,
		"listID", listID,
		"products", len(products))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleNpcSell sends SellList packet with player's sellable items.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleNpcSell(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	sellableItems := player.Inventory().GetSellableItems()

	var items []serverpackets.SellListItem
	for _, item := range sellableItems {
		itemDef := skilldata.GetItemDef(item.ItemID())
		sellPrice := int64(0)
		if itemDef != nil {
			sellPrice = itemDef.Price() / 2 // Sell at 50% of base price
		}
		if sellPrice <= 0 {
			continue // Skip items with no sell value
		}

		items = append(items, serverpackets.SellListItem{
			Item:      item,
			SellPrice: sellPrice,
		})
	}

	playerAdena := player.Inventory().GetAdena()
	sellListPkt := serverpackets.NewSellList(playerAdena, items)

	pktData, err := sellListPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SellList: %w", err)
	}

	slog.Debug("sent SellList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"items", len(items))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestBuyItem processes RequestBuyItem packet (opcode 0x1F).
// Player confirms purchase of items from NPC shop.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleRequestBuyItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBuyItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestBuyItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Validate buylist exists
	if skilldata.GetBuylistProducts(pkt.ListID) == nil {
		slog.Warn("invalid buylist ID", "listID", pkt.ListID, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Calculate total cost and validate items
	var totalCost int64
	for _, entry := range pkt.Items {
		product := skilldata.FindProductInBuylist(pkt.ListID, entry.ItemID)
		if product == nil {
			slog.Warn("item not in buylist",
				"itemID", entry.ItemID,
				"listID", pkt.ListID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		price := product.Price
		if price <= 0 {
			// Use item's base price from item data
			if itemDef := skilldata.GetItemDef(entry.ItemID); itemDef != nil {
				price = itemDef.Price()
			}
		}

		totalCost += price * int64(entry.Count)
	}

	// Check Adena
	playerAdena := player.Inventory().GetAdena()
	if playerAdena < totalCost {
		slog.Debug("not enough adena for purchase",
			"character", player.Name(),
			"have", playerAdena,
			"need", totalCost)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Deduct Adena
	if err := player.Inventory().RemoveAdena(int32(totalCost)); err != nil {
		slog.Error("failed to remove adena", "error", err, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Create items
	for _, entry := range pkt.Items {
		tmpl := db.ItemDefToTemplate(entry.ItemID)
		if tmpl == nil {
			slog.Error("item template not found", "itemID", entry.ItemID)
			continue
		}

		objectID := world.IDGenerator().NextItemID()
		item, err := model.NewItem(objectID, entry.ItemID, int64(player.CharacterID()), entry.Count, tmpl)
		if err != nil {
			slog.Error("failed to create item",
				"itemID", entry.ItemID,
				"error", err)
			continue
		}

		if err := player.Inventory().AddItem(item); err != nil {
			slog.Error("failed to add item to inventory",
				"itemID", entry.ItemID,
				"error", err)
			continue
		}

		slog.Debug("item purchased",
			"character", player.Name(),
			"itemID", entry.ItemID,
			"count", entry.Count)
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("purchase completed",
		"character", player.Name(),
		"items", len(pkt.Items),
		"totalCost", totalCost)

	return totalBytes, true, nil
}

// handleRequestSellItem processes RequestSellItem packet (opcode 0x1E).
// Player confirms selling items to NPC.
//
// Phase 8.3: NPC Shops.
func (h *Handler) handleRequestSellItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSellItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSellItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Calculate total sell value and validate items
	var totalValue int64
	type sellEntry struct {
		item      *model.Item
		sellPrice int64
		count     int32
	}
	var entries []sellEntry

	for _, entry := range pkt.Items {
		item := player.Inventory().GetItem(uint32(entry.ObjectID))
		if item == nil {
			slog.Warn("sell: item not in inventory",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Cannot sell equipped items
		if item.IsEquipped() {
			slog.Debug("sell: cannot sell equipped item",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Cannot sell Adena
		if item.ItemID() == model.AdenaItemID {
			continue
		}

		// Validate count
		if entry.Count > item.Count() {
			slog.Warn("sell: not enough items",
				"objectID", entry.ObjectID,
				"have", item.Count(),
				"want", entry.Count,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Calculate sell price
		itemDef := skilldata.GetItemDef(item.ItemID())
		sellPrice := int64(0)
		if itemDef != nil {
			sellPrice = itemDef.Price() / 2
		}

		totalValue += sellPrice * int64(entry.Count)
		entries = append(entries, sellEntry{
			item:      item,
			sellPrice: sellPrice,
			count:     entry.Count,
		})
	}

	// Process sales
	for _, se := range entries {
		if se.count >= se.item.Count() {
			// Remove entire item
			player.Inventory().RemoveItem(se.item.ObjectID())
		} else {
			// Decrease count (stackable items)
			if err := se.item.SetCount(se.item.Count() - se.count); err != nil {
				slog.Error("failed to decrease item count", "error", err)
			}
		}
	}

	// Add Adena
	if totalValue > 0 {
		if err := player.Inventory().AddAdena(int32(totalValue)); err != nil {
			// If no Adena item exists yet, create one
			tmpl := db.ItemDefToTemplate(model.AdenaItemID)
			if tmpl != nil {
				objectID := world.IDGenerator().NextItemID()
				adenaItem, err := model.NewItem(objectID, model.AdenaItemID, int64(player.CharacterID()), int32(totalValue), tmpl)
				if err != nil {
					slog.Error("failed to create adena item", "error", err)
				} else {
					if err := player.Inventory().AddItem(adenaItem); err != nil {
						slog.Error("failed to add adena to inventory", "error", err)
					}
				}
			}
		}
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("items sold",
		"character", player.Name(),
		"items", len(entries),
		"totalValue", totalValue)

	return totalBytes, true, nil
}

// buildBuyListProducts converts buylist products from data package into
// BuyListProduct slice for the BuyList server packet.
//
// Phase 8.3: NPC Shops.
func buildBuyListProducts(listID int32) []serverpackets.BuyListProduct {
	dataProducts := skilldata.GetBuylistProducts(listID)
	if dataProducts == nil {
		return nil
	}

	products := make([]serverpackets.BuyListProduct, 0, len(dataProducts))

	for _, dp := range dataProducts {
		itemDef := skilldata.GetItemDef(dp.ItemID)
		if itemDef == nil {
			continue
		}

		price := dp.Price
		if price <= 0 {
			price = itemDef.Price()
		}

		tmpl := db.ItemDefToTemplate(dp.ItemID)
		var type1, type2 int16
		var bodyPart int32
		if tmpl != nil {
			type1 = tmpl.Type1
			type2 = tmpl.Type2
			bodyPart = tmpl.BodyPartMask
		}

		products = append(products, serverpackets.BuyListProduct{
			ItemID:       dp.ItemID,
			Price:        price,
			Count:        dp.Count,
			RestockDelay: dp.RestockDelay,
			Type1:        type1,
			Type2:        type2,
			BodyPart:     bodyPart,
			Weight:       itemDef.Weight(),
		})
	}

	return products
}


// handleNpcWarehouseDeposit sends WareHouseDepositList showing items player can deposit.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleNpcWarehouseDeposit(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	// NPC distance validation — prevent remote warehouse access.
	if player.Location().DistanceSquared(npc.Location()) > maxNpcInteractionDistanceSquared {
		slog.Debug("warehouse deposit: NPC too far", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	depositableItems := player.Inventory().GetDepositableItems()

	playerAdena := player.Inventory().GetAdena()
	pkt := serverpackets.NewWareHouseDepositList(
		serverpackets.WarehouseTypePrivate,
		playerAdena,
		depositableItems,
	)

	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing WareHouseDepositList: %w", err)
	}

	slog.Debug("sent WareHouseDepositList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"items", len(depositableItems))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleNpcWarehouseWithdraw sends WareHouseWithdrawalList showing items in player's warehouse.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleNpcWarehouseWithdraw(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	// NPC distance validation — prevent remote warehouse access.
	if player.Location().DistanceSquared(npc.Location()) > maxNpcInteractionDistanceSquared {
		slog.Debug("warehouse withdraw: NPC too far", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	warehouseItems := player.Inventory().GetWarehouseItems()

	playerAdena := player.Inventory().GetAdena()
	pkt := serverpackets.NewWareHouseWithdrawalList(
		serverpackets.WarehouseTypePrivate,
		playerAdena,
		warehouseItems,
	)

	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing WareHouseWithdrawalList: %w", err)
	}

	slog.Debug("sent WareHouseWithdrawalList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"items", len(warehouseItems))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleNpcClanWarehouseDeposit opens clan warehouse deposit UI with privilege check.
func (h *Handler) handleNpcClanWarehouseDeposit(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	// NPC distance validation — prevent remote warehouse access.
	if player.Location().DistanceSquared(npc.Location()) > maxNpcInteractionDistanceSquared {
		slog.Debug("clan warehouse deposit: NPC too far", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		slog.Debug("clan warehouse: player has no clan", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		slog.Warn("clan warehouse: clan not found", "clanID", clanID, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	// Clan warehouse deposit is available to all members (no privilege check needed for deposits).
	depositableItems := player.Inventory().GetDepositableItems()
	playerAdena := player.Inventory().GetAdena()
	pkt := serverpackets.NewWareHouseDepositList(
		serverpackets.WarehouseTypeClan,
		playerAdena,
		depositableItems,
	)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing clan WareHouseDepositList: %w", err)
	}
	slog.Debug("sent clan WareHouseDepositList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"items", len(depositableItems))
	return copy(buf, pktData), true, nil
}

// handleNpcClanWarehouseWithdraw opens clan warehouse withdraw UI with privilege check.
func (h *Handler) handleNpcClanWarehouseWithdraw(player *model.Player, npc *model.Npc, buf []byte) (int, bool, error) {
	// NPC distance validation — prevent remote warehouse access.
	if player.Location().DistanceSquared(npc.Location()) > maxNpcInteractionDistanceSquared {
		slog.Debug("clan warehouse withdraw: NPC too far", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		slog.Debug("clan warehouse: player has no clan", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		slog.Warn("clan warehouse: clan not found", "clanID", clanID, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	// Check ACCESS_WAREHOUSE privilege for withdrawals.
	member := c.Member(player.CharacterID())
	if member == nil {
		slog.Warn("clan warehouse: member not found", "character", player.Name(), "clanID", clanID)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	if !member.Privileges().Has(clan.PrivCLViewWarehouse) {
		slog.Debug("clan warehouse: insufficient privileges",
			"character", player.Name(),
			"privileges", member.Privileges())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		return copy(buf, afData), true, nil
	}

	// Return clan warehouse items.
	// TODO: integrate Clan.Warehouse with WareHouseWithdrawalList when Clan model holds Warehouse ref.
	playerAdena := player.Inventory().GetAdena()
	pkt := serverpackets.NewWareHouseWithdrawalList(
		serverpackets.WarehouseTypeClan,
		playerAdena,
		nil,
	)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing clan WareHouseWithdrawalList: %w", err)
	}
	slog.Debug("sent clan WareHouseWithdrawalList",
		"character", player.Name(),
		"npcID", npc.TemplateID())
	return copy(buf, pktData), true, nil
}

// handleNpcMultisell sends MultiSellList for the specified multisell list ID.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleNpcMultisell(player *model.Player, npc *model.Npc, listIDStr string, buf []byte) (int, bool, error) {
	listID, err := strconv.ParseInt(listIDStr, 10, 32)
	if err != nil {
		slog.Warn("invalid multisell listID", "listID", listIDStr, "error", err)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	entries := skilldata.GetMultisellEntries(int32(listID))
	if entries == nil {
		slog.Warn("multisell list not found", "listID", listID, "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	pkt := serverpackets.NewMultiSellList(int32(listID), entries, 1)

	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing MultiSellList: %w", err)
	}

	slog.Debug("sent MultiSellList",
		"character", player.Name(),
		"npcID", npc.TemplateID(),
		"listID", listID,
		"entries", len(entries))

	n := copy(buf, pktData)
	return n, true, nil
}

// handleWarehouseDeposit processes SendWareHouseDepositList packet (opcode 0x31).
// Player confirms depositing items into warehouse.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleWarehouseDeposit(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSendWareHouseDepositList(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SendWareHouseDepositList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Personal warehouse capacity limit (Java: WAREHOUSE_SLOTS_DWARF=120, others=100).
	const warehouseSlotsDefault int32 = 100
	const warehouseSlotsDwarf int32 = 120
	const raceDwarf int32 = 4
	maxSlots := warehouseSlotsDefault
	if player.RaceID() == raceDwarf {
		maxSlots = warehouseSlotsDwarf
	}
	currentCount := int32(player.Inventory().WarehouseCount())
	newItems := int32(len(pkt.Items))
	if currentCount+newItems > maxSlots {
		slog.Debug("warehouse capacity exceeded",
			"character", player.Name(),
			"current", currentCount,
			"adding", newItems,
			"max", maxSlots)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Calculate and deduct deposit fee BEFORE depositing items.
	// Fee must be deducted first so that depositing Adena doesn't
	// drain the balance needed for the fee.
	fee := newItems * 30
	playerAdena := player.Inventory().GetAdena()
	if playerAdena < int64(fee) {
		slog.Debug("not enough adena for warehouse deposit fee",
			"character", player.Name(),
			"have", playerAdena,
			"fee", fee)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}
	if fee > 0 {
		if err := player.Inventory().RemoveAdena(fee); err != nil {
			slog.Error("failed to deduct warehouse fee", "error", err, "character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}
	}

	// Process each deposit
	for _, entry := range pkt.Items {
		objectID := uint32(entry.ObjectID)
		item := player.Inventory().GetItem(objectID)
		if item == nil {
			slog.Warn("warehouse deposit: item not found",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		if entry.Count == item.Count() {
			// Full stack deposit
			if err := player.Inventory().DepositToWarehouse(objectID, entry.Count); err != nil {
				slog.Error("warehouse deposit failed",
					"objectID", entry.ObjectID,
					"error", err)
				af := serverpackets.NewActionFailed()
				afData, _ := af.Write()
				n := copy(buf, afData)
				return n, true, nil
			}
		} else {
			// Partial stack deposit (needs new objectID)
			newObjectID := world.IDGenerator().NextItemID()
			if err := player.Inventory().DepositToWarehouseSplit(objectID, entry.Count, newObjectID); err != nil {
				slog.Error("warehouse deposit split failed",
					"objectID", entry.ObjectID,
					"error", err)
				af := serverpackets.NewActionFailed()
				afData, _ := af.Write()
				n := copy(buf, afData)
				return n, true, nil
			}
		}
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("warehouse deposit completed",
		"character", player.Name(),
		"items", len(pkt.Items),
		"fee", fee)

	return totalBytes, true, nil
}

// handleWarehouseWithdraw processes SendWareHouseWithDrawList packet (opcode 0x32).
// Player confirms withdrawing items from warehouse.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleWarehouseWithdraw(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSendWareHouseWithDrawList(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SendWareHouseWithDrawList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Process each withdrawal
	for _, entry := range pkt.Items {
		objectID := uint32(entry.ObjectID)
		whItem := player.Inventory().GetWarehouseItem(objectID)
		if whItem == nil {
			slog.Warn("warehouse withdraw: item not found",
				"objectID", entry.ObjectID,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		newObjectID := world.IDGenerator().NextItemID()
		if err := player.Inventory().WithdrawFromWarehouse(objectID, entry.Count, newObjectID); err != nil {
			slog.Error("warehouse withdraw failed",
				"objectID", entry.ObjectID,
				"error", err)
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("warehouse withdraw completed",
		"character", player.Name(),
		"items", len(pkt.Items))

	return totalBytes, true, nil
}

// handleMultiSellChoose processes MultiSellChoose packet (opcode 0xA7).
// Player confirms a multisell exchange.
//
// Phase 8: Trade System Foundation.
func (h *Handler) handleMultiSellChoose(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseMultiSellChoose(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing MultiSellChoose: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// NPC distance validation — prevent remote multisell exploitation.
	target := player.Target()
	if target == nil {
		slog.Debug("multisell: no target NPC", "character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}
	playerLoc := player.Location()
	targetLoc := target.Location()
	if playerLoc.DistanceSquared(targetLoc) > maxNpcInteractionDistanceSquared {
		slog.Debug("multisell: NPC too far",
			"character", player.Name(),
			"distSq", playerLoc.DistanceSquared(targetLoc))
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Find the multisell entry
	entry := skilldata.FindMultisellEntry(pkt.ListID, pkt.EntryID)
	if entry == nil {
		slog.Warn("multisell entry not found",
			"listID", pkt.ListID,
			"entryID", pkt.EntryID,
			"character", player.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Check all ingredients (amount × ingredient count)
	for _, ing := range entry.Ingredients {
		needed := ing.Count * int64(pkt.Amount)
		have := player.Inventory().CountItemsByID(ing.ItemID)
		if have < needed {
			slog.Debug("multisell: not enough ingredients",
				"character", player.Name(),
				"itemID", ing.ItemID,
				"have", have,
				"need", needed)
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}
	}

	// Remove ingredients — abort entirely if any removal fails to prevent item dupe.
	for _, ing := range entry.Ingredients {
		needed := ing.Count * int64(pkt.Amount)
		removed := player.Inventory().RemoveItemsByID(ing.ItemID, needed)
		if removed < needed {
			slog.Error("multisell: failed to remove all ingredients — aborting exchange",
				"itemID", ing.ItemID,
				"needed", needed,
				"removed", removed,
				"character", player.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}
	}

	// Add products
	for _, prod := range entry.Productions {
		tmpl := db.ItemDefToTemplate(prod.ItemID)
		if tmpl == nil {
			slog.Error("multisell: product template not found", "itemID", prod.ItemID)
			continue
		}

		totalCount := prod.Count * int64(pkt.Amount)

		// Check if item already exists in inventory (merge stacks)
		existing := player.Inventory().FindItemByItemID(prod.ItemID)
		if existing != nil && !existing.IsEquipped() {
			newCount := int64(existing.Count()) + totalCount
			if err := existing.SetCount(int32(newCount)); err != nil {
				slog.Error("multisell: failed to update item count", "error", err)
			}
			continue
		}

		objectID := world.IDGenerator().NextItemID()
		item, err := model.NewItem(objectID, prod.ItemID, int64(player.CharacterID()), int32(totalCount), tmpl)
		if err != nil {
			slog.Error("multisell: failed to create product item",
				"itemID", prod.ItemID,
				"error", err)
			continue
		}

		if err := player.Inventory().AddItem(item); err != nil {
			slog.Error("multisell: failed to add product to inventory",
				"itemID", prod.ItemID,
				"error", err)
		}
	}

	// Send updated inventory
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(player.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("multisell exchange completed",
		"character", player.Name(),
		"listID", pkt.ListID,
		"entryID", pkt.EntryID,
		"amount", pkt.Amount)

	return totalBytes, true, nil
}

// =============================================================================
// Private Store handlers (Phase 8.1)
// =============================================================================

// handleRequestPrivateStoreManageSell processes RequestPrivateStoreManageSell (0x73).
// Opens the sell store management UI for the player.
//
// Phase 8.1: Private Store System.
// Java reference: RequestPrivateStoreManageSell.java
func (h *Handler) handleRequestPrivateStoreManageSell(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Cannot open store while in combat
	if player.HasAttackStance() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Set manage mode
	player.SetPrivateStoreType(model.StoreSellManage)

	// Get sellable items from inventory
	sellableItems := player.Inventory().GetSellableItems()

	// Get current store items (if re-opening)
	var storeItems []*model.TradeItem
	if sl := player.SellList(); sl != nil {
		storeItems = sl.Items()
	}

	pkt := &serverpackets.PrivateStoreManageListSell{
		ObjectID:      player.ObjectID(),
		PackageSale:   false,
		PlayerAdena:   player.Inventory().GetAdena(),
		SellableItems: sellableItems,
		StoreItems:    storeItems,
	}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing PrivateStoreManageListSell: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleSetPrivateStoreListSell processes SetPrivateStoreListSell (0x74).
// Sets up the sell store items and activates the store.
//
// Phase 8.1: Private Store System.
// Java reference: SetPrivateStoreListSell.java
func (h *Handler) handleSetPrivateStoreListSell(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSetPrivateStoreListSell(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SetPrivateStoreListSell: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Must be in sell manage mode
	if player.PrivateStoreType() != model.StoreSellManage {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Empty list — cancel store
	if len(pkt.Items) == 0 {
		player.SetPrivateStoreType(model.StoreNone)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Create and populate trade list
	tradeList := model.NewTradeList()
	tradeList.SetPackaged(pkt.PackageSale)

	for _, entry := range pkt.Items {
		// Validate item exists in inventory
		item := player.Inventory().GetItem(uint32(entry.ObjectID))
		if item == nil {
			slog.Warn("sell store: item not in inventory",
				"objectID", entry.ObjectID,
				"character", player.Name())
			continue
		}

		if item.IsEquipped() {
			slog.Debug("sell store: cannot sell equipped item",
				"objectID", entry.ObjectID,
				"character", player.Name())
			continue
		}

		if item.ItemID() == model.AdenaItemID {
			continue
		}

		if item.Template() != nil && !item.Template().Tradeable {
			slog.Debug("sell store: item not tradeable",
				"objectID", entry.ObjectID,
				"character", player.Name())
			continue
		}

		// Validate count
		sellCount := entry.Count
		if sellCount > item.Count() {
			sellCount = item.Count()
		}

		tmpl := item.Template()
		var type2 int16
		var bodyPart int32
		if tmpl != nil {
			type2 = tmpl.Type2
			bodyPart = tmpl.BodyPartMask
		}

		ti := &model.TradeItem{
			ObjectID: item.ObjectID(),
			ItemID:   item.ItemID(),
			Count:    sellCount,
			Price:    entry.Price,
			Enchant:  item.Enchant(),
			Type2:    type2,
			BodyPart: bodyPart,
		}

		if err := tradeList.AddItem(ti); err != nil {
			slog.Warn("sell store: failed to add item",
				"objectID", entry.ObjectID,
				"error", err,
				"character", player.Name())
			continue
		}
	}

	if tradeList.ItemCount() == 0 {
		player.SetPrivateStoreType(model.StoreNone)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Activate store
	player.SetSellList(tradeList)
	if pkt.PackageSale {
		player.SetPrivateStoreType(model.StorePackageSell)
	} else {
		player.SetPrivateStoreType(model.StoreSell)
	}

	// Store owners sit down (Java: Player.sitDown + broadcast)
	player.SetSitting(true)
	sitPkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeSitting)
	sitData, _ := sitPkt.Write()
	h.clientManager.BroadcastToVisibleNear(player, sitData, len(sitData))

	// Broadcast UserInfo so nearby players see store icon
	userInfo := serverpackets.NewUserInfo(player)
	uiData, _ := userInfo.Write()
	h.clientManager.BroadcastToVisibleNear(player, uiData, len(uiData))

	slog.Info("private sell store opened",
		"character", player.Name(),
		"items", tradeList.ItemCount(),
		"package", pkt.PackageSale)

	return 0, true, nil
}

// handleRequestPrivateStoreQuitSell processes RequestPrivateStoreQuitSell (0x76).
// Closes the sell store.
//
// Phase 8.1: Private Store System.
func (h *Handler) handleRequestPrivateStoreQuitSell(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.ClosePrivateStore()

	// Stand up and broadcast state change
	player.SetSitting(false)
	standPkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeStanding)
	standData, _ := standPkt.Write()
	h.clientManager.BroadcastToVisibleNear(player, standData, len(standData))

	// Broadcast UserInfo to remove store icon
	userInfo := serverpackets.NewUserInfo(player)
	uiData, _ := userInfo.Write()
	h.clientManager.BroadcastToVisibleNear(player, uiData, len(uiData))

	slog.Debug("private sell store closed", "character", player.Name())

	af := serverpackets.NewActionFailed()
	afData, _ := af.Write()
	n := copy(buf, afData)
	return n, true, nil
}

// handleSetPrivateStoreMsgSell processes SetPrivateStoreMsgSell (0x77).
// Sets the sell store message (title shown above player's head).
//
// Phase 8.1: Private Store System.
func (h *Handler) handleSetPrivateStoreMsgSell(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSetPrivateStoreMsgSell(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SetPrivateStoreMsgSell: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.SetStoreMessage(pkt.Message)

	if sl := player.SellList(); sl != nil {
		sl.SetTitle(pkt.Message)
	}

	// Broadcast store message to nearby players
	msgPkt := &serverpackets.PrivateStoreMsgSell{
		ObjectID: player.ObjectID(),
		Message:  pkt.Message,
	}
	msgData, err := msgPkt.Write()
	if err != nil {
		slog.Error("failed to serialize PrivateStoreMsgSell", "error", err)
		return 0, true, nil
	}

	// Broadcast store message to nearby players so they see the title
	h.clientManager.BroadcastToVisibleNear(player, msgData, len(msgData))

	n := copy(buf, msgData)
	return n, true, nil
}

// handleRequestPrivateStoreBuy processes RequestPrivateStoreBuy (0x79).
// Buyer purchases items from a sell store.
//
// Phase 8.1: Private Store System.
// Java reference: RequestPrivateStoreBuy.java
func (h *Handler) handleRequestPrivateStoreBuy(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPrivateStoreBuy(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestPrivateStoreBuy: %w", err)
	}

	buyer := client.ActivePlayer()
	if buyer == nil {
		return 0, true, nil
	}

	// Find seller by ObjectID
	sellerObj, sellerFound := world.Instance().GetObject(uint32(pkt.StorePlayerID))
	if !sellerFound || sellerObj == nil {
		slog.Warn("private store buy: seller not found",
			"sellerObjectID", pkt.StorePlayerID,
			"buyer", buyer.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	sellerPlayer, ok := sellerObj.Data.(*model.Player)
	if !ok {
		slog.Warn("private store buy: target is not a player",
			"sellerObjectID", pkt.StorePlayerID)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	sellerStoreType := sellerPlayer.PrivateStoreType()
	if sellerStoreType != model.StoreSell && sellerStoreType != model.StorePackageSell {
		slog.Warn("private store buy: seller not in sell mode",
			"sellerObjectID", pkt.StorePlayerID,
			"storeType", sellerStoreType)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	sellList := sellerPlayer.SellList()
	if sellList == nil {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Package sell: must buy all items
	if sellList.IsPackaged() && len(pkt.Items) != sellList.ItemCount() {
		slog.Warn("private store buy: package sell requires all items",
			"requested", len(pkt.Items),
			"available", sellList.ItemCount(),
			"buyer", buyer.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Calculate total cost and validate
	var totalCost int64
	type buyOp struct {
		tradeItem *model.TradeItem
		count     int32
	}
	var ops []buyOp

	for _, entry := range pkt.Items {
		ti := sellList.FindItem(uint32(entry.ObjectID))
		if ti == nil {
			slog.Warn("private store buy: item not in sell list",
				"objectID", entry.ObjectID,
				"buyer", buyer.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Price must match (anti-cheat)
		if entry.Price != ti.Price {
			slog.Warn("private store buy: price mismatch",
				"expected", ti.Price,
				"got", entry.Price,
				"buyer", buyer.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		count := entry.Count
		if count > ti.Count {
			count = ti.Count
		}
		if count <= 0 {
			continue
		}

		// Overflow protection
		itemCost := int64(count) * ti.Price
		if ti.Price > 0 && itemCost/ti.Price != int64(count) {
			slog.Warn("private store buy: price overflow",
				"count", count,
				"price", ti.Price,
				"buyer", buyer.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		totalCost += itemCost
		ops = append(ops, buyOp{tradeItem: ti, count: count})
	}

	// Check buyer has enough Adena
	if buyer.Inventory().GetAdena() < totalCost {
		slog.Debug("private store buy: not enough adena",
			"buyer", buyer.Name(),
			"have", buyer.Inventory().GetAdena(),
			"need", totalCost)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Execute transaction: transfer items and Adena
	// Deduct Adena from buyer
	if err := buyer.Inventory().RemoveAdena(int32(totalCost)); err != nil {
		slog.Error("private store buy: failed to remove buyer adena",
			"error", err,
			"buyer", buyer.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Add Adena to seller
	if err := sellerPlayer.Inventory().AddAdena(int32(totalCost)); err != nil {
		slog.Error("private store buy: failed to add seller adena",
			"error", err,
			"seller", sellerPlayer.Name())
		// Rollback buyer's adena
		if rbErr := buyer.Inventory().AddAdena(int32(totalCost)); rbErr != nil {
			slog.Error("private store buy: rollback failed", "error", rbErr)
		}
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Transfer items: remove from seller, create for buyer
	for _, op := range ops {
		sellerItem := sellerPlayer.Inventory().GetItem(op.tradeItem.ObjectID)
		if sellerItem == nil {
			continue
		}

		if op.count >= sellerItem.Count() {
			// Transfer entire item
			sellerPlayer.Inventory().RemoveItem(op.tradeItem.ObjectID)
			sellerItem.SetLocation(model.ItemLocationInventory)
			if err := buyer.Inventory().AddItem(sellerItem); err != nil {
				slog.Error("private store buy: failed to add item to buyer",
					"objectID", op.tradeItem.ObjectID,
					"error", err)
			}
		} else {
			// Partial transfer: decrease seller count, create new for buyer
			if err := sellerItem.SetCount(sellerItem.Count() - op.count); err != nil {
				slog.Error("private store buy: failed to decrease seller count", "error", err)
				continue
			}

			tmpl := sellerItem.Template()
			newObjID := world.IDGenerator().NextItemID()
			newItem, err := model.NewItem(newObjID, sellerItem.ItemID(), buyer.CharacterID(), op.count, tmpl)
			if err != nil {
				slog.Error("private store buy: failed to create item for buyer", "error", err)
				continue
			}
			if err := buyer.Inventory().AddItem(newItem); err != nil {
				slog.Error("private store buy: failed to add new item to buyer", "error", err)
			}
		}

		// Update trade list
		sellList.UpdateItemCount(op.tradeItem.ObjectID, op.count)
	}

	// If store is empty, close it
	if sellList.ItemCount() == 0 {
		sellerPlayer.ClosePrivateStore()
		sellerPlayer.SetSitting(false)
		standPkt := serverpackets.NewChangeWaitType(sellerPlayer, serverpackets.WaitTypeStanding)
		standData, _ := standPkt.Write()
		h.clientManager.BroadcastToVisibleNear(sellerPlayer, standData, len(standData))
		sellerUI := serverpackets.NewUserInfo(sellerPlayer)
		sellerUIData, _ := sellerUI.Write()
		h.clientManager.BroadcastToVisibleNear(sellerPlayer, sellerUIData, len(sellerUIData))
	}

	// Send updated inventory to buyer
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(buyer.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize buyer InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("private store purchase completed",
		"buyer", buyer.Name(),
		"seller", sellerPlayer.Name(),
		"totalCost", totalCost,
		"items", len(ops))

	return totalBytes, true, nil
}

// handleRequestPrivateStoreManageBuy processes RequestPrivateStoreManageBuy (0x90).
// Opens the buy store management UI for the player.
//
// Phase 8.1: Private Store System.
func (h *Handler) handleRequestPrivateStoreManageBuy(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Cannot open store while in combat
	if player.HasAttackStance() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Set manage mode
	player.SetPrivateStoreType(model.StoreBuyManage)

	// Get current store items (if re-opening)
	var storeItems []*model.TradeItem
	if bl := player.BuyList(); bl != nil {
		storeItems = bl.Items()
	}

	pkt := &serverpackets.PrivateStoreBuyManageList{
		ObjectID:    player.ObjectID(),
		PlayerAdena: player.Inventory().GetAdena(),
		StoreItems:  storeItems,
	}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing PrivateStoreBuyManageList: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleSetPrivateStoreListBuy processes SetPrivateStoreListBuy (0x91).
// Sets up the buy store items and activates the store.
//
// Phase 8.1: Private Store System.
func (h *Handler) handleSetPrivateStoreListBuy(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSetPrivateStoreListBuy(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SetPrivateStoreListBuy: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Must be in buy manage mode
	if player.PrivateStoreType() != model.StoreBuyManage {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Empty list — cancel store
	if len(pkt.Items) == 0 {
		player.SetPrivateStoreType(model.StoreNone)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Calculate total cost (buyer must have enough Adena to cover all requests)
	var totalCost int64
	tradeList := model.NewTradeList()

	for _, entry := range pkt.Items {
		// Validate item template exists
		tmpl := db.ItemDefToTemplate(entry.ItemID)
		if tmpl == nil {
			slog.Warn("buy store: unknown item template",
				"itemID", entry.ItemID,
				"character", player.Name())
			continue
		}

		// Overflow protection
		itemCost := int64(entry.Count) * entry.Price
		if entry.Price > 0 && itemCost/entry.Price != int64(entry.Count) {
			slog.Warn("buy store: price overflow",
				"itemID", entry.ItemID,
				"character", player.Name())
			continue
		}
		totalCost += itemCost

		ti := &model.TradeItem{
			ItemID:     entry.ItemID,
			Count:      entry.Count,
			StoreCount: entry.Count,
			Price:      entry.Price,
			Type2:      tmpl.Type2,
			BodyPart:   tmpl.BodyPartMask,
		}

		if err := tradeList.AddItem(ti); err != nil {
			slog.Warn("buy store: failed to add item",
				"itemID", entry.ItemID,
				"error", err,
				"character", player.Name())
			continue
		}
	}

	if tradeList.ItemCount() == 0 {
		player.SetPrivateStoreType(model.StoreNone)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Check buyer has enough Adena to cover total buy cost
	if player.Inventory().GetAdena() < totalCost {
		slog.Debug("buy store: not enough adena to open",
			"character", player.Name(),
			"have", player.Inventory().GetAdena(),
			"need", totalCost)
		player.SetPrivateStoreType(model.StoreNone)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Activate store
	player.SetBuyList(tradeList)
	player.SetPrivateStoreType(model.StoreBuy)

	// Store owners sit down (Java: Player.sitDown + broadcast)
	player.SetSitting(true)
	sitPkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeSitting)
	sitData, _ := sitPkt.Write()
	h.clientManager.BroadcastToVisibleNear(player, sitData, len(sitData))

	// Broadcast UserInfo so nearby players see buy-store icon
	userInfo := serverpackets.NewUserInfo(player)
	uiData, _ := userInfo.Write()
	h.clientManager.BroadcastToVisibleNear(player, uiData, len(uiData))

	slog.Info("private buy store opened",
		"character", player.Name(),
		"items", tradeList.ItemCount(),
		"totalCost", totalCost)

	return 0, true, nil
}

// handleRequestPrivateStoreQuitBuy processes RequestPrivateStoreQuitBuy (0x93).
// Closes the buy store.
//
// Phase 8.1: Private Store System.
func (h *Handler) handleRequestPrivateStoreQuitBuy(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.ClosePrivateStore()

	// Stand up and broadcast state change
	player.SetSitting(false)
	standPkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeStanding)
	standData, _ := standPkt.Write()
	h.clientManager.BroadcastToVisibleNear(player, standData, len(standData))

	// Broadcast UserInfo to remove store icon
	userInfo := serverpackets.NewUserInfo(player)
	uiData, _ := userInfo.Write()
	h.clientManager.BroadcastToVisibleNear(player, uiData, len(uiData))

	slog.Debug("private buy store closed", "character", player.Name())

	af := serverpackets.NewActionFailed()
	afData, _ := af.Write()
	n := copy(buf, afData)
	return n, true, nil
}

// handleSetPrivateStoreMsgBuy processes SetPrivateStoreMsgBuy (0x94).
// Sets the buy store message (title shown above player's head).
//
// Phase 8.1: Private Store System.
func (h *Handler) handleSetPrivateStoreMsgBuy(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSetPrivateStoreMsgBuy(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SetPrivateStoreMsgBuy: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.SetStoreMessage(pkt.Message)

	if bl := player.BuyList(); bl != nil {
		bl.SetTitle(pkt.Message)
	}

	// Broadcast store message to nearby players
	msgPkt := &serverpackets.PrivateStoreMsgBuy{
		ObjectID: player.ObjectID(),
		Message:  pkt.Message,
	}
	msgData, err := msgPkt.Write()
	if err != nil {
		slog.Error("failed to serialize PrivateStoreMsgBuy", "error", err)
		return 0, true, nil
	}

	// Broadcast store message to nearby players so they see the title
	h.clientManager.BroadcastToVisibleNear(player, msgData, len(msgData))

	n := copy(buf, msgData)
	return n, true, nil
}

// handleRequestPrivateStoreSell processes RequestPrivateStoreSell (0x96).
// Seller sells items to a buyer's buy store.
//
// Phase 8.1: Private Store System.
// Java reference: RequestPrivateStoreSell.java
func (h *Handler) handleRequestPrivateStoreSell(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPrivateStoreSell(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestPrivateStoreSell: %w", err)
	}

	seller := client.ActivePlayer()
	if seller == nil {
		return 0, true, nil
	}

	// Find buyer by ObjectID
	buyerObj, buyerFound := world.Instance().GetObject(uint32(pkt.StorePlayerID))
	if !buyerFound || buyerObj == nil {
		slog.Warn("private store sell: buyer not found",
			"buyerObjectID", pkt.StorePlayerID,
			"seller", seller.Name())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	buyerPlayer, ok := buyerObj.Data.(*model.Player)
	if !ok {
		slog.Warn("private store sell: target is not a player",
			"buyerObjectID", pkt.StorePlayerID)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	if buyerPlayer.PrivateStoreType() != model.StoreBuy {
		slog.Warn("private store sell: buyer not in buy mode",
			"buyerObjectID", pkt.StorePlayerID,
			"storeType", buyerPlayer.PrivateStoreType())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	buyList := buyerPlayer.BuyList()
	if buyList == nil {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Validate and calculate total payment
	var totalPayment int64
	type sellOp struct {
		sellerItem *model.Item
		tradeItem  *model.TradeItem
		count      int32
	}
	var ops []sellOp

	for _, entry := range pkt.Items {
		// Find in buyer's buy list by itemID
		ti := buyList.FindItemByID(entry.ItemID)
		if ti == nil {
			slog.Warn("private store sell: item not in buy list",
				"itemID", entry.ItemID,
				"seller", seller.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Price must match
		if entry.Price != ti.Price {
			slog.Warn("private store sell: price mismatch",
				"expected", ti.Price,
				"got", entry.Price,
				"seller", seller.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		// Find item in seller's inventory
		sellerItem := seller.Inventory().GetItem(uint32(entry.ObjectID))
		if sellerItem == nil {
			slog.Warn("private store sell: seller doesn't have item",
				"objectID", entry.ObjectID,
				"seller", seller.Name())
			af := serverpackets.NewActionFailed()
			afData, _ := af.Write()
			n := copy(buf, afData)
			return n, true, nil
		}

		count := entry.Count
		if count > ti.Count {
			count = ti.Count // Cannot sell more than buyer wants
		}
		if count > sellerItem.Count() {
			count = sellerItem.Count()
		}
		if count <= 0 {
			continue
		}

		itemPayment := int64(count) * ti.Price
		totalPayment += itemPayment
		ops = append(ops, sellOp{sellerItem: sellerItem, tradeItem: ti, count: count})
	}

	// Check buyer has enough Adena
	if buyerPlayer.Inventory().GetAdena() < totalPayment {
		slog.Debug("private store sell: buyer not enough adena",
			"buyer", buyerPlayer.Name(),
			"have", buyerPlayer.Inventory().GetAdena(),
			"need", totalPayment)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Execute transaction
	// Deduct Adena from buyer
	if err := buyerPlayer.Inventory().RemoveAdena(int32(totalPayment)); err != nil {
		slog.Error("private store sell: failed to remove buyer adena", "error", err)
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Add Adena to seller
	if err := seller.Inventory().AddAdena(int32(totalPayment)); err != nil {
		slog.Error("private store sell: failed to add seller adena", "error", err)
		// Rollback buyer's adena
		if rbErr := buyerPlayer.Inventory().AddAdena(int32(totalPayment)); rbErr != nil {
			slog.Error("private store sell: rollback failed", "error", rbErr)
		}
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Transfer items: remove from seller, create for buyer
	for _, op := range ops {
		if op.count >= op.sellerItem.Count() {
			// Transfer entire item
			seller.Inventory().RemoveItem(op.sellerItem.ObjectID())
			op.sellerItem.SetLocation(model.ItemLocationInventory)
			if err := buyerPlayer.Inventory().AddItem(op.sellerItem); err != nil {
				slog.Error("private store sell: failed to add item to buyer",
					"objectID", op.sellerItem.ObjectID(),
					"error", err)
			}
		} else {
			// Partial: decrease seller, create for buyer
			if err := op.sellerItem.SetCount(op.sellerItem.Count() - op.count); err != nil {
				slog.Error("private store sell: failed to decrease seller count", "error", err)
				continue
			}

			tmpl := op.sellerItem.Template()
			newObjID := world.IDGenerator().NextItemID()
			newItem, err := model.NewItem(newObjID, op.sellerItem.ItemID(), buyerPlayer.CharacterID(), op.count, tmpl)
			if err != nil {
				slog.Error("private store sell: failed to create item for buyer", "error", err)
				continue
			}
			if err := buyerPlayer.Inventory().AddItem(newItem); err != nil {
				slog.Error("private store sell: failed to add item to buyer", "error", err)
			}
		}

		// Update buy list
		buyList.UpdateItemCount(op.tradeItem.ObjectID, op.count)
	}

	// If buy store is empty, close it
	if buyList.ItemCount() == 0 {
		buyerPlayer.ClosePrivateStore()
		buyerPlayer.SetSitting(false)
		standPkt := serverpackets.NewChangeWaitType(buyerPlayer, serverpackets.WaitTypeStanding)
		standData, _ := standPkt.Write()
		h.clientManager.BroadcastToVisibleNear(buyerPlayer, standData, len(standData))
		buyerUI := serverpackets.NewUserInfo(buyerPlayer)
		buyerUIData, _ := buyerUI.Write()
		h.clientManager.BroadcastToVisibleNear(buyerPlayer, buyerUIData, len(buyerUIData))
	}

	// Send updated inventory to seller
	totalBytes := 0
	invPkt := serverpackets.NewInventoryItemList(seller.Inventory().GetItems())
	invPkt.ShowWindow = false
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("failed to serialize seller InventoryItemList", "error", err)
	} else {
		n := copy(buf[totalBytes:], invData)
		totalBytes += n
	}

	slog.Info("private store sell completed",
		"seller", seller.Name(),
		"buyer", buyerPlayer.Name(),
		"totalPayment", totalPayment,
		"items", len(ops))

	return totalBytes, true, nil
}

// handleRequestRecipeBookOpen opens recipe book for the player.
// Phase 15: Recipe/Craft System.
func (h *Handler) handleRequestRecipeBookOpen(
	_ context.Context,
	client *GameClient,
	body, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestRecipeBookOpen(body)
	if err != nil {
		return 0, true, fmt.Errorf("parse RequestRecipeBookOpen: %w", err)
	}

	recipeIDs := player.GetRecipeBook(pkt.IsDwarvenCraft)

	resp := &serverpackets.RecipeBookItemList{
		IsDwarvenCraft: pkt.IsDwarvenCraft,
		MaxMP:          player.MaxMP(),
		RecipeIDs:      recipeIDs,
	}
	data, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("write RecipeBookItemList: %w", err)
	}

	n := copy(buf, data)
	return n, true, nil
}

// handleRequestRecipeItemMakeInfo sends recipe crafting info to client.
// Phase 15: Recipe/Craft System.
func (h *Handler) handleRequestRecipeItemMakeInfo(
	_ context.Context,
	client *GameClient,
	body, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestRecipeItemMakeInfo(body)
	if err != nil {
		return 0, true, fmt.Errorf("parse RequestRecipeItemMakeInfo: %w", err)
	}

	recipe := skilldata.GetRecipeTemplate(pkt.RecipeListID)
	if recipe == nil {
		slog.Warn("recipe not found for make info",
			"recipeListID", pkt.RecipeListID,
			"player", player.Name())
		return 0, true, nil
	}

	resp := &serverpackets.RecipeItemMakeInfo{
		RecipeListID:   pkt.RecipeListID,
		IsDwarvenCraft: recipe.IsDwarven,
		CurrentMP:      player.CurrentMP(),
		MaxMP:          player.MaxMP(),
		Success:        false, // info only, not a result
	}
	data, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("write RecipeItemMakeInfo: %w", err)
	}

	n := copy(buf, data)
	return n, true, nil
}

// handleRequestRecipeItemMakeSelf handles crafting request from the player.
// Phase 15: Recipe/Craft System.
func (h *Handler) handleRequestRecipeItemMakeSelf(
	_ context.Context,
	client *GameClient,
	body, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestRecipeItemMakeSelf(body)
	if err != nil {
		return 0, true, fmt.Errorf("parse RequestRecipeItemMakeSelf: %w", err)
	}

	if h.craftController == nil {
		slog.Warn("craft controller not initialized", "player", player.Name())
		return 0, true, nil
	}

	result, err := h.craftController.Craft(player, pkt.RecipeListID)
	if err != nil {
		slog.Warn("craft failed",
			"player", player.Name(),
			"recipeListID", pkt.RecipeListID,
			"error", err)

		// Отправляем RecipeItemMakeInfo с Success=false
		recipe := skilldata.GetRecipeTemplate(pkt.RecipeListID)
		isDwarven := false
		if recipe != nil {
			isDwarven = recipe.IsDwarven
		}
		resp := &serverpackets.RecipeItemMakeInfo{
			RecipeListID:   pkt.RecipeListID,
			IsDwarvenCraft: isDwarven,
			CurrentMP:      player.CurrentMP(),
			MaxMP:          player.MaxMP(),
			Success:        false,
		}
		data, wErr := resp.Write()
		if wErr != nil {
			return 0, true, fmt.Errorf("write RecipeItemMakeInfo (error): %w", wErr)
		}
		n := copy(buf, data)
		return n, true, nil
	}

	resp := &serverpackets.RecipeItemMakeInfo{
		RecipeListID:   pkt.RecipeListID,
		IsDwarvenCraft: result.Recipe.IsDwarven,
		CurrentMP:      player.CurrentMP(),
		MaxMP:          player.MaxMP(),
		Success:        result.Success,
	}
	data, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("write RecipeItemMakeInfo: %w", err)
	}

	n := copy(buf, data)
	return n, true, nil
}

// handleExtendedPacket dispatches extended client packets (opcode 0xD0).
// Extended packets have a 2-byte sub-opcode after the main opcode.
// Phase 20: Duel System (first 0xD0 packets).
func (h *Handler) handleExtendedPacket(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	if len(data) < 2 {
		return 0, true, fmt.Errorf("extended packet too short: %d bytes", len(data))
	}

	subOpcode := int16(data[0]) | int16(data[1])<<8 // LE read
	subBody := data[2:]

	switch subOpcode {
	case clientpackets.SubOpcodeRequestDuelStart:
		return h.handleRequestDuelStart(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestDuelAnswerStart:
		return h.handleRequestDuelAnswerStart(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestDuelSurrender:
		return h.handleRequestDuelSurrender(ctx, client, subBody, buf)
	// Phase 28: Augmentation System
	case clientpackets.SubOpcodeRequestConfirmTargetItem:
		return h.handleRequestConfirmTargetItem(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestConfirmRefinerItem:
		return h.handleRequestConfirmRefinerItem(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestConfirmGemStone:
		return h.handleRequestConfirmGemStone(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestRefine:
		return h.handleRequestRefine(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestConfirmCancelItem:
		return h.handleRequestConfirmCancelItem(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestRefineCancel:
		return h.handleRequestRefineCancel(ctx, client, subBody, buf)
	// Phase 32: Cursed Weapons
	case clientpackets.SubOpcodeRequestCursedWeaponList:
		return h.handleRequestCursedWeaponList(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestCursedWeaponLocation:
		return h.handleRequestCursedWeaponLocation(ctx, client, subBody, buf)
	// Phase 18: Clan extended packets (moved from main dispatch)
	case clientpackets.SubOpcodeRequestPledgeMemberInfo:
		return h.handleRequestPledgeMemberInfo(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestPledgeSetMemberPowerGrade:
		return h.handleRequestPledgeSetMemberPowerGrade(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestPledgeReorganizeMember:
		return h.handleRequestPledgeReorganizeMember(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestPledgeWarList:
		return h.handleRequestPledgeWarList(ctx, client, subBody, buf)
	// Phase 36: Auto SoulShot
	case clientpackets.SubOpcodeRequestAutoSoulShot:
		return h.handleRequestAutoSoulShot(ctx, client, subBody, buf)
	// Phase 49: Enchant Skill
	case clientpackets.SubOpcodeRequestExEnchantSkillInfo:
		return h.handleRequestExEnchantSkillInfo(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExEnchantSkill:
		return h.handleRequestExEnchantSkill(ctx, client, subBody, buf)
	// Phase 49: Ground-targeted skill
	case clientpackets.SubOpcodeRequestExMagicSkillUseGround:
		return h.handleRequestExMagicSkillUseGround(ctx, client, subBody, buf)
	// Phase 49: Manor
	case clientpackets.SubOpcodeRequestManorList:
		return h.handleRequestManorList(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestProcureCropList:
		return h.handleRequestProcureCropList(ctx, client, subBody, buf)
	// Phase 50: Party Leader Change
	case clientpackets.SubOpcodeRequestChangePartyLeader:
		return h.handleRequestChangePartyLeader(ctx, client, subBody, buf)
	// Phase 50: Stubs (acknowledged, not yet fully implemented)
	case clientpackets.SubOpcodeRequestExPledgeCrestLarge:
		return h.handleRequestExPledgeCrestLarge(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExSetPledgeCrestLarge:
		return h.handleRequestExSetPledgeCrestLarge(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestOlympiadObserverEnd:
		return h.handleRequestOlympiadObserverEnd(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestOlympiadMatchList:
		return h.handleRequestOlympiadMatchList(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExMPCCShowPartyMembersInfo:
		return h.handleRequestExMPCCShowPartyMembersInfo(ctx, client, subBody, buf)
	// Phase 50: Manor management stubs
	case clientpackets.SubOpcodeRequestSetSeed:
		return h.handleRequestSetSeed(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestSetCrop:
		return h.handleRequestSetCrop(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExShowManorSeedInfo:
		return h.handleRequestExShowManorSeedInfo(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExShowCropInfo:
		return h.handleRequestExShowCropInfo(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExShowSeedSetting:
		return h.handleRequestExShowSeedSetting(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestExShowCropSetting:
		return h.handleRequestExShowCropSetting(ctx, client, subBody, buf)
	// Phase 50: Party Room stubs
	case clientpackets.SubOpcodeRequestOustFromPartyRoom:
		return h.handleRequestOustFromPartyRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestDismissPartyRoom:
		return h.handleRequestDismissPartyRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestWithdrawPartyRoom:
		return h.handleRequestWithdrawPartyRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestListPartyMatchingWaitingRoom:
		return h.handleRequestListPartyMatchingWaitingRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestAskJoinPartyRoom:
		return h.handleRequestAskJoinPartyRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeConfirmJoinPartyRoom:
		return h.handleConfirmJoinPartyRoom(ctx, client, subBody, buf)
	case clientpackets.SubOpcodeRequestListPartyMatching:
		return h.handleRequestListPartyMatching(ctx, client, subBody, buf)
	default:
		slog.Warn("unknown extended packet sub-opcode",
			"subOpcode", fmt.Sprintf("0x%04X", subOpcode),
			"client", client.IP())
		return 0, true, nil
	}
}

// handleRequestDropItem processes the RequestDropItem packet (opcode 0x12).
// Removes item from inventory and drops it on the ground.
func (h *Handler) handleRequestDropItem(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestDropItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestDropItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for drop item")
	}

	inv := player.Inventory()
	if inv == nil {
		return 0, false, fmt.Errorf("no inventory for drop item")
	}

	// Validate count
	if pkt.Count <= 0 {
		slog.Warn("invalid drop count", "count", pkt.Count, "account", client.AccountName())
		return 0, false, nil
	}

	objectID := uint32(pkt.ObjectID)

	// Find item in inventory
	item := inv.GetItem(objectID)
	if item == nil {
		slog.Warn("item not found for drop", "objectID", objectID, "account", client.AccountName())
		return 0, false, nil
	}

	// Cannot drop equipped items
	if item.IsEquipped() {
		slog.Warn("cannot drop equipped item", "objectID", objectID, "account", client.AccountName())
		return 0, false, nil
	}

	// Cannot drop quest items
	if tmpl := item.Template(); tmpl != nil && tmpl.Type == model.ItemTypeQuestItem {
		slog.Warn("cannot drop quest item", "objectID", objectID, "account", client.AccountName())
		return 0, false, nil
	}

	// Remove from inventory
	if item.Template() != nil && item.Template().Stackable {
		removed := inv.RemoveItemsByID(item.Template().ItemID, pkt.Count)
		if removed <= 0 {
			slog.Warn("insufficient items to drop", "objectID", objectID, "count", pkt.Count)
			return 0, false, nil
		}
	} else {
		inv.RemoveItem(objectID)
	}

	// Create dropped item on ground
	dropLoc := model.NewLocation(pkt.X, pkt.Y, pkt.Z, 0)
	droppedItem := model.NewDroppedItem(world.IDGenerator().NextItemID(), item, dropLoc, player.ObjectID())

	// Add to world (NewDroppedItem already sets WorldObject.Data)
	if err := world.Instance().AddObject(droppedItem.WorldObject); err != nil {
		slog.Error("adding dropped item to world", "error", err)
		return 0, false, nil
	}

	// Broadcast ItemOnGround to nearby players
	dropPkt := serverpackets.NewItemOnGround(droppedItem)
	dropData, err := dropPkt.Write()
	if err != nil {
		slog.Error("serializing ItemOnGround", "error", err)
		return 0, false, nil
	}
	h.clientManager.BroadcastToVisible(player, dropData, len(dropData))

	// Send updated inventory to player
	invPkt := serverpackets.NewInventoryItemList(inv.GetItems())
	invData, err := invPkt.Write()
	if err != nil {
		slog.Error("serializing InventoryItemList", "error", err)
		return 0, false, nil
	}
	n := copy(buf, invData)

	slog.Debug("item dropped",
		"objectID", objectID,
		"count", pkt.Count,
		"x", pkt.X, "y", pkt.Y, "z", pkt.Z,
		"account", client.AccountName())

	return n, true, nil
}

// handleRequestSocialAction processes the RequestSocialAction packet (opcode 0x1B).
// Validates player state and broadcasts the social action animation to nearby players.
//
// Validation rules:
//   - Player must be alive (not dead)
//   - Player must not be in private store mode
//   - Player must not be fishing
//   - ActionID must be in range [2..16]
//   - ActionID 15 (Charm) requires hero status
func (h *Handler) handleRequestSocialAction(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSocialAction(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSocialAction: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for social action")
	}

	// Validate action ID range (2-16 for Interlude)
	if pkt.ActionID < serverpackets.MinSocialActionID || pkt.ActionID > serverpackets.MaxSocialActionID {
		slog.Warn("invalid social action ID",
			"actionID", pkt.ActionID,
			"account", client.AccountName())
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Dead players cannot perform social actions
	if player.IsDead() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Players in private store mode cannot perform social actions
	if player.IsInStoreMode() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Players currently fishing cannot perform social actions
	if player.IsFishing() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Charm (actionID 15) is hero-only emote
	if pkt.ActionID == serverpackets.SocialActionCharm && !player.IsHero() {
		af := serverpackets.NewActionFailed()
		afData, _ := af.Write()
		n := copy(buf, afData)
		return n, true, nil
	}

	// Broadcast SocialAction to nearby players (including sender)
	sa := serverpackets.NewSocialAction(int32(player.ObjectID()), pkt.ActionID)
	saData, err := sa.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SocialAction: %w", err)
	}

	h.clientManager.BroadcastToVisible(player, saData, len(saData))

	return 0, false, nil
}

// handleRequestTargetCanceld processes RequestTargetCanceld (opcode 0x37).
// Clears the player's current target.
func (h *Handler) handleRequestTargetCanceld(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	_, err := clientpackets.ParseRequestTargetCanceld(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestTargetCanceld: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	player.ClearTarget()

	// Send TargetUnselected to the player (objectID=0 means no target)
	targetPkt := serverpackets.NewMyTargetSelected(0)
	pktData, err := targetPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing MyTargetSelected: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleAppearing processes Appearing (opcode 0x30).
// Sent by client after teleport. Broadcasts character info to nearby players.
func (h *Handler) handleAppearing(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	// Broadcast CharInfo to nearby so they see the player appear
	charInfo := serverpackets.NewCharInfo(player)
	charData, err := charInfo.Write()
	if err != nil {
		slog.Error("serializing CharInfo for appearing", "error", err)
		return 0, false, nil
	}
	h.clientManager.BroadcastToVisibleExcept(player, player, charData, len(charData))

	return 0, false, nil
}

// handleChangeMoveType2 processes ChangeMoveType2 (opcode 0x1C).
// Toggles walk/run mode and broadcasts to nearby players.
func (h *Handler) handleChangeMoveType2(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseChangeMoveType2(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ChangeMoveType2: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	running := pkt.TypeRun == 1
	player.SetRunning(running)

	// Broadcast to nearby players
	moveType := int32(0)
	if running {
		moveType = 1
	}
	changePkt := &serverpackets.ChangeMoveType{
		ObjectID: int32(player.ObjectID()),
		MoveType: moveType,
	}
	changeData, err := changePkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ChangeMoveType: %w", err)
	}
	h.clientManager.BroadcastToVisible(player, changeData, len(changeData))

	return 0, false, nil
}

// handleChangeWaitType2 processes ChangeWaitType2 (opcode 0x1D).
// Toggles sit/stand and broadcasts to nearby players.
func (h *Handler) handleChangeWaitType2(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseChangeWaitType2(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing ChangeWaitType2: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	if pkt.TypeStand == 1 {
		// Stand up
		player.SetSitting(false)
		changePkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeStanding)
		changeData, err := changePkt.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ChangeWaitType: %w", err)
		}
		h.clientManager.BroadcastToVisible(player, changeData, len(changeData))
	} else {
		// Sit down
		player.SetSitting(true)
		changePkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeSitting)
		changeData, err := changePkt.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ChangeWaitType: %w", err)
		}
		h.clientManager.BroadcastToVisible(player, changeData, len(changeData))
	}

	return 0, false, nil
}

// handleRequestSkillList processes RequestSkillList (opcode 0x3F).
// Sends the player's skill list.
func (h *Handler) handleRequestSkillList(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	skillList := serverpackets.NewSkillList(player.Skills())
	skillData, err := skillList.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SkillList: %w", err)
	}
	n := copy(buf, skillData)
	return n, true, nil
}

// handleRequestItemList processes RequestItemList (opcode 0x0F).
// Sends the full inventory item list to the player.
func (h *Handler) handleRequestItemList(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	inv := player.Inventory()
	if inv == nil {
		return 0, false, nil
	}

	invPkt := serverpackets.NewInventoryItemList(inv.GetItems())
	invData, err := invPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}
	n := copy(buf, invData)
	return n, true, nil
}

// handleRequestUnEquipItem processes RequestUnEquipItem (opcode 0x11).
// Unequips an item from the specified slot.
func (h *Handler) handleRequestUnEquipItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestUnEquipItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestUnEquipItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	inv := player.Inventory()
	if inv == nil {
		return 0, false, nil
	}

	item := inv.UnequipItem(pkt.Slot)
	if item == nil {
		slog.Debug("nothing equipped in slot", "slot", pkt.Slot)
		return 0, false, nil
	}

	// Send updated inventory
	invPkt := serverpackets.NewInventoryItemList(inv.GetItems())
	invData, err := invPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}

	// Broadcast UserInfo for visual update
	userInfo := serverpackets.NewUserInfo(player)
	uiData, err := userInfo.Write()
	if err != nil {
		slog.Error("serializing UserInfo after unequip", "error", err)
	} else {
		h.clientManager.BroadcastToVisible(player, uiData, len(uiData))
	}

	n := copy(buf, invData)
	return n, true, nil
}

// handleRequestDestroyItem processes RequestDestroyItem (opcode 0x59).
// Destroys an item from inventory.
func (h *Handler) handleRequestDestroyItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestDestroyItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestDestroyItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	inv := player.Inventory()
	if inv == nil {
		return 0, false, nil
	}

	objID := uint32(pkt.ObjectID)
	item := inv.GetItem(objID)
	if item == nil {
		slog.Warn("item not found for destroy", "objectID", pkt.ObjectID, "account", client.AccountName())
		return 0, false, nil
	}

	// Cannot destroy equipped items
	if item.IsEquipped() {
		slog.Warn("cannot destroy equipped item", "objectID", pkt.ObjectID)
		return 0, false, nil
	}

	// Remove from inventory
	if item.Template() != nil && item.Template().Stackable && pkt.Count > 0 {
		inv.RemoveItemsByID(item.Template().ItemID, int64(pkt.Count))
	} else {
		inv.RemoveItem(objID)
	}

	// Send updated inventory
	invPkt := serverpackets.NewInventoryItemList(inv.GetItems())
	invData, err := invPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}
	n := copy(buf, invData)

	slog.Debug("item destroyed", "objectID", pkt.ObjectID, "count", pkt.Count, "account", client.AccountName())
	return n, true, nil
}

// handleStartRotating processes StartRotating (opcode 0x4A).
// Broadcasts character rotation start to nearby players.
func (h *Handler) handleStartRotating(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseStartRotating(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing StartRotating: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	rotPkt := &serverpackets.StartRotation{
		ObjectID: int32(player.ObjectID()),
		Degree:   pkt.Degree,
		Side:     pkt.Side,
		Speed:    0,
	}
	rotData, err := rotPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing StartRotation: %w", err)
	}
	h.clientManager.BroadcastToVisibleExcept(player, player, rotData, len(rotData))

	return 0, false, nil
}

// handleFinishRotating processes FinishRotating (opcode 0x4B).
// Broadcasts character rotation stop to nearby players.
func (h *Handler) handleFinishRotating(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseFinishRotating(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing FinishRotating: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	rotPkt := &serverpackets.StopRotation{
		ObjectID: int32(player.ObjectID()),
		Degree:   pkt.Degree,
		Speed:    0,
	}
	rotData, err := rotPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing StopRotation: %w", err)
	}
	h.clientManager.BroadcastToVisibleExcept(player, player, rotData, len(rotData))

	return 0, false, nil
}

// tradeRequestDistance is the maximum distance (game units) for initiating a trade.
const tradeRequestDistance = 150

// handleTradeRequest handles TradeRequest (0x15) — player initiates trade with target.
// Validates distance, store mode, and existing trade, then sends request to target.
// Java reference: TradeRequest.java
func (h *Handler) handleTradeRequest(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseTradeRequest(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing TradeRequest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	// Нельзя торговать, находясь в private store
	if player.PrivateStoreType() != model.StoreNone {
		return 0, true, nil
	}

	// Нельзя начать новый трейд, пока в активном
	if player.IsProcessingTransaction() {
		return 0, true, nil
	}

	// Находим целевого игрока
	targetClient := h.clientManager.GetClientByObjectID(uint32(pkt.TargetObjectID))
	if targetClient == nil {
		return 0, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		return 0, true, nil
	}

	// Нельзя торговать с самим собой
	if player.ObjectID() == targetPlayer.ObjectID() {
		return 0, true, nil
	}

	// Целевой игрок в store mode — нельзя
	if targetPlayer.PrivateStoreType() != model.StoreNone {
		return 0, true, nil
	}

	// Целевой игрок уже в трейде
	if targetPlayer.IsProcessingTransaction() {
		return 0, true, nil
	}

	// Проверяем дистанцию
	loc := player.Location()
	tLoc := targetPlayer.Location()
	dx := float64(loc.X - tLoc.X)
	dy := float64(loc.Y - tLoc.Y)
	dz := float64(loc.Z - tLoc.Z)
	distSq := dx*dx + dy*dy + dz*dz
	if distSq > tradeRequestDistance*tradeRequestDistance {
		return 0, true, nil
	}

	// Устанавливаем trade request state на целевом игроке
	targetPlayer.OnTransactionRequest(player)

	// Отправляем SendTradeRequest целевому игроку
	reqPkt := serverpackets.NewSendTradeRequest(int32(player.ObjectID()))
	reqData, err := reqPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SendTradeRequest: %w", err)
	}
	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), reqData, len(reqData)); err != nil {
		slog.Warn("trade request: send to target",
			"target", targetPlayer.Name(),
			"error", err)
	}

	slog.Debug("trade request sent",
		"from", player.Name(),
		"to", targetPlayer.Name())

	return 0, true, nil
}

// handleAnswerTradeRequest handles AnswerTradeRequest (0x44) — player responds to trade request.
// On accept: creates trade lists for both players and sends TradeStart.
// On reject: sends TradeDone(0) to requester.
// Java reference: AnswerTradeRequest.java
func (h *Handler) handleAnswerTradeRequest(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAnswerTradeRequest(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AnswerTradeRequest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	requester := player.ActiveRequester()
	if requester == nil {
		return 0, true, nil
	}

	player.OnTransactionResponse()

	// Проверяем не истёк ли запрос
	if player.IsRequestExpired() {
		player.SetActiveRequester(nil)
		return 0, true, nil
	}

	if pkt.Response != 1 {
		// Отклонить: отправляем TradeDone(0) инициатору
		player.SetActiveRequester(nil)

		cancelPkt := serverpackets.NewTradeDonePacket(0)
		cancelData, writeErr := cancelPkt.Write()
		if writeErr != nil {
			slog.Error("trade reject: serialize TradeDone",
				"error", writeErr)
			return 0, true, nil
		}
		if err := h.clientManager.SendToPlayer(requester.ObjectID(), cancelData, len(cancelData)); err != nil {
			slog.Warn("trade reject: send TradeDone to requester",
				"error", err)
		}

		slog.Debug("trade rejected",
			"player", player.Name(),
			"requester", requester.Name())
		return 0, true, nil
	}

	// Принять трейд: создаём trade lists для обоих
	playerList := model.NewP2PTradeList(player, requester)
	requesterList := model.NewP2PTradeList(requester, player)

	player.SetActiveTradeList(playerList)
	player.SetActiveRequester(nil)
	requester.SetActiveTradeList(requesterList)

	// Отправляем TradeStart обоим игрокам (tradeable items из инвентаря)
	h.sendTradeStart(player, requester)
	h.sendTradeStart(requester, player)

	slog.Debug("trade started",
		"player1", player.Name(),
		"player2", requester.Name())

	return 0, true, nil
}

// sendTradeStart sends TradeStart packet to player showing partner's ObjectID and player's tradeable items.
func (h *Handler) sendTradeStart(player, partner *model.Player) {
	inv := player.Inventory()
	allItems := inv.GetItems()

	// Фильтруем tradeable items (не equipped, не quest items)
	tradeable := make([]*model.Item, 0, len(allItems))
	for _, item := range allItems {
		if item.IsEquipped() {
			continue
		}
		if tmpl := item.Template(); tmpl != nil && tmpl.Type == model.ItemTypeQuestItem {
			continue
		}
		tradeable = append(tradeable, item)
	}

	pkt := serverpackets.NewTradeStart(int32(partner.ObjectID()), tradeable)
	pktData, err := pkt.Write()
	if err != nil {
		slog.Error("trade start: serialize",
			"player", player.Name(),
			"error", err)
		return
	}
	if err := h.clientManager.SendToPlayer(player.ObjectID(), pktData, len(pktData)); err != nil {
		slog.Warn("trade start: send",
			"player", player.Name(),
			"error", err)
	}
}

// handleAddTradeItem handles AddTradeItem (0x16) — add item to active trade window.
// Validates item is tradeable, not equipped, not quest item, then adds to trade list.
// Sends TradeOwnAdd to self, TradeOtherAdd to partner.
// Java reference: AddTradeItem.java
func (h *Handler) handleAddTradeItem(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAddTradeItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AddTradeItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	if pkt.Count < 1 {
		return 0, true, nil
	}

	tradeList := player.ActiveTradeList()
	if tradeList == nil {
		return 0, true, nil
	}

	// Находим предмет в инвентаре
	inv := player.Inventory()
	item := inv.GetItem(uint32(pkt.ObjectID))
	if item == nil {
		return 0, true, nil
	}

	// Нельзя добавлять equipped предметы
	if item.IsEquipped() {
		return 0, true, nil
	}

	// Нельзя добавлять quest items
	if tmpl := item.Template(); tmpl != nil && tmpl.Type == model.ItemTypeQuestItem {
		return 0, true, nil
	}

	// Добавляем в trade list
	tradeItem, addErr := tradeList.AddItem(pkt.ObjectID, pkt.Count)
	if addErr != nil {
		slog.Warn("add trade item",
			"player", player.Name(),
			"error", addErr)
		return 0, true, nil
	}

	// Получаем template для пакета
	tmpl := item.Template()
	var type1, type2 int16
	var bodyPart int32
	if tmpl != nil {
		type1 = tmpl.Type1
		type2 = tmpl.Type2
		bodyPart = tmpl.BodyPartMask
	}

	// Отправляем TradeOwnAdd игроку
	ownAdd := serverpackets.NewTradeOwnAdd(
		type1,
		int32(item.ObjectID()),
		item.ItemID(),
		tradeItem.Count,
		type2,
		bodyPart,
		int16(tradeItem.Enchant),
	)
	ownData, err := ownAdd.Write()
	if err != nil {
		slog.Error("trade own add: serialize", "error", err)
		return 0, true, nil
	}
	if err := h.clientManager.SendToPlayer(player.ObjectID(), ownData, len(ownData)); err != nil {
		slog.Warn("trade own add: send",
			"player", player.Name(),
			"error", err)
	}

	// Отправляем TradeOtherAdd партнёру
	partner := tradeList.Partner()
	if partner != nil {
		otherAdd := serverpackets.NewTradeOtherAdd(
			type1,
			int32(item.ObjectID()),
			item.ItemID(),
			tradeItem.Count,
			type2,
			bodyPart,
			int16(tradeItem.Enchant),
		)
		otherData, err := otherAdd.Write()
		if err != nil {
			slog.Error("trade other add: serialize", "error", err)
			return 0, true, nil
		}
		if err := h.clientManager.SendToPlayer(partner.ObjectID(), otherData, len(otherData)); err != nil {
			slog.Warn("trade other add: send",
				"partner", partner.Name(),
				"error", err)
		}
	}

	slog.Debug("trade item added",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"count", tradeItem.Count)

	return 0, true, nil
}

// handleTradeDone handles TradeDone (0x17) — confirm or cancel active trade.
// response=1: confirm trade. If both confirm, exchange items.
// response=0: cancel trade for both players.
// Java reference: TradeDone.java
func (h *Handler) handleTradeDone(_ context.Context, client *GameClient, data, _ []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseTradeDone(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing TradeDone: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	tradeList := player.ActiveTradeList()
	if tradeList == nil {
		return 0, true, nil
	}

	partner := tradeList.Partner()
	if partner == nil {
		player.CancelActiveTrade()
		return 0, true, nil
	}

	// Cancel
	if pkt.Response != 1 {
		h.cancelTrade(player, partner)
		return 0, true, nil
	}

	// Confirm
	if !tradeList.Confirm() {
		// Уже подтверждён — игнорируем
		return 0, true, nil
	}

	// Уведомляем партнёра что мы подтвердили
	otherDone := serverpackets.NewTradeOtherDone()
	otherDoneData, err := otherDone.Write()
	if err != nil {
		slog.Error("trade other done: serialize", "error", err)
		return 0, true, nil
	}
	if err := h.clientManager.SendToPlayer(partner.ObjectID(), otherDoneData, len(otherDoneData)); err != nil {
		slog.Warn("trade other done: send",
			"partner", partner.Name(),
			"error", err)
	}

	// Проверяем подтвердил ли партнёр
	partnerList := partner.ActiveTradeList()
	if partnerList == nil || !partnerList.IsConfirmed() {
		return 0, true, nil
	}

	// Оба подтвердили — обмениваемся предметами
	h.executeTrade(player, partner, tradeList, partnerList)

	return 0, true, nil
}

// cancelTrade cancels P2P trade for both players and notifies them.
func (h *Handler) cancelTrade(player, partner *model.Player) {
	player.CancelActiveTrade()

	// Отправляем TradeDone(0) обоим
	cancelPkt := serverpackets.NewTradeDonePacket(0)
	cancelData, err := cancelPkt.Write()
	if err != nil {
		slog.Error("trade cancel: serialize", "error", err)
		return
	}

	if err := h.clientManager.SendToPlayer(player.ObjectID(), cancelData, len(cancelData)); err != nil {
		slog.Warn("trade cancel: send to player",
			"player", player.Name(),
			"error", err)
	}
	if err := h.clientManager.SendToPlayer(partner.ObjectID(), cancelData, len(cancelData)); err != nil {
		slog.Warn("trade cancel: send to partner",
			"partner", partner.Name(),
			"error", err)
	}

	slog.Debug("trade cancelled",
		"player", player.Name(),
		"partner", partner.Name())
}

// executeTrade exchanges items between two players.
// Locks trade lists in ObjectID order to prevent deadlocks.
func (h *Handler) executeTrade(player, partner *model.Player, playerList, partnerList *model.P2PTradeList) {
	// Deadlock prevention: блокируем в порядке ObjectID (меньший первым)
	if player.ObjectID() < partner.ObjectID() {
		playerList.Lock()
		partnerList.Lock()
	} else {
		partnerList.Lock()
		playerList.Lock()
	}

	playerItems := playerList.Items()
	partnerItems := partnerList.Items()

	playerInv := player.Inventory()
	partnerInv := partner.Inventory()

	// Перемещаем предметы от player к partner
	for _, ti := range playerItems {
		item := playerInv.RemoveItem(uint32(ti.ObjectID))
		if item == nil {
			slog.Warn("trade execute: item not found",
				"player", player.Name(),
				"objectID", ti.ObjectID)
			continue
		}
		if err := partnerInv.AddItem(item); err != nil {
			slog.Error("trade execute: add item to partner",
				"partner", partner.Name(),
				"objectID", ti.ObjectID,
				"error", err)
		}
	}

	// Перемещаем предметы от partner к player
	for _, ti := range partnerItems {
		item := partnerInv.RemoveItem(uint32(ti.ObjectID))
		if item == nil {
			slog.Warn("trade execute: item not found",
				"partner", partner.Name(),
				"objectID", ti.ObjectID)
			continue
		}
		if err := playerInv.AddItem(item); err != nil {
			slog.Error("trade execute: add item to player",
				"player", player.Name(),
				"objectID", ti.ObjectID,
				"error", err)
		}
	}

	// Очищаем trade state
	player.CancelActiveTrade()

	// Отправляем TradeDone(1) обоим — успех
	successPkt := serverpackets.NewTradeDonePacket(1)
	successData, err := successPkt.Write()
	if err != nil {
		slog.Error("trade success: serialize", "error", err)
		return
	}

	if err := h.clientManager.SendToPlayer(player.ObjectID(), successData, len(successData)); err != nil {
		slog.Warn("trade success: send to player",
			"player", player.Name(),
			"error", err)
	}
	if err := h.clientManager.SendToPlayer(partner.ObjectID(), successData, len(successData)); err != nil {
		slog.Warn("trade success: send to partner",
			"partner", partner.Name(),
			"error", err)
	}

	slog.Info("trade completed",
		"player", player.Name(),
		"partner", partner.Name(),
		"playerItems", len(playerItems),
		"partnerItems", len(partnerItems))
}

// handleRequestEnchantItem handles RequestEnchantItem (0x58) -- enchant an item.
//
// Validates scroll/item compatibility (type, grade, max level), consumes the scroll,
// rolls enchant chance via enchant.TryEnchant, and applies the result.
//
// Java reference: RequestEnchantItem.java, EnchantScroll.java
func (h *Handler) handleRequestEnchantItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestEnchantItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestEnchantItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, nil
	}

	// Получаем ObjectID скролла из activeEnchantItemID (установлен в handleUseItem)
	scrollObjID := player.ActiveEnchantItemID()
	if scrollObjID == 0 {
		slog.Warn("enchant: no active scroll",
			"player", player.Name())
		return h.sendEnchantResult(buf, 0)
	}

	// Всегда сбрасываем enchant state после обработки
	defer player.SetActiveEnchantItemID(0)

	inv := player.Inventory()

	// Находим скролл в инвентаре
	scroll := inv.GetItem(uint32(scrollObjID))
	if scroll == nil {
		slog.Warn("enchant: scroll not found",
			"player", player.Name(),
			"scrollObjID", scrollObjID)
		return h.sendEnchantResult(buf, 0)
	}

	// Проверяем что скролл -- действительно enchant scroll
	scrollInfo, ok := enchant.IsScroll(scroll.ItemID())
	if !ok {
		slog.Warn("enchant: item is not a scroll",
			"player", player.Name(),
			"scrollItemID", scroll.ItemID())
		return h.sendEnchantResult(buf, 0)
	}

	// Находим предмет для заточки
	item := inv.GetItem(uint32(pkt.ObjectID))
	if item == nil {
		slog.Warn("enchant: target item not found",
			"player", player.Name(),
			"objectID", pkt.ObjectID)
		return h.sendEnchantResult(buf, 0)
	}

	// Валидация: тип, грейд, max level, enchantability
	if reason := enchant.Validate(scrollInfo, item); reason != "" {
		slog.Warn("enchant: validation failed",
			"player", player.Name(),
			"itemID", item.ItemID(),
			"scrollID", scroll.ItemID(),
			"reason", reason)
		return h.sendEnchantResult(buf, 0)
	}

	// Потребляем 1 скролл
	inv.RemoveItemsByID(scroll.ItemID(), 1)

	// Выполняем попытку заточки
	result := enchant.TryEnchant(scrollInfo, item)

	if result.Success {
		if err := item.SetEnchant(result.NewEnchant); err != nil {
			slog.Error("enchant: set enchant level",
				"player", player.Name(),
				"error", err)
			return h.sendEnchantResult(buf, 0)
		}

		slog.Info("enchant success",
			"player", player.Name(),
			"item", item.Name(),
			"newEnchant", result.NewEnchant)

		h.sendInventoryUpdateModify(client, item)
		return h.sendEnchantResult(buf, result.NewEnchant)
	}

	// Провал заточки
	slog.Info("enchant failed",
		"player", player.Name(),
		"item", item.Name(),
		"enchant", item.Enchant(),
		"scrollType", scrollInfo.ScrollType)

	if result.Destroyed {
		// Обычный скролл: предмет уничтожается
		inv.RemoveItem(item.ObjectID())
		h.sendInventoryUpdateRemove(client, item)
	} else if result.NewEnchant == 0 && item.Enchant() > 0 {
		// Blessed скролл: enchant → 0
		if err := item.SetEnchant(0); err != nil {
			slog.Error("enchant blessed fail: reset enchant",
				"player", player.Name(),
				"error", err)
		}
		h.sendInventoryUpdateModify(client, item)
	}
	// Crystal scroll fail: ничего не происходит с предметом (safe fail)

	return h.sendEnchantResult(buf, 0)
}

// sendEnchantResult writes EnchantResult packet to buf.
func (h *Handler) sendEnchantResult(buf []byte, newEnchant int32) (int, bool, error) {
	resultPkt := serverpackets.NewEnchantResult(newEnchant)
	resultData, err := resultPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing EnchantResult: %w", err)
	}
	n := copy(buf, resultData)
	return n, true, nil
}

// sendInventoryUpdateModify sends an InventoryUpdate with modify change for a single item.
func (h *Handler) sendInventoryUpdateModify(client *GameClient, item *model.Item) {
	changes := []serverpackets.InvUpdateEntry{
		{ChangeType: serverpackets.InvUpdateModify, Item: item},
	}
	invUpdate := serverpackets.NewInventoryUpdate(changes...)
	invData, err := invUpdate.Write()
	if err != nil {
		slog.Error("serializing InventoryUpdate (modify)",
			"error", err)
		return
	}
	if err := h.clientManager.SendToPlayer(client.ActivePlayer().ObjectID(), invData, len(invData)); err != nil {
		slog.Error("sending InventoryUpdate (modify)",
			"error", err)
	}
}

// sendInventoryUpdateRemove sends an InventoryUpdate with remove change for a single item.
func (h *Handler) sendInventoryUpdateRemove(client *GameClient, item *model.Item) {
	changes := []serverpackets.InvUpdateEntry{
		{ChangeType: serverpackets.InvUpdateRemove, Item: item},
	}
	invUpdate := serverpackets.NewInventoryUpdate(changes...)
	invData, err := invUpdate.Write()
	if err != nil {
		slog.Error("serializing InventoryUpdate (remove)",
			"error", err)
		return
	}
	if err := h.clientManager.SendToPlayer(client.ActivePlayer().ObjectID(), invData, len(invData)); err != nil {
		slog.Error("sending InventoryUpdate (remove)",
			"error", err)
	}
}

// --- Shortcut handlers (Phase 34) ---

// handleRequestShortCutReg registers a shortcut in the action bar (C2S 0x33).
//
// Reference: L2J_Mobius RequestShortcutReg.java
func (h *Handler) handleRequestShortCutReg(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestShortCutReg(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestShortCutReg: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Validate page/slot bounds
	if pkt.Page < 0 || pkt.Page >= model.MaxShortcutPages || pkt.Slot < 0 || pkt.Slot >= model.MaxShortcutsPerBar {
		return 0, true, nil
	}

	// Java: for NONE type, just ignore
	if pkt.Type == model.ShortcutTypeNone {
		return 0, true, nil
	}

	level := int32(-1)

	switch pkt.Type {
	case model.ShortcutTypeSkill:
		// Verify player knows this skill and use actual server-side level
		skillLevel := player.GetSkillLevel(pkt.ID)
		if skillLevel <= 0 {
			return 0, true, nil
		}
		level = skillLevel

	case model.ShortcutTypeItem:
		// Verify item exists in inventory
		inv := player.Inventory()
		if inv == nil || inv.GetItem(uint32(pkt.ID)) == nil {
			return 0, true, nil
		}

	case model.ShortcutTypeAction, model.ShortcutTypeMacro, model.ShortcutTypeRecipe:
		// No additional validation needed
	}

	sc := &model.Shortcut{
		Slot:  pkt.Slot,
		Page:  pkt.Page,
		Type:  pkt.Type,
		ID:    pkt.ID,
		Level: level,
	}

	player.RegisterShortcut(sc)

	// Send ShortCutRegister confirmation to client
	regPkt := serverpackets.NewShortCutRegister(sc)
	regData, err := regPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ShortCutRegister: %w", err)
	}
	n := copy(buf, regData)

	return n, true, nil
}

// handleRequestShortCutDel deletes a shortcut from the action bar (C2S 0x35).
//
// Reference: L2J_Mobius RequestShortcutDel.java
func (h *Handler) handleRequestShortCutDel(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestShortCutDel(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestShortCutDel: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if pkt.Page < 0 || pkt.Page >= model.MaxShortcutPages || pkt.Slot < 0 || pkt.Slot >= model.MaxShortcutsPerBar {
		return 0, true, nil
	}

	player.DeleteShortcut(pkt.Slot, pkt.Page)

	// Java behaviour: re-send full shortcut list after deletion
	initPkt := serverpackets.NewShortCutInit(player.GetShortcuts())
	initData, err := initPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ShortCutInit: %w", err)
	}
	n := copy(buf, initData)

	return n, true, nil
}
