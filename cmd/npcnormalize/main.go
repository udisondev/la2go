// Command npcnormalize converts CamelCase NPC type names to snake_case
// in Go generated source files (npc_data_generated.go, skill_data_generated.go).
//
// This ensures NpcType values match HTML directory names (e.g. "clan_hall_doorman/").
//
// Usage:
//
//	go run ./cmd/npcnormalize
package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// camelToSnake converts CamelCase to snake_case.
//
//	"ClanHallDoorman" → "clan_hall_doorman"
//	"Folk" → "folk"
//	"BabyPet" → "baby_pet"
func camelToSnake(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)

	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}

	return b.String()
}

// npcTypeMapping defines the CamelCase → snake_case conversion for all 46 NPC types.
// VillageMaster* subtypes all map to "village_master".
var npcTypeMapping = map[string]string{
	"Adventurer":           "adventurer",
	"Artefact":             "artefact",
	"Auctioneer":           "auctioneer",
	"BabyPet":              "baby_pet",
	"BroadcastingTower":    "broadcasting_tower",
	"CastleDoorman":        "castle_doorman",
	"Chest":                "chest",
	"ClanHallDoorman":      "clan_hall_doorman",
	"ClanHallManager":      "clan_hall_manager",
	"ControlTower":         "control_tower",
	"DawnPriest":           "dawn_priest",
	"Defender":             "defender",
	"Doorman":              "doorman",
	"DungeonGatekeeper":    "dungeon_gatekeeper",
	"DuskPriest":           "dusk_priest",
	"EffectPoint":          "effect_point",
	"FeedableBeast":        "feedable_beast",
	"FestivalGuide":        "festival_guide",
	"FestivalMonster":      "festival_monster",
	"Fisherman":            "fisherman",
	"FlameTower":           "flame_tower",
	"FlyTerrainObject":     "fly_terrain_object",
	"Folk":                 "folk",
	"FriendlyMob":          "friendly_mob",
	"GrandBoss":            "grand_boss",
	"Guard":                "guard",
	"Merchant":             "merchant",
	"Monster":              "monster",
	"OlympiadManager":      "olympiad_manager",
	"Pet":                  "pet",
	"PetManager":           "pet_manager",
	"RaceManager":          "race_manager",
	"RaidBoss":             "raid_boss",
	"RiftInvader":          "rift_invader",
	"Servitor":             "servitor",
	"SignsPriest":          "signs_priest",
	"TamedBeast":           "tamed_beast",
	"Teleporter":           "teleporter",
	"Trainer":              "trainer",
	"VillageMasterDElf":    "village_master",
	"VillageMasterDwarf":   "village_master",
	"VillageMasterFighter": "village_master",
	"VillageMasterMystic":  "village_master",
	"VillageMasterOrc":     "village_master",
	"VillageMasterPriest":  "village_master",
	"Warehouse":            "warehouse",
}

func init() {
	// Verify all mappings match camelToSnake (except VillageMaster* overrides).
	for k, v := range npcTypeMapping {
		if strings.HasPrefix(k, "VillageMaster") {
			continue
		}
		if got := camelToSnake(k); got != v {
			panic(fmt.Sprintf("mapping mismatch for %q: camelToSnake=%q, mapping=%q", k, got, v))
		}
	}
}

func main() {
	files := []string{
		"internal/data/npc_data_generated.go",
		"internal/data/skill_data_generated.go",
	}

	totalReplacements := 0
	for _, path := range files {
		n, err := processFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", path, err)
			os.Exit(1)
		}
		totalReplacements += n
		fmt.Printf("%s: %d replacements\n", path, n)
	}

	fmt.Printf("\nTotal: %d replacements across %d files\n", totalReplacements, len(files))

	// Print mapping summary.
	fmt.Println("\nMapping applied:")
	keys := make([]string, 0, len(npcTypeMapping))
	for k := range npcTypeMapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-25s → %s\n", k, npcTypeMapping[k])
	}
}

// npcTypeRe matches npcType field assignments in generated Go code.
// Handles both:
//   - npcType:  "SomeType",        (struct literal)
//   - "npcType": "SomeType"        (map literal, skill_data)
var npcTypeRe = regexp.MustCompile(`("npcType":\s*"|npcType:\s+")([A-Z][a-zA-Z]+)"`)

func processFile(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(raw)
	count := 0

	result := npcTypeRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := npcTypeRe.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		prefix := sub[1]  // everything before the value
		oldType := sub[2] // the CamelCase type name

		newType, ok := npcTypeMapping[oldType]
		if !ok {
			fmt.Fprintf(os.Stderr, "WARNING: unknown NpcType %q in %s\n", oldType, path)
			return match
		}

		if oldType != newType {
			count++
		}
		return prefix + newType + `"`
	})

	if count == 0 {
		fmt.Printf("%s: no changes needed\n", path)
		return 0, nil
	}

	if err := os.WriteFile(path, []byte(result), 0o644); err != nil {
		return 0, fmt.Errorf("writing %s: %w", path, err)
	}

	return count, nil
}
