# Архитектура тестирования la2go

## Обзор

Проект использует трёхуровневую архитектуру тестирования по best practices (etcd, CockroachDB, Traefik):

```
la2go/
├── internal/*/         # Unit тесты (*_test.go)
├── internal/testutil/  # Общие test helpers
├── tests/integration/  # Integration тесты
└── tests/e2e/          # End-to-end тесты
```

## 1. Unit тесты (`internal/*_test.go`)

**Расположение:** Рядом с кодом в пакете
**Скорость:** Быстрые (~миллисекунды)
**Зависимости:** Нет внешних зависимостей (используют mocks, net.Pipe)
**Запуск:** `make test-unit` или `go test -short ./internal/...`

### Примеры:
- `internal/crypto/blowfish_test.go` — тесты шифрования
- `internal/login/session_manager_test.go` — тесты управления сессиями
- `internal/protocol/packet_test.go` — тесты протокола

## 2. Integration тесты (`tests/integration/`)

**Скорость:** Средние (~секунды)
**Зависимости:** Требуют PostgreSQL (env var `DB_ADDR`)
**Запуск:** `make test-integration` (skip если нет DB_ADDR)

### Suite структура (testify/suite):
```go
type DatabaseSuite struct {
    IntegrationSuite
}

func (s *DatabaseSuite) SetupSuite() {
    // Один раз перед всеми тестами
}

func (s *DatabaseSuite) SetupTest() {
    // Перед каждым тестом (cleanup data)
}
```

### Текущие suites:
- `DatabaseSuite` — CRUD операции с БД, concurrent tests
- `LoginServerSuite` — TCP подключения клиентов, Init пакеты
- `GSListenerSuite` — GS↔LS relay, InitLS пакеты

## 3. E2E тесты (`tests/e2e/`)

**Скорость:** Медленные (~секунды-минуты)
**Зависимости:** PostgreSQL + все компоненты (LS + GS + клиент)
**Запуск:** `make test-e2e`

### Статус:
- ✅ Структура готова
- ⏳ Ожидает реализации Phase 4+ (GameServer)

## Test Helpers (`internal/testutil/`)

Централизованные утилиты для переиспользования в тестах:

### `fixtures.go`
Предварительно сгенерированные данные:
```go
testutil.Fixtures.RSAKey         // RSA-2048 ключ
testutil.Fixtures.RSAKey512      // RSA-512 ключ
testutil.Fixtures.BlowfishKey    // Blowfish ключ
testutil.Fixtures.SessionKey     // SessionKey
testutil.Fixtures.ValidAccount   // "testuser"
testutil.Fixtures.ValidPassword  // "testpass"
```

### `mocks.go`
In-memory имплементации:
```go
mockDB := testutil.NewMockDB()
mockDB.CreateAccount(ctx, "user", "hash", "127.0.0.1")
acc, _ := mockDB.GetAccount(ctx, "user")
```

### `netutil.go`
Сетевые утилиты:
```go
client, server := testutil.PipeConn(t)        // net.Pipe
listener, addr := testutil.ListenTCP(t)       // TCP listener на случайном порту
fakeAddr := testutil.TCPAddr("127.0.0.1:123") // Fake net.Addr
```

### `assertions.go`
L2 протокол assertions:
```go
testutil.AssertPacketOpcode(t, 0x00, packet)
testutil.AssertInt32LE(t, 12345, packet, offset)
testutil.AssertUTF16String(t, "hello", packet, offset)
testutil.AssertBytesEqual(t, expected, actual, "message")
```

### `protocol.go`
L2 пакеты builders:
```go
testutil.EncodeUTF16LE("hello")
testutil.MakeBlowFishKeyPacket(key)
testutil.MakeGameServerAuthPacket(serverID, hexID)
testutil.MakePlayerAuthRequestPacket(account, ...)
```

### `context.go`
Context helpers:
```go
ctx := testutil.ContextWithTimeout(t, 5*time.Second)
```

## Taskfile команды

```bash
# Список доступных команд
task --list

# Unit тесты (быстро, без БД)
task test-unit

# Integration + unit (skip integration если нет DB_ADDR)
task test

# Integration тесты (требуют DB_ADDR)
DB_ADDR="postgres://..." task test-integration

# E2E тесты
DB_ADDR="postgres://..." task test-e2e

# Все тесты
DB_ADDR="postgres://..." task test-all

# Все тесты с Docker тестовой БД
task test-with-db

# Coverage report
task test-coverage

# Управление test DB
task test-db-up
task test-db-down

# Дополнительные команды
task quick           # Быстрые unit тесты (без race detector)
task test-compile    # Компиляция тестов без запуска
task test-clean      # Очистка артефактов
task test-list       # Список всех тестов
```

## Environment Variables

### `DB_ADDR`
PostgreSQL connection string для integration/e2e тестов.

**Примеры:**
```bash
# Локальный PostgreSQL
DB_ADDR="postgres://la2go:password@localhost:5432/la2go_test?sslmode=disable"

# Docker контейнер
DB_ADDR="postgres://la2go:testpass@localhost:5433/la2go_test?sslmode=disable"
```

**Поведение:**
- Не задан → integration/e2e тесты skip'нуться
- Задан → тесты запустятся с реальной БД

## CI/CD Example (GitHub Actions)

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Install Task
        run: |
          sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
      - run: task test-unit

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17-alpine
        env:
          POSTGRES_USER: la2go
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: la2go_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Install Task
        run: |
          sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
      - run: task test-integration
        env:
          DB_ADDR: postgres://la2go:testpass@localhost:5432/la2go_test?sslmode=disable
```

## Best Practices применённые

1. ✅ **Разделение уровней:** unit / integration / e2e
2. ✅ **testify/suite:** Setup/Teardown hooks
3. ✅ **Environment-based skip:** гибче чем build tags
4. ✅ **Централизованные helpers:** нет дублирования
5. ✅ **t.Helper():** правильные stack traces
6. ✅ **testing.TB:** универсальность (T + B)
7. ✅ **t.Cleanup:** автоматическая очистка
8. ✅ **Race detector:** `-race` флаг
9. ✅ **Reproducible:** `-count=1` (без кэша)
10. ✅ **Makefile:** удобные команды

## Масштабируемость для Phase 4+ (GameServer)

При добавлении GameServer:

### Unit тесты:
```
internal/gameserver/
├── packet_handlers_test.go
├── world_test.go
└── character_test.go
```

### Integration тесты:
```go
// tests/integration/game_server_test.go
type GameServerSuite struct {
    IntegrationSuite
    server *gameserver.Server
}

func (s *GameServerSuite) TestCharacterSelection() {
    // ...
}
```

### E2E тесты:
```go
// tests/e2e/full_flow_test.go
func TestFullLoginFlow(t *testing.T) {
    // 1. Start LoginServer
    // 2. Start gslistener
    // 3. Start GameServer
    // 4. Client: Init → Auth → ServerList → PlayOk
    // 5. GS: PlayerAuthRequest → PlayerInGame
    // 6. Client → GS: CharSelection → EnterWorld
    // 7. Verify character in world
}
```

Структура готова к расширению — достаточно добавить новые файлы.

## Troubleshooting

### Integration тесты skip'нуты
Проверьте что `DB_ADDR` задан:
```bash
echo $DB_ADDR
```

### Database connection refused
Запустите test database:
```bash
make test-db-up
```

### Тесты падают с timeout
Увеличьте timeout в `testutil.ContextWithTimeout()` или в коде теста.

### Race detector находит проблемы
Это хорошо! Фикс race conditions критичен для корректности сервера.
