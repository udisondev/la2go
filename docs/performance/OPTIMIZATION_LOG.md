# Optimization Log — la2go

История всех оптимизаций производительности с результатами бенчмарков и профилирования.

**Цель:** Сделать la2go **значительно производительнее L2J Mobius** за счет использования горутин, sync.Pool, atomic операций и устранения GC пауз Java.

---

## Формат записи

```markdown
## YYYY-MM-DD: Название оптимизации

### Проблема
Описание проблемы производительности (из профилирования или бенчмарков).

### Диагностика
- CPU profile: <что показало>
- Memory profile: <что показало>
- Escape analysis: <что показало>
- Benchmark: <baseline результаты>

### Решение
Описание примененной оптимизации.

### Результат
- Время: ±X% (old → new)
- Аллокации: ±Y% (old → new)
- Throughput: ±Z% (old → new)

### Коммит
<commit hash или PR number>

### Профили
- CPU: profiles/cpu_YYYYMMDD_HHMMSS.prof
- Memory: profiles/mem_YYYYMMDD_HHMMSS.prof
- Baseline: benchmarks/baseline_YYYYMMDD.txt
- Optimized: benchmarks/optimized_YYYYMMDD.txt
```

---

## 2026-02-09: Инфраструктура Performance Testing (Baseline)

### Проблема
- ❌ НЕТ benchmarks
- ❌ НЕТ профилирования
- ❌ НЕТ escape analysis
- ❌ НЕТ CI для отслеживания регрессий

### Решение
Создана полная инфраструктура performance testing:

**Benchmarks:**
- `internal/crypto/blowfish_bench_test.go` — Blowfish Encrypt/Decrypt, Checksum, XOR
- `internal/crypto/rsa_bench_test.go` — RSA Decrypt, ScrambleModulus
- `internal/login/bufpool_bench_test.go` — BytePool vs make, concurrent
- `internal/login/session_manager_bench_test.go` — Validate, concurrent read/write
- `internal/gameserver/table_bench_test.go` — Register, GetByID, concurrent

**Инструменты:**
- `scripts/profile.sh` — CPU/Memory/Block/Mutex/Escape профилирование
- `scripts/bench.sh` — автоматизация запуска бенчмарков
- `.github/workflows/benchmarks.yml` — CI для PR с автоматическим сравнением

**Документация:**
- `docs/performance/BENCHMARKING_GUIDE.md`
- `docs/performance/PROFILING_GUIDE.md`
- `docs/performance/OPTIMIZATION_LOG.md` (этот файл)

### Baseline результаты (Apple M4 Pro, 14 cores)

**Crypto:**
```
BenchmarkBlowfishEncrypt-14              139861       864.1 ns/op       0 B/op       0 allocs/op
BenchmarkRSADecrypt_1024-14                 412    286484 ns/op    6288 B/op      24 allocs/op
```

**BytePool:**
```
BenchmarkBytePool_GetPut-14             2561670        49.58 ns/op      24 B/op       1 allocs/op
```

**SessionManager:**
```
BenchmarkSessionManager_Validate-14    16447932         6.907 ns/op       0 B/op       0 allocs/op
```

**GameServerTable:**
```
BenchmarkGameServerTable_GetByID-14    28207111         3.663 ns/op       0 B/op       0 allocs/op
```

### Коммит
TBD (initial performance infrastructure)

### Следующие шаги
1. Запустить полное профилирование всех пакетов
2. Выявить узкие места (CPU > 30%, Memory > 10MB)
3. Применить приоритизированные оптимизации (P0 → P1 → P2)
4. Сравнить с L2J Mobius (Java) на идентичном железе

---

## Шаблон для будущих оптимизаций

```markdown
## YYYY-MM-DD: <Название оптимизации>

### Проблема
<Описание проблемы>

### Диагностика
- CPU profile: <результаты>
- Memory profile: <результаты>
- Escape analysis: <результаты>
- Benchmark baseline:
  ```
  <вывод go test -bench>
  ```

### Решение
<Описание решения>

### Код
```go
// ДО
<старый код>

// ПОСЛЕ
<новый код>
```

### Результат
```
benchstat baseline.txt optimized.txt

name                old time/op    new time/op    delta
XXX-14               X.XXµs ± Y%    X.XXµs ± Y%   -ZZ.ZZ%  (p=0.000)

