## Context

`gadgad` is a new Go library (greenfield). This design covers the first and
most algorithmically complex piece: the GADDAG index and its supporting
dictionary loader. There is no existing code to integrate with or migrate from.

The GADDAG (Graph Automaton for Dictionary Directed by Gordon) was introduced
by Steven Gordon (1994) as a solution to efficient Scrabble move generation. It
encodes every word in a dictionary as multiple rotations — one per letter
position — using a reserved separator character (`+`). All rotations are stored
in a single minimized directed acyclic word graph (DAWG), enabling bidirectional
traversal from any anchor letter without scanning the full dictionary.

## Goals / Non-Goals

**Goals:**
- Correct GADDAG construction from an arbitrary `[]string` word list
- DAWG minimization (shared subgraph compression) for space efficiency
- Single-anchor unconstrained traversal returning all words containing a given
  letter at any position
- Pluggable dictionary loading via a `Source` interface (`FileSource`,
  `SliceSource`)
- Benchmark baselines for construction time and memory on a standard wordlist

**Non-Goals:**
- Query constraints (available letter sets, exclusions, length bounds) — Layer 2
- Multi-anchor patterns — Layer 2
- Dynamic updates (insert/delete after construction) — explicitly deferred;
  would require a different structure
- Game-specific logic of any kind

## Decisions

### D1: GADDAG over plain DAWG or trie

**Decision:** Use GADDAG.

A plain trie is the simplest option but requires a full scan for non-prefix
queries. A DAWG adds minimization but still only supports prefix traversal
efficiently. A reverse-trie augmentation handles suffixes, but infix queries
(anchor in the middle) still require either two traversals merged in memory or
a scan. GADDAG encodes the rotation trick directly into the graph, making
any-position anchor traversal a single graph walk.

**Alternatives considered:**
- *Trie only*: Simple to build, easy to understand, but O(n) for infix queries.
  Unsuitable for the stated use case.
- *DAWG + reverse DAWG*: Two structures in memory, awkward API, still doesn't
  generalize cleanly to arbitrary anchor positions.
- *Suffix array*: Excellent for substring queries but doesn't naturally
  incorporate letter-set constraints during search; better fit for text search
  than constrained word generation.

### D2: Daciuk algorithm for DAWG minimization

**Decision:** Use Daciuk et al. (2000) incremental construction algorithm.

The Daciuk algorithm builds a minimal DAWG incrementally from a
lexicographically sorted input. It is well-described, has known correctness
properties, and produces an optimal minimal automaton. Appel-Jacobson (1988)
is an alternative but is oriented toward a specific Scrabble engine and less
general.

**Approach:** Generate all rotations for all words, sort them lexicographically,
then run Daciuk's algorithm to produce the minimized graph in a single pass.
The separator character `+` (Unicode U+002B, sorts before all uppercase ASCII
letters) naturally ensures rotations group correctly during sorting.

**Alternatives considered:**
- *Appel-Jacobson*: Tightly coupled to a specific Scrabble board representation;
  less appropriate for a general-purpose library.
- *Post-hoc minimization*: Build a full trie first, then minimize. Correct but
  uses peak memory proportional to the unminimized trie, which is significantly
  larger.

### D3: Node representation as struct with edge map

**Decision:** Represent each GADDAG node as a struct with a `map[rune]*Node`
for child edges, a terminal flag, and a separator flag (marks the `+` node).

```
Node {
    children  map[rune]*Node
    terminal  bool
}
```

This is idiomatic Go, readable, and straightforward to minimize. At scale
(TWL06, ~180k words), the minimized graph has on the order of tens of thousands
of nodes — small enough that map overhead per node is acceptable.

**Alternatives considered:**
- *Array-indexed children (26-slot array)*: Faster edge lookup (O(1) vs
  O(1) average for map, but lower constant), but wastes memory on sparse nodes
  and couples the structure to the Latin alphabet. An explicit `rune`-keyed map
  keeps the structure encoding-agnostic.
- *Packed edge list (sorted slice)*: More cache-friendly, but complicates
  minimization equality checks and edge insertion during construction.

