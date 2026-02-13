package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeCharInfo is the opcode for CharInfo packet (S2C 0x03).
	// Java: ServerPackets.CHAR_INFO(0x03)
	OpcodeCharInfo = 0x03
)

// CharInfo packet (S2C 0x03) sends information about another player character.
// Sent when player enters visibility range of another player.
// Similar to UserInfo but for OTHER players (not self).
// Java: CharInfo.java
type CharInfo struct {
	Player     *model.Player
	GMSeeInvis bool // true when the receiving client is GM and can see invisible players
}

// NewCharInfo creates CharInfo packet from Player model.
func NewCharInfo(player *model.Player) CharInfo {
	return CharInfo{
		Player: player,
	}
}

// paperdollItemID returns the item template ID for a paperdoll slot, or 0 if empty.
func paperdollItemID(inv *model.Inventory, slot int32) int32 {
	if inv == nil {
		return 0
	}
	item := inv.GetPaperdollItem(slot)
	if item == nil {
		return 0
	}
	return item.ItemID()
}

// paperdollAugID returns the augmentation ID for a paperdoll slot, or 0 if empty.
func paperdollAugID(inv *model.Inventory, slot int32) int32 {
	if inv == nil {
		return 0
	}
	item := inv.GetPaperdollItem(slot)
	if item == nil {
		return 0
	}
	return item.AugmentationID()
}

