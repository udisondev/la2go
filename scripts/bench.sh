#!/bin/bash
# bench.sh — автоматизация запуска бенчмарков la2go

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Директории
BENCH_DIR="benchmarks"
mkdir -p "$BENCH_DIR"

# Временная метка
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Функция вывода помощи
usage() {
    cat <<EOF
Использование: $0 [OPTION]

Автоматизация запуска бенчмарков la2go.

ОПЦИИ:
  all         Запустить все бенчмарки (./...)
  crypto      Только crypto бенчмарки
  login       Только login бенчмарки
  gameserver  Только gameserver бенчмарки
  quick       Быстрый прогон всех бенчмарков (benchtime=100ms, count=1)
  full        Полный прогон с профилированием (benchtime=1s, count=10)
  ci          CI режим (benchtime=500ms, count=5, no progress)
  compare     Сравнить текущие результаты с baseline

ПРИМЕРЫ:
  $0 all
  $0 crypto
  $0 quick
  $0 full
  $0 compare

EOF
    exit 1
}

# Проверка аргументов
if [ $# -lt 1 ]; then
    usage
fi

MODE="$1"

# Функция запуска бенчмарков
run_bench() {
    local package="$1"
    local benchtime="${2:-1s}"
    local count="${3:-10}"
    local output_file="$4"

    echo -e "${YELLOW}Запуск бенчмарков: $package${NC}"
    echo -e "${YELLOW}  benchtime: $benchtime, count: $count${NC}"

    go test -bench=. -benchmem -benchtime="$benchtime" -count="$count" "$package" | tee -a "$output_file"

    echo -e "${GREEN}✓ Завершено: $package${NC}\n"
}

case "$MODE" in
    all)
        echo -e "${BLUE}[ALL BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/all_${TIMESTAMP}.txt"

        echo -e "${YELLOW}Запуск всех бенчмарков (может занять несколько минут)...${NC}\n"

        run_bench "./internal/crypto" "1s" "10" "$OUTPUT_FILE"
        run_bench "./internal/login" "1s" "10" "$OUTPUT_FILE"
        run_bench "./internal/gameserver" "1s" "10" "$OUTPUT_FILE"

        echo -e "${GREEN}Все бенчмарки завершены!${NC}"
        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"
        ;;

    crypto)
        echo -e "${BLUE}[CRYPTO BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/crypto_${TIMESTAMP}.txt"

        run_bench "./internal/crypto" "1s" "10" "$OUTPUT_FILE"

        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"
        ;;

    login)
        echo -e "${BLUE}[LOGIN BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/login_${TIMESTAMP}.txt"

        run_bench "./internal/login" "1s" "10" "$OUTPUT_FILE"

        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"
        ;;

    gameserver)
        echo -e "${BLUE}[GAMESERVER BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/gameserver_${TIMESTAMP}.txt"

        run_bench "./internal/gameserver" "1s" "10" "$OUTPUT_FILE"

        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"
        ;;

    quick)
        echo -e "${BLUE}[QUICK BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/quick_${TIMESTAMP}.txt"

        echo -e "${YELLOW}Быстрый прогон всех бенчмарков...${NC}\n"

        run_bench "./internal/crypto" "100ms" "1" "$OUTPUT_FILE"
        run_bench "./internal/login" "100ms" "1" "$OUTPUT_FILE"
        run_bench "./internal/gameserver" "100ms" "1" "$OUTPUT_FILE"

        echo -e "${GREEN}Быстрый прогон завершен!${NC}"
        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"
        ;;

    full)
        echo -e "${BLUE}[FULL BENCHMARKS WITH PROFILING]${NC}"
        OUTPUT_FILE="$BENCH_DIR/full_${TIMESTAMP}.txt"
        PROFILE_DIR="profiles"
        mkdir -p "$PROFILE_DIR"

        echo -e "${YELLOW}Полный прогон с профилированием...${NC}\n"

        # Crypto с профилированием
        echo -e "${YELLOW}=== Crypto (CPU + Memory profiling) ===${NC}"
        go test -bench=. -benchmem -benchtime=1s -count=10 \
            -cpuprofile="$PROFILE_DIR/crypto_cpu_${TIMESTAMP}.prof" \
            -memprofile="$PROFILE_DIR/crypto_mem_${TIMESTAMP}.prof" \
            ./internal/crypto | tee -a "$OUTPUT_FILE"

        # Login с профилированием
        echo -e "\n${YELLOW}=== Login (CPU + Memory profiling) ===${NC}"
        go test -bench=. -benchmem -benchtime=1s -count=10 \
            -cpuprofile="$PROFILE_DIR/login_cpu_${TIMESTAMP}.prof" \
            -memprofile="$PROFILE_DIR/login_mem_${TIMESTAMP}.prof" \
            ./internal/login | tee -a "$OUTPUT_FILE"

        # GameServer с профилированием
        echo -e "\n${YELLOW}=== GameServer (CPU + Memory profiling) ===${NC}"
        go test -bench=. -benchmem -benchtime=1s -count=10 \
            -cpuprofile="$PROFILE_DIR/gameserver_cpu_${TIMESTAMP}.prof" \
            -memprofile="$PROFILE_DIR/gameserver_mem_${TIMESTAMP}.prof" \
            ./internal/gameserver | tee -a "$OUTPUT_FILE"

        echo -e "\n${GREEN}Полный прогон завершен!${NC}"
        echo -e "${GREEN}Результаты: $OUTPUT_FILE${NC}"
        echo -e "${GREEN}Профили: $PROFILE_DIR/*_${TIMESTAMP}.prof${NC}"
        echo -e "${YELLOW}Для просмотра профилей: go tool pprof -http=:8080 <file>.prof${NC}"
        ;;

    ci)
        echo -e "${BLUE}[CI BENCHMARKS]${NC}"
        OUTPUT_FILE="$BENCH_DIR/ci_${TIMESTAMP}.txt"

        echo -e "${YELLOW}CI режим (no progress output)...${NC}\n"

        # Запускаем без tee для CI
        go test -bench=. -benchmem -benchtime=500ms -count=5 ./... > "$OUTPUT_FILE"

        echo -e "${GREEN}CI бенчмарки завершены!${NC}"
        echo -e "${GREEN}Результаты сохранены: $OUTPUT_FILE${NC}"

        # Выводим summary
        echo -e "\n${YELLOW}Summary:${NC}"
        grep "^Benchmark" "$OUTPUT_FILE" | tail -20
        ;;

    compare)
        echo -e "${BLUE}[COMPARE BENCHMARKS]${NC}"

        # Ищем baseline файл
        BASELINE="$BENCH_DIR/baseline.txt"

        if [ ! -f "$BASELINE" ]; then
            echo -e "${RED}Ошибка: baseline файл не найден: $BASELINE${NC}"
            echo -e "${YELLOW}Создайте baseline командой:${NC}"
            echo -e "  $0 all"
            echo -e "  mv $BENCH_DIR/all_*.txt $BASELINE"
            exit 1
        fi

        # Запускаем текущие бенчмарки
        CURRENT="$BENCH_DIR/current_${TIMESTAMP}.txt"
        echo -e "${YELLOW}Запуск текущих бенчмарков...${NC}\n"

        run_bench "./internal/crypto" "1s" "10" "$CURRENT"
        run_bench "./internal/login" "1s" "10" "$CURRENT"
        run_bench "./internal/gameserver" "1s" "10" "$CURRENT"

        echo -e "\n${YELLOW}Сравнение с baseline...${NC}\n"

        # Проверяем наличие benchstat
        if ! command -v benchstat &> /dev/null; then
            echo -e "${YELLOW}benchstat не найден. Установка...${NC}"
            go install golang.org/x/perf/cmd/benchstat@latest
        fi

        benchstat "$BASELINE" "$CURRENT"

        echo -e "\n${GREEN}Текущие результаты сохранены: $CURRENT${NC}"
        echo -e "${YELLOW}Для обновления baseline:${NC}"
        echo -e "  mv $CURRENT $BASELINE"
        ;;

    *)
        echo -e "${RED}Неизвестный режим: $MODE${NC}"
        usage
        ;;
esac
