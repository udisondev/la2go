package data

// SeedTemplate — exported view of a seed for use outside the data package.
// Phase 27: Manor System.
type SeedTemplate struct {
	SeedID        int32
	CropID        int32
	CastleID      int32
	MatureID      int32
	Reward1       int32
	Reward2       int32
	Level         int32
	IsAlternative bool
	LimitSeeds    int32
	LimitCrops    int32
}

// defaultLimitSeeds is used when seedDef.limitSeeds is zero (generated data not yet updated).
const defaultLimitSeeds int32 = 5000

// defaultLimitCrops is used when seedDef.limitCrops is zero (generated data not yet updated).
const defaultLimitCrops int32 = 5000

func seedDefToTemplate(def *seedDef) *SeedTemplate {
	ls := def.limitSeeds
	if ls <= 0 {
		ls = defaultLimitSeeds
	}
	lc := def.limitCrops
	if lc <= 0 {
		lc = defaultLimitCrops
	}
	return &SeedTemplate{
		SeedID:        def.seedID,
		CropID:        def.cropID,
		CastleID:      def.castleID,
		MatureID:      def.matureID,
		Reward1:       def.reward1,
		Reward2:       def.reward2,
		Level:         def.level,
		IsAlternative: def.isAlternative,
		LimitSeeds:    ls,
		LimitCrops:    lc,
	}
}

// GetSeedTemplate returns an exported seed by seedID.
// Returns nil if not found.
func GetSeedTemplate(seedID int32) *SeedTemplate {
	def := SeedTable[seedID]
	if def == nil {
		return nil
	}
	return seedDefToTemplate(def)
}

// GetSeedByCropID returns the first seed template for the given cropID.
// Returns nil if not found.
func GetSeedByCropID(cropID int32) *SeedTemplate {
	for _, def := range SeedTable {
		if def.cropID == cropID {
			return seedDefToTemplate(def)
		}
	}
	return nil
}

// GetSeedByCropAndCastle returns the seed template for the given cropID and castleID.
// Returns nil if not found.
func GetSeedByCropAndCastle(cropID, castleID int32) *SeedTemplate {
	for _, def := range SeedTable {
		if def.cropID == cropID && def.castleID == castleID {
			return seedDefToTemplate(def)
		}
	}
	return nil
}

// GetSeedsByCastle returns all seed templates for the given castle.
func GetSeedsByCastle(castleID int32) []*SeedTemplate {
	var result []*SeedTemplate
	for _, def := range SeedTable {
		if def.castleID == castleID {
			result = append(result, seedDefToTemplate(def))
		}
	}
	return result
}

// GetAllCropIDs returns a set of all known crop IDs.
func GetAllCropIDs() map[int32]struct{} {
	result := make(map[int32]struct{}, len(SeedTable))
	for _, def := range SeedTable {
		result[def.cropID] = struct{}{}
	}
	return result
}

// GetAllSeedIDs returns a set of all known seed IDs.
func GetAllSeedIDs() map[int32]struct{} {
	result := make(map[int32]struct{}, len(SeedTable))
	for _, def := range SeedTable {
		result[def.seedID] = struct{}{}
	}
	return result
}

// SeedReferencePrice returns the reference price for a seed item.
// Uses the item's base price from ItemTable.
func SeedReferencePrice(seedID int32) int64 {
	def := ItemTable[seedID]
	if def == nil {
		return 1
	}
	if def.price <= 0 {
		return 1
	}
	return def.price
}

// CropReferencePrice returns the reference price for a crop item.
// Uses the item's base price from ItemTable.
func CropReferencePrice(cropID int32) int64 {
	def := ItemTable[cropID]
	if def == nil {
		return 1
	}
	if def.price <= 0 {
		return 1
	}
	return def.price
}

// SeedMaxPrice returns the maximum price for a seed (reference * 10).
func SeedMaxPrice(seedID int32) int64 {
	return SeedReferencePrice(seedID) * 10
}

// SeedMinPrice returns the minimum price for a seed (reference * 0.6).
func SeedMinPrice(seedID int32) int64 {
	return SeedReferencePrice(seedID) * 6 / 10
}

// CropMaxPrice returns the maximum price for a crop (reference * 10).
func CropMaxPrice(cropID int32) int64 {
	return CropReferencePrice(cropID) * 10
}

// CropMinPrice returns the minimum price for a crop (reference * 0.6).
func CropMinPrice(cropID int32) int64 {
	return CropReferencePrice(cropID) * 6 / 10
}
