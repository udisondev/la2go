package constants

// IsPlayerObjectID returns true if objectID is in Player range.
// Player range: 268435456-536870911 (0x10000000-0x1FFFFFFF)
func IsPlayerObjectID(objectID uint32) bool {
	return objectID >= ObjectIDPlayerStart && objectID <= ObjectIDPlayerEnd
}

// IsNpcObjectID returns true if objectID is in NPC range.
// NPC range: 536870912+ (0x20000000+)
func IsNpcObjectID(objectID uint32) bool {
	return objectID >= ObjectIDNpcStart
}

// IsItemObjectID returns true if objectID is in Item (on ground) range.
// Item range: 1-268435455 (0x00000001-0x0FFFFFFF)
func IsItemObjectID(objectID uint32) bool {
	return objectID >= ObjectIDItemStart && objectID <= ObjectIDItemEnd
}
