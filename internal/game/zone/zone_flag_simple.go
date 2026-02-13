package zone

import "github.com/udisondev/la2go/internal/model"

// --- NoStoreZone ---
// Java reference: NoStoreZone.java — sets ZoneId.NO_STORE for players.

type NoStoreZone struct{ *BaseZone }

func NewNoStoreZone(base *BaseZone) *NoStoreZone {
	z := &NoStoreZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *NoStoreZone) IsPeace() bool   { return false }
func (z *NoStoreZone) AllowsPvP() bool { return true }
func (z *NoStoreZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoStore, true)
}
func (z *NoStoreZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoStore, false)
}

// --- NoLandingZone ---
// Java reference: NoLandingZone.java — sets ZoneId.NO_LANDING for players.

type NoLandingZone struct{ *BaseZone }

func NewNoLandingZone(base *BaseZone) *NoLandingZone {
	z := &NoLandingZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *NoLandingZone) IsPeace() bool   { return false }
func (z *NoLandingZone) AllowsPvP() bool { return true }
func (z *NoLandingZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoLanding, true)
}
func (z *NoLandingZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoLanding, false)
}

// --- NoSummonFriendZone ---
// Java reference: NoSummonFriendZone.java — sets ZoneId.NO_SUMMON_FRIEND for ALL creatures.

type NoSummonFriendZone struct{ *BaseZone }

func NewNoSummonFriendZone(base *BaseZone) *NoSummonFriendZone {
	z := &NoSummonFriendZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *NoSummonFriendZone) IsPeace() bool   { return false }
func (z *NoSummonFriendZone) AllowsPvP() bool { return true }
func (z *NoSummonFriendZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, true)
}
func (z *NoSummonFriendZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, false)
}

// --- NoRestartZone ---
// Java reference: NoRestartZone.java — sets ZoneId.NO_RESTART for players.

type NoRestartZone struct{ *BaseZone }

func NewNoRestartZone(base *BaseZone) *NoRestartZone {
	z := &NoRestartZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *NoRestartZone) IsPeace() bool   { return false }
func (z *NoRestartZone) AllowsPvP() bool { return true }
func (z *NoRestartZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoRestart, true)
}
func (z *NoRestartZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoRestart, false)
}

// --- HqZone ---
// Java reference: HqZone.java — sets ZoneId.HQ. Zone where 'Build Headquarters' is allowed.

type HqZone struct{ *BaseZone }

func NewHqZone(base *BaseZone) *HqZone {
	z := &HqZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *HqZone) IsPeace() bool   { return false }
func (z *HqZone) AllowsPvP() bool { return true }
func (z *HqZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDHQ, true)
}
func (z *HqZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDHQ, false)
}

// --- ScriptZone ---
// Java reference: ScriptZone.java — sets ZoneId.SCRIPT.

type ScriptZone struct{ *BaseZone }

func NewScriptZone(base *BaseZone) *ScriptZone {
	z := &ScriptZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *ScriptZone) IsPeace() bool   { return false }
func (z *ScriptZone) AllowsPvP() bool { return true }
func (z *ScriptZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDScript, true)
}
func (z *ScriptZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDScript, false)
}

// --- NoPvPZone ---
// Java reference: NoPvPZone.java — sets ZoneId.NO_PVP.

type NoPvPZone struct{ *BaseZone }

func NewNoPvPZone(base *BaseZone) *NoPvPZone {
	z := &NoPvPZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *NoPvPZone) IsPeace() bool   { return false }
func (z *NoPvPZone) AllowsPvP() bool { return false }
func (z *NoPvPZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoPVP, true)
}
func (z *NoPvPZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDNoPVP, false)
}

// --- DerbyTrackZone ---
// Java reference: DerbyTrackZone.java — sets ZoneId.MONSTER_TRACK.

type DerbyTrackZone struct{ *BaseZone }

func NewDerbyTrackZone(base *BaseZone) *DerbyTrackZone {
	z := &DerbyTrackZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}
func (z *DerbyTrackZone) IsPeace() bool   { return false }
func (z *DerbyTrackZone) AllowsPvP() bool { return true }
func (z *DerbyTrackZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDMonsterTrack, true)
}
func (z *DerbyTrackZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDMonsterTrack, false)
}
