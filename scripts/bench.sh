#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BENCHMARKS_MD="$REPO_ROOT/BENCHMARKS.md"
BENCHTIME="${BENCHTIME:-5s}"

cd "$REPO_ROOT"

echo "Running benchmarks (-benchmem -benchtime=${BENCHTIME})..."
bench_output=$(go test -bench=. -benchmem -benchtime="$BENCHTIME" ./index/ 2>&1)

# Extract benchmark result lines by name. The GOMAXPROCS suffix (-N) is kept
# as part of the canonical output.
small_build=$(printf '%s\n' "$bench_output" | grep '^BenchmarkBuild_SmallWords'  || true)
traverse_e=$(printf '%s\n'  "$bench_output" | grep '^BenchmarkTraverse_E'        || true)
twl06_build=$(printf '%s\n' "$bench_output" | grep '^BenchmarkBuild_TWL06-'      || true)
twl06_trav=$(printf '%s\n'  "$bench_output" | grep '^BenchmarkTraverse_TWL06_S'  || true)

if [ -z "$small_build" ] || [ -z "$traverse_e" ]; then
    echo "error: expected benchmark output not found. Raw output:" >&2
    printf '%s\n' "$bench_output" >&2
    exit 1
fi

commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
date_str=$(date +%Y-%m-%d)
small_count=$(grep -c '^[A-Z]' "$REPO_ROOT/testdata/small_words.txt" 2>/dev/null || echo "?")

# Write replacement block to a temp file.
tmpfile=$(mktemp)
trap 'rm -f "$tmpfile"' EXIT

{
    echo "Recorded at commit \`${commit}\` on ${date_str}. Run with \`-benchmem -benchtime=${BENCHTIME}\`."
    echo ""
    echo "### Small dictionary (\`testdata/small_words.txt\`, ${small_count} words)"
    echo ""
    echo '```'
    printf '%s\n' "$small_build" "$traverse_e"
    echo '```'

    if [ -n "$twl06_build" ] && [ -n "$twl06_trav" ]; then
        twl06_count=$(grep -c '^[A-Z]' "$REPO_ROOT/testdata/twl06.txt" 2>/dev/null || echo "?")
        echo ""
        echo "### Full dictionary (\`testdata/twl06.txt\`, ${twl06_count} words)"
        echo ""
        echo '```'
        printf '%s\n' "$twl06_build" "$twl06_trav"
        echo '```'
    else
        echo ""
        echo "_Full dictionary benchmarks skipped (\`testdata/twl06.txt\` not present)._"
    fi
} > "$tmpfile"

# Splice replacement between markers (markers are preserved; content between them is replaced).
awk -v newfile="$tmpfile" '
/<!-- bench:current:start -->/ {
    print
    while ((getline line < newfile) > 0) print line
    close(newfile)
    skip = 1
    next
}
/<!-- bench:current:end -->/ {
    skip = 0
}
!skip { print }
' "$BENCHMARKS_MD" > "${BENCHMARKS_MD}.tmp" && mv "${BENCHMARKS_MD}.tmp" "$BENCHMARKS_MD"

echo "Updated ${BENCHMARKS_MD} (commit ${commit}, ${date_str})"
