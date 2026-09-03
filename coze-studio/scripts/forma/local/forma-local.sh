#!/usr/bin/env bash
# Forma local development launcher (macOS / Linux / Git Bash).
# Thin wrapper around the shared Node core.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE="${SCRIPT_DIR}/forma-local.mjs"

if ! command -v node >/dev/null 2>&1; then
  echo "BLOCKED: Node.js is required (rush.json requires >=21)." >&2
  exit 2
fi

if [[ ! -f "$CORE" ]]; then
  echo "BLOCKED: missing shared launcher: $CORE" >&2
  exit 2
fi

exec node "$CORE" "$@"
