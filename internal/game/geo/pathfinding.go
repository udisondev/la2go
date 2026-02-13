package geo

import (
	"container/heap"
	"math"
)

// Point3D represents a 3D coordinate in world space.
type Point3D struct {
	X, Y, Z int32
}

// FindPath finds a path from start to end using A* pathfinding.
// Returns a slice of Point3D waypoints (world coordinates) or nil if no path found.
// maxIterations limits CPU usage (default: MaxPathfindIterations = 7000).
func (e *Engine) FindPath(sx, sy, sz, ex, ey, ez int32) []Point3D {
	if !e.IsLoaded() {
		// No geodata — direct path
		return []Point3D{{X: ex, Y: ey, Z: ez}}
	}

	gsx := GeoX(sx)
	gsy := GeoY(sy)
	gex := GeoX(ex)
	gey := GeoY(ey)

	startZ := e.getNearestZ(gsx, gsy, sz)
	endZ := e.getNearestZ(gex, gey, ez)

	// Same cell — already there
	if gsx == gex && gsy == gey {
		return []Point3D{{X: ex, Y: ey, Z: endZ}}
	}

	// Run A*
	result := e.astar(gsx, gsy, startZ, gex, gey, endZ)
	if result == nil {
		return nil // No path found
	}

	// Convert to world coordinates
	path := make([]Point3D, 0, 32)
	for n := result; n != nil; n = n.parent {
		path = append(path, Point3D{
			X: WorldX(n.x),
			Y: WorldY(n.y),
			Z: n.z,
		})
	}

	// Reverse (A* builds path backward)
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Anti-zigzag: remove unnecessary intermediate waypoints (up to 3 passes)
	path = e.smoothPath(path)

	return path
}

// smoothPath removes unnecessary intermediate waypoints from an A* path.
// If waypoint N can be reached directly from N-2 (no wall), waypoint N-1 is removed.
// Runs up to 3 passes to progressively simplify the path.
//
// Java reference: GeoEngine.reducePointList (lines 724-780).
func (e *Engine) smoothPath(path []Point3D) []Point3D {
	for pass := range 3 {
		_ = pass
		if len(path) <= 2 {
			return path
		}

		changed := false
		smoothed := make([]Point3D, 0, len(path))
		smoothed = append(smoothed, path[0])

		for i := 1; i < len(path)-1; i++ {
			prev := smoothed[len(smoothed)-1]
			next := path[i+1]

			// Check if direct movement from prev to next is possible
			if e.CanMoveToTarget(prev.X, prev.Y, prev.Z, next.X, next.Y, next.Z) {
				// Skip intermediate point path[i]
				changed = true
				continue
			}
			smoothed = append(smoothed, path[i])
		}
		smoothed = append(smoothed, path[len(path)-1])
		path = smoothed

		if !changed {
			break
		}
	}
	return path
}

// geoNode represents a node in the A* search graph.
type geoNode struct {
	x, y, z int32
	parent  *geoNode
	gCost   float64 // Actual cost from start
	hCost   float64 // Heuristic cost to target
	fCost   float64 // gCost + hCost
	index   int     // heap index
}

// astar implements the A* algorithm on geodata cells.
func (e *Engine) astar(sx, sy, sz, tx, ty, tz int32) *geoNode {
	start := &geoNode{x: sx, y: sy, z: sz}
	start.hCost = heuristic(sx, sy, sz, tx, ty, tz)
	start.fCost = start.hCost

	openList := &nodeHeap{}
	heap.Init(openList)
	heap.Push(openList, start)

	closed := make(map[nodeKey]struct{}, 256)

	for i := range MaxPathfindIterations {
		_ = i
		if openList.Len() == 0 {
			return nil
		}

		current := heap.Pop(openList).(*geoNode)

		// Goal reached (allow Z tolerance of 64 units)
		if current.x == tx && current.y == ty && abs32(current.z-tz) < 64 {
			return current
		}

		key := nodeKey{current.x, current.y, current.z}
		if _, exists := closed[key]; exists {
			continue
		}
		closed[key] = struct{}{}

		// Expand neighbors
		nswe := e.getNSWE(current.x, current.y, current.z)
		e.expandNeighbors(current, nswe, tx, ty, tz, openList, closed)
	}

	return nil // Max iterations exceeded
}

