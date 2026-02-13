package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeUserInfo is the opcode for UserInfo packet (S2C 0x04).
	// Java reference: ServerPackets.USER_INFO(0x04)
	OpcodeUserInfo = 0x04
)

// UserInfo packet (S2C 0x04) sends complete character information to the owning client.
// Sent when character enters world, changes equipment, levels up, etc.
// Java reference: UserInfo.java
type UserInfo struct {
	Player *model.Player
}

// NewUserInfo creates UserInfo packet from Player model.
func NewUserInfo(player *model.Player) UserInfo {
	return UserInfo{
		Player: player,
	}
}

// paperdollOrder defines the slot order for paperdoll fields in UserInfo/CharInfo packets.
// Matches Java Inventory.PAPERDOLL_* write order in UserInfo.writeImpl().
// Note: RHAND appears twice (index 7 and 14) for two-hand weapon display.
var paperdollOrder = [17]int32{
	model.PaperdollUnder,   // 0: Underwear
	model.PaperdollREar,    // 1: Right Ear  (Java: PAPERDOLL_REAR)
	model.PaperdollLEar,    // 2: Left Ear   (Java: PAPERDOLL_LEAR)
	model.PaperdollNeck,    // 3: Necklace
	model.PaperdollRFinger, // 4: Right Ring (Java: PAPERDOLL_RFINGER)
	model.PaperdollLFinger, // 5: Left Ring  (Java: PAPERDOLL_LFINGER)
	model.PaperdollHead,    // 6: Helmet
	model.PaperdollRHand,   // 7: Right Hand (weapon)
	model.PaperdollLHand,   // 8: Left Hand  (shield)
	model.PaperdollGloves,  // 9: Gloves
	model.PaperdollChest,   // 10: Chest
	model.PaperdollLegs,    // 11: Legs
	model.PaperdollFeet,    // 12: Boots
	model.PaperdollCloak,   // 13: Cloak
	model.PaperdollRHand,   // 14: RHAND again (two-hand weapon)
	model.PaperdollHair,    // 15: Hair accessory
	model.PaperdollFace,    // 16: Face accessory
}

