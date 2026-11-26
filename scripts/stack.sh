#!/usr/bin/env bash
# Local e2e stack for validation: Postgres, API, Go worker, and two test sites
# (badsite: expired self-signed cert, no headers; goodsite: valid cert, all headers).
# It never reads credentials from the shell: the API key and DB password are local-only.
#
#   scripts/stack.sh start    # build, start, return when the API is healthy
#   scripts/stack.sh status
#   scripts/stack.sh logs [service]
#   scripts/stack.sh stop     # stop and delete containers and volumes
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

COMPOSE=(docker compose -f docker-compose.yml)
# Reason: behind a TLS-intercepting proxy, image builds need its CA; set LANE_PROXY_CA to use it.
if [ -z "${LANE_PROXY_CA:-}" ] && [ -f /root/.ccr/ca-bundle.crt ]; then export LANE_PROXY_CA=/root/.ccr/ca-bundle.crt; fi
if [ -n "${LANE_PROXY_CA:-}" ]; then COMPOSE+=(-f docker-compose.proxy.yml); fi

ensure_docker() {
  if docker info >/dev/null 2>&1; then return; fi
  # Reason: cloud sessions ship the Docker binaries without a running daemon.
  if command -v dockerd >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    echo "[stack] starting dockerd"
    (sudo -n dockerd >/tmp/agent-lane-dockerd.log 2>&1 &)
    for _ in $(seq 1 30); do docker info >/dev/null 2>&1 && return; sleep 1; done
  fi
  echo "[stack] Docker is not available" >&2
  exit 1
}

case "${1:-}" in
  start)
    ensure_docker
    "${COMPOSE[@]}" up -d --build --wait --wait-timeout "${STACK_TIMEOUT:-600}"
    echo "[stack] API on http://127.0.0.1:${LANE_API_PORT:-8080} (x-api-key: ${LANE_API_KEY:-local-dev-key})"
    ;;
  status) "${COMPOSE[@]}" ps ;;
  logs) shift; "${COMPOSE[@]}" logs --no-color --tail 200 "$@" ;;
  stop) "${COMPOSE[@]}" down -v --remove-orphans ;;
  *) echo "usage: $0 start|status|logs|stop" >&2; exit 2 ;;
esac
