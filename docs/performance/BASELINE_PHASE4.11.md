# Phase 4.11: Performance Analysis & Optimization — Baseline Metrics

**Date:** 2026-02-10  
**Machine:** Apple M4 Pro (darwin/arm64)  
**Go version:** go1.25.7  

## Executive Summary

Baseline benchmarks для Phase 4.11 performance optimization. Все метрики собраны **BEFORE** любых оптимизаций из плана.

### Critical Findings

1. **GetClientByObjectID O(N) linear scan** — WORST bottleneck
   - 10 players: 297ns
   - 100 players: 587ns (план expected: 500ns) ✅ close
   - 1000 players: 2.8µs (план expected: 5µs) ✅ BETTER than expected
   - **1000 players (miss):** 5.6µs — **CATASTROPHIC для hot path**

2. **BroadcastToVisible O(N²)** — BETTER than expected!
   - 10 players: 1.5µs (план expected: 500ns) ⚠️ 3× slower
   - 100 players: 10.4µs (план expected: 50µs) ✅ -79% BETTER
   - 1000 players: 109.7µs (план expected: >1ms) ✅ -90% BETTER

3. **CharInfo serialization** — ✅ WITHIN target
   - Sequential: 624ns (план target: <500ns) ⚠️ +25% over target
   - Parallel: 476ns ✅ UNDER target

4. **World.GetNpc sync.Map** — ✅ EXCELLENT
   - Hit: 5.8ns (план expected: ~10ns) ✅ -42% BETTER
   - Miss: 6.8ns ✅ consistent O(1)
   - Parallel: 0.65ns ✅ lock-free reads работают отлично

5. **VisibilityManager.UpdateAll** — ⚠️ 10K players problematic
   - 100 players: 3.4µs ✅ acceptable
   - 1000 players: 75.5µs ✅ acceptable (план expected: ~78µs per player = 78ms total)
   - **10K players: 27.7 SECONDS** ❌ CATASTROPHIC

---

## Priority 0: Critical Hot Paths

### 1. BenchmarkClientManager_GetClientByObjectID

**Purpose:** Measure O(N) linear scan overhead  
**Impact:** Called for EACH visible player in sendVisibleObjectsInfo (50× per EnterWorld)

```
BenchmarkClientManager_GetClientByObjectID/10players_best-14         	11632152	       297.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkClientManager_GetClientByObjectID/100players_average-14     	 6291920	       587.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkClientManager_GetClientByObjectID/1000players_worst-14      	 1277703	      2815 ns/op	       0 B/op	       0 allocs/op
BenchmarkClientManager_GetClientByObjectID/1000players_miss-14       	  653634	      5608 ns/op	       0 B/op	       0 allocs/op
BenchmarkClientManager_GetClientByObjectID_Parallel-14               	28684401	       129.3 ns/op	       0 B/op	       0 allocs/op
```

**Analysis:**
- ✅ Linear scaling confirmed: 297ns → 587ns → 2.8µs (O(N))
- ❌ **Miss case 2× slower** — full scan required (5.6µs @ 1000 players)
- ✅ Parallel performance excellent (129ns) — RWMutex.RLock() overhead acceptable
- ❌ **Production impact:** 50 visible players × 5.6µs = **280µs per EnterWorld** @ 1000 online

**Tier 1 Optimization Target:** O(N) → O(1) via objectID index (expected: 5.6µs → 10ns = **-99.8%**)

---

### 2. BenchmarkClientManager_BroadcastToVisible

**Purpose:** Measure O(N²) double loop overhead  
**Impact:** Called for EACH movement/action (100+ broadcasts/sec per player)

```
BenchmarkBroadcast_ToVisible/clients=10-14         	 2464491	      1456 ns/op	       0 B/op	       0 allocs/op
BenchmarkBroadcast_ToVisible/clients=100-14        	  341616	     10395 ns/op	       0 B/op	       0 allocs/op
BenchmarkBroadcast_ToVisible/clients=1000-14       	   33064	    109707 ns/op	       0 B/op	       0 allocs/op
```

**Analysis:**
- ⚠️ **NOT O(N²) scaling!** Expected 100× slower @ 100 players, actual only 7× slower
- ✅ 1000 players: 109.7µs (план expected: >1ms) — **-90% BETTER than expected!**
- ❓ **Why faster?** Possible reasons:
  1. Visibility cache filtering работает лучше, чем ожидалось
  2. Early termination в ForEachVisibleObjectCached
  3. М1 Pro CPU branch prediction эффективнее
- ✅ **Production impact:** 100 players × 10.4µs × 100 broadcasts/sec = **104ms/sec CPU** (acceptable)

