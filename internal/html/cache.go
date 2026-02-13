package html

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

const maxHTMLFileSize = 8192

// DialogData is a flexible key→value map for template substitution.
// Keys match L2J variable names exactly (e.g. "objectId", "npcname", "castleName").
// Handler fills in the keys relevant to the current context.
type DialogData map[string]any

// Cache loads .htm files from a directory and stores compiled text/template objects.
// Files must already be converted from L2J %var% to Go {{index . "var"}} syntax
// (use cmd/htmlconvert for batch conversion).
type Cache struct {
	htmlDir   string
	templates map[string]*template.Template
	mu        sync.RWMutex
	lazy      bool
}

// NewCache creates a new HTML template cache.
// If lazy is false, all .htm files are loaded from htmlDir at creation time.
// If lazy is true, files are loaded on first access (cache miss).
func NewCache(htmlDir string, lazy bool) (*Cache, error) {
	c := &Cache{
		htmlDir:   htmlDir,
		templates: make(map[string]*template.Template),
		lazy:      lazy,
	}

	if !lazy {
		if err := c.preload(); err != nil {
			return nil, fmt.Errorf("preloading HTML templates: %w", err)
		}
	}

	return c, nil
}

// Get returns a compiled template by relative path (e.g. "merchant/30001.htm").
func (c *Cache) Get(path string) (*template.Template, error) {
	if strings.Contains(path, "..") {
		return nil, fmt.Errorf("path traversal denied: %s", path)
	}

	c.mu.RLock()
	tmpl, ok := c.templates[path]
	c.mu.RUnlock()
	if ok {
		return tmpl, nil
	}

	if !c.lazy {
		return nil, fmt.Errorf("template not found: %s", path)
	}

	return c.loadAndCache(path)
}

// Execute renders a template with the given data and returns the HTML string.
func (c *Cache) Execute(path string, data DialogData) (string, error) {
	tmpl, err := c.Get(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any(data)); err != nil {
		return "", fmt.Errorf("executing template %s: %w", path, err)
	}

	return buf.String(), nil
}

// Exists returns true if the template is cached or the file exists on disk.
func (c *Cache) Exists(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}

	c.mu.RLock()
	_, ok := c.templates[path]
	c.mu.RUnlock()
	if ok {
		return true
	}

	if c.lazy {
		fullPath := filepath.Join(c.htmlDir, path)
		_, err := os.Stat(fullPath)
		return err == nil
	}

	return false
}

// preload walks htmlDir and loads all .htm files into cache.
func (c *Cache) preload() error {
	info, err := os.Stat(c.htmlDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("HTML directory does not exist, skipping preload", "dir", c.htmlDir)
			return nil
		}
		return fmt.Errorf("stat html dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("html dir is not a directory: %s", c.htmlDir)
	}

	count := 0
	err = filepath.WalkDir(c.htmlDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".htm") {
			return nil
		}

		relPath, err := filepath.Rel(c.htmlDir, path)
		if err != nil {
			return fmt.Errorf("computing relative path for %s: %w", path, err)
		}

		if _, err := c.loadFile(relPath); err != nil {
			slog.Warn("failed to load HTML template", "path", relPath, "error", err)
			return nil // skip broken files, don't fail entire preload
		}

		count++
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking html dir: %w", err)
	}

	slog.Info("HTML templates preloaded", "count", count, "dir", c.htmlDir)
	return nil
}

// loadAndCache loads a file from disk, compiles it, and stores in cache.
func (c *Cache) loadAndCache(path string) (*template.Template, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock.
	if tmpl, ok := c.templates[path]; ok {
		return tmpl, nil
	}

	return c.loadFile(path)
}

// loadFile reads and compiles the template, stores it in cache.
// Caller must hold c.mu write lock (or be called during init).
func (c *Cache) loadFile(path string) (*template.Template, error) {
	fullPath := filepath.Join(c.htmlDir, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Size() > maxHTMLFileSize {
		return nil, fmt.Errorf("file too large (%d bytes, max %d): %s", info.Size(), maxHTMLFileSize, path)
	}

	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	// option "missingkey=zero" makes missing variables render as "" instead of error.
	tmpl, err := template.New(path).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", path, err)
	}

	c.templates[path] = tmpl
	return tmpl, nil
}
