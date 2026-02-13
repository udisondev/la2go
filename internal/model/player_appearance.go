package model

// --- Appearance getters/setters ---

// Title returns the player's title (shown above character).
func (p *Player) Title() string {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.title
}

// SetTitle sets the player's title.
func (p *Player) SetTitle(title string) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.title = title
}

// IsFemale returns true if the player character is female.
func (p *Player) IsFemale() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.isFemale
}

// SetIsFemale sets the player's sex.
func (p *Player) SetIsFemale(female bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.isFemale = female
}

// HairStyle returns the hair style index.
func (p *Player) HairStyle() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hairStyle
}

// SetHairStyle sets the hair style index.
func (p *Player) SetHairStyle(style int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.hairStyle = style
}

// HairColor returns the hair color index.
func (p *Player) HairColor() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hairColor
}

// SetHairColor sets the hair color index.
func (p *Player) SetHairColor(color int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.hairColor = color
}

// Face returns the face type index.
func (p *Player) Face() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.face
}

// SetFace sets the face type index.
func (p *Player) SetFace(face int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.face = face
}

// NameColor returns the name color in BGR format (default 0xFFFFFF = white).
func (p *Player) NameColor() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.nameColor
}

// SetNameColor sets the name color in BGR format.
func (p *Player) SetNameColor(color int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.nameColor = color
}

// TitleColor returns the title color in BGR format (default 0xFFFF77 = light yellow).
func (p *Player) TitleColor() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.titleColor
}

// SetTitleColor sets the title color in BGR format.
func (p *Player) SetTitleColor(color int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.titleColor = color
}

// --- PvP getters/setters ---

// PvPKills returns the player's PvP kill count.
func (p *Player) PvPKills() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pvpKills
}

// SetPvPKills sets the player's PvP kill count.
func (p *Player) SetPvPKills(count int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pvpKills = count
}

// PvPFlag returns the PvP flag state (0=unflagged, 1=flagged).
func (p *Player) PvPFlag() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pvpFlag
}

// SetPvPFlag sets the PvP flag state.
func (p *Player) SetPvPFlag(flag int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pvpFlag = flag
}

// --- Movement state ---

// IsRunning returns true if the player is running (vs walking).
func (p *Player) IsRunning() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.running
}

// SetRunning sets the running/walking state.
func (p *Player) SetRunning(running bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.running = running
}

// IsSitting returns true if the player is sitting.
func (p *Player) IsSitting() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.sitting
}

// SetSitting sets the sitting/standing state.
func (p *Player) SetSitting(sitting bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.sitting = sitting
}

// --- Noble/Hero ---

// IsNoble returns true if the player has noble status.
func (p *Player) IsNoble() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.noble
}

// SetNoble sets noble status.
func (p *Player) SetNoble(noble bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.noble = noble
}

// IsHero returns true if the player has hero status.
func (p *Player) IsHero() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.hero
}

// SetHero sets hero status.
func (p *Player) SetHero(hero bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.hero = hero
}

// --- Fishing ---

// IsFishing returns true if the player is currently fishing.
func (p *Player) IsFishing() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.fishing
}

// SetFishing sets fishing state.
func (p *Player) SetFishing(fishing bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.fishing = fishing
}

// FishX returns the fishing spot X coordinate.
func (p *Player) FishX() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.fishX
}

// FishY returns the fishing spot Y coordinate.
func (p *Player) FishY() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.fishY
}

// FishZ returns the fishing spot Z coordinate.
func (p *Player) FishZ() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.fishZ
}

// SetFishLoc sets the fishing spot coordinates.
func (p *Player) SetFishLoc(x, y, z int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.fishX = x
	p.fishY = y
	p.fishZ = z
}

// --- Clan display ---

// PledgeClass returns the clan pledge class (shown above CP bar).
// Java reference: Player.getPledgeClass()
func (p *Player) PledgeClass() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pledgeClass
}

// SetPledgeClass sets the pledge class.
func (p *Player) SetPledgeClass(class int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pledgeClass = class
}

// PledgeType returns the sub-pledge type.
func (p *Player) PledgeType() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.pledgeType
}

// SetPledgeType sets the sub-pledge type.
func (p *Player) SetPledgeType(pledgeType int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.pledgeType = pledgeType
}

// --- Recommendations ---

// RecomLeft returns remaining recommendations the player can give.
func (p *Player) RecomLeft() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.recomLeft
}

// SetRecomLeft sets remaining recommendations.
func (p *Player) SetRecomLeft(left int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.recomLeft = left
}

// RecomHave returns the number of recommendations received.
func (p *Player) RecomHave() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.recomHave
}

// SetRecomHave sets received recommendations count.
func (p *Player) SetRecomHave(have int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.recomHave = have
}

// --- Abnormal visual effects ---

// AbnormalVisualEffects returns the bitmask of active visual effects.
func (p *Player) AbnormalVisualEffects() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.abnormalVisualEffects
}

// SetAbnormalVisualEffects sets the visual effects bitmask.
func (p *Player) SetAbnormalVisualEffects(effects int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.abnormalVisualEffects = effects
}

// --- Team ---

// TeamID returns the team ID (0=none, 1=blue, 2=red).
func (p *Player) TeamID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.teamID
}

// SetTeamID sets the team ID.
func (p *Player) SetTeamID(id int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.teamID = id
}

// --- Mount ---

// MountType returns the mount type (0=none, 1=strider, 2=wyvern, 3=wolf).
func (p *Player) MountType() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.mountType
}

// SetMountType sets the mount type.
func (p *Player) SetMountType(mountType int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.mountType = mountType
}

// IsMounted returns true if the player is on a mount.
func (p *Player) IsMounted() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.mountType != 0
}

// MountNpcID returns the NPC template ID of the mount (0 if not mounted).
func (p *Player) MountNpcID() int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()
	return p.mountNpcID
}

// SetMountNpcID sets the mount NPC template ID.
func (p *Player) SetMountNpcID(npcID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()
	p.mountNpcID = npcID
}
