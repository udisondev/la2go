# Profiling Guide — la2go

Руководство по профилированию и анализу производительности la2go с помощью pprof и escape analysis.

## Содержание

- [Введение](#введение)
- [CPU Profiling](#cpu-profiling)
- [Memory Profiling](#memory-profiling)
- [Escape Analysis](#escape-analysis)
- [Block Profiling](#block-profiling)
- [Mutex Profiling](#mutex-profiling)
- [Helper скрипты](#helper-скрипты)

---

## Введение

Профилирование позволяет:

1. **Найти узкие места** (bottlenecks) в коде
2. **Выявить неожиданные аллокации** и memory leaks
3. **Обнаружить lock contention** в параллельном коде
4. **Оптимизировать escape to heap** переменных

**Workflow:**
1. Запустить бенчмарки → обнаружить медленные операции
2. Запустить профилирование → найти конкретные функции-виновники
3. Применить оптимизацию
4. Re-benchmark → подтвердить улучшение

---

## CPU Profiling

### Что измеряет

CPU profiling показывает, **где программа тратит больше всего процессорного времени**.

### Как запустить

```bash
# Вручную
go test -cpuprofile=cpu.prof -bench=. ./internal/crypto
go tool pprof -http=:8080 cpu.prof

# Через helper скрипт
./scripts/profile.sh cpu ./internal/crypto
```

### Интерфейс pprof

Откроется браузер на `http://localhost:8080` с интерактивным профилем.

**Основные view:**

1. **Top** — функции, потребляющие больше всего CPU
2. **Graph** — граф вызовов с процентами времени
3. **Flame Graph** — визуализация stack traces
4. **Source** — исходный код с аннотациями времени

### Пример анализа

**Top view:**
```
Showing nodes accounting for 850ms, 85% of 1000ms total
      flat  flat%   sum%        cum   cum%
     450ms 45.00% 45.00%      450ms 45.00%  crypto/blowfish.(*Cipher).Encrypt
     200ms 20.00% 65.00%      200ms 20.00%  runtime.memmove
     150ms 15.00% 80.00%      600ms 60.00%  la2go/internal/crypto.BlowfishCipher.Encrypt
      50ms  5.00% 85.00%       50ms  5.00%  runtime.checksum
```

**Расшифровка:**
- **flat** — время в самой функции (без вызовов)
- **flat%** — процент от общего времени
- **sum%** — кумулятивный процент
- **cum** — время с учетом всех вызовов (cumulative)
- **cum%** — процент cumulative времени

**Вывод:** `crypto/blowfish.(*Cipher).Encrypt` потребляет 45% времени — это основной bottleneck.

### Что искать

✅ **Нормально:**
- Время распределено по многим функциям
- Горячие функции — ожидаемые (crypto, I/O)

❌ **Проблемы:**
- Одна функция > 30% времени (если это не криптография)
- Неожиданные функции в топе (runtime.newobject, runtime.memmove)
- Глубокие вызовы через interface{} (dynamic dispatch)

### Команды в pprof CLI

```bash
# Top 20 функций по cumulative time
(pprof) top20 -cum

# Детали функции
(pprof) list BlowfishEncrypt

# Граф вызовов
(pprof) web

# Выход
(pprof) quit
```

---

## Memory Profiling

### Что измеряет

Memory profiling показывает, **где программа аллоцирует память в heap**.

### Как запустить

```bash
# Вручную
go test -memprofile=mem.prof -bench=. ./internal/crypto
go tool pprof -http=:8080 mem.prof

# Через helper скрипт
./scripts/profile.sh mem ./internal/crypto
```

### Режимы анализа

**1. alloc_space** — количество аллокаций (по умолчанию)
```bash
go tool pprof -alloc_space mem.prof
```

**2. inuse_space** — используемая память (для memory leaks)
```bash
go tool pprof -inuse_space mem.prof
```

### Пример анализа

**Top view (alloc_space):**
```
Showing nodes accounting for 512MB, 80% of 640MB total
      flat  flat%   sum%        cum   cum%
     256MB 40.00% 40.00%      256MB 40.00%  make([]byte, size)
     128MB 20.00% 60.00%      128MB 20.00%  runtime.makeslice
     128MB 20.00% 80.00%      384MB 60.00%  la2go/internal/login.Client.handlePacket
```

**Вывод:** `make([]byte, size)` создает 256MB аллокаций — нужен sync.Pool.

### Что искать

✅ **Хорошо:**
- Малое количество аллокаций в горячих путях
- Аллокации в init/setup функциях (не в loops)

❌ **Плохо:**
- Аллокации в loops (> 10MB)
- Неожиданные аллокации (interface{} boxing, string concat)
- Memory leaks (inuse_space растет без остановки)

### Проверка memory leaks

```bash
# Запустить тест 2 раза и сравнить inuse_space
go test -memprofile=mem1.prof -bench=. ./internal/login
go test -memprofile=mem2.prof -bench=. ./internal/login

# Если inuse_space в mem2.prof значительно больше → leak
go tool pprof -inuse_space mem1.prof
go tool pprof -inuse_space mem2.prof
```

---

## Escape Analysis

### Что измеряет

Escape analysis показывает, **какие переменные "убегают" из стека в heap**.

**Почему это важно:**
- Stack allocations — быстрые (bump pointer)
- Heap allocations — медленные (GC overhead)

### Как запустить

```bash
# Вручную
go build -gcflags="-m -m" ./internal/crypto 2>&1 | grep "escapes to heap"

# Через helper скрипт
./scripts/profile.sh escape ./internal/crypto
```

### Пример вывода

```
./internal/crypto/blowfish.go:45:10: data escapes to heap:
./internal/crypto/blowfish.go:45:10:   flow: ~r0 = &data:
./internal/crypto/blowfish.go:45:10:     from data (spill) at ./internal/crypto/blowfish.go:45:10
./internal/crypto/blowfish.go:45:10:     from ~r0 = <N> (assign-pair) at ./internal/crypto/blowfish.go:45:3

./internal/protocol/packet.go:21:10: buf escapes to heap:
./internal/protocol/packet.go:21:10:   flow: {heap} = buf:
./internal/protocol/packet.go:21:10:     from buf (interface-converted) at ./internal/protocol/packet.go:22:20
```

**Расшифровка:**
- `data escapes to heap` — переменная `data` аллоцирована в heap
- `flow: ~r0 = &data` — возвращаемый pointer на `data` → heap
- `interface-converted` — конвертация в `interface{}` → heap

### Типичные причины escape

1. **Возврат pointer на локальную переменную**
   ```go
   // ❌ Escapes to heap
   func bad() *int {
       x := 42
       return &x  // Pointer на stack переменную → heap
   }

   // ✅ Остается на stack
   func good() int {
       x := 42
       return x  // Возвращаем значение, не pointer
   }
   ```

2. **Interface{} conversion**
   ```go
   // ❌ Escapes to heap
   func bad(data []byte) {
       log.Printf("data: %v", data)  // []byte → interface{} → heap
   }

   // ✅ Остается на stack (если возможно)
   func good(data []byte) {
       // Избегаем interface{} в горячих путях
   }
   ```

3. **Слишком большие структуры**
   ```go
   // ❌ Escapes to heap (структура > 64KB)
   type BigStruct struct {
       data [100000]byte
   }

   func bad() BigStruct {
       return BigStruct{}  // Слишком большая для stack → heap
   }

   // ✅ Используем pointer для больших структур
   func good() *BigStruct {
       return &BigStruct{}  // Явно в heap (но 1 аллокация)
   }
   ```

4. **Slice append без capacity**
   ```go
   // ❌ Много аллокаций (slice grows)
   func bad() []byte {
       var buf []byte
       for i := range 100 {
           buf = append(buf, byte(i))  // Reslice → heap
       }
       return buf
   }

   // ✅ Preallocate capacity
   func good() []byte {
       buf := make([]byte, 0, 100)  // 1 аллокация
       for i := range 100 {
           buf = append(buf, byte(i))
       }
       return buf
   }
   ```

### Как исправить

1. **Передавать pointer вместо value** (если структура > 128 bytes)
2. **Использовать value receiver** (если структура < 128 bytes)
3. **Избегать interface{}** в горячих путях
4. **Preallocate slices** с известным capacity

---

## Block Profiling

### Что измеряет

Block profiling показывает, **где горутины блокируются** (channel operations, select, sync primitives).

### Как запустить

```bash
# Только для Concurrent бенчмарков
go test -blockprofile=block.prof -bench=Concurrent ./internal/login
go tool pprof -http=:8080 block.prof

# Через helper скрипт
./scripts/profile.sh block ./internal/login
```

### Что искать

❌ **Проблемы:**
- Горутины блокируются на channel send/receive
- Долгие ожидания на `sync.Cond.Wait`
- Unbuffered channels в горячих путях

**Решения:**
- Использовать buffered channels
- Уменьшить granularity locks
- Рефакторинг на lock-free структуры (sync.Map, atomic)

---

## Mutex Profiling

### Что измеряет

Mutex profiling показывает, **где происходит lock contention** (конкуренция за mutex).

### Как запустить

```bash
# Только для Concurrent бенчмарков
go test -mutexprofile=mutex.prof -bench=Concurrent ./internal/gameserver
go tool pprof -http=:8080 mutex.prof

# Через helper скрипт
./scripts/profile.sh mutex ./internal/gameserver
```

### Пример анализа

**Top view:**
```
Showing nodes accounting for 500ms, 90% of 555ms total
      flat  flat%   sum%        cum   cum%
     300ms 54.05% 54.05%      300ms 54.05%  sync.(*RWMutex).RLock
     200ms 36.04% 90.09%      200ms 36.04%  sync.(*Mutex).Lock
```

**Вывод:** Много времени тратится на RLock — возможно, слишком частые reads или длинная critical section.

### Что искать

❌ **Проблемы:**
- Lock contention > 10% времени
- Множественные locks в цикле
- Read locks на immutable данные

**Решения:**
1. **Использовать sync.Map** вместо `map + sync.RWMutex`
2. **Batch операции** (один lock вместо N)
3. **Copy-on-write** для read-heavy workloads
4. **atomic операции** вместо mutex (где возможно)

---

## Helper скрипты

### profile.sh

```bash
# CPU profiling
./scripts/profile.sh cpu ./internal/crypto

# Memory profiling
./scripts/profile.sh mem ./internal/login

# Escape analysis
./scripts/profile.sh escape ./internal/protocol

# Все профили сразу
./scripts/profile.sh all ./internal/crypto
```

### Результаты сохраняются в `profiles/`

```
profiles/
├── cpu_20260209_120000.prof
├── mem_20260209_120100.prof
├── block_20260209_120200.prof
├── mutex_20260209_120300.prof
└── escape_20260209_120400.log
```

---

## Workflow оптимизации

### Шаг 1: Baseline benchmark

```bash
go test -bench=BenchmarkBlowfish -benchmem -count=10 ./internal/crypto > baseline.txt
```

### Шаг 2: CPU profiling

```bash
./scripts/profile.sh cpu ./internal/crypto
```

Найти функцию-bottleneck (например, `Encrypt` занимает 45% времени).

### Шаг 3: Memory profiling

```bash
./scripts/profile.sh mem ./internal/crypto
```

Найти неожиданные аллокации (например, `make([]byte)` в loop).

### Шаг 4: Escape analysis

```bash
./scripts/profile.sh escape ./internal/crypto
```

Найти переменные, убегающие в heap.

### Шаг 5: Применить оптимизацию

```go
// ДО: 256 B/op, 1 allocs/op
func encryptBad(data []byte) {
    buf := make([]byte, 256)  // Escapes to heap
    // ...
}

// ПОСЛЕ: 0 B/op, 0 allocs/op
var bufPool = sync.Pool{
    New: func() any { return make([]byte, 256) },
}

func encryptGood(data []byte) {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // ...
}
```

### Шаг 6: Re-benchmark

```bash
go test -bench=BenchmarkBlowfish -benchmem -count=10 ./internal/crypto > optimized.txt
benchstat baseline.txt optimized.txt
```

**Ожидаемый результат:**
```
name                old time/op    new time/op    delta
BlowfishEncrypt-14   1.23µs ± 2%    0.85µs ± 3%  -30.89%  (p=0.000)

name                old alloc/op   new alloc/op   delta
BlowfishEncrypt-14    256B ± 0%       0B ± 0%  -100.00%  (p=0.000)
```

### Шаг 7: Документировать в OPTIMIZATION_LOG.md

```markdown
## 2026-02-09: Blowfish Encrypt — sync.Pool

### Проблема
CPU profile: Encrypt занимает 45% времени.
Memory profile: 256 B/op аллокации в loop.
Escape analysis: buf escapes to heap.

### Решение
Использовать sync.Pool для переиспользования буферов.

### Результат
- Время: -31% (1.23µs → 0.85µs)
- Аллокации: -100% (256B → 0B)
- Коммит: abc123

### Профили
- CPU: profiles/cpu_20260209_120000.prof
- Memory: profiles/mem_20260209_120100.prof
```

---

## Частые проблемы и решения

### Проблема 1: Слишком много heap allocations

**Симптом:**
```
BenchmarkEncrypt-14    100000   10234 ns/op   512 B/op   4 allocs/op
                                               ^^^^^^^^^  ^^^^^^^^^^^^^
```

**Диагностика:**
```bash
./scripts/profile.sh mem ./internal/crypto
```

**Решение:**
- Использовать sync.Pool
- Preallocate slices
- Избегать interface{} boxing

---

### Проблема 2: Lock contention

**Симптом:**
```
BenchmarkValidate_Concurrent-14    50000   45678 ns/op
```
(слишком медленно для простой операции)

**Диагностика:**
```bash
./scripts/profile.sh mutex ./internal/login
```

**Решение:**
- Использовать sync.Map вместо map + RWMutex
- Batch операции (один lock вместо N)
- Atomic операции

---

### Проблема 3: Горячий путь с interface{}

**Симптом:**
```
escape analysis: data escapes to heap (interface-converted)
```

**Решение:**
```go
// ❌ ПЛОХО
func logData(data []byte) {
    log.Printf("data: %v", data)  // interface{} → heap
}

// ✅ ХОРОШО
func logData(data []byte) {
    if debug {  // Условное логирование
        log.Printf("data: %v", data)
    }
}

// ИЛИ: специализированная функция без interface{}
func processData(data []byte) {
    // Прямая работа с []byte, без interface{}
}
```

---

## Дополнительные ресурсы

- [Go pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Escape Analysis Guide](https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html)
- [Benchmarking Guide](./BENCHMARKING_GUIDE.md)
- [Optimization Log](./OPTIMIZATION_LOG.md)

---

## Следующие шаги

1. Запустите `./scripts/profile.sh all ./internal/crypto` для baseline профилей
2. Используйте Flame Graph для визуализации bottlenecks
3. Применяйте оптимизации итеративно (одна за раз)
4. Всегда документируйте результаты в OPTIMIZATION_LOG.md