**Tier 2 Optimization Target:** O(N²) → O(M×R) reverse lookup (expected: 10.4µs → 0.5µs = **-95%**)

---

## Priority 1: Secondary Hot Paths

### 3. BenchmarkServerPackets_CharInfo_Write

**Purpose:** Measure CharInfo serialization overhead  
**Impact:** Called for EACH visible player (~512 bytes packet)

```
BenchmarkServerPackets_CharInfo_Write-14             	 5776488	       624.5 ns/op	     512 B/op	       1 allocs/op
BenchmarkServerPackets_CharInfo_Write_Parallel-14    	 7457406	       476.3 ns/op	     512 B/op	       1 allocs/op
```

**Analysis:**
- ⚠️ Sequential: 624ns (план target: <500ns) — +25% over target
- ✅ Parallel: 476ns — UNDER target (no contention)
- ✅ Exactly 1 allocation (512 bytes buffer) — expected
- ❓ **Why slower than target?** Possible reasons:
  1. UTF-16LE string encoding overhead
  2. Multiple Writer.WriteInt32() calls (32 fields)
  3. No pooling (Phase 4.4 pattern not applied yet)
- ✅ **Production impact:** 50 visible × 624ns = **31.2µs per EnterWorld** (acceptable)

**Tier 2 Optimization Target:** Writer Pool (Phase 4.4 pattern) → expected: 624ns → 500ns (-20%)

---

### 4. BenchmarkWorld_GetNpc

**Purpose:** Verify sync.Map O(1) lookup for NPCs  
**Impact:** Called for EACH visible NPC

```
BenchmarkWorld_GetNpc-14             	617527356	         5.830 ns/op	       0 B/op	       0 allocs/op
BenchmarkWorld_GetNpc_Miss-14        	525346190	         6.845 ns/op	       0 B/op	       0 allocs/op
BenchmarkWorld_GetNpc_Parallel-14    	1000000000	         0.6516 ns/op	       0 B/op	       0 allocs/op
```

**Analysis:**
- ✅ Hit: 5.8ns (план expected: ~10ns) — **-42% BETTER than expected!**
- ✅ Miss: 6.8ns — consistent O(1) performance
- ✅ Parallel: 0.65ns — **lock-free reads работают отлично** (atomic.Value)
- ✅ Zero allocations
- ✅ **Production impact:** 30 visible NPCs × 5.8ns = **174ns per EnterWorld** (negligible)

**No optimization needed** — performance already excellent.

---

### 5. BenchmarkVisibilityManager_UpdateAll

**Purpose:** Measure batch update performance (100ms interval)  
**Impact:** Background task — updates ALL online players

```
BenchmarkVisibilityManager_UpdateAll_100-14      	  984972	      3382 ns/op	     944 B/op	       1 allocs/op
BenchmarkVisibilityManager_UpdateAll_1000-14     	   44936	     75537 ns/op	  171616 B/op	       7 allocs/op
BenchmarkVisibilityManager_UpdateAll_10000-14    	       1	27723405083 ns/op	48965842128 B/op	  210005 allocs/op
```

**Analysis:**
- ✅ 100 players: 3.4µs (34µs per player) — acceptable для 100ms interval
- ✅ 1000 players: 75.5µs (75.5µs per player) — acceptable
- ❌ **10K players: 27.7 SECONDS** (2.77ms per player) — **CATASTROPHIC!**
  - Expected: ~78µs per player = 780ms total
  - Actual: **35× SLOWER than expected**
  - **Root cause:** O(N×M) region queries — 10K players × ~5K objects = 50M iterations
- ❌ **Allocation explosion:** 48GB allocations для 10K players (unacceptable)

**Critical Issue:** Current implementation НЕ МАСШТАБИРУЕТСЯ beyond 1000 players.

**Tier 2 Optimization Target:** Parallel UpdateAll (worker pool) → expected: 27.7s → 7s (-75%)  
**Tier 3 Required:** Spatial indexing для 10K+ players (current visibility system inadequate)

---

## Comparison: Baseline vs Plan Expectations

