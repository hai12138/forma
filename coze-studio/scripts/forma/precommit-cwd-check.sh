#!/usr/bin/env bash
# Verifies Forma repo-root pre-commit wrapper resolves into coze-studio/.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
WRAPPER="$ROOT/.git/hooks/pre-commit"
IMPL="$ROOT/coze-studio/common/git-hooks/pre-commit"

test -f "$WRAPPER"
test -f "$IMPL"
grep -q 'coze-studio' "$WRAPPER"
grep -q 'cd "\$COZE_ROOT"' "$WRAPPER" || grep -q 'cd "$COZE_ROOT"' "$WRAPPER"
echo "forma pre-commit CWD resolution OK"
