# Benchmarking & Profiling — la2go

Руководство по запуску бенчмарков и профилированию производительности la2go.

## Быстрый старт

Все задачи доступны через [Task](https://taskfile.dev) (аналог Makefile):

```bash
# Показать все доступные задачи
task --list

# Быстрый прогон всех бенчмарков (100ms, 1 итерация)
task bench-quick

# Полный прогон с профилированием (1s, 10 итераций)
task bench-full

# CPU профилирование crypto пакета
task profile-cpu

# CPU профилирование login пакета
PACKAGE=./internal/login task profile-cpu
```

## Benchmark задачи

### `task bench`
Запустить все бенчмарки (crypto, login, gameserver) с дефолтными настройками (1s, 10 итераций).

```bash
task bench
```

Результаты сохраняются в `benchmarks/all_YYYYMMDD_HHMMSS.txt`.

### `task bench-crypto`
Запустить только crypto бенчмарки.

```bash
task bench-crypto
```

Результаты: `benchmarks/crypto_YYYYMMDD_HHMMSS.txt`.

### `task bench-login`
Запустить только login бенчмарки.

```bash
task bench-login
```

Результаты: `benchmarks/login_YYYYMMDD_HHMMSS.txt`.

### `task bench-gameserver`
Запустить только gameserver бенчмарки.

```bash
task bench-gameserver
```

Результаты: `benchmarks/gameserver_YYYYMMDD_HHMMSS.txt`.

### `task bench-quick`
Быстрый прогон всех бенчмарков: 100ms benchtime, 1 итерация.

**Использование:** для быстрой проверки после изменений (CI precheck).

```bash
task bench-quick
```

Результаты: `benchmarks/quick_YYYYMMDD_HHMMSS.txt`.

### `task bench-full`
Полный прогон с CPU и Memory профилированием: 1s benchtime, 10 итераций.

**Использование:** перед оптимизацией для получения baseline + профилей.

```bash
task bench-full
```

**Результаты:**
- Бенчмарки: `benchmarks/full_YYYYMMDD_HHMMSS.txt`
- CPU профили: `profiles/crypto_cpu_YYYYMMDD_HHMMSS.prof`, `profiles/login_cpu_YYYYMMDD_HHMMSS.prof`, etc.
- Memory профили: `profiles/crypto_mem_YYYYMMDD_HHMMSS.prof`, `profiles/login_mem_YYYYMMDD_HHMMSS.prof`, etc.

**Просмотр профилей:**

```bash
go tool pprof -http=:8080 profiles/crypto_cpu_20260209_123045.prof
```

### `task bench-ci`
CI режим: 500ms benchtime, 5 итераций, без интерактивного вывода.

**Использование:** в CI/CD pipeline для проверки регрессий производительности.

```bash
task bench-ci
```

Результаты: `benchmarks/ci_YYYYMMDD_HHMMSS.txt`.

### `task bench-compare`
Сравнить текущие бенчмарки с baseline с помощью `benchstat`.

**Workflow:**

```bash
# 1. Создать baseline (до оптимизации)
task bench
mv benchmarks/all_*.txt benchmarks/baseline.txt

# 2. Внести изменения в код

# 3. Сравнить с baseline
task bench-compare
```

**Пример вывода:**

```
name                  old time/op    new time/op    delta
Blowfish/Encrypt-8       156ns ± 2%     142ns ± 1%   -8.97%  (p=0.000 n=10+10)
Blowfish/Decrypt-8       158ns ± 1%     144ns ± 2%   -8.86%  (p=0.000 n=10+10)

name                  old alloc/op   new alloc/op   delta
SessionKey/New-8        48.0B ± 0%     32.0B ± 0%  -33.33%  (p=0.000 n=10+10)

name                  old allocs/op  new allocs/op  delta
SessionKey/New-8         2.00 ± 0%      1.00 ± 0%  -50.00%  (p=0.000 n=10+10)
```

**Обновить baseline:**

```bash
# Если текущие результаты стали новым baseline
mv benchmarks/current_YYYYMMDD_HHMMSS.txt benchmarks/baseline.txt
```

## Profiling задачи

Все profiling задачи используют пакет по умолчанию `./internal/crypto`. Для профилирования другого пакета используйте переменную `PACKAGE`:

```bash
PACKAGE=./internal/login task profile-cpu
```

### `task profile-cpu`
CPU профилирование — где процессор тратит больше всего времени.

```bash
task profile-cpu

# Другой пакет
PACKAGE=./internal/login task profile-cpu
```

**Результат:** `profiles/cpu_YYYYMMDD_HHMMSS.prof`

Автоматически открывает `pprof` в браузере на `:8080`.

**Что искать:**
- Горячие функции (high % flat, cumulative)
- Неожиданно медленные операции (alloc, syscall, lock)

### `task profile-mem`
Memory профилирование — где происходят heap аллокации.

```bash
task profile-mem
```

**Результат:** `profiles/mem_YYYYMMDD_HHMMSS.prof`

Автоматически открывает `pprof` в браузере на `:8080`.

**Что искать:**
- Функции с большим числом аллокаций
- Непредвиденные heap escapes
- Возможности для `sync.Pool` или buffer reuse

**Переключить вид на inuse_space:**

```bash
go tool pprof -http=:8080 -inuse_space profiles/mem_YYYYMMDD_HHMMSS.prof
```

### `task profile-block`
Block профилирование — где горутины блокируются (channels, I/O).

**Использование:** для выявления узких мест concurrency.

```bash
task profile-block
```

**Результат:** `profiles/block_YYYYMMDD_HHMMSS.prof`

**Что искать:**
- Длительные channel operations
- I/O блокировки (syscall)
- Неоптимальная работа с channels (buffered vs unbuffered)

### `task profile-mutex`
Mutex профилирование — lock contention (конкуренция за мьютексы).

**Использование:** для выявления проблем с блокировками.

```bash
task profile-mutex
```

**Результат:** `profiles/mutex_YYYYMMDD_HHMMSS.prof`

**Что искать:**
- Высокая contention на мьютексах
- Длительное удержание lock
- Возможности для RWMutex, atomic, lock-free структур

### `task profile-escape`
Escape analysis — какие переменные убегают в heap (должны быть на stack).

**Использование:** для оптимизации аллокаций.

```bash
task profile-escape
```

**Результаты:**
- Полный лог: `profiles/escape_YYYYMMDD_HHMMSS.log`
- Только "escapes to heap": `profiles/escape_critical_YYYYMMDD_HHMMSS.txt`

**Что искать:**
- Неожиданные "escapes to heap" для простых типов
- Interface conversions
- Closure captures

**Пример критичного escape:**

```
./internal/crypto/blowfish.go:42:6: b escapes to heap:
./internal/crypto/blowfish.go:42:6:   flow: ~r1 = &b:
./internal/crypto/blowfish.go:42:6:     from &b (address-of) at ./internal/crypto/blowfish.go:50:9
./internal/crypto/blowfish.go:42:6:     from return &b (return) at ./internal/crypto/blowfish.go:50:2
```

**Исправление:** изменить API чтобы принимать `*Blowfish` вместо возвращать `*Blowfish`.

### `task profile-all`
Запустить все профили (cpu, mem, block, mutex) для пакета.

**Использование:** для комплексного анализа производительности.

```bash
task profile-all

# Другой пакет
PACKAGE=./internal/login task profile-all
```

**Результаты:** `profiles/cpu_YYYYMMDD_HHMMSS.prof`, `profiles/mem_YYYYMMDD_HHMMSS.prof`, etc.

### `task profile-compare`
Сравнить два набора бенчмарков с помощью `benchstat`.

**Workflow:**

```bash
# 1. Сохранить baseline
./scripts/profile.sh bench ./internal/crypto
mv profiles/bench_*.txt benchmarks/baseline.txt

# 2. Оптимизировать код

# 3. Сохранить optimized
./scripts/profile.sh bench ./internal/crypto
mv profiles/bench_*.txt benchmarks/optimized.txt

# 4. Сравнить
BASELINE=benchmarks/baseline.txt OPTIMIZED=benchmarks/optimized.txt task profile-compare
```

**Альтернатива (через скрипт напрямую):**

```bash
./scripts/profile.sh compare benchmarks/baseline.txt benchmarks/optimized.txt
```

## Прямой вызов скриптов

Если нужен более гибкий контроль, можно вызывать скрипты напрямую:

### `scripts/bench.sh`

```bash
# Usage
./scripts/bench.sh [all|crypto|login|gameserver|quick|full|ci|compare]

# Examples
./scripts/bench.sh all
./scripts/bench.sh crypto
./scripts/bench.sh quick
./scripts/bench.sh full
./scripts/bench.sh compare
```

### `scripts/profile.sh`

```bash
# Usage
./scripts/profile.sh [cpu|mem|block|mutex|escape|all|bench|compare] [PACKAGE]

# Examples
./scripts/profile.sh cpu ./internal/crypto
./scripts/profile.sh mem ./internal/login
./scripts/profile.sh escape ./internal/protocol
./scripts/profile.sh all ./internal/crypto
./scripts/profile.sh bench ./...
./scripts/profile.sh compare baseline.txt optimized.txt
```

## Best Practices

### 1. Baseline перед оптимизацией

```bash
task bench-full
mv benchmarks/full_*.txt benchmarks/baseline.txt
mv profiles/*_*.prof profiles/baseline/
```

### 2. Изолируйте оптимизации

Оптимизируйте **одну** вещь за раз, чтобы понять эффект:

```bash
# До оптимизации
task bench > before.txt

# Оптимизация X

# После оптимизации
task bench > after.txt

# Сравнение
./scripts/profile.sh compare before.txt after.txt
```

### 3. Проверяйте корректность после оптимизации

```bash
# После оптимизации ВСЕГДА запускайте тесты
task test-all

# Проверьте что поведение не изменилось
task test-coverage
```

### 4. CI Integration

Добавьте в `.github/workflows/bench.yml`:

```yaml
- name: Run benchmarks
  run: task bench-ci

- name: Check for regressions
  run: |
    # Скачать baseline из предыдущего релиза
    # Сравнить с текущими результатами
    # Fail если деградация > 10%
```

### 5. Избегайте micro-optimizations

- **НЕ** оптимизируйте код который не в hotpath
- **НЕ** оптимизируйте без профилирования (guess)
- **НЕ** жертвуйте читаемостью ради 1-2% производительности

**Приоритизируйте:**
1. Correctness (корректность)
2. Readability (читаемость)
3. Performance (производительность)

## Интерпретация результатов

### Benchmark output

```
BenchmarkBlowfishEncrypt-8    10000000    156 ns/op    48 B/op    2 allocs/op
```

- `BenchmarkBlowfishEncrypt-8` — имя бенчмарка, `-8` = GOMAXPROCS
- `10000000` — число итераций (автоматически подстраивается для достижения benchtime)
- `156 ns/op` — время на операцию
- `48 B/op` — heap аллокаций на операцию
- `2 allocs/op` — число аллокаций на операцию

**Хорошие цели:**
- **Crypto:** < 200 ns/op, 0 allocs/op (после оптимизации buffer reuse)
- **Packet write:** < 100 ns/op, 0 allocs/op (sync.Pool)
- **Packet read:** < 50 ns/op, 0 allocs/op (buffer from pool)

### benchstat output

```
name                  old time/op    new time/op    delta
Blowfish/Encrypt-8       156ns ± 2%     142ns ± 1%   -8.97%  (p=0.000 n=10+10)
```

- `old time/op` — baseline
- `new time/op` — optimized
- `delta` — изменение в %
- `(p=0.000 n=10+10)` — статистическая значимость (p < 0.05 = значимо)

**Интерпретация:**
- `-8.97%` — улучшение на ~9%
- `p=0.000` — **статистически значимо** (не случайность)
- `n=10+10` — 10 итераций baseline, 10 итераций optimized

**Незначимые изменения (~):**

```
name                  old time/op    new time/op    delta
SessionKey/New-8        48.0ns ± 1%    49.0ns ± 2%   ~     (p=0.100 n=10+10)
```

`~` = изменение **статистически незначимо** (шум).

## Troubleshooting

### benchstat не найден

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Профили не открываются в браузере

Запустить вручную:

```bash
go tool pprof -http=:8080 profiles/cpu_20260209_123045.prof
```

### Недостаточно samples в профиле

Увеличить benchtime:

```bash
go test -bench=. -benchtime=10s -cpuprofile=cpu.prof ./internal/crypto
```

### Escape analysis слишком verbose

Фильтровать только критичное:

```bash
go build -gcflags="-m -m" ./internal/crypto 2>&1 | grep "escapes to heap"
```

## Дополнительные ресурсы

- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking Go Code](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [benchstat Guide](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Escape Analysis](https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html)