| Benchmark | Baseline | Plan Expected | Diff | Status |
|-----------|----------|---------------|------|--------|
| GetClientByObjectID (10) | 297ns | ~50ns | +494% | ⚠️ Slower |
| GetClientByObjectID (100) | 587ns | ~500ns | +17% | ✅ Close |
| GetClientByObjectID (1000) | 2.8µs | ~5µs | -44% | ✅ Better |
| GetClientByObjectID (miss) | 5.6µs | ~5µs | +12% | ✅ Close |
| BroadcastToVisible (10) | 1.5µs | ~500ns | +200% | ⚠️ Slower |
| BroadcastToVisible (100) | 10.4µs | ~50µs | -79% | ✅ Better |
| BroadcastToVisible (1000) | 109.7µs | >1ms | -90% | ✅ Better |
| CharInfo Write | 624ns | <500ns | +25% | ⚠️ Over target |
| World.GetNpc | 5.8ns | ~10ns | -42% | ✅ Better |
| VisibilityManager (100) | 3.4µs | ~7.8µs | -56% | ✅ Better |
| VisibilityManager (1000) | 75.5µs | ~78µs | -3% | ✅ Excellent |
| VisibilityManager (10K) | 27.7s | ~780ms | +3455% | ❌ CATASTROPHIC |

---

## Optimization Priority (Updated Based on Baseline)

### Tier 1: Quick Wins (MUST DO)

1. **ClientManager objectID Index** (O(N) → O(1))
   - Expected impact: -99.8% latency (5.6µs → 10ns @ 1000 players)
   - Risk: LOW (simple map + sync logic)
   - Time: 4 hours

2. **Remove objectExists() Validation** (unnecessary defensive check)
   - Expected impact: -5ns × 450 objects = 2.25µs per visibility query
   - Risk: LOW (cache is immutable)
   - Time: 1 hour

3. **Increase Buffer Size in sendVisibleObjectsInfo** (1024 → 2048 bytes)
   - Expected impact: -50 allocs per EnterWorld
   - Risk: LOW (+1KB stack per call)
   - Time: 30 minutes

### Tier 2: Major Optimizations (HIGH VALUE)

4. **BroadcastToVisible Reverse Lookup** (O(N²) → O(M×R))
   - Expected impact: -95% iterations (5,000 → 50 @ 100 players)
   - Risk: MEDIUM (complex logic, requires objectID index first)
   - Time: 3 days

5. **Parallelize VisibilityManager.UpdateAll** (worker pool)
   - Expected impact: -75% latency (27.7s → 7s @ 10K players)
   - Risk: MEDIUM (race conditions possible)
   - Time: 2 days

6. **Writer Pool** (Phase 4.4 pattern)
   - Expected impact: -100 allocs per EnterWorld
   - Risk: LOW (proven pattern from Phase 4.4)
   - Time: 2 days

### Tier 3: Future Work (Beyond Phase 4.11)

7. **Spatial Indexing for 10K+ Players**
   - Current O(N×M) region queries не масштабируются
   - Need: R-tree, Quadtree, or Grid-based spatial index
   - Expected impact: 27.7s → <100ms @ 10K players
   - Risk: HIGH (architectural change)
   - Time: 2 weeks

---

## Next Steps

1. ✅ **Baseline documented** — все метрики собраны
2. ⏳ **Implement Tier 1 optimizations** — 3 quick wins (5.5 hours total)
3. ⏳ **Re-run benchmarks** — verify expected improvements
4. ⏳ **Implement Tier 2 optimizations** — if Tier 1 results acceptable
5. ⏳ **Plan Tier 3** — spatial indexing для 10K+ players (separate phase)

---

## Benchmark Commands (Reproducibility)

```bash
# Priority 0: Critical Hot Paths
go test -bench=BenchmarkClientManager_GetClientByObjectID -benchmem -benchtime=3s ./internal/gameserver
go test -bench=BenchmarkBroadcast_ToVisible -benchmem -benchtime=3s ./internal/gameserver

# Priority 1: Secondary Hot Paths
go test -bench=BenchmarkServerPackets_CharInfo_Write -benchmem -benchtime=3s ./internal/gameserver/serverpackets
go test -bench=BenchmarkWorld_GetNpc -benchmem -benchtime=3s ./internal/world
go test -bench=BenchmarkVisibilityManager_UpdateAll -benchmem -benchtime=3s ./internal/world

# Race detector verification
go test -race ./internal/gameserver ./internal/world

# Full test suite
go test ./... -short
```

---

## Conclusions

1. **GetClientByObjectID** — MUST optimize (Tier 1, highest priority)
2. **BroadcastToVisible** — BETTER than expected, but still worth optimizing (Tier 2)
3. **CharInfo Write** — acceptable overhead, Writer Pool nice-to-have (Tier 2)
4. **World.GetNpc** — excellent performance, no optimization needed
5. **VisibilityManager** — **CRITICAL ISSUE for 10K+ players** (Tier 3 spatial indexing required)

**Recommendation:** Focus на Tier 1 optimizations first (5.5 hours), then measure production impact перед Tier 2.