### D4: Separator character `+`

**Decision:** Use `+` (U+002B) as the GADDAG separator.

It sorts before all uppercase ASCII letters (`A`–`Z` are U+0041–U+005A),
which is required for the Daciuk sort order to work correctly. It is visually
distinct and unlikely to appear in any natural-language word.

### D5: Iterator as range-over-func (Go 1.23+)

**Decision:** The traversal result is an `iter.Seq[string]` (Go 1.23 standard
library `iter` package), enabling idiomatic `for word := range traverse(...)`
usage.

The module targets Go 1.23 from the start. This avoids a future breaking change
to the public API and makes the iterator pattern consistent across all layers.
A `Collect(iter.Seq[string]) []string` helper in the `index` package provides
eager materialization.

### D6: dict.Source interface decouples loading from construction

**Decision:** `index.Build` accepts a `dict.Source`, not a `[]string` directly.

This keeps the loading concern out of the index package and makes it easy to
add new source types (embedded wordlists, HTTP, gzip, etc.) without changing
the index API. `SliceSource` serves as the zero-friction path for tests.

## Risks / Trade-offs

**Minimization correctness is hard to verify by inspection.**
→ Mitigation: property-based tests assert that every word in the input is
reachable via every anchor traversal after construction. This catches
minimization bugs that unit tests on small inputs might miss.

**Memory usage grows with dictionary size during construction (all rotations in
memory before minimization).**
→ Mitigation: rotations are generated and sorted before the Daciuk pass;
peak memory is bounded by the total number of rotation strings. For TWL06 this
is manageable (~tens of MB). Benchmarks will establish the baseline.

**`map[rune]*Node` has higher allocator pressure than array-indexed edges.**
→ Mitigation: the minimized graph is built once and then read-only. GC pressure
during construction is acceptable; steady-state query performance is not
affected by construction allocations.

**GADDAG space is larger than a plain DAWG** (each word stored N times, once
per letter). For TWL06 this is still well within practical limits, and the
query expressiveness justifies the trade-off.

### D7: ASCII-only alphabet, uppercase normalization

**Decision:** Word lists are assumed to be ASCII-encodeable. All words are
normalized to uppercase at load time by `dict.Source` implementations. The
index only stores and matches uppercase A–Z runes.

The `rune`-keyed edge map in D3 is retained for implementation simplicity, but
the valid edge domain is restricted to `[A-Z]` plus the `+` separator. Non-ASCII
runes encountered during construction SHALL cause `index.Build` to return an
error; they are not silently skipped. This keeps the error surface explicit and
makes the ASCII assumption testable.

**Alternatives considered:**
- *Silent skip*: Non-ASCII words are dropped without error. Simpler, but masks
  bad dictionary files.
- *Full Unicode support*: Deferred — would require specifying collation order
  for the Daciuk sort, which is language-dependent and out of scope for v1.

### D8: Two traversal methods — raw and word-level

**Decision:** Expose two traversal methods with distinct contracts:

- `TraverseRaw(anchor rune) iter.Seq[string]` — yields raw rotation strings
  with the `+` separator present (e.g., `"RAC+ED"` for the `R`-anchor rotation
  of `CARED`, where `R` is the first character). The anchor letter is always
  the first character of the rotation; the segment before `+` is the reversed
  prefix ending at the first letter of the original word. Intended for Layer 2
  consumers that need to know the structural boundary between the backward and
  forward segments.
- `Traverse(anchor rune) iter.Seq[string]` — yields complete reconstructed word
  strings with no `+` visible (e.g., `"CARED"`). The high-level convenience
  form; suitable for direct use and for `Collect`.

`Collect` and all other higher-level helpers operate on `Traverse`, never on
`TraverseRaw`.

**Alternatives considered:**
- *Single method with a `Segment{Before, After []rune}` return type*: Structured
  and type-safe, but adds a new type to the public API and makes the common case
  (just give me words) more verbose.
- *Hide `+` entirely, pass boundary info via callback parameter*: Avoids exposing
  the separator character but couples callers to a callback signature, making
  composition with `iter.Seq` awkward.
