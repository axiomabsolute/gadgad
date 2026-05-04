## Context

A post-implementation code review of the Layer 1 GADDAG index (`c8cf565`) identified ten findings: allocation-heavy patterns on the build and query hot paths, a misplaced helper, a redundant call in the `dict` package, and a missing test case. All changes are confined to `index/` and `dict/`; no public API is removed. The one public API addition (`RotationToWord`) formalizes logic that already exists internally and was already being duplicated in test code.

## Goals / Non-Goals

**Goals:**
- Eliminate unnecessary per-call allocations on the build hot path (`insertSuffix`, `commonPrefixLen`) and the query hot path (`TraverseRaw`)
- Cache graph metrics in `Index` so `Stats()` is a free struct copy
- Export `RotationToWord` as the single canonical decoder for `TraverseRaw` output
- Consolidate string validation and normalization into the files where they are used
- Fix a misleading error message and add a missing edge-case test

**Non-Goals:**
- No algorithmic changes to the GADDAG construction or traversal
- No changes to `TraverseRaw` or `Traverse` signatures
- No serialization or persistence of the index

## Decisions

### D1 — Buffer reuse in `insertSuffix` via caller-owned slices

`insertSuffix` currently allocates `path`, `edges`, and `runes` on every call (~1.25M times for a full dictionary). The fix lifts all three allocations to `Build`, passing them as parameters and re-slicing to `[:0]` before each call.

**Why not a struct receiver?** The buffers are ephemeral to the build loop; packing them into a struct adds indirection without benefit. Plain parameters keep the ownership model explicit.

**Aliasing safety:** `lastPath` and `lastEdges` are read by `replaceOrRegister` on the *next* iteration, before `insertSuffix` overwrites the buffer. The ordering is: read last → re-slice → write new. No aliasing hazard.

### D2 — Package-level `traversalOrder` for DFS iteration

`TraverseRaw` currently calls `alphabetRunes()` on every recursive DFS step, producing a 27-element allocation per node visited. A package-level `var traversalOrder []rune` (Separator followed by A–Z) is initialized once at startup and iterated read-only in all traversals.

**Why not a constant array?** Go does not support constant slices; a package-level `var` is the idiomatic equivalent.

### D3 — Byte comparison in `commonPrefixLen`

`isASCIIUpper` guarantees all characters in rotation strings are in A–Z (single-byte UTF-8), and `Separator` (`+`) is also single-byte. Rune conversion is therefore a no-op correctness-wise and purely wasteful. Direct byte indexing (`a[i] != b[i]`) is safe and avoids two heap allocations per call.

### D4 — Stats cached at build time

`countNodesEdges` is an O(V+E) DFS with a `map` allocation on every `Stats()` call. Since the index is immutable after `Build`, node and edge counts are fixed. They are computed once during `Build` (the graph walk already happens there implicitly via minimization) and stored on `Index`.

**Alternative:** a lazy `sync.Once` cache. Rejected — `Index` is already immutable; eager computation at build time is simpler and has no concurrency concern.

### D5 — Export `RotationToWord`

The function exists in production code as `rotationToWord` and was duplicated verbatim in `index_test.go`. Exporting it removes the duplication, creates a stable public contract for callers using `TraverseRaw`, and allows the test to verify the function directly rather than shadowing it.

## Risks / Trade-offs

- **Buffer reuse (D1):** if a future refactor moves `replaceOrRegister` to after the next `insertSuffix` call, the aliasing invariant breaks silently. The call ordering in `Build` is the load-bearing constraint; it should be preserved carefully.
- **`RotationToWord` as public API (D5):** exporting it is a commitment. If the rotation format changes in a future layer, `RotationToWord` becomes a breaking change. Acceptable given `TraverseRaw` already exposes the format in yielded strings.
- **`traversalOrder` is a package-level mutable slice (D2):** callers who modify it would corrupt all traversals. Since it is unexported and only ranged over (never written to) in production code, the risk is contained.
