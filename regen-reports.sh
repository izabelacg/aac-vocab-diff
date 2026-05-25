#!/usr/bin/env bash
# regen-reports.sh — build the binary and diff all consecutive .ce pairs in a directory.
#
# Usage:
#   ./regen-reports.sh                    # diffs files in diff/testdata/
#   ./regen-reports.sh /path/to/dir       # diffs files in the given directory
#
# Files are sorted by filename (oldest first) and diffed as consecutive
# pairs: 1→2, 2→3, 3→4, etc.  Reports are written to tmp/reports/.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="${1:-"$SCRIPT_DIR/diff/testdata"}"
OUT_DIR="$SCRIPT_DIR/tmp/reports"
BIN="$SCRIPT_DIR/bin/aac-vocab-diff"

# ── Build ────────────────────────────────────────────────────────────────────

echo "==> Building..."
make -C "$SCRIPT_DIR" build
echo

# ── Resolve source directory ─────────────────────────────────────────────────

if [[ ! -d "$SRC_DIR" ]]; then
    echo "error: $SRC_DIR is not a directory" >&2
    exit 1
fi
# Resolve symlinks so find works correctly on macOS (/tmp → /private/tmp).
SRC_DIR="$(cd "$SRC_DIR" && pwd -P)"

# ── Collect .ce files sorted by modification time, oldest first ──────────────

FILES=()
while IFS= read -r f; do
    FILES+=("$f")
done < <(python3 - "$SRC_DIR" << 'PYEOF'
import os, glob, sys
files = glob.glob(os.path.join(sys.argv[1], "*.ce"))
files.sort(key=os.path.basename)
print("\n".join(files))
PYEOF
)

# To use a specific list instead of auto-discovery, uncomment and edit this line,
# then comment out the while-loop block below it.
# FILES=("diff/testdata/file_a.ce" "diff/testdata/file_b.ce" "diff/testdata/file_c.ce")

if [[ ${#FILES[@]} -lt 2 ]]; then
    echo "error: need at least 2 .ce files in $SRC_DIR, found ${#FILES[@]}" >&2
    exit 1
fi

# ── Generate consecutive-pair reports ────────────────────────────────────────

mkdir -p "$OUT_DIR"

echo "Found ${#FILES[@]} files — generating $((${#FILES[@]} - 1)) report(s) in $OUT_DIR"
echo

for (( i = 0; i < ${#FILES[@]} - 1; i++ )); do
    OLD="${FILES[$i]}"
    NEW="${FILES[$((i+1))]}"
    OLD_BASE="$(basename "$OLD" .ce)"
    NEW_BASE="$(basename "$NEW" .ce)"
    OUT="$OUT_DIR/${OLD_BASE}_vs_${NEW_BASE}.html"

    echo "  $(basename "$OLD") → $(basename "$NEW")"
    "$BIN" "$OLD" "$NEW" --report "$OUT"
done

echo
echo "Done → $OUT_DIR"
