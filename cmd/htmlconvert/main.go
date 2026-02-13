// htmlconvert converts L2J HTML dialog files from %var% format to Go template syntax.
//
// Usage:
//
//	go run ./cmd/htmlconvert -dir data/html
//	go run ./cmd/htmlconvert -dir data/html -dry-run
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// l2jVarPattern matches L2J %variable% placeholders.
// Captures: alphanumeric, dots, underscores between % signs.
// Avoids matching HTML width=100% by requiring at least 2 chars inside.
var l2jVarPattern = regexp.MustCompile(`%([a-zA-Z][a-zA-Z0-9_.]+)%`)

func main() {
	dir := flag.String("dir", "data/html", "directory with .htm files")
	dryRun := flag.Bool("dry-run", false, "show changes without writing")
	flag.Parse()

	stats, err := convert(*dir, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("files scanned:   %d\n", stats.scanned)
	fmt.Printf("files converted: %d\n", stats.converted)
	fmt.Printf("vars replaced:   %d\n", stats.varsReplaced)
	fmt.Printf("unique vars:     %d\n", len(stats.uniqueVars))
	if *dryRun {
		fmt.Println("(dry-run, no files modified)")
	}
}

type convertStats struct {
	scanned      int
	converted    int
	varsReplaced int
	uniqueVars   map[string]int
}

func convert(dir string, dryRun bool) (convertStats, error) {
	stats := convertStats{uniqueVars: make(map[string]int)}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".htm") {
			return nil
		}

		stats.scanned++

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		content := string(raw)
		converted, count, vars := convertContent(content)

		if count == 0 {
			return nil
		}

		stats.converted++
		stats.varsReplaced += count
		for v, c := range vars {
			stats.uniqueVars[v] += c
		}

		if dryRun {
			rel, _ := filepath.Rel(dir, path)
			fmt.Printf("  %s: %d replacements\n", rel, count)
			return nil
		}

		if err := os.WriteFile(path, []byte(converted), d.Type().Perm()); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		return nil
	})

	return stats, err
}

func convertContent(content string) (string, int, map[string]int) {
	vars := make(map[string]int)
	count := 0

	result := l2jVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name from %varName%.
		varName := match[1 : len(match)-1]
		vars[varName]++
		count++
		return `{{index . "` + varName + `"}}`
	})

	return result, count, vars
}
