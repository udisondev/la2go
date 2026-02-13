package gameserver

import (
	"context"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// handleRequestCursedWeaponList handles C2S 0xD0:0x22.
// Sends the list of all cursed weapon item IDs.
// Java: RequestCursedWeaponList → ExCursedWeaponList.
func (h *Handler) handleRequestCursedWeaponList(_ context.Context, client *GameClient, _ []byte, buf []byte) (int, bool, error) {
	if h.cursedMgr == nil {
		return 0, true, nil
	}

	ids := h.cursedMgr.CursedWeaponIDs()
	pkt := &serverpackets.ExCursedWeaponList{WeaponIDs: ids}

	data, err := pkt.Write()
	if err != nil {
		slog.Error("write ExCursedWeaponList", "error", err)
		return 0, true, nil
	}

	n := copy(buf, data)
	return n, true, nil
}

// handleRequestCursedWeaponLocation handles C2S 0xD0:0x23.
// Sends positions of active cursed weapons.
// Java: RequestCursedWeaponLocation → ExCursedWeaponLocation.
func (h *Handler) handleRequestCursedWeaponLocation(_ context.Context, client *GameClient, _ []byte, buf []byte) (int, bool, error) {
	if h.cursedMgr == nil {
		return 0, true, nil
	}

	infos := h.cursedMgr.LocationInfo()

	weapons := make([]serverpackets.CursedWeaponLocationInfo, len(infos))
	for i, info := range infos {
		weapons[i] = serverpackets.CursedWeaponLocationInfo{
			ItemID:    info.ItemID,
			Activated: info.Activated,
			X:         info.X,
			Y:         info.Y,
			Z:         info.Z,
		}
	}

	pkt := &serverpackets.ExCursedWeaponLocation{Weapons: weapons}

	data, err := pkt.Write()
	if err != nil {
		slog.Error("write ExCursedWeaponLocation", "error", err)
		return 0, true, nil
	}

	n := copy(buf, data)
	return n, true, nil
}
