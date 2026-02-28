#!/usr/bin/env bash
set -euo pipefail

# =====================================================
# CONFIGURATION
# =====================================================
APP_DIR="/opt/intergrated_health_portal"
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="intergrated-health-portal"
ENV_FILE="app.env"

# GHCR (GitHub Container Registry)
GHCR_USER="mohuganda"
GHCR_TOKEN_FILE="$APP_DIR/.secrets/ghcr_token"

# Deployment mode: pull | build
MODE="pull"

# Optional behavior
TAIL_LOGS="true"
PRUNE_IMAGES="true"
LOG_TAIL_LINES=120

# =====================================================
# UTILITIES
# =====================================================
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

fail() {
  echo "❌ ERROR: $*"
  exit 1
}

# =====================================================
# PRECHECKS
# =====================================================

# Ensure bash is used
if [ -z "${BASH_VERSION:-}" ]; then
  fail "This script must be run with bash (not sh)."
fi

# Ensure docker exists
command -v docker >/dev/null 2>&1 || fail "Docker is not installed."

# Ensure docker is accessible (no sudo needed)
if ! docker ps >/dev/null 2>&1; then
  fail "Docker permission denied. Add user to docker group."
fi

# Ensure app directory exists
[[ -d "$APP_DIR" ]] || fail "App directory not found: $APP_DIR"

cd "$APP_DIR"

# Ensure compose file exists
[[ -f "$COMPOSE_FILE" ]] || fail "Compose file not found: $COMPOSE_FILE"

# Ensure env file exists
[[ -f "$ENV_FILE" ]] || fail "Env file not found: $ENV_FILE"

# Ensure GHCR token exists
[[ -f "$GHCR_TOKEN_FILE" ]] || fail "GHCR token file not found: $GHCR_TOKEN_FILE"

# =====================================================
# GHCR LOGIN
# =====================================================
log "Logging into GHCR..."

if ! cat "$GHCR_TOKEN_FILE" | docker login ghcr.io -u "$GHCR_USER" --password-stdin >/dev/null 2>&1; then
  fail "GHCR login failed. Check token."
fi

log "GHCR login successful."

# =====================================================
# DEPLOYMENT
# =====================================================
log "Starting deployment..."
log "Project: $PROJECT_NAME"
log "Mode: $MODE"
log "Using env file: $ENV_FILE"

if [[ "$MODE" == "pull" ]]; then
  log "Pulling latest images..."
  docker compose \
    --env-file "$ENV_FILE" \
    -p "$PROJECT_NAME" \
    -f "$COMPOSE_FILE" \
    pull
elif [[ "$MODE" == "build" ]]; then
  log "Building images locally (with --pull)..."
  docker compose \
    --env-file "$ENV_FILE" \
    -p "$PROJECT_NAME" \
    -f "$COMPOSE_FILE" \
    build --pull
else
  fail "Invalid MODE: $MODE (use 'pull' or 'build')"
fi

log "Recreating containers..."
docker compose \
  --env-file "$ENV_FILE" \
  -p "$PROJECT_NAME" \
  -f "$COMPOSE_FILE" \
  up -d --force-recreate --remove-orphans

# =====================================================
# STATUS
# =====================================================
log "Container status:"
docker compose \
  --env-file "$ENV_FILE" \
  -p "$PROJECT_NAME" \
  -f "$COMPOSE_FILE" \
  ps

# =====================================================
# LOGS
# =====================================================
if [[ "$TAIL_LOGS" == "true" ]]; then
  log "Recent logs (last $LOG_TAIL_LINES lines):"
  docker compose \
    --env-file "$ENV_FILE" \
    -p "$PROJECT_NAME" \
    -f "$COMPOSE_FILE" \
    logs --tail="$LOG_TAIL_LINES" || true
fi

# =====================================================
# CLEANUP
# =====================================================
if [[ "$PRUNE_IMAGES" == "true" ]]; then
  log "Cleaning up dangling images..."
  docker image prune -f
fi

log "Deployment completed successfully ✅"