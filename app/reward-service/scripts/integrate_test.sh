#!/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $SCRIPT_DIR/..
docker compose  -p w-book -f docker-compose-integration.yml down -v
docker compose  -p w-book -f docker-compose-integration.yml up -d

go test -race -coverprofile=coverage.out ./... 
# 提取总覆盖率
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# 检查覆盖率是否大于等于 50
if (( $(echo "$COVERAGE >= 30" | bc -l) )); then
  echo "Coverage check passed: $COVERAGE%"
  docker compose  -p w-book -f docker-compose-integration.yml down -v
  exit 0
else
  echo "Coverage check failed: $COVERAGE%"
  docker compose  -p w-book -f docker-compose-integration.yml down -v
  exit 1
fi

