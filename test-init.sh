#!/bin/bash
# Скрипт для тестирования Init пакета

set -e

echo "=== Building loginserver ==="
go build -o loginserver ./cmd/loginserver

echo ""
echo "=== Building test-init-real ==="
go build -o test-init-real ./cmd/test-init-real

echo ""
echo "=== Starting loginserver in background ==="
./loginserver > loginserver.log 2>&1 &
LOGINSERVER_PID=$!
echo "LoginServer PID: $LOGINSERVER_PID"

# Даём серверу время запуститься
sleep 2

echo ""
echo "=== Running test-init-real ==="
./test-init-real || true

echo ""
echo "=== Killing loginserver ==="
kill $LOGINSERVER_PID || true
wait $LOGINSERVER_PID 2>/dev/null || true

echo ""
echo "=== LoginServer logs (last 50 lines) ==="
tail -50 loginserver.log

echo ""
echo "=== Cleanup ==="
rm -f loginserver test-init-real loginserver.log

echo ""
echo "=== Done ==="
