package data

// spawnDef — определение одного спавна NPC (generated).
type spawnDef struct {
	npcID        int32
	count        int32
	respawnDelay int32 // seconds
	respawnRand  int32 // seconds (random addition)
	chaseRange   int32
	// Fixed location (either x/y/z or territory)
	x, y, z int32
	heading int32
	// Territory spawn (polygon area)
	territory *territoryDef
}

// territoryDef — polygon territory for random spawn location.
type territoryDef struct {
	minZ, maxZ int32
	nodes      []pointDef
}

// pointDef — 2D point.
type pointDef struct {
	x, y int32
}