// Write serializes CharInfo packet to binary format.
// Field order matches Java CharInfo.writeImpl exactly.
func (p *CharInfo) Write() ([]byte, error) {
	w := packet.NewWriter(512)

	pl := p.Player
	loc := pl.Location()
	inv := pl.Inventory()

	// Speed calculations (Java constructor values)
	moveMultiplier := 1.0
	pAtkSpd := int32(pl.GetPAtkSpd())
	mAtkSpd := pl.GetMAtkSpd()
	runSpd := int32(120)
	walkSpd := int32(80)
	swimRunSpd := runSpd
	swimWalkSpd := walkSpd
	flyRunSpd := int32(0)
	flyWalkSpd := int32(0)

	// Collision from template (use female values when applicable)
	collisionRadius := 8.0
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

	isCursedWeapon := pl.CursedWeaponEquippedID() != 0

	// isFemale as int32 for packet
	isFemale := int32(0)
	if pl.IsFemale() {
		isFemale = 1
	}

	// Title — visible title (or "Invisible" if GM see invis)
	title := pl.Title()
	if p.GMSeeInvis {
		title = "Invisible"
	}

	// Standing/Sitting (Java: !isSitting())
	standing := byte(1)
	if pl.IsSitting() {
		standing = 0
	}

	// Running
	runningByte := byte(0)
	if pl.IsRunning() {
		runningByte = 1
	}

	// In combat
	inCombat := byte(0)
	if pl.HasAttackStance() {
		inCombat = 1
	}

	// AlikeDead (dead and not in olympiad)
	alikeDead := byte(0)
	if pl.IsDead() {
		alikeDead = 1
	}

	// Invisible (only show if NOT gmSeeInvis)
	invisByte := byte(0)
	if !p.GMSeeInvis && pl.IsInvisible() {
		invisByte = 1
	}

	// Abnormal visual effects (add STEALTH mask if gmSeeInvis)
	abnormalEffects := pl.AbnormalVisualEffects()
	if p.GMSeeInvis {
		abnormalEffects |= 0x1000 // AbnormalVisualEffect.STEALTH mask
	}

	// Noble
	noble := byte(0)
	if pl.IsNoble() {
		noble = 1
	}

	// Hero
	hero := byte(0)
	if pl.IsHero() {
		hero = 1
	}

	// Fishing
	fishingByte := byte(0)
	if pl.IsFishing() {
		fishingByte = 1
	}

	// Enchant effect (0 if mounted)
	enchantEffect := byte(pl.GetEnchantEffect())

	w.WriteByte(OpcodeCharInfo)

	// --- Position ---
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)
	w.WriteInt(0) // vehicleId — vehicle system not implemented
	w.WriteInt(int32(pl.ObjectID()))
	w.WriteString(pl.Name())
	w.WriteInt(pl.RaceID())
	w.WriteInt(isFemale)
	w.WriteInt(pl.BaseClassID())

	// --- 12 paperdoll item display IDs ---
	w.WriteInt(paperdollItemID(inv, model.PaperdollUnder))
	w.WriteInt(paperdollItemID(inv, model.PaperdollHead))
	w.WriteInt(paperdollItemID(inv, model.PaperdollRHand))
	w.WriteInt(paperdollItemID(inv, model.PaperdollLHand))
	w.WriteInt(paperdollItemID(inv, model.PaperdollGloves))
	w.WriteInt(paperdollItemID(inv, model.PaperdollChest))
	w.WriteInt(paperdollItemID(inv, model.PaperdollLegs))
	w.WriteInt(paperdollItemID(inv, model.PaperdollFeet))
	w.WriteInt(paperdollItemID(inv, model.PaperdollCloak))
	w.WriteInt(paperdollItemID(inv, model.PaperdollRHand))  // duplicate (Java: RHAND again)
	w.WriteInt(paperdollItemID(inv, model.PaperdollHair))
	w.WriteInt(paperdollItemID(inv, model.PaperdollHair2))

	// --- c6 new h's: enchant/augmentation section 1 ---
	w.WriteShort(0) // UNDER enchant
	w.WriteShort(0) // HEAD enchant
	w.WriteShort(0) // RHAND enchant
	w.WriteShort(0) // LHAND enchant
	w.WriteInt(paperdollAugID(inv, model.PaperdollRHand))
	w.WriteShort(0) // GLOVES enchant
	w.WriteShort(0) // CHEST enchant
	w.WriteShort(0) // LEGS enchant
	w.WriteShort(0) // FEET enchant
	w.WriteShort(0) // CLOAK enchant
	w.WriteShort(0) // RHAND enchant (duplicate)
	w.WriteShort(0) // HAIR enchant
	w.WriteShort(0) // HAIR2 enchant
	// Java writes 4 more zeros here (enchant for DECO1-DECO4 or reserved slots)
	w.WriteShort(0) // reserved enchant slot 1
	w.WriteShort(0) // reserved enchant slot 2
	w.WriteShort(0) // reserved enchant slot 3
	w.WriteShort(0) // reserved enchant slot 4

	// --- c6 new h's: enchant/augmentation section 2 ---
	w.WriteInt(paperdollAugID(inv, model.PaperdollRHand)) // duplicate augmentation
	w.WriteShort(0)
	w.WriteShort(0)
	w.WriteShort(0)
	w.WriteShort(0)

	// --- PvP/Karma (first occurrence) ---
	w.WriteInt(pl.PvPFlag())
	w.WriteInt(pl.Karma())

	// --- Attack speed ---
	w.WriteInt(mAtkSpd)
	w.WriteInt(pAtkSpd)

	// --- PvP/Karma (duplicate — Java writes twice) ---
	w.WriteInt(pl.PvPFlag())
	w.WriteInt(pl.Karma())

	// --- Movement speeds ---
	w.WriteInt(runSpd)
	w.WriteInt(walkSpd)
	w.WriteInt(swimRunSpd)
	w.WriteInt(swimWalkSpd)
	w.WriteInt(flyRunSpd)
	w.WriteInt(flyWalkSpd)
	w.WriteInt(flyRunSpd)   // duplicate (Java writes twice)
	w.WriteInt(flyWalkSpd)  // duplicate (Java writes twice)
	w.WriteDouble(moveMultiplier)
	w.WriteDouble(pl.GetAttackSpeedMultiplier())

	// --- Collision ---
	w.WriteDouble(collisionRadius)
	w.WriteDouble(collisionHeight)

	// --- Appearance ---
	w.WriteInt(pl.HairStyle())
	w.WriteInt(pl.HairColor())
	w.WriteInt(pl.Face())

	// --- Title ---
	w.WriteString(title)

	// --- Clan (zeroed if cursed weapon equipped) ---
	if !isCursedWeapon {
		w.WriteInt(pl.ClanID())
		w.WriteInt(0) // clanCrestId — loaded from Clan table
		w.WriteInt(0) // allyId — loaded from Clan table
		w.WriteInt(0) // allyCrestId — loaded from Clan table
	} else {
		w.WriteInt(0)
		w.WriteInt(0)
		w.WriteInt(0)
		w.WriteInt(0)
	}

	// --- Relation (unknown in CharInfo, always 0 per Java) ---
	w.WriteInt(0)

	// --- Status bytes ---
	w.WriteByte(standing)  // standing (Java: !isSitting())
	w.WriteByte(runningByte) // running
	w.WriteByte(inCombat)  // isInCombat
	w.WriteByte(alikeDead) // isAlikeDead
	w.WriteByte(invisByte) // isInvisible
	w.WriteByte(byte(pl.MountType())) // mountType
	w.WriteByte(byte(pl.PrivateStoreType()))

	// --- Cubics — not implemented, send count 0 ---
	w.WriteShort(0)

	// --- Party match --- not implemented
	w.WriteByte(0)

	// --- Abnormal visual effects ---
	w.WriteInt(abnormalEffects)

	// --- Recommendations ---
	w.WriteByte(byte(pl.RecomLeft()))
	w.WriteShort(int16(pl.RecomHave()))

	// --- Current class ID ---
	w.WriteInt(pl.ClassID())

	// --- CP ---
	w.WriteInt(pl.MaxCP())
	w.WriteInt(pl.CurrentCP())

	// --- Enchant effect ---
	w.WriteByte(enchantEffect)

	// --- Team ---
	w.WriteByte(byte(pl.TeamID()))

	// --- Large clan crest — loaded from Clan table ---
	w.WriteInt(0)

	// --- Noble / Hero ---
	w.WriteByte(noble)
	w.WriteByte(hero)

	// --- Fishing ---
	w.WriteByte(fishingByte)
	w.WriteInt(pl.FishX())
	w.WriteInt(pl.FishY())
	w.WriteInt(pl.FishZ())

	// --- Name color ---
	w.WriteInt(pl.NameColor())

	// --- Heading ---
	w.WriteInt(int32(loc.Heading))

	// --- Pledge ---
	w.WriteInt(pl.PledgeClass())
	w.WriteInt(pl.PledgeType())

	// --- Title color ---
	w.WriteInt(pl.TitleColor())

	// --- Cursed weapon level ---
	cursedLevel := int32(0)
	if pl.IsCursedWeaponEquipped() {
		cursedLevel = 1
	}
	w.WriteInt(cursedLevel)

	return w.Bytes(), nil
}
