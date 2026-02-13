package data

import "log/slog"

// ItemTable — глобальный registry всех item templates.
// map[itemID]*itemDef
var ItemTable map[int32]*itemDef

// GetItemDef возвращает itemDef по item ID.
func GetItemDef(itemID int32) *itemDef {
	if ItemTable == nil {
		return nil
	}
	return ItemTable[itemID]
}

// LoadItemTemplates строит ItemTable из Go-литералов (itemDefs).
func LoadItemTemplates() error {
	ItemTable = make(map[int32]*itemDef, len(itemDefs))

	for i := range itemDefs {
		ItemTable[itemDefs[i].id] = &itemDefs[i]
	}

	slog.Info("loaded item templates", "count", len(ItemTable))
	return nil
}

// ItemDef accessor methods
func (d *itemDef) ID() int32           { return d.id }
func (d *itemDef) Name() string        { return d.name }
func (d *itemDef) Type() string        { return d.itemType }
func (d *itemDef) WeaponType() string  { return d.weaponType }
func (d *itemDef) ArmorType() string   { return d.armorType }
func (d *itemDef) BodyPart() string    { return d.bodyPart }
func (d *itemDef) Material() string    { return d.material }
func (d *itemDef) Weight() int32       { return d.weight }
func (d *itemDef) Price() int64        { return d.price }
func (d *itemDef) IsStackable() bool   { return d.stackable }
func (d *itemDef) IsTradeable() bool   { return d.tradeable }
func (d *itemDef) IsDroppable() bool   { return d.droppable }
func (d *itemDef) IsSellable() bool    { return d.sellable }
func (d *itemDef) PAtk() int32         { return d.pAtk }
func (d *itemDef) MAtk() int32         { return d.mAtk }
func (d *itemDef) PDef() int32         { return d.pDef }
func (d *itemDef) MDef() int32         { return d.mDef }
func (d *itemDef) PAtkSpd() int32      { return d.pAtkSpd }
func (d *itemDef) CritRate() int32     { return d.critRate }
func (d *itemDef) AttackRange() int32  { return d.attackRange }
func (d *itemDef) RandomDamage() int32 { return d.randomDamage }
func (d *itemDef) SoulShots() int32    { return d.soulshots }
func (d *itemDef) SpiritShots() int32  { return d.spiritshots }
func (d *itemDef) IsMagicWeapon() bool { return d.magicWeapon }
func (d *itemDef) EtcItemType() string  { return d.etcItemType }
func (d *itemDef) CrystalType() string  { return d.crystalType }
func (d *itemDef) Handler() string      { return d.handler }
func (d *itemDef) IsQuestItem() bool    { return d.questItem }
func (d *itemDef) DefaultAction() string { return d.defaultAction }
func (d *itemDef) IsEnchantable() bool  { return d.enchantable }
func (d *itemDef) ItemSkillID() int32   { return d.itemSkillID }
func (d *itemDef) ItemSkillLevel() int32 { return d.itemSkillLevel }
func (d *itemDef) ReuseDelay() int32    { return d.reuseDelay }
func (d *itemDef) IsOlyRestricted() bool { return d.olyRestricted }
func (d *itemDef) IsForNpc() bool       { return d.forNpc }

// GetWeaponCritRate returns weapon critical rate for given itemID, or 0 if not found.
func GetWeaponCritRate(itemID int32) int32 {
	def := GetItemDef(itemID)
	if def == nil {
		return 0
	}
	return def.critRate
}

// GetWeaponRandomDamage returns weapon random damage for given itemID, or 0 if not found.
func GetWeaponRandomDamage(itemID int32) int32 {
	def := GetItemDef(itemID)
	if def == nil {
		return 0
	}
	return def.randomDamage
}
