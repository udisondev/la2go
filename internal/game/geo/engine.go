package geo

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
)

// Engine is the main GeoEngine for pathfinding and LOS checks.
// Thread-safe: regions are loaded once and never modified.
type Engine struct {
	regions [GeoRegionsX * GeoRegionsY]atomic.Pointer[Region]
	loaded  atomic.Int32
}

// NewEngine creates an empty GeoEngine (no regions loaded).
func NewEngine() *Engine {
	return &Engine{}
}

// LoadGeodata loads all .l2j files from the given directory.
// File naming convention: "<regionX>_<regionY>.l2j"
func (e *Engine) LoadGeodata(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading geodata dir %s: %w", dir, err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".l2j" {
			continue
		}

		var rx, ry int
		base := name[:len(name)-len(ext)]
		if _, err := fmt.Sscanf(base, "%d_%d", &rx, &ry); err != nil {
			slog.Warn("skip geodata file (bad name)", "file", name)
			continue
		}

		if rx < 0 || rx >= GeoRegionsX || ry < 0 || ry >= GeoRegionsY {
			slog.Warn("skip geodata file (out of range)", "file", name, "rx", rx, "ry", ry)
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("reading geodata %s: %w", name, err)
		}

		region, err := LoadRegion(data)
		if err != nil {
			return fmt.Errorf("parsing geodata %s: %w", name, err)
		}

		e.regions[rx*GeoRegionsY+ry].Store(region)
		loaded++
	}

	e.loaded.Store(int32(loaded))
	slog.Info("geodata loaded", "regions", loaded, "dir", dir)
	return nil
}

// IsLoaded returns true if any geodata regions are loaded.
func (e *Engine) IsLoaded() bool {
	return e.loaded.Load() > 0
}

// getRegion returns the region for given geo coordinates (nil if not loaded).
func (e *Engine) getRegion(geoX, geoY int32) *Region {
	rx, ry := RegionXY(geoX, geoY)
	if rx < 0 || rx >= GeoRegionsX || ry < 0 || ry >= GeoRegionsY {
		return nil
	}
	return e.regions[rx*GeoRegionsY+ry].Load()
}

// HasGeoPos returns true if geodata exists at the given world position.
func (e *Engine) HasGeoPos(worldX, worldY int32) bool {
	gx := GeoX(worldX)
	gy := GeoY(worldY)
	region := e.getRegion(gx, gy)
	if region == nil {
		return false
	}
	localX := gx % RegionCellsX
	localY := gy % RegionCellsY
	return region.HasGeoData(localX, localY)
}

// GetHeight returns the geodata Z height at world (x, y, z).
// Returns worldZ unchanged if no geodata is loaded for this position.
func (e *Engine) GetHeight(worldX, worldY, worldZ int32) int32 {
	gx := GeoX(worldX)
	gy := GeoY(worldY)
	region := e.getRegion(gx, gy)
	if region == nil {
		return worldZ
	}
	localX := gx % RegionCellsX
	localY := gy % RegionCellsY
	return region.GetNearestZ(localX, localY, worldZ)
}

// getNearestZ returns nearest Z from geodata for geo coordinates.
func (e *Engine) getNearestZ(geoX, geoY int32, worldZ int32) int32 {
	if geoX < 0 || geoY < 0 {
		return worldZ
	}
	region := e.getRegion(geoX, geoY)
	if region == nil {
		return worldZ
	}
	localX := geoX % RegionCellsX
	localY := geoY % RegionCellsY
	return region.GetNearestZ(localX, localY, worldZ)
}

// getNextHigherZ returns the next higher Z from geodata for geo coordinates.
// Used for LOS: when movement is blocked, we need the wall top height.
func (e *Engine) getNextHigherZ(geoX, geoY int32, worldZ int32) int32 {
	if geoX < 0 || geoY < 0 {
		return worldZ
	}
	region := e.getRegion(geoX, geoY)
	if region == nil {
		return worldZ
	}
	localX := geoX % RegionCellsX
	localY := geoY % RegionCellsY
	return region.GetNextHigherZ(localX, localY, worldZ)
}

// getNSWE returns the NSWE mask at geo coordinates.
func (e *Engine) getNSWE(geoX, geoY int32, worldZ int32) byte {
	if geoX < 0 || geoY < 0 {
		return NSWEAll
	}
	region := e.getRegion(geoX, geoY)
	if region == nil {
		return NSWEAll
	}
	localX := geoX % RegionCellsX
	localY := geoY % RegionCellsY
	return region.GetNSWE(localX, localY, worldZ)
}

// hasGeoData returns true if geo coordinates have per-cell data.
func (e *Engine) hasGeoData(geoX, geoY int32) bool {
	if geoX < 0 || geoY < 0 {
		return false
	}
	region := e.getRegion(geoX, geoY)
	if region == nil {
		return false
	}
	localX := geoX % RegionCellsX
	localY := geoY % RegionCellsY
	return region.HasGeoData(localX, localY)
}
