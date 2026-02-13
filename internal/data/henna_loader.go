package data

import "log/slog"

// HennaTable — глобальный registry всех henna templates.
// map[dyeID]*hennaDef
var HennaTable map[int32]*hennaDef

// hennaByClass — precomputed map: classID → []*hennaDef.
// O(1) lookup for class-specific henna lists.
var hennaByClass map[int32][]*hennaDef

// GetHennaDef возвращает hennaDef по dye ID.
// Returns nil если хенна не найдена.
func GetHennaDef(dyeID int32) *hennaDef {
	if HennaTable == nil {
		return nil
	}
	return HennaTable[dyeID]
}

// GetHennaListForClass возвращает все хенны, доступные для classID.
// Returns nil если нет доступных хенн.
func GetHennaListForClass(classID int32) []*hennaDef {
	if hennaByClass == nil {
		return nil
	}
	return hennaByClass[classID]
}

// LoadHennaTemplates строит HennaTable из Go-литералов (hennaDefs).
func LoadHennaTemplates() error {
	HennaTable = make(map[int32]*hennaDef, len(hennaDefs))
	hennaByClass = make(map[int32][]*hennaDef)

	for i := range hennaDefs {
		def := &hennaDefs[i]
		HennaTable[def.dyeID] = def

		for _, classID := range def.classIDs {
			hennaByClass[classID] = append(hennaByClass[classID], def)
		}
	}

	slog.Info("loaded henna templates", "count", len(HennaTable), "classes", len(hennaByClass))
	return nil
}
