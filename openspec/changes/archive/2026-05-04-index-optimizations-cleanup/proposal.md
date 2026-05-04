## Why

The initial GADDAG implementation is correct but contains several allocation-heavy patterns on the build and query hot paths, a handful of misplaced helpers, and minor API rough edges identified during a post-implementation code review. Addressing them now, before any consumers exist, costs nothing in migration and keeps the codebase clean as Layer 2 work begins.

## What Changes

- Replace `words[:0]` in-place filter with `make([]string, 0, len(words))` for clarity and safety (`index/index.go`)
- Promote the per-DFS-call `alphabetRunes()` allocation to a package-level `traversalOrder` slice (`index/index.go`)
- Replace `[]rune` conversion in `commonPrefixLen` with byte comparison; rotation strings are guaranteed single-byte ASCII (`index/index.go`)
- Lift `path`, `edges`, and `runes` buffer allocations out of `insertSuffix` and reuse across the build loop (`index/index.go`)
- Cache `NodeCount` and `EdgeCount` in `Index` at build time so `Stats()` is a free struct copy (`index/index.go`, `index/index_test.go`)
- Export `rotationToWord` as `RotationToWord`; add focused unit tests; remove the verbatim copy from the test file (`index/index.go`, `index/index_test.go`)
- Move `isASCIIUpper` from `node.go` to `index.go` where it is used (`index/node.go`, `index/index.go`)
- Simplify `FileSource.Words` to normalize first and check the result; eliminates redundant `TrimSpace` call (`dict/file.go`)
- Fix imprecise error message in `Build`: "contains characters outside A–Z" (`index/index.go`)
- Add single-letter word roundtrip test to `index_test.go` (`index/index_test.go`)

## Capabilities

### New Capabilities

- `rotation-word-reconstruction`: Export `RotationToWord` as a public function so callers working with `TraverseRaw` output can reconstruct original words without duplicating the decoding logic.

### Modified Capabilities

*(none — all other changes are implementation improvements with no spec-level behavior changes)*

## Impact

- `index/index.go`: majority of changes (dedup, traversal order, prefix comparison, buffer reuse, stats caching, error message, `RotationToWord` export, `isASCIIUpper` move)
- `index/node.go`: remove `isASCIIUpper`
- `index/index_test.go`: remove duplicate `rotationToWord`, add tests for `RotationToWord` and single-letter roundtrip, update consistency test
- `dict/file.go`: simplify `Words` normalization path
- No public API removals; one addition (`RotationToWord`)
- No dependency changes
