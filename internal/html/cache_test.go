package html

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("creating dirs for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}
	return dir
}

func TestCache_LoadAndExecute(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"test.htm": `<html><body>Hello, {{index . "name"}}!</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	result, err := cache.Execute("test.htm", DialogData{"name": "TestPlayer"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	expected := `<html><body>Hello, TestPlayer!</body></html>`
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestCache_AllVarSubstitution(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"vars.htm": `<html><body>{{index . "npcname"}}:<br>Hello, {{index . "name"}}! <a action="bypass -h npc_{{index . "objectId"}}_Shop">Shop</a></body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	data := DialogData{
		"objectId": "12345",
		"npcname":  "Grocer",
		"name":     "Hero",
	}
	result, err := cache.Execute("vars.htm", data)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(result, "Grocer:") {
		t.Errorf("expected npcname substitution, got: %s", result)
	}
	if !strings.Contains(result, "Hello, Hero!") {
		t.Errorf("expected name substitution, got: %s", result)
	}
	if !strings.Contains(result, "npc_12345_Shop") {
		t.Errorf("expected objectId substitution, got: %s", result)
	}
}

func TestCache_MissingKey(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"missing.htm": `<html><body>{{index . "castleName"}} castle</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	// Missing key should render as empty string (missingkey=zero).
	result, err := cache.Execute("missing.htm", DialogData{"objectId": "1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(result, " castle") {
		t.Errorf("expected empty substitution for missing key, got: %s", result)
	}
}

func TestCache_LazyLoad(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"lazy.htm": `<html><body>Lazy: {{index . "npcname"}}</body></html>`,
	})

	cache, err := NewCache(dir, true)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	// Template not yet loaded.
	cache.mu.RLock()
	_, loaded := cache.templates["lazy.htm"]
	cache.mu.RUnlock()
	if loaded {
		t.Fatal("expected template not to be loaded yet in lazy mode")
	}

	// Access triggers load.
	result, err := cache.Execute("lazy.htm", DialogData{"npcname": "Guard"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(result, "Lazy: Guard") {
		t.Errorf("unexpected result: %s", result)
	}

	// Now cached.
	cache.mu.RLock()
	_, loaded = cache.templates["lazy.htm"]
	cache.mu.RUnlock()
	if !loaded {
		t.Fatal("expected template to be cached after first access")
	}
}

func TestCache_PathTraversal(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"safe.htm": `ok`,
	})

	cache, err := NewCache(dir, true)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	_, err = cache.Get("../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

func TestCache_FileTooLarge(t *testing.T) {
	content := strings.Repeat("x", maxHTMLFileSize+1)
	dir := setupTestDir(t, map[string]string{
		"big.htm": content,
	})

	_, err := NewCache(dir, false)
	// Preload skips broken files, so no error from NewCache.
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	// Lazy load should fail.
	cache, err := NewCache(dir, true)
	if err != nil {
		t.Fatalf("NewCache lazy: %v", err)
	}

	_, err = cache.Get("big.htm")
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' error, got: %v", err)
	}
}

func TestCache_Exists(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"exists.htm": `<html></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	if !cache.Exists("exists.htm") {
		t.Error("expected Exists to return true for loaded template")
	}
	if cache.Exists("notexists.htm") {
		t.Error("expected Exists to return false for missing template")
	}
	if cache.Exists("../escape.htm") {
		t.Error("expected Exists to return false for path traversal")
	}
}

func TestCache_PreloadSubdirs(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"merchant/30001.htm": `<html><body>{{index . "npcname"}}</body></html>`,
		"guard/30002.htm":    `<html><body>Guard {{index . "npcname"}}</body></html>`,
	})

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	if !cache.Exists("merchant/30001.htm") {
		t.Error("expected merchant/30001.htm to be loaded")
	}
	if !cache.Exists("guard/30002.htm") {
		t.Error("expected guard/30002.htm to be loaded")
	}
}

func TestCache_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	cache, err := NewCache(dir, false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	if cache.Exists("anything.htm") {
		t.Error("expected empty cache")
	}
}

func TestCache_NonExistentDir(t *testing.T) {
	cache, err := NewCache("/nonexistent/path", false)
	if err != nil {
		t.Fatalf("NewCache should not error on missing dir: %v", err)
	}

	if cache.Exists("anything.htm") {
		t.Error("expected empty cache for nonexistent dir")
	}
}

func TestCache_PreloadRealData(t *testing.T) {
	dataDir := "../../data/html"

	cache, err := NewCache(dataDir, false)
	if err != nil {
		t.Fatalf("preload failed: %v", err)
	}

	knownFiles := []string{
		"npcdefault.htm",
		"noquest.htm",
		"merchant/30001.htm",
	}
	for _, f := range knownFiles {
		if !cache.Exists(f) {
			t.Errorf("expected %s to be loaded", f)
		}
	}

	cache.mu.RLock()
	count := len(cache.templates)
	cache.mu.RUnlock()

	t.Logf("loaded %d templates from %s", count, dataDir)
	if count < 100 {
		t.Errorf("expected at least 100 templates, got %d", count)
	}
}

func TestCache_ExecuteRealMerchant(t *testing.T) {
	cache, err := NewCache("../../data/html", false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	data := DialogData{
		"objectId": "99999",
		"npcname":  "Grocer",
		"name":     "TestPlayer",
	}

	result, err := cache.Execute("merchant/30001.htm", data)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(result, "npc_99999_") {
		t.Errorf("expected objectId substitution, got: %s", result)
	}
	if !strings.Contains(result, "Trader Lector") {
		t.Errorf("expected NPC text preserved, got: %s", result)
	}

	t.Logf("Result:\n%s", result)
}

func TestCache_ExecuteRealTeleporter(t *testing.T) {
	cache, err := NewCache("../../data/html", false)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}

	data := DialogData{
		"objectId": "12345",
		"npcname":  "Gatekeeper",
		"name":     "Hero",
	}

	result, err := cache.Execute("teleporter/30006.htm", data)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(result, "npc_12345_") {
		t.Errorf("expected objectId substitution, got: %s", result)
	}

	t.Logf("Result:\n%s", result)
}

func BenchmarkCache_Execute(b *testing.B) {
	dir := b.TempDir()
	content := `<html><body>{{index . "npcname"}}:<br>Hello, {{index . "name"}}!<br><a action="bypass -h npc_{{index . "objectId"}}_Shop">Shop</a></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "bench.htm"), []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}

	cache, err := NewCache(dir, false)
	if err != nil {
		b.Fatal(err)
	}

	data := DialogData{
		"objectId": "12345",
		"npcname":  "Merchant",
		"name":     "TestPlayer",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, err := cache.Execute("bench.htm", data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
