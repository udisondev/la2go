package html

import (
	"strings"
	"testing"
)

func TestDialogManager_ResolutionOrder(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"merchant/30001.htm": `<html><body>Merchant dialog for {{index . "npcname"}}</body></html>`,
		"default/30002.htm":  `<html><body>Default dialog for {{index . "npcname"}}</body></html>`,
		"npcdefault.htm":     `<html><body>{{index . "npcname"}}: Generic fallback</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	mgr := NewDialogManager(cache)
	data := DialogData{"npcname": "TestNPC", "objectId": "100"}

	// 1. Type-specific file exists → use it.
	result, err := mgr.GetNpcDialog("merchant", 30001, data)
	if err != nil {
		t.Fatalf("GetNpcDialog merchant: %v", err)
	}
	if !strings.Contains(result, "Merchant dialog") {
		t.Errorf("expected type-specific template, got: %s", result)
	}

	// 2. No type-specific, but default/<id>.htm exists → use it.
	result, err = mgr.GetNpcDialog("guard", 30002, data)
	if err != nil {
		t.Fatalf("GetNpcDialog default: %v", err)
	}
	if !strings.Contains(result, "Default dialog") {
		t.Errorf("expected default template, got: %s", result)
	}

	// 3. No type-specific, no default → npcdefault.htm.
	result, err = mgr.GetNpcDialog("folk", 99999, data)
	if err != nil {
		t.Fatalf("GetNpcDialog fallback: %v", err)
	}
	if !strings.Contains(result, "Generic fallback") {
		t.Errorf("expected npcdefault.htm, got: %s", result)
	}
}

func TestDialogManager_Fallback(t *testing.T) {
	dir := setupTestDir(t, map[string]string{})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	mgr := NewDialogManager(cache)
	data := DialogData{"npcname": "Guard"}

	// No templates at all → hardcoded fallback.
	result, err := mgr.GetNpcDialog("guard", 30001, data)
	if err != nil {
		t.Fatalf("GetNpcDialog: %v", err)
	}
	if !strings.Contains(result, "Guard:") {
		t.Errorf("expected NPC name in fallback, got: %s", result)
	}
	if !strings.Contains(result, "nothing to say") {
		t.Errorf("expected fallback text, got: %s", result)
	}
}

func TestDialogManager_FallbackEmptyName(t *testing.T) {
	mgr := NewDialogManager(nil)
	data := DialogData{}
	result := mgr.FallbackHTML(data)
	if !strings.Contains(result, "NPC:") {
		t.Errorf("expected default 'NPC' name, got: %s", result)
	}
}

func TestDialogManager_DialogPage(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"merchant/30001-1.htm": `<html><body>{{index . "npcname"}}: Page 1 content</body></html>`,
		"default/30002-2.htm":  `<html><body>{{index . "npcname"}}: Default page 2</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	mgr := NewDialogManager(cache)
	data := DialogData{"npcname": "TestNPC", "objectId": "100"}

	// Type-specific page.
	result, err := mgr.GetDialogPage("merchant", 30001, 1, data)
	if err != nil {
		t.Fatalf("GetDialogPage: %v", err)
	}
	if !strings.Contains(result, "Page 1 content") {
		t.Errorf("expected page 1 content, got: %s", result)
	}

	// Default page.
	result, err = mgr.GetDialogPage("guard", 30002, 2, data)
	if err != nil {
		t.Fatalf("GetDialogPage default: %v", err)
	}
	if !strings.Contains(result, "Default page 2") {
		t.Errorf("expected default page 2, got: %s", result)
	}

	// Missing page → error.
	_, err = mgr.GetDialogPage("merchant", 30001, 99, data)
	if err == nil {
		t.Fatal("expected error for missing page")
	}
}

func TestDialogManager_EmptyNpcType(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"default/30001.htm": `<html><body>Default: {{index . "npcname"}}</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	mgr := NewDialogManager(cache)
	data := DialogData{"npcname": "TestNPC"}

	// Empty npcType should skip type-specific and go to default.
	result, err := mgr.GetNpcDialog("", 30001, data)
	if err != nil {
		t.Fatalf("GetNpcDialog empty type: %v", err)
	}
	if !strings.Contains(result, "Default: TestNPC") {
		t.Errorf("expected default template, got: %s", result)
	}
}

func TestDialogManager_VariableSubstitution(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"folk/30001.htm": `<html><body>{{index . "npcname"}}: Hello {{index . "name"}}. <a action="bypass -h npc_{{index . "objectId"}}_Quest">Quest</a></body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	mgr := NewDialogManager(cache)
	data := DialogData{
		"objectId": "54321",
		"npcname":  "Elder",
		"name":     "Hero",
	}

	result, err := mgr.GetNpcDialog("folk", 30001, data)
	if err != nil {
		t.Fatalf("GetNpcDialog: %v", err)
	}

	if !strings.Contains(result, "Elder:") {
		t.Errorf("missing npcname, got: %s", result)
	}
	if !strings.Contains(result, "Hello Hero") {
		t.Errorf("missing name, got: %s", result)
	}
	if !strings.Contains(result, "npc_54321_Quest") {
		t.Errorf("missing objectId in bypass, got: %s", result)
	}
}
