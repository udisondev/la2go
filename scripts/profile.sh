#!/bin/bash
# profile.sh — helper скрипт для профилирования la2go

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Директории
PROFILE_DIR="profiles"
mkdir -p "$PROFILE_DIR"

# Функция вывода помощи
usage() {
    cat <<EOF
Использование: $0 [OPTION] PACKAGE

Профилирование Go пакетов la2go.

ОПЦИИ:
  cpu         CPU профилирование (pprof)
  mem         Memory профилирование (pprof)
  block       Block профилирование (горутины blocking)
  mutex       Mutex профилирование (lock contention)
  escape      Escape analysis (переменные, убегающие в heap)
  all         Запустить все профили (кроме escape)
  bench       Запустить все бенчмарки и сохранить результаты
  compare     Сравнить два набора бенчмарков с помощью benchstat

ПРИМЕРЫ:
  $0 cpu ./internal/crypto
  $0 mem ./internal/login
  $0 escape ./internal/protocol
  $0 all ./internal/crypto
  $0 bench ./...
  $0 compare baseline.txt optimized.txt

EOF
    exit 1
}

# Проверка аргументов
if [ $# -lt 1 ]; then
    usage
fi

MODE="$1"
PACKAGE="${2:-.}"

# Временная метка для имен файлов
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

case "$MODE" in
    cpu)
        echo -e "${BLUE}[CPU PROFILING]${NC} $PACKAGE"
        PROFILE_FILE="$PROFILE_DIR/cpu_${TIMESTAMP}.prof"

        echo -e "${YELLOW}Запуск бенчмарков с CPU профилированием...${NC}"
        go test -cpuprofile="$PROFILE_FILE" -bench=. -benchtime=1s "$PACKAGE"

        echo -e "${GREEN}CPU профиль сохранен: $PROFILE_FILE${NC}"
        echo -e "${YELLOW}Открытие pprof в браузере на :8080...${NC}"
        go tool pprof -http=:8080 "$PROFILE_FILE"
        ;;

    mem)
        echo -e "${BLUE}[MEMORY PROFILING]${NC} $PACKAGE"
        PROFILE_FILE="$PROFILE_DIR/mem_${TIMESTAMP}.prof"

        echo -e "${YELLOW}Запуск бенчмарков с Memory профилированием...${NC}"
        go test -memprofile="$PROFILE_FILE" -bench=. -benchtime=1s "$PACKAGE"

        echo -e "${GREEN}Memory профиль сохранен: $PROFILE_FILE${NC}"
        echo -e "${YELLOW}Открытие pprof в браузере на :8080...${NC}"
        go tool pprof -http=:8080 "$PROFILE_FILE"
        ;;

    block)
        echo -e "${BLUE}[BLOCK PROFILING]${NC} $PACKAGE"
        PROFILE_FILE="$PROFILE_DIR/block_${TIMESTAMP}.prof"

        echo -e "${YELLOW}Запуск бенчмарков с Block профилированием...${NC}"
        go test -blockprofile="$PROFILE_FILE" -bench=Concurrent "$PACKAGE"

        echo -e "${GREEN}Block профиль сохранен: $PROFILE_FILE${NC}"
        echo -e "${YELLOW}Открытие pprof в браузере на :8080...${NC}"
        go tool pprof -http=:8080 "$PROFILE_FILE"
        ;;

    mutex)
        echo -e "${BLUE}[MUTEX PROFILING]${NC} $PACKAGE"
        PROFILE_FILE="$PROFILE_DIR/mutex_${TIMESTAMP}.prof"

        echo -e "${YELLOW}Запуск бенчмарков с Mutex профилированием...${NC}"
        go test -mutexprofile="$PROFILE_FILE" -bench=Concurrent "$PACKAGE"

        echo -e "${GREEN}Mutex профиль сохранен: $PROFILE_FILE${NC}"
        echo -e "${YELLOW}Открытие pprof в браузере на :8080...${NC}"
        go tool pprof -http=:8080 "$PROFILE_FILE"
        ;;

    escape)
        echo -e "${BLUE}[ESCAPE ANALYSIS]${NC} $PACKAGE"
        ESCAPE_FILE="$PROFILE_DIR/escape_${TIMESTAMP}.log"

        echo -e "${YELLOW}Запуск escape analysis...${NC}"
        go build -gcflags="-m -m" "$PACKAGE" 2>&1 | tee "$ESCAPE_FILE"

        echo -e "${GREEN}Escape analysis сохранен: $ESCAPE_FILE${NC}"

        # Фильтруем только "escapes to heap"
        CRITICAL_FILE="$PROFILE_DIR/escape_critical_${TIMESTAMP}.txt"
        grep "escapes to heap" "$ESCAPE_FILE" > "$CRITICAL_FILE" || true

        echo -e "${YELLOW}Критичные escapes (только 'escapes to heap'):${NC}"
        cat "$CRITICAL_FILE"
        echo -e "${GREEN}Сохранено в: $CRITICAL_FILE${NC}"
        ;;

    all)
        echo -e "${BLUE}[ALL PROFILING]${NC} $PACKAGE"

        # CPU
        echo -e "\n${YELLOW}=== CPU PROFILING ===${NC}"
        CPU_FILE="$PROFILE_DIR/cpu_${TIMESTAMP}.prof"
        go test -cpuprofile="$CPU_FILE" -bench=. -benchtime=1s "$PACKAGE"
        echo -e "${GREEN}CPU профиль: $CPU_FILE${NC}"

        # Memory
        echo -e "\n${YELLOW}=== MEMORY PROFILING ===${NC}"
        MEM_FILE="$PROFILE_DIR/mem_${TIMESTAMP}.prof"
        go test -memprofile="$MEM_FILE" -bench=. -benchtime=1s "$PACKAGE"
        echo -e "${GREEN}Memory профиль: $MEM_FILE${NC}"

        # Block
        echo -e "\n${YELLOW}=== BLOCK PROFILING ===${NC}"
        BLOCK_FILE="$PROFILE_DIR/block_${TIMESTAMP}.prof"
        go test -blockprofile="$BLOCK_FILE" -bench=Concurrent "$PACKAGE" || true
        echo -e "${GREEN}Block профиль: $BLOCK_FILE${NC}"

        # Mutex
        echo -e "\n${YELLOW}=== MUTEX PROFILING ===${NC}"
        MUTEX_FILE="$PROFILE_DIR/mutex_${TIMESTAMP}.prof"
        go test -mutexprofile="$MUTEX_FILE" -bench=Concurrent "$PACKAGE" || true
        echo -e "${GREEN}Mutex профиль: $MUTEX_FILE${NC}"

        echo -e "\n${GREEN}Все профили сохранены в $PROFILE_DIR${NC}"
        echo -e "${YELLOW}Для просмотра используйте: go tool pprof -http=:8080 <profile_file>${NC}"
        ;;

    bench)
        echo -e "${BLUE}[BENCHMARKS]${NC} $PACKAGE"
        BENCH_FILE="$PROFILE_DIR/bench_${TIMESTAMP}.txt"

        echo -e "${YELLOW}Запуск бенчмарков (10 итераций)...${NC}"
        go test -bench=. -benchmem -count=10 "$PACKAGE" | tee "$BENCH_FILE"

        echo -e "${GREEN}Результаты сохранены: $BENCH_FILE${NC}"
        echo -e "${YELLOW}Для сравнения с baseline используйте:${NC}"
        echo -e "  benchstat baseline.txt $BENCH_FILE"
        ;;

    compare)
        if [ $# -ne 3 ]; then
            echo -e "${RED}Ошибка: требуется 2 файла для сравнения${NC}"
            echo "Использование: $0 compare <baseline.txt> <optimized.txt>"
            exit 1
        fi

        BASELINE="$2"
        OPTIMIZED="$3"

        if [ ! -f "$BASELINE" ]; then
            echo -e "${RED}Ошибка: файл $BASELINE не найден${NC}"
            exit 1
        fi

        if [ ! -f "$OPTIMIZED" ]; then
            echo -e "${RED}Ошибка: файл $OPTIMIZED не найден${NC}"
            exit 1
        fi

        echo -e "${BLUE}[BENCHSTAT COMPARISON]${NC}"
        echo -e "${YELLOW}Baseline:   $BASELINE${NC}"
        echo -e "${YELLOW}Optimized:  $OPTIMIZED${NC}\n"

        # Проверяем наличие benchstat
        if ! command -v benchstat &> /dev/null; then
            echo -e "${YELLOW}benchstat не найден. Установка...${NC}"
            go install golang.org/x/perf/cmd/benchstat@latest
        fi

        benchstat "$BASELINE" "$OPTIMIZED"
        ;;

    *)
        echo -e "${RED}Неизвестный режим: $MODE${NC}"
        usage
        ;;
esac
