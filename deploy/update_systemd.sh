#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  update_systemd.sh --service <name> --binary <install-path> --new-binary <path> [options]

Required:
  --service       systemd service name (e.g. sub2api)
  --binary        target binary path used by systemd ExecStart
  --new-binary    new binary file to deploy

Options:
  --health-url    health check URL (e.g. http://127.0.0.1:8080/health)
  --timeout       seconds to wait for service/health (default: 20)
  --keep-backups  number of backups to keep (default: 5)

Example:
  ./update_systemd.sh \
    --service sub2api \
    --binary /opt/sub2api/server \
    --new-binary /tmp/server \
    --health-url http://127.0.0.1:8080/health
EOF
}

SERVICE=""
BIN_PATH=""
NEW_BIN=""
HEALTH_URL=""
TIMEOUT=20
KEEP_BACKUPS=5

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service)
      SERVICE="$2"; shift 2 ;;
    --binary)
      BIN_PATH="$2"; shift 2 ;;
    --new-binary)
      NEW_BIN="$2"; shift 2 ;;
    --health-url)
      HEALTH_URL="$2"; shift 2 ;;
    --timeout)
      TIMEOUT="$2"; shift 2 ;;
    --keep-backups)
      KEEP_BACKUPS="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      echo "Unknown argument: $1" >&2
      usage; exit 1 ;;
  esac
done

if [[ -z "$SERVICE" || -z "$BIN_PATH" || -z "$NEW_BIN" ]]; then
  usage
  exit 1
fi

if [[ ! -f "$NEW_BIN" ]]; then
  echo "New binary not found: $NEW_BIN" >&2
  exit 1
fi

if [[ ! -x "$NEW_BIN" ]]; then
  echo "New binary is not executable: $NEW_BIN" >&2
  exit 1
fi

timestamp=$(date +%Y%m%d%H%M%S)
backup="${BIN_PATH}.bak.${timestamp}"

echo "Stopping service: $SERVICE"
systemctl stop "$SERVICE"

if [[ -f "$BIN_PATH" ]]; then
  echo "Backing up current binary to: $backup"
  cp -p "$BIN_PATH" "$backup"
fi

echo "Installing new binary: $NEW_BIN -> $BIN_PATH"
install -m 755 "$NEW_BIN" "$BIN_PATH"

echo "Starting service: $SERVICE"
systemctl start "$SERVICE"

echo "Waiting for service to be active..."
active=0
for ((i=0; i<TIMEOUT; i++)); do
  if systemctl is-active --quiet "$SERVICE"; then
    active=1
    break
  fi
  sleep 1
done

if [[ "$active" -ne 1 ]]; then
  echo "Service did not become active, rolling back" >&2
  if [[ -f "$backup" ]]; then
    install -m 755 "$backup" "$BIN_PATH"
    systemctl start "$SERVICE" || true
  fi
  exit 1
fi

if [[ -n "$HEALTH_URL" ]]; then
  echo "Checking health: $HEALTH_URL"
  ok=0
  for ((i=0; i<TIMEOUT; i++)); do
    if curl -fsS --connect-timeout 3 --max-time 5 "$HEALTH_URL" >/dev/null; then
      ok=1
      break
    fi
    sleep 1
  done
  if [[ "$ok" -ne 1 ]]; then
    echo "Health check failed, rolling back" >&2
    if [[ -f "$backup" ]]; then
      install -m 755 "$backup" "$BIN_PATH"
      systemctl start "$SERVICE" || true
    fi
    exit 1
  fi
fi

if [[ -f "$BIN_PATH" ]]; then
  dir=$(dirname "$BIN_PATH")
  pattern=$(basename "$BIN_PATH").bak.*
  backups=("$dir"/$pattern)
  if [[ ${#backups[@]} -gt $KEEP_BACKUPS ]]; then
    excess=$(( ${#backups[@]} - KEEP_BACKUPS ))
    ls -1t "${backups[@]}" | tail -n "$excess" | while read -r f; do
      rm -f "$f"
    done
  fi
fi

echo "Update completed"