// Write serializes UserInfo packet to binary format matching Java UserInfo.writeImpl() exactly.
// Java reference: UserInfo.java (L2J Mobius CT_0_Interlude)
func (p *UserInfo) Write() ([]byte, error) {
	w := packet.NewWriter(600)

	pl := p.Player
	inv := pl.Inventory()
	loc := pl.Location()

	// Pre-compute speeds (Java constructor logic).
	// moveMultiplier = runSpeed / baseRunSpeed; base run speed ≈120 for all classes.
	moveMultiplier := 1.0
	runSpd := int32(120)
	walkSpd := int32(80)
	swimRunSpd := runSpd
	swimWalkSpd := walkSpd
	flyRunSpd := int32(0)
	flyWalkSpd := int32(0)

	// Collision from template (use female values when applicable).
	collisionRadius := 9.0  // Default male Human Fighter
	collisionHeight := 23.0
	tmpl := data.GetTemplate(uint8(pl.ClassID()))
	if tmpl != nil {
		if pl.IsFemale() {
			collisionRadius = float64(tmpl.CollisionRadiusFemale)
			collisionHeight = float64(tmpl.CollisionHeightFemale)
		} else {
			collisionRadius = float64(tmpl.CollisionRadiusMale)
			collisionHeight = float64(tmpl.CollisionHeightMale)
		}
	}

	// GM flag
	gmFlag := int32(0)
	if pl.IsGM() {
		gmFlag = 1
	}

	// Active weapon check (Java: activeWeaponItem != null ? 40 : 20)
	activeWeaponFlag := int32(20) // no weapon
	if pl.GetEquippedWeapon() != nil {
		activeWeaponFlag = 40
	}

	// Title — override with "[Invisible]" for invisible GMs
	title := pl.Title()
	if pl.IsGM() && pl.IsInvisible() {
		title = "[Invisible]"
	}

	// isFemale as int32 for packet
	isFemale := int32(0)
	if pl.IsFemale() {
		isFemale = 1
	}

	// Relation flags (siege state).
	// Java: _relation = isClanLeader() ? 0x40 : 0; if siegeState==1: |= 0x180; if siegeState==2: |= 0x80
	// Siege state not yet wired — using 0.
	relation := int32(0)

	// Abnormal visual effects (add STEALTH mask if invisible).
	abnormalEffects := pl.AbnormalVisualEffects()
	if pl.IsInvisible() {
		abnormalEffects |= 0x1000 // AbnormalVisualEffect.STEALTH mask
	}

	// Water zone
	waterZone := byte(0)
	if pl.IsInsideZone(model.ZoneIDWater) {
		waterZone = 1
	}

	// hasDwarvenCraft
	hasDwarvenCraft := byte(0)
	if pl.HasDwarvenCraft() {
		hasDwarvenCraft = 1
	}

	// isRunning
	isRunning := byte(0)
	if pl.IsRunning() {
		isRunning = 1
	}

	// Noble
	noble := byte(0)
	if pl.IsNoble() {
		noble = 1
	}

	// Hero
	hero := byte(0)
	if pl.IsHero() || (pl.IsGM() && pl.IsInvisible()) {
		hero = 1
	}

	// Fishing
	fishingByte := byte(0)
	if pl.IsFishing() {
		fishingByte = 1
	}

	// Mount NPC ID (Java: mountNpcId > 0 ? mountNpcId + 1000000 : 0)
	mountNpcDisplay := int32(0)
	if pl.MountNpcID() > 0 {
		mountNpcDisplay = pl.MountNpcID() + 1000000
	}

	// Enchant effect (0 if mounted)
	enchantEffect := byte(pl.GetEnchantEffect())

	// --- Packet body (exact Java field order) ---

	w.WriteByte(OpcodeUserInfo)

	// Position
	w.WriteInt(loc.X)                        // writeInt(x)
	w.WriteInt(loc.Y)                        // writeInt(y)
	w.WriteInt(loc.Z)                        // writeInt(z)
	w.WriteInt(0)                            // writeInt(vehicle objectId) — vehicle system not implemented
	w.WriteInt(int32(pl.ObjectID()))         // writeInt(objectId)
	w.WriteString(pl.Name())                 // writeString(visibleName)
	w.WriteInt(pl.RaceID())                  // writeInt(race ordinal)
	w.WriteInt(isFemale)                     // writeInt(isFemale)
	w.WriteInt(pl.BaseClassID())             // writeInt(baseClass)
	w.WriteInt(pl.Level())                   // writeInt(level)
	w.WriteLong(pl.Experience())             // writeLong(exp)

	// Base stats (STR/DEX/CON/INT/WIT/MEN)
	w.WriteInt(int32(pl.GetSTR()))           // writeInt(STR)
	w.WriteInt(int32(pl.GetDEX()))           // writeInt(DEX)
	w.WriteInt(int32(pl.GetCON()))           // writeInt(CON)
	w.WriteInt(int32(pl.GetINT()))           // writeInt(INT)
	w.WriteInt(int32(pl.GetWIT()))           // writeInt(WIT)
	w.WriteInt(int32(pl.GetMEN()))           // writeInt(MEN)

	// Vitals
	w.WriteInt(pl.MaxHP())                   // writeInt(maxHp)
	w.WriteInt(pl.CurrentHP())               // writeInt(currentHp)
	w.WriteInt(pl.MaxMP())                   // writeInt(maxMp)
	w.WriteInt(pl.CurrentMP())               // writeInt(currentMp)
	w.WriteInt(int32(pl.SP()))               // writeInt(sp)
	w.WriteInt(pl.GetCurrentLoad())          // writeInt(currentLoad)
	w.WriteInt(pl.GetMaxLoad())              // writeInt(maxLoad)
	w.WriteInt(activeWeaponFlag)             // writeInt(weapon ? 40 : 20)

	// --- Paperdoll ObjectIDs (17 slots in Java order) ---
	for _, slot := range paperdollOrder {
		item := inv.GetPaperdollItem(slot)
		if item != nil {
			w.WriteInt(int32(item.ObjectID()))
		} else {
			w.WriteInt(0)
		}
	}

	// --- Paperdoll ItemDisplayIDs (17 slots in Java order) ---
	for _, slot := range paperdollOrder {
		item := inv.GetPaperdollItem(slot)
		if item != nil {
			w.WriteInt(item.ItemID())
		} else {
			w.WriteInt(0)
		}
	}

	// --- C6 augmentation shorts ---
	// 14 shorts (element attributes placeholder)
	for range 14 {
		w.WriteShort(0)
	}
	// Augmentation ID for RHAND
	rhandItem := inv.GetPaperdollItem(model.PaperdollRHand)
	rhandAugID := int32(0)
	if rhandItem != nil {
		rhandAugID = rhandItem.AugmentationID()
	}
	w.WriteInt(rhandAugID)

	// 12 more shorts
	for range 12 {
		w.WriteShort(0)
	}
	// Augmentation ID for RHAND (again, for two-hand)
	w.WriteInt(rhandAugID)

	// 4 more shorts
	for range 4 {
		w.WriteShort(0)
	}
	// --- End C6 augmentation shorts ---

	// Combat stats
	w.WriteInt(pl.GetPAtk())                 // writeInt(pAtk)
	w.WriteInt(int32(pl.GetPAtkSpd()))       // writeInt(pAtkSpd)
	w.WriteInt(pl.GetPDef())                 // writeInt(pDef)
	w.WriteInt(pl.GetEvasionRate())          // writeInt(evasionRate)
	w.WriteInt(pl.GetAccuracy())             // writeInt(accuracy)
	w.WriteInt(pl.GetCriticalHit())          // writeInt(criticalHit)
	w.WriteInt(pl.GetMAtk())                 // writeInt(mAtk)
	w.WriteInt(pl.GetMAtkSpd())              // writeInt(mAtkSpd)
	w.WriteInt(int32(pl.GetPAtkSpd()))       // writeInt(pAtkSpd) — yes, pAtkSpd again!
	w.WriteInt(pl.GetMDef())                 // writeInt(mDef)

	// PvP/Karma
	w.WriteInt(pl.PvPFlag())                 // writeInt(pvpFlag)
	w.WriteInt(pl.Karma())                   // writeInt(karma)

	// Movement speeds
	w.WriteInt(runSpd)                       // writeInt(runSpd)
	w.WriteInt(walkSpd)                      // writeInt(walkSpd)
	w.WriteInt(swimRunSpd)                   // writeInt(swimRunSpd)
	w.WriteInt(swimWalkSpd)                  // writeInt(swimWalkSpd)
	w.WriteInt(flyRunSpd)                    // writeInt(flyRunSpd)
	w.WriteInt(flyWalkSpd)                   // writeInt(flyWalkSpd)
	w.WriteInt(flyRunSpd)                    // writeInt(flyRunSpd) duplicate
	w.WriteInt(flyWalkSpd)                   // writeInt(flyWalkSpd) duplicate
	w.WriteDouble(moveMultiplier)            // writeDouble(moveMultiplier)
	w.WriteDouble(pl.GetAttackSpeedMultiplier()) // writeDouble(attackSpeedMultiplier)
	w.WriteDouble(collisionRadius)           // writeDouble(collisionRadius)
	w.WriteDouble(collisionHeight)           // writeDouble(collisionHeight)

	// Appearance
	w.WriteInt(pl.HairStyle())               // writeInt(hairStyle)
	w.WriteInt(pl.HairColor())               // writeInt(hairColor)
	w.WriteInt(pl.Face())                    // writeInt(face)
	w.WriteInt(gmFlag)                       // writeInt(isGM / builder level)

	// Title
	w.WriteString(title)                     // writeString(title)

	// Clan
	w.WriteInt(pl.ClanID())                  // writeInt(clanId)
	w.WriteInt(0)                            // writeInt(clanCrestId) — loaded from Clan table, not on Player
	w.WriteInt(0)                            // writeInt(allyId) — loaded from Clan table
	w.WriteInt(0)                            // writeInt(allyCrestId) — loaded from Clan table

	// Relation (siege flags)
	w.WriteInt(relation)                     // writeInt(relation)

	// Mount, Store, DwarvenCraft — BYTES (not ints!)
	_ = w.WriteByte(byte(pl.MountType()))    // writeByte(mountType)
	_ = w.WriteByte(byte(pl.PrivateStoreType())) // writeByte(privateStoreType)
	_ = w.WriteByte(hasDwarvenCraft)         // writeByte(hasDwarvenCraft)

	// PK/PvP kills
	w.WriteInt(pl.PKKills())                 // writeInt(pkKills)
	w.WriteInt(pl.PvPKills())                // writeInt(pvpKills)

	// Cubics — cubics system not implemented, send count 0
	w.WriteShort(0)                          // writeShort(cubics.size())

	// Party match room — not implemented
	_ = w.WriteByte(0)                       // writeByte(isInPartyMatchRoom)

	// Abnormal visual effects
	w.WriteInt(abnormalEffects)              // writeInt(abnormalVisualEffects)

	// Water zone
	_ = w.WriteByte(waterZone)               // writeByte(isInsideZone(WATER))

	// Clan privileges — loaded from Clan table, not on Player
	w.WriteInt(0)                            // writeInt(clanPrivileges mask)

	// Recommendations
	w.WriteShort(int16(pl.RecomLeft()))      // writeShort(recomLeft)
	w.WriteShort(int16(pl.RecomHave()))      // writeShort(recomHave)

	// Mount NPC
	w.WriteInt(mountNpcDisplay)              // writeInt(mountNpcId + 1000000 or 0)

	// Inventory limit
	w.WriteShort(int16(pl.GetInventoryLimit())) // writeShort(inventoryLimit)

	// Class ID
	w.WriteInt(pl.ClassID())                 // writeInt(playerClass id)

	// Special effects
	w.WriteInt(0)                            // writeInt(0) special effects

	// CP
	w.WriteInt(pl.MaxCP())                   // writeInt(maxCp)
	w.WriteInt(pl.CurrentCP())               // writeInt(currentCp)

	// Enchant effect
	_ = w.WriteByte(enchantEffect)           // writeByte(enchantEffect)

	// Team
	_ = w.WriteByte(byte(pl.TeamID()))       // writeByte(teamId)

	// Large clan crest — loaded from Clan table
	w.WriteInt(0)                            // writeInt(clanCrestLargeId)

	// Noble
	_ = w.WriteByte(noble)                   // writeByte(isNoble)

	// Hero aura
	_ = w.WriteByte(hero)                    // writeByte(isHero)

	// Fishing
	_ = w.WriteByte(fishingByte)             // writeByte(isFishing)
	w.WriteInt(pl.FishX())                   // writeInt(fishX)
	w.WriteInt(pl.FishY())                   // writeInt(fishY)
	w.WriteInt(pl.FishZ())                   // writeInt(fishZ)

	// Name color (BGR format)
	w.WriteInt(pl.NameColor())               // writeInt(nameColor)

	// Running state
	_ = w.WriteByte(isRunning)               // writeByte(isRunning)

	// Pledge
	w.WriteInt(pl.PledgeClass())             // writeInt(pledgeClass)
	w.WriteInt(pl.PledgeType())              // writeInt(pledgeType)

	// Title color (BGR format)
	w.WriteInt(pl.TitleColor())              // writeInt(titleColor)

	// Cursed weapon level — CursedWeaponsManager lookup not wired to Player, use simplified
	cursedLevel := int32(0)
	if pl.IsCursedWeaponEquipped() {
		cursedLevel = 1
	}
	w.WriteInt(cursedLevel)                  // writeInt(cursedWeaponLevel)

	return w.Bytes(), nil
}
