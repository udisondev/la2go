package model

// Location представляет координаты в игровом мире.
// Value type, передаётся по значению (immutable).
type Location struct {
	X       int32
	Y       int32
	Z       int32
	Heading uint16 // 0-65535
}

// NewLocation создаёт Location с указанными координатами.
func NewLocation(x, y, z int32, heading uint16) Location {
	// Heading already 0-65535 по типу uint16, no need to clamp
	return Location{X: x, Y: y, Z: z, Heading: heading}
}

// WithHeading возвращает новый Location с обновлённым направлением (immutable pattern).
func (l Location) WithHeading(heading uint16) Location {
	l.Heading = heading
	return l
}

// WithCoordinates возвращает новый Location с обновлёнными координатами (immutable pattern).
func (l Location) WithCoordinates(x, y, z int32) Location {
	l.X = x
	l.Y = y
	l.Z = z
	return l
}

// DistanceSquared возвращает квадрат расстояния до другой точки (без sqrt для производительности).
func (l Location) DistanceSquared(other Location) int64 {
	dx := int64(l.X - other.X)
	dy := int64(l.Y - other.Y)
	dz := int64(l.Z - other.Z)
	return dx*dx + dy*dy + dz*dz
}
