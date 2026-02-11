// Unified code generator: parses L2J Mobius XML data files and generates Go literals.
//
// Usage:
//
//	go run ./cmd/gendata all                    # generate everything
//	go run ./cmd/gendata skills npcs items      # generate only specified categories
//	go run ./cmd/gendata --list                 # list available generators
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	javaDataDir = "../L2J_Mobius_CT_0_Interlude/dist/game/data"
	outputDir   = "internal/data"
)

type generator struct {
	name     string
	desc     string
	generate func(javaDir, outDir string) error
}

var generators []generator

func registerGenerator(name, desc string, fn func(javaDir, outDir string) error) {
	generators = append(generators, generator{name: name, desc: desc, generate: fn})
}

func init() {
	registerGenerator("skills", "Skill definitions (stats/skills/*.xml)", generateSkills)
	registerGenerator("skilltrees", "Skill trees (stats/players/skillTrees/**/*.xml)", generateSkillTrees)
	registerGenerator("npcs", "NPC templates (stats/npcs/*.xml)", generateNpcs)
	registerGenerator("items", "Item templates (stats/items/*.xml)", generateItems)
	registerGenerator("spawns", "Spawn definitions (spawns/**/*.xml)", generateSpawns)
	registerGenerator("zones", "Zone definitions (zones/*.xml)", generateZones)
	registerGenerator("buylists", "Buylist definitions (buylists/*.xml)", generateBuylists)
	registerGenerator("teleporters", "Teleporter definitions (teleporters/**/*.xml)", generateTeleporters)
	registerGenerator("multisell", "Multisell lists (multisell/*.xml)", generateMultisell)
	registerGenerator("doors", "Door definitions (stats/doors/Doors.xml)", generateDoors)
	registerGenerator("armorsets", "Armor set definitions (stats/armorsets/*.xml)", generateArmorsets)
	registerGenerator("recipes", "Recipe definitions (Recipes.xml)", generateRecipes)
	registerGenerator("augmentation", "Augmentation data (stats/augmentation/*.xml)", generateAugmentation)
	registerGenerator("playerconfig", "Player config (initialEquipment, shortcuts, karma/exp loss)", generatePlayerConfig)
	registerGenerator("categorydata", "Category data (CategoryData.xml)", generateCategoryData)
	registerGenerator("pets", "Pet data (stats/pets/*.xml)", generatePets)
	registerGenerator("fishing", "Fishing data (stats/fishing/*.xml)", generateFishing)
	registerGenerator("seeds", "Seed data (Seeds.xml)", generateSeeds)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	if args[0] == "--list" {
		printList()
		return
	}

	// Determine which generators to run
	var toRun []generator
	if args[0] == "all" {
		toRun = generators
	} else {
		genMap := make(map[string]generator, len(generators))
		for _, g := range generators {
			genMap[g.name] = g
		}
		for _, name := range args {
			g, ok := genMap[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown generator: %s\n", name)
				printList()
				os.Exit(1)
			}
			toRun = append(toRun, g)
		}
	}

	// Run generators sequentially
	totalStart := time.Now()
	for _, g := range toRun {
		start := time.Now()
		fmt.Printf("[gendata] running %s...\n", g.name)
		if err := g.generate(javaDataDir, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "[gendata] FAILED %s: %v\n", g.name, err)
			os.Exit(1)
		}
		fmt.Printf("[gendata] %s done (%s)\n", g.name, time.Since(start).Round(time.Millisecond))
	}
	fmt.Printf("[gendata] all done (%s)\n", time.Since(totalStart).Round(time.Millisecond))
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/gendata <all | name1 name2 ...>")
	fmt.Fprintln(os.Stderr, "       go run ./cmd/gendata --list")
}

func printList() {
	names := make([]string, 0, len(generators))
	maxLen := 0
	for _, g := range generators {
		names = append(names, g.name)
		if len(g.name) > maxLen {
			maxLen = len(g.name)
		}
	}
	sort.Strings(names)

	genMap := make(map[string]generator, len(generators))
	for _, g := range generators {
		genMap[g.name] = g
	}

	fmt.Println("Available generators:")
	for _, name := range names {
		g := genMap[name]
		padding := strings.Repeat(" ", maxLen-len(name)+2)
		fmt.Printf("  %s%s%s\n", name, padding, g.desc)
	}
}
