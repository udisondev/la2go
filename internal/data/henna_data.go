package data

import "slices"

// hennaDef — определение хенны (generated from hennaList.xml).
type hennaDef struct {
	dyeID       int32
	dyeName     string
	dyeItemID   int32
	statSTR     int32
	statCON     int32
	statDEX     int32
	statINT     int32
	statMEN     int32
	statWIT     int32
	wearCount   int32
	wearFee     int64
	cancelCount int32
	cancelFee   int64
	classIDs    []int32
}

// HennaDef accessor methods — provide read access to hennaDef fields.

func (d *hennaDef) DyeID() int32       { return d.dyeID }
func (d *hennaDef) DyeName() string     { return d.dyeName }
func (d *hennaDef) DyeItemID() int32    { return d.dyeItemID }
func (d *hennaDef) StatSTR() int32      { return d.statSTR }
func (d *hennaDef) StatCON() int32      { return d.statCON }
func (d *hennaDef) StatDEX() int32      { return d.statDEX }
func (d *hennaDef) StatINT() int32      { return d.statINT }
func (d *hennaDef) StatMEN() int32      { return d.statMEN }
func (d *hennaDef) StatWIT() int32      { return d.statWIT }
func (d *hennaDef) WearCount() int32    { return d.wearCount }
func (d *hennaDef) WearFee() int64      { return d.wearFee }
func (d *hennaDef) CancelCount() int32  { return d.cancelCount }
func (d *hennaDef) CancelFee() int64    { return d.cancelFee }
func (d *hennaDef) ClassIDs() []int32   { return d.classIDs }

// IsAllowedClass returns true if classID is in the allowed list.
func (d *hennaDef) IsAllowedClass(classID int32) bool {
	return slices.Contains(d.classIDs, classID)
}
