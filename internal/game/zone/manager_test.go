package zone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestManager создает Manager с вручную заданными зонами (без data.ZoneTable).
func newTestManager(zones ...Zone) *Manager {
	m := NewManager()

	for _, z := range zones {
		m.zones = append(m.zones, z)
		m.byID[z.ID()] = z
		m.byType[z.ZoneType()] = append(m.byType[z.ZoneType()], z)
	}

	m.buildGrid()

	return m
}

func makeTownZone(id int32, name string, nodesX, nodesY []int32, minZ, maxZ int32) *PeaceZone {
	return &PeaceZone{
		BaseZone: &BaseZone{
			id:       id,
			name:     name,
			zoneType: TypeTown,
			shape:    "NPoly",
			minZ:     minZ,
			maxZ:     maxZ,
			nodesX:   nodesX,
			nodesY:   nodesY,
			params:   map[string]string{},
		},
	}
}

func makeSiegeZone(id int32, name string, nodesX, nodesY []int32, minZ, maxZ int32) *PvPZone {
	return &PvPZone{
		BaseZone: &BaseZone{
			id:       id,
			name:     name,
			zoneType: TypeSiege,
			shape:    "NPoly",
			minZ:     minZ,
			maxZ:     maxZ,
			nodesX:   nodesX,
			nodesY:   nodesY,
			params:   map[string]string{},
		},
	}
}

func TestGetZonesAt(t *testing.T) {
	// Квадрат 0..1000, 0..1000
	town := makeTownZone(1, "Giran",
		[]int32{0, 1000, 1000, 0},
		[]int32{0, 0, 1000, 1000},
		-1000, 1000,
	)

	// Квадрат 500..1500, 500..1500 — пересекается с town
	siege := makeSiegeZone(2, "Siege Area",
		[]int32{500, 1500, 1500, 500},
		[]int32{500, 500, 1500, 1500},
		-1000, 1000,
	)

	m := newTestManager(town, siege)

	// Точка в пересечении обеих зон.
	zones := m.GetZonesAt(750, 750, 0)
	require.Len(t, zones, 2, "point in overlap should match both zones")

	// Точка только в town.
	zones = m.GetZonesAt(100, 100, 0)
	require.Len(t, zones, 1)
	assert.Equal(t, int32(1), zones[0].ID())

	// Точка только в siege.
	zones = m.GetZonesAt(1200, 1200, 0)
	require.Len(t, zones, 1)
	assert.Equal(t, int32(2), zones[0].ID())

	// Точка вне обеих зон.
	zones = m.GetZonesAt(2000, 2000, 0)
	assert.Empty(t, zones)
}

func TestIsInPeaceZone(t *testing.T) {
	town := makeTownZone(1, "Aden",
		[]int32{0, 1000, 1000, 0},
		[]int32{0, 0, 1000, 1000},
		-500, 500,
	)

	siege := makeSiegeZone(2, "Aden Siege",
		[]int32{2000, 3000, 3000, 2000},
		[]int32{2000, 2000, 3000, 3000},
		-500, 500,
	)

	m := newTestManager(town, siege)

	assert.True(t, m.IsInPeaceZone(500, 500, 0), "inside town should be peace")
	assert.False(t, m.IsInPeaceZone(2500, 2500, 0), "inside siege should not be peace")
	assert.False(t, m.IsInPeaceZone(5000, 5000, 0), "outside all zones should not be peace")
}

func TestGetZonesByType(t *testing.T) {
	town1 := makeTownZone(1, "Giran",
		[]int32{0, 100, 100, 0}, []int32{0, 0, 100, 100}, -1000, 1000)
	town2 := makeTownZone(2, "Aden",
		[]int32{200, 300, 300, 200}, []int32{200, 200, 300, 300}, -1000, 1000)
	siege := makeSiegeZone(3, "Siege",
		[]int32{400, 500, 500, 400}, []int32{400, 400, 500, 500}, -1000, 1000)

	m := newTestManager(town1, town2, siege)

	towns := m.GetZonesByType(TypeTown)
	assert.Len(t, towns, 2, "should have 2 town zones")

	sieges := m.GetZonesByType(TypeSiege)
	assert.Len(t, sieges, 1, "should have 1 siege zone")

	waters := m.GetZonesByType(TypeWater)
	assert.Empty(t, waters, "should have no water zones")
}

func TestGetZoneByID(t *testing.T) {
	town := makeTownZone(42, "Gludin",
		[]int32{0, 100, 100, 0}, []int32{0, 0, 100, 100}, -1000, 1000)

	m := newTestManager(town)

	z := m.GetZoneByID(42)
	require.NotNil(t, z)
	assert.Equal(t, "Gludin", z.Name())

	assert.Nil(t, m.GetZoneByID(999), "non-existent zone should return nil")
}

func TestNegativeCoordinatesGrid(t *testing.T) {
	// Зона в отрицательных координатах.
	town := makeTownZone(1, "NegTown",
		[]int32{-2000, -1000, -1000, -2000},
		[]int32{-2000, -2000, -1000, -1000},
		-500, 500,
	)

	m := newTestManager(town)

	zones := m.GetZonesAt(-1500, -1500, 0)
	require.Len(t, zones, 1, "should find zone at negative coords")
	assert.Equal(t, int32(1), zones[0].ID())

	zones = m.GetZonesAt(0, 0, 0)
	assert.Empty(t, zones, "should not find zone at origin")
}

func TestZoneInterfaceCompliance(t *testing.T) {
	// Проверяем, что все типы зон реализуют интерфейс Zone.
	var _ Zone = (*PeaceZone)(nil)
	var _ Zone = (*PvPZone)(nil)
	var _ Zone = (*DamageZone)(nil)
	var _ Zone = (*WaterZone)(nil)
	var _ Zone = (*EffectZone)(nil)
}
