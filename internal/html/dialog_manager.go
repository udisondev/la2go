package html

import (
	"fmt"
	"strconv"
)

// DialogManager resolves NPC dialog HTML by NPC type and ID.
// Uses a resolution order to find the best matching template.
type DialogManager struct {
	cache *Cache
}

// NewDialogManager creates a new DialogManager backed by the given Cache.
func NewDialogManager(cache *Cache) *DialogManager {
	return &DialogManager{cache: cache}
}

// GetNpcDialog returns rendered HTML for the given NPC.
//
// Resolution order:
//  1. <npcType>/<npcID>.htm  (e.g. "merchant/30001.htm")
//  2. default/<npcID>.htm
//  3. npcdefault.htm
//
// Returns FallbackHTML if nothing found.
func (m *DialogManager) GetNpcDialog(npcType string, npcID int32, data DialogData) (string, error) {
	npcIDStr := strconv.FormatInt(int64(npcID), 10)

	// 1. Type-specific: <npcType>/<npcID>.htm
	if npcType != "" {
		path := npcType + "/" + npcIDStr + ".htm"
		if m.cache.Exists(path) {
			return m.cache.Execute(path, data)
		}
	}

	// 2. Default: default/<npcID>.htm
	defaultPath := "default/" + npcIDStr + ".htm"
	if m.cache.Exists(defaultPath) {
		return m.cache.Execute(defaultPath, data)
	}

	// 3. Global fallback: npcdefault.htm
	if m.cache.Exists("npcdefault.htm") {
		return m.cache.Execute("npcdefault.htm", data)
	}

	return m.FallbackHTML(data), nil
}

// GetDialogPage returns a specific dialog page for "Chat N" bypass.
//
// Resolution order:
//  1. <npcType>/<npcID>-<page>.htm
//  2. default/<npcID>-<page>.htm
//
// Returns error if page not found.
func (m *DialogManager) GetDialogPage(npcType string, npcID int32, page int, data DialogData) (string, error) {
	npcIDStr := strconv.FormatInt(int64(npcID), 10)
	pageStr := strconv.Itoa(page)

	// 1. Type-specific: <npcType>/<npcID>-<page>.htm
	if npcType != "" {
		path := npcType + "/" + npcIDStr + "-" + pageStr + ".htm"
		if m.cache.Exists(path) {
			return m.cache.Execute(path, data)
		}
	}

	// 2. Default: default/<npcID>-<page>.htm
	defaultPath := "default/" + npcIDStr + "-" + pageStr + ".htm"
	if m.cache.Exists(defaultPath) {
		return m.cache.Execute(defaultPath, data)
	}

	return "", fmt.Errorf("dialog page not found: npcType=%s npcID=%d page=%d", npcType, npcID, page)
}

// ExecuteLink renders an HTML file by direct relative path.
// Used by RequestLinkHtml (opcode 0x20) for NPC dialog link navigation.
func (m *DialogManager) ExecuteLink(link string, data DialogData) (string, error) {
	return m.cache.Execute(link, data)
}

// FallbackHTML returns a hardcoded fallback dialog when no template is found.
func (m *DialogManager) FallbackHTML(data DialogData) string {
	name, _ := data["npcname"].(string)
	if name == "" {
		name = "NPC"
	}
	return "<html><body>" + name + ":<br>I have nothing to say to you.<br></body></html>"
}
