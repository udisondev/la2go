package commands

import "github.com/udisondev/la2go/internal/gameserver/admin"

// RegisterAll registers all admin and user commands into the handler.
func RegisterAll(h *admin.Handler, clientMgr ClientManager) {
	// Admin commands (// prefix)
	h.RegisterAdmin(NewTeleport(clientMgr))
	h.RegisterAdmin(&Spawn{})
	h.RegisterAdmin(&Delete{})
	h.RegisterAdmin(NewKill(clientMgr))
	h.RegisterAdmin(NewHeal(clientMgr))
	h.RegisterAdmin(NewRes(clientMgr))
	h.RegisterAdmin(&SetLevel{})
	h.RegisterAdmin(&GiveItem{})
	h.RegisterAdmin(NewAnnounce(clientMgr))
	h.RegisterAdmin(NewKick(clientMgr))
	h.RegisterAdmin(NewBan(clientMgr))
	h.RegisterAdmin(NewJail(clientMgr))
	h.RegisterAdmin(&Invisible{})
	h.RegisterAdmin(&Invul{})
	h.RegisterAdmin(&Speed{})
	h.RegisterAdmin(NewInfo(clientMgr))

	// User commands (/ prefix)
	h.RegisterUser(&Loc{})
	h.RegisterUser(&GameTime{})
	h.RegisterUser(&Unstuck{})
	h.RegisterUser(NewOnline(clientMgr))
}