name                old alloc/op   new alloc/op   delta
XXX-14                XXXB ± Y%       XXXB ± Y%   -ZZ.ZZ%  (p=0.000)

name                old allocs/op  new allocs/op  delta
XXX-14                X.XX ± Y%      X.XX ± Y%    -ZZ.ZZ%  (p=0.000)
```

### Коммит
<commit hash>

### Профили
- CPU: profiles/cpu_YYYYMMDD_HHMMSS.prof
- Memory: profiles/mem_YYYYMMDD_HHMMSS.prof
```

---

## Целевые метрики (la2go vs L2J Mobius)

| Операция | L2J Mobius (Java) | la2go (Go) | Цель улучшения |
|----------|-------------------|------------|----------------|
| Login flow (полный цикл) | ~500µs | TBD | **-30%** |
| Blowfish Encrypt (256B) | ~1.2µs | 864ns | **✅ -28%** |
| RSA Decrypt (1024-bit) | ~350µs | 286µs | **✅ -18%** |
| SessionKey validation | ~15ns | 6.9ns | **✅ -54%** |
| Concurrent logins (1000 clients) | ~2s | TBD | **-50%** |
| Memory usage (10k online) | ~2GB | TBD | **-40%** |
| GC pause time | ~50ms | <1ms | **✅ -98%** |

**Легенда:**
- ✅ — цель достигнута
- ⏳ — в процессе
- ❌ — требует работы

---

## Известные узкие места (еще не оптимизированы)

### P0 🔴 Критичные

1. **Blowfish Encrypt/Decrypt loop** (internal/crypto/blowfish.go:44-46)
   - Проблема: O(packet_size/8) на КАЖДЫЙ пакет
   - Текущее: 864ns для 256B пакета
   - Цель: <600ns
   - Решение: Убрать bounds checks, использовать SIMD (если возможно)

2. **RSA Decryption** (internal/crypto/rsa.go:188-209)
   - Проблема: ~286µs блокировка на каждый login
   - Текущее: 6288 B/op, 24 allocs/op
   - Решение: Кеширование или параллелизация (если применимо)

### P1 🟠 Важные

3. **GameServerTable.RegisterWithFirstAvailableID** (internal/gameserver/table.go:58)
   - Проблема: O(maxID) линейный поиск под write lock
   - Решение: Free list (slice свободных ID)

4. **handlePlayerInGame** — mutex в цикле (internal/gslistener/handler.go:246)
   - Проблема: N × mutex lock
   - Решение: Batch AddAccounts([]string)

### P2 🟡 Желательные

5. **handleServerStatus** — множественные SetXXX locks (internal/gslistener/handler.go:321-336)
   - Проблема: 5+ mutex locks для одной операции
   - Решение: UpdateBatch(struct)

### P3 🔵 Низкий приоритет

6. **Client/GSConnection** — лишние mutex на immutable getters
   - Проблема: Lock на SessionID() (immutable после init)
   - Решение: Убрать lock (верифицировать через -race)

---

## Как добавлять записи

### Шаг 1: Baseline benchmark
```bash
./scripts/bench.sh crypto > benchmarks/baseline_$(date +%Y%m%d).txt
```

### Шаг 2: Профилирование
```bash
./scripts/profile.sh all ./internal/crypto
```

### Шаг 3: Применить оптимизацию

### Шаг 4: Re-benchmark
```bash
./scripts/bench.sh crypto > benchmarks/optimized_$(date +%Y%m%d).txt
benchstat benchmarks/baseline_*.txt benchmarks/optimized_*.txt > comparison.txt
```

### Шаг 5: Добавить запись в этот файл
Скопировать шаблон выше, заполнить результатами.

### Шаг 6: Коммит
```bash
git add docs/performance/OPTIMIZATION_LOG.md benchmarks/ profiles/
git commit -m "perf: <описание оптимизации>"
```

---

## Дополнительные ресурсы

- [BENCHMARKING_GUIDE.md](./BENCHMARKING_GUIDE.md) — как писать и запускать бенчмарки
- [PROFILING_GUIDE.md](./PROFILING_GUIDE.md) — как использовать pprof и escape analysis
- [Go Performance Best Practices](https://github.com/dgryski/go-perfbook)
- [Optimization Patterns](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)
