#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BENCHMARKS_MD="$REPO_ROOT/BENCHMARKS.md"
BENCHTIME="${BENCHTIME:-5s}"

cd "$REPO_ROOT"

echo "Running index benchmarks (-benchmem -benchtime=${BENCHTIME})..."
index_output=$(go test -bench=. -benchmem -benchtime="$BENCHTIME" ./index/ 2>&1)

echo "Running query benchmarks (-benchmem -benchtime=${BENCHTIME})..."
query_output=$(go test -bench=. -benchmem -benchtime="$BENCHTIME" ./query/ 2>&1)

# --- index benchmark lines ---
small_build=$(printf '%s\n' "$index_output" | grep '^BenchmarkBuild_SmallWords'    || true)
traverse_e=$(printf '%s\n'  "$index_output" | grep '^BenchmarkTraverse_E'           || true)
twl06_build=$(printf '%s\n' "$index_output" | grep '^BenchmarkBuild_TWL06-'         || true)
twl06_trav=$(printf '%s\n'  "$index_output" | grep '^BenchmarkTraverse_TWL06_S-'    || true)
twl06_trat=$(printf '%s\n'  "$index_output" | grep '^BenchmarkTraverseAt_TWL06_S-'  || true)

# --- query benchmark lines (all require testdata/twl06.txt) ---
q_crossword=$(printf '%s\n'  "$query_output" | grep '^BenchmarkSearch_Crossword_TWL06-'        || true)
q_scrabble=$(printf '%s\n'   "$query_output" | grep '^BenchmarkSearch_Scrabble_TWL06-'         || true)
q_anagram_c=$(printf '%s\n'  "$query_output" | grep '^BenchmarkSearch_Anagram_TWL06_Common-'   || true)
q_anagram_m=$(printf '%s\n'  "$query_output" | grep '^BenchmarkSearch_Anagram_TWL06_Moderate-' || true)

if [ -z "$small_build" ] || [ -z "$traverse_e" ]; then
    echo "error: expected index benchmark output not found. Raw output:" >&2
    printf '%s\n' "$index_output" >&2
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
    echo "### Index layer (\`index/\`)"
    echo ""
    echo "#### Small dictionary (\`testdata/small_words.txt\`, ${small_count} words)"
    echo ""
    echo '```'
    printf '%s\n' "$small_build" "$traverse_e"
    echo '```'

    if [ -n "$twl06_build" ] || [ -n "$twl06_trav" ] || [ -n "$twl06_trat" ]; then
        twl06_count=$(grep -c '^[A-Z]' "$REPO_ROOT/testdata/twl06.txt" 2>/dev/null || echo "?")
        echo ""
        echo "#### Full dictionary (\`testdata/twl06.txt\`, ${twl06_count} words)"
        echo ""
        echo '```'
        [ -n "$twl06_build" ] && printf '%s\n' "$twl06_build"
        [ -n "$twl06_trav"  ] && printf '%s\n' "$twl06_trav"
        [ -n "$twl06_trat"  ] && printf '%s\n' "$twl06_trat"
        echo '```'
    else
        echo ""
        echo "_Full dictionary index benchmarks skipped (\`testdata/twl06.txt\` not present)._"
    fi

    echo ""
    echo "### Query layer (\`query/\`)"
    echo ""
    if [ -n "$q_crossword" ] || [ -n "$q_scrabble" ] || [ -n "$q_anagram_c" ] || [ -n "$q_anagram_m" ]; then
        twl06_count=$(grep -c '^[A-Z]' "$REPO_ROOT/testdata/twl06.txt" 2>/dev/null || echo "?")
        echo "#### Full dictionary (\`testdata/twl06.txt\`, ${twl06_count} words)"
        echo ""
        echo '```'
        [ -n "$q_crossword"  ] && printf '%s\n' "$q_crossword"
        [ -n "$q_scrabble"   ] && printf '%s\n' "$q_scrabble"
        [ -n "$q_anagram_c"  ] && printf '%s\n' "$q_anagram_c"
        [ -n "$q_anagram_m"  ] && printf '%s\n' "$q_anagram_m"
        echo '```'
        echo ""
        echo "_Note: \`Anagram_TWL06_Common\` (rack \`AEINTRS\`) exercises the worst-case path through_"
        echo "_\`Traverse\` on a high-frequency anchor. \`Anagram_TWL06_Moderate\` (rack \`ABORVXZ\`) reflects_"
        echo "_typical performance when the rack contains at least one rare letter._"
    else
        echo "_Query benchmarks skipped (\`testdata/twl06.txt\` not present)._"
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
