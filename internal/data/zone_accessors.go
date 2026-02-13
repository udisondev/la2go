package data

// NodeX returns X coordinate of a node.
func (p pointDef) NodeX() int32 { return p.x }

// NodeY returns Y coordinate of a node.
func (p pointDef) NodeY() int32 { return p.y }

// SpawnX returns X coordinate of spawn.
func (s zoneSpawnDef) SpawnX() int32 { return s.x }

// SpawnY returns Y coordinate of spawn.
func (s zoneSpawnDef) SpawnY() int32 { return s.y }

// SpawnZ returns Z coordinate of spawn.
func (s zoneSpawnDef) SpawnZ() int32 { return s.z }

// SpawnType returns spawn type.
func (s zoneSpawnDef) SpawnType() string { return s.spawnType }
