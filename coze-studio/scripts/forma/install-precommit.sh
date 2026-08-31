#!/bin/bash
# Install Forma-aware pre-commit into .git/hooks (repo root = Forma workspace).
set -euo pipefail
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FORMA_ROOT="$( cd "$SCRIPT_DIR/../../.." && pwd )"
HOOK_DST="$FORMA_ROOT/.git/hooks/pre-commit"
mkdir -p "$(dirname "$HOOK_DST")"
cat > "$HOOK_DST" <<'EOF'
#!/bin/bash
set -e
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
COZE_ROOT="$( cd "$SCRIPT_DIR/../../coze-studio" && pwd )"
IMPL="$COZE_ROOT/common/git-hooks/pre-commit"
if [[ -f "$IMPL" ]]; then
  cd "$COZE_ROOT"
  "$IMPL" "$@"
else
  echo "missing coze-studio/common/git-hooks/pre-commit" >&2
  exit 1
fi
EOF
chmod +x "$HOOK_DST"
echo "Installed Forma pre-commit wrapper -> $HOOK_DST"
