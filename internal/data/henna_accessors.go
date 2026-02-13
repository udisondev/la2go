package data

// HennaInfo — exported view of a henna template for use outside the data package.
// Phase 13: Henna System.
type HennaInfo struct {
	DyeID       int32
	DyeItemID   int32
	StatSTR     int32
	StatCON     int32
	StatDEX     int32
	StatINT     int32
	StatMEN     int32
	StatWIT     int32
	WearCount   int32
	WearFee     int64
	CancelCount int32
	CancelFee   int64
}

// IsAllowedClass checks if the henna is allowed for classID.
func (h *HennaInfo) IsAllowedClass(classID int32) bool {
	def := GetHennaDef(h.DyeID)
	if def == nil {
		return false
	}
	return def.IsAllowedClass(classID)
}

// GetHennaInfo returns an exported HennaInfo for a dye ID.
// Returns nil if not found.
func GetHennaInfo(dyeID int32) *HennaInfo {
	def := GetHennaDef(dyeID)
	if def == nil {
		return nil
	}
	return hennaDefToInfo(def)
}

// GetHennaInfoListForClass returns all hennas available for a class as exported structs.
func GetHennaInfoListForClass(classID int32) []*HennaInfo {
	defs := GetHennaListForClass(classID)
	if len(defs) == 0 {
		return nil
	}

	result := make([]*HennaInfo, len(defs))
	for i, def := range defs {
		result[i] = hennaDefToInfo(def)
	}
	return result
}

func hennaDefToInfo(def *hennaDef) *HennaInfo {
	return &HennaInfo{
		DyeID:       def.dyeID,
		DyeItemID:   def.dyeItemID,
		StatSTR:     def.statSTR,
		StatCON:     def.statCON,
		StatDEX:     def.statDEX,
		StatINT:     def.statINT,
		StatMEN:     def.statMEN,
		StatWIT:     def.statWIT,
		WearCount:   def.wearCount,
		WearFee:     def.wearFee,
		CancelCount: def.cancelCount,
		CancelFee:   def.cancelFee,
	}
}