// expandNeighbors adds valid adjacent cells to the open list.
func (e *Engine) expandNeighbors(
	current *geoNode,
	nswe byte,
	tx, ty, tz int32,
	openList *nodeHeap,
	closed map[nodeKey]struct{},
) {
	type dir struct {
		dx, dy   int32
		flag     byte
		diagonal bool
	}

	// Cardinal directions
	cardinals := [4]dir{
		{0, -1, NSWENorth, false},
		{1, 0, NSWEEast, false},
		{0, 1, NSWESouth, false},
		{-1, 0, NSWEWest, false},
	}

	cardinalNodes := [4]*geoNode{nil, nil, nil, nil} // N, E, S, W

	for i, d := range cardinals {
		if nswe&d.flag == 0 {
			continue
		}
		nx := current.x + d.dx
		ny := current.y + d.dy
		nz := e.getNearestZ(nx, ny, current.z)

		key := nodeKey{nx, ny, nz}
		if _, exists := closed[key]; exists {
			continue
		}

		stepZ := abs32(nz - current.z)
		weight := WeightLow
		if stepZ > 16 || e.getNSWE(nx, ny, nz) != NSWEAll {
			weight = WeightHigh
		}

		gCost := current.gCost + weight
		node := &geoNode{
			x: nx, y: ny, z: nz,
			parent: current,
			gCost:  gCost,
			hCost:  heuristic(nx, ny, nz, tx, ty, tz),
		}
		node.fCost = node.gCost + node.hCost

		cardinalNodes[i] = node
		heap.Push(openList, node)
	}

	// Diagonal directions (anti-corner-cut: both adjacent cardinals must be passable)
	diagonals := [4]struct {
		dx, dy int32
		adj1   int // cardinal index 1
		adj2   int // cardinal index 2
	}{
		{1, -1, 0, 1},  // NE: need N(0) and E(1)
		{1, 1, 1, 2},   // SE: need E(1) and S(2)
		{-1, 1, 2, 3},  // SW: need S(2) and W(3)
		{-1, -1, 3, 0}, // NW: need W(3) and N(0)
	}

	for _, d := range diagonals {
		if cardinalNodes[d.adj1] == nil || cardinalNodes[d.adj2] == nil {
			continue
		}

		nx := current.x + d.dx
		ny := current.y + d.dy
		nz := e.getNearestZ(nx, ny, current.z)

		key := nodeKey{nx, ny, nz}
		if _, exists := closed[key]; exists {
			continue
		}

		stepZ := abs32(nz - current.z)
		weight := WeightDiagonal
		if stepZ > 16 || e.getNSWE(nx, ny, nz) != NSWEAll {
			weight = WeightHigh
		}

		gCost := current.gCost + weight
		node := &geoNode{
			x: nx, y: ny, z: nz,
			parent: current,
			gCost:  gCost,
			hCost:  heuristic(nx, ny, nz, tx, ty, tz),
		}
		node.fCost = node.gCost + node.hCost

		heap.Push(openList, node)
	}
}

// heuristic calculates the 3D Euclidean distance with reduced Z weight.
// Java reference: NodeBuffer.getCost() — Z is scaled 1/16.
func heuristic(x, y, z, tx, ty, tz int32) float64 {
	dx := float64(x - tx)
	dy := float64(y - ty)
	dz := float64(z - tz)
	return math.Sqrt(dx*dx + dy*dy + dz*dz/256.0)
}

// nodeKey uniquely identifies a cell position for the closed set.
type nodeKey struct {
	x, y, z int32
}

// nodeHeap implements container/heap for A* open list (min-heap by fCost).
type nodeHeap []*geoNode

func (h nodeHeap) Len() int            { return len(h) }
func (h nodeHeap) Less(i, j int) bool  { return h[i].fCost < h[j].fCost }
func (h nodeHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *nodeHeap) Push(x any)         { n := x.(*geoNode); n.index = len(*h); *h = append(*h, n) }
func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // GC
	node.index = -1
	*h = old[:n-1]
	return node
}
