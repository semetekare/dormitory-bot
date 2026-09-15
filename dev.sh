#!/usr/bin/env bash
# dormitory-bot/dev.sh — one-command local development
# Usage: ./dev.sh [up|down|reset|logs|token|auth]
set -e

COMPOSE_FILE="docker-compose.mock.yml"

cmd="${1:-up}"

case "$cmd" in
  up)
    echo "=== Starting mock backend (postgres + redis + app) ==="
    echo MOCK_MODE="true"> ~/.env
    docker compose -f "$COMPOSE_FILE" up -d --build postgres redis app
    echo ""
    echo "=== Backend started on :8080 ==="
    echo ""
    echo "  Frontend:  cd frontend && npm run dev"
    echo "  Open:      http://localhost:5173?role=employ"
    echo "  Swagger:   http://localhost:8080/swagger/index.html"
    echo ""
    echo "  Other roles:  ?role=resident   ?role=stud_council"
    echo ""
    echo "  Logs:     ./dev.sh logs"
    echo "  JWT (curl): curl -s 'http://localhost:8080/api/internal/dev/auth?role=employ' | jq ."
    ;;

  down)
    docker compose -f "$COMPOSE_FILE" down
    ;;

  reset)
    echo "=== Resetting (removing volumes) ==="
    docker compose -f "$COMPOSE_FILE" down -v
    ;;

  rebuild)
    echo "=== Rebuilding and restarting ==="
    docker compose -f "$COMPOSE_FILE" down
    docker compose -f "$COMPOSE_FILE" up -d --build postgres redis app
    ;;

  logs)
    docker compose -f "$COMPOSE_FILE" logs -f app
    ;;

  token)
    echo "=== Launch token (5 min, for old flow) ==="
    curl -s http://localhost:8080/api/internal/dev/launch-token \
      -H 'Content-Type: application/json' \
      -d "{\"phone\":\"${2:-79025643215}\"}" | jq .
    ;;

  auth)
    echo "=== Dev JWT (30 day TTL) ==="
    curl -s "http://localhost:8080/api/internal/dev/auth?role=${2:-employ}" | jq .
    ;;

  *)
    echo "Usage: ./dev.sh [command]"
    echo ""
    echo "Commands:"
    echo "  up        Start mock backend (postgres + redis + app)"
    echo "  down      Stop all containers"
    echo "  reset     Stop + remove volumes (clean DB)"
    echo "  rebuild   Rebuild image and restart"
    echo "  logs      Follow app logs"
    echo "  token     Get launch_token (old 5min flow)"
    echo "  auth      Get dev JWT (30 day TTL)"
    echo ""
    echo "After 'up', start frontend: cd frontend && npm run dev"
    echo "Then open:  http://localhost:5173?role=employ"
    echo "Swagger:    http://localhost:8080/swagger/index.html"
    ;;
esac
