#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  cp configs/.env.example .env
  echo "created .env from configs/.env.example"
fi

docker compose -f deployments/docker-compose.yml up -d

echo "waiting for mysql..."
for _ in $(seq 1 60); do
  status="$(docker inspect --format='{{.State.Health.Status}}' otome-mysql 2>/dev/null || true)"
  [ "$status" = "healthy" ] && break
  sleep 2
done
if [ "$status" != "healthy" ]; then
  echo "mysql did not become healthy in time" >&2
  exit 1
fi

go run ./cmd/migrate up
exec go run ./cmd/api
