# Benchmarking Guide — la2go

Руководство по написанию и запуску бенчмарков для la2go.

## Содержание

- [Введение](#введение)
- [Структура бенчмарков](#структура-бенчмарков)
- [Написание бенчмарков](#написание-бенчмарков)
- [Запуск бенчмарков](#запуск-бенчмарков)
- [Интерпретация результатов](#интерпретация-результатов)
- [Best Practices](#best-practices)

---

## Введение

Бенчмарки в la2go используются для:

1. **Измерения производительности** критичных операций (криптография, I/O, concurrency)
2. **Выявления узких мест** (bottlenecks) в коде
3. **Предотвращения регрессий** производительности в CI/CD
4. **Валидации оптимизаций** (сравнение до/после)

**Философия:** Мы НЕ оптимизируем то, что НЕ измерили. Бенчмарки — основа для принятия решений.

---

## Структура бенчмарков

Бенчмарки находятся рядом с тестируемым кодом в `*_bench_test.go` файлах:

```
internal/
├── crypto/
│   ├── blowfish.go
│   ├── blowfish_test.go
│   ├── blowfish_bench_test.go      ← Benchmarks
│   ├── rsa.go
│   ├── rsa_test.go
│   └── rsa_bench_test.go           ← Benchmarks
├── login/
│   ├── bufpool.go
│   ├── bufpool_test.go
│   ├── bufpool_bench_test.go       ← Benchmarks
│   ├── session_manager.go
│   └── session_manager_bench_test.go ← Benchmarks
└── gameserver/
    ├── table.go
    ├── table_test.go
    └── table_bench_test.go         ← Benchmarks
```

**Принцип:** Один файл с бенчмарками на один Go файл с кодом.

---

## Написание бенчмарков

### Базовый шаблон

```go
func BenchmarkFunctionName(b *testing.B) {
    b.ReportAllocs()  // ВСЕГДА включать для отслеживания аллокаций

    // Setup (вне loop) — подготовка данных
    data := make([]byte, 256)
    cipher, _ := NewBlowfishCipher(key)

    b.ResetTimer()  // Сброс таймера после setup
    for range b.N {
        // Код для бенчмарка (измеряемая часть)
        cipher.Encrypt(data, 0, len(data))
    }
}
```

### Обязательные элементы

1. **`b.ReportAllocs()`** — отслеживание heap allocations
2. **`b.ResetTimer()`** — сброс таймера после setup
3. **Setup вне loop** — подготовка данных НЕ должна входить в измерение

### Бенчмарк с разными размерами

```go
func BenchmarkEncrypt_Sizes(b *testing.B) {
    sizes := []int{64, 128, 256, 512, 1024, 2048}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            b.ReportAllocs()

            data := make([]byte, size)
            cipher, _ := NewBlowfishCipher(key)

            b.SetBytes(int64(size))  // Для расчета throughput (MB/s)

            b.ResetTimer()
            for range b.N {
                cipher.Encrypt(data, 0, size)
            }
        })
    }
}
```

**Результат:**
```
BenchmarkEncrypt_Sizes/size=64-14      1000000    1234 ns/op   51.8 MB/s
BenchmarkEncrypt_Sizes/size=256-14      500000    2456 ns/op  104.2 MB/s
```

### Параллельный бенчмарк

```go
func BenchmarkSessionManager_Validate_Concurrent(b *testing.B) {
    b.ReportAllocs()

    sm := NewSessionManager()
    // Setup: заполняем SessionManager
    for i := range 1000 {
        sm.Store(fmt.Sprintf("user_%d", i), key, nil)
    }

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            sm.Validate("user_500", key, false)
        }
    })
}
```

**Результат покажет производительность под параллельной нагрузкой.**

---

## Запуск бенчмарков

### Вручную

```bash
# Запустить все бенчмарки в пакете
go test -bench=. -benchmem ./internal/crypto

# Запустить конкретный бенчмарк
go test -bench=BenchmarkBlowfishEncrypt$ -benchmem ./internal/crypto

# Запустить с указанным временем
go test -bench=. -benchmem -benchtime=2s ./internal/crypto

# Запустить N раз для стабильности
go test -bench=. -benchmem -count=10 ./internal/crypto
```

### Через helper скрипты

```bash
# Запустить все бенчмарки (crypto, login, gameserver)
./scripts/bench.sh all

# Только crypto бенчмарки
./scripts/bench.sh crypto

# Быстрый прогон (100ms, count=1)
./scripts/bench.sh quick

# Полный прогон с профилированием
./scripts/bench.sh full

# Сравнить с baseline
./scripts/bench.sh compare
```

### В CI/CD

GitHub Actions автоматически запускает бенчмарки на каждый PR:

- `.github/workflows/benchmarks.yml`
- Результаты публикуются в комментарии к PR
- Сравнение с baseline (main branch)

---

## Интерпретация результатов

### Пример вывода

```
BenchmarkBlowfishEncrypt-14        139861       864.1 ns/op       0 B/op       0 allocs/op
```

**Расшифровка:**
- `BenchmarkBlowfishEncrypt` — имя бенчмарка
- `-14` — количество GOMAXPROCS (CPU cores)
- `139861` — количество итераций за benchtime (по умолчанию 1s)
- `864.1 ns/op` — время одной операции (наносекунды)
- `0 B/op` — heap allocations на операцию (байты)
- `0 allocs/op` — количество аллокаций на операцию

### Что искать

✅ **Хорошо:**
- `0 B/op, 0 allocs/op` — нет аллокаций в heap
- Стабильные результаты при повторных запусках

❌ **Плохо:**
- Высокие аллокации (> 100 B/op) в горячих путях
- Нестабильные результаты (разброс > 10%)

### Сравнение с помощью benchstat

```bash
# Baseline (до оптимизации)
go test -bench=. -benchmem -count=10 ./internal/crypto > baseline.txt

# После оптимизации
go test -bench=. -benchmem -count=10 ./internal/crypto > optimized.txt

# Сравнение
benchstat baseline.txt optimized.txt
```

**Пример вывода:**
```
name                    old time/op    new time/op    delta
BlowfishEncrypt-14       1.23µs ± 2%    0.85µs ± 3%  -30.89%  (p=0.000 n=10+10)

name                    old alloc/op   new alloc/op   delta
BlowfishEncrypt-14        128B ± 0%       64B ± 0%  -50.00%  (p=0.000 n=10+10)

name                    old allocs/op  new allocs/op  delta
BlowfishEncrypt-14        2.00 ± 0%      1.00 ± 0%  -50.00%  (p=0.000 n=10+10)
```

**Расшифровка:**
- **delta** — процентное изменение (+ медленнее, - быстрее)
- **p-value** — статистическая значимость (< 0.05 = значимое изменение)
- **n=10+10** — количество измерений в каждой группе

---

## Best Practices

### 1. Всегда используйте `b.ReportAllocs()`

Аллокации — один из главных источников проблем производительности.

```go
func BenchmarkBad(b *testing.B) {
    // ❌ БЕЗ b.ReportAllocs() — не видим аллокаций
    for range b.N {
        _ = make([]byte, 256)
    }
}

func BenchmarkGood(b *testing.B) {
    b.ReportAllocs()  // ✅ Видим: 256 B/op, 1 allocs/op
    for range b.N {
        _ = make([]byte, 256)
    }
}
```

### 2. Избегайте compiler optimizations

Компилятор может оптимизировать "мертвый код":

```go
func BenchmarkBad(b *testing.B) {
    b.ReportAllocs()
    for range b.N {
        _ = expensiveComputation()  // ❌ Может быть оптимизировано
    }
}

func BenchmarkGood(b *testing.B) {
    b.ReportAllocs()
    var result int
    for range b.N {
        result = expensiveComputation()  // ✅ Сохраняем результат
    }
    _ = result  // Используем после loop
}
```

### 3. Запускайте несколько раз

Один прогон может быть нестабильным:

```bash
# Плохо: один прогон
go test -bench=. ./internal/crypto

# Хорошо: 10 прогонов
go test -bench=. -count=10 ./internal/crypto
```

### 4. Используйте `b.SetBytes()` для throughput

```go
func BenchmarkEncrypt(b *testing.B) {
    b.ReportAllocs()

    data := make([]byte, 1024)
    b.SetBytes(1024)  // Указываем обрабатываемый объем данных

    b.ResetTimer()
    for range b.N {
        encrypt(data)
    }
}
```

**Результат:**
```
BenchmarkEncrypt-14    100000   10234 ns/op   100.0 MB/s   0 B/op   0 allocs/op
                                               ^^^^^^^^^^
```

### 5. Бенчмаркайте разные сценарии

```go
// Best case
BenchmarkSessionManager_Validate_Empty

// Average case
BenchmarkSessionManager_Validate_1000Accounts

// Worst case
BenchmarkSessionManager_Validate_50000Accounts

// Concurrent load
BenchmarkSessionManager_Validate_Concurrent
```

### 6. Документируйте результаты

После каждой оптимизации сохраняйте результаты в `OPTIMIZATION_LOG.md`:

```markdown
## 2026-02-09: Blowfish Encrypt Optimization

### Проблема
Blowfish encryption занимает 1.2µs на пакет.

### Решение
Убрали bounds checks через `_ = data[:size]` hint.

### Результат
- Время: -30% (1.2µs → 0.85µs)
- Аллокации: без изменений (0 allocs)
- Коммит: abc123
```

---

## Частые ошибки

### 1. Setup внутри loop

```go
// ❌ ПЛОХО
func BenchmarkBad(b *testing.B) {
    b.ReportAllocs()
    for range b.N {
        data := make([]byte, 256)  // Аллокация входит в измерение!
        encrypt(data)
    }
}

// ✅ ХОРОШО
func BenchmarkGood(b *testing.B) {
    b.ReportAllocs()
    data := make([]byte, 256)  // Аллокация ВНЕ loop

    b.ResetTimer()
    for range b.N {
        encrypt(data)
    }
}
```

### 2. Не сбрасывать таймер после setup

```go
// ❌ ПЛОХО
func BenchmarkBad(b *testing.B) {
    b.ReportAllocs()
    // Долгая подготовка данных
    data := generateLargeDataset()  // Входит в измерение!

    for range b.N {
        process(data)
    }
}

// ✅ ХОРОШО
func BenchmarkGood(b *testing.B) {
    b.ReportAllocs()
    data := generateLargeDataset()

    b.ResetTimer()  // Сбрасываем таймер!
    for range b.N {
        process(data)
    }
}
```

### 3. Игнорировать аллокации

```go
// ❌ ПЛОХО — 256 B/op, 1 allocs/op в горячем пути
func encryptBad(data []byte) {
    buf := make([]byte, 256)  // Аллокация на каждый вызов
    // ...
}

// ✅ ХОРОШО — 0 B/op, 0 allocs/op
var bufPool = sync.Pool{
    New: func() any { return make([]byte, 256) },
}

func encryptGood(data []byte) {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // ...
}
```

---

## Дополнительные ресурсы

- [Go Testing Package — Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [How to Write Benchmarks in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [benchstat command](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Profiling Guide](./PROFILING_GUIDE.md) — как использовать pprof
- [Optimization Log](./OPTIMIZATION_LOG.md) — история оптимизаций

---

## Следующие шаги

1. Прочитайте [PROFILING_GUIDE.md](./PROFILING_GUIDE.md) для глубокого анализа производительности
2. Запустите `./scripts/bench.sh quick` для быстрой проверки
3. При оптимизации всегда создавайте baseline для сравнения
4. Документируйте результаты в [OPTIMIZATION_LOG.md](./OPTIMIZATION_LOG.md)
