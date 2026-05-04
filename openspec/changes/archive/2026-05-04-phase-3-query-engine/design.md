## Context

Layer 1 (`index/`) provides `Traverse(anchor)` and `TraverseRaw(anchor)`, which return all words containing a given letter at any position. To support structured queries (fixed positions, rack constraints, length bounds), a Layer 2 query engine is needed.

TWL06 has ~178k words with an average length of 8.87, yielding ~1.58M total rotations. For common anchor letters (e.g. 'S'), `TraverseRaw` would yield ~150k rotation strings — most of which would be discarded by a naïve post-filter. Keeping rotation-format internals out of Layer 2 is an explicit design goal.

## Goals / Non-Goals

**Goals:**
- Add `TraverseAt(anchor, position)` to Layer 1 to prune wrong-position branches during DFS, returning decoded words — Layer 2 never parses rotation strings.
- Implement a `query/` package covering crossword, Scrabble, Wordle, and anagram query shapes.
- Lazy `iter.Seq[string]` output; filters applied inline as words are yielded.
- Auto anchor selection that minimizes search fan-out.

**Non-Goals:**
- In-traversal rack pruning (cursor/guided traversal API) — deferred pending Phase 3 benchmarks.
- Top-level `gadgad.go` convenience wrapper — Phase 4.
- Game adapters (Layer 3) — future.

## Decisions

### Decision: `TraverseAt` in Layer 1, not `TraverseRaw` parsing in Layer 2

**Chosen:** Add `index.Index.TraverseAt(anchor rune, position int) iter.Seq[string]` to Layer 1. During DFS it only follows branches where traversal depth equals `position`; returns fully decoded words.

**Rejected:** Layer 2 calls `TraverseRaw` and parses prefix lengths to determine anchor position. This couples Layer 2 to the GADDAG rotation format (`prefix.len - 1 == anchor_position`), which is a Layer 1 implementation detail. Any internal encoding change would silently break Layer 2.

**Rationale:** `TraverseAt` keeps the rotation encoding entirely within Layer 1. It also prunes ~8/9 of rotations for the average anchor (those at the wrong position) without generating or allocating their strings — concrete efficiency gain for common-letter anchors on large dictionaries.

### Decision: Rack constraint applied post-traversal per word

**Chosen:** After `TraverseAt` yields a candidate word, Layer 2 counts the letters needed for all `Free` slots and verifies they are available in the rack (with multiplicity).

**Rejected:** In-traversal pruning via a cursor/guided traversal API that prunes edges at each free slot based on remaining rack tiles.

**Rationale:** After position + fixed-slot filters, the remaining candidate set is small. Per-word rack checks are O(word_length) and the constant factor is low. In-traversal pruning adds significant API complexity for uncertain gain. Phase 3 benchmarks will provide evidence if this needs revisiting.

### Decision: `Free` vs. `Any` slot distinction

**Chosen:** Two slot kinds for unconstrained positions:
- `Free` — any letter, consumed from the `Available` rack. Used for Scrabble tiles.
- `Any` — any letter, ignores the `Available` set entirely. Used for crossword blanks where a letter is known to exist but isn't from the player's rack.

**Rationale:** Without this distinction, crossword queries (blank = "any letter, including ones I don't have") would require setting `Available = nil` globally, which removes rack constraints for all free slots — breaking mixed patterns like `__A??ED` where some blanks are rack tiles and some are known-letter positions.

### Decision: Anchor auto-selection uses letter frequency heuristic

**Chosen:** Among all `Fixed` slots, pick the letter with the lowest frequency in English (or a static TWL06-derived frequency table). For all-`Free` patterns (anagram), pick the rarest letter in `Available`.

**Rationale:** Rarer letters have fewer rotations in the GADDAG, so `TraverseAt` yields fewer candidates before filtering. This is a static heuristic with no runtime cost beyond a map lookup.

### Decision: Lazy iterator via `iter.Seq[string]` (range-over-func)

**Chosen:** `query.Search` returns `iter.Seq[string]` directly. An `Iterator` wrapper type with `Collect() []string` is provided for callers who prefer eager materialization.

**Rationale:** Consistent with how Layer 1 (`Traverse`, `TraverseRaw`) already works. Go 1.23+ range-over-func is already a stated dependency of this library.

## Risks / Trade-offs

- **Rack post-filter may be slow for all-Free queries with large Available sets** → Deferred; benchmarks in Phase 3 will measure. A guided traversal API is the planned mitigation if needed.
- **Auto anchor selection is a heuristic, not optimal** → For adversarial patterns (all rare letters), a worse anchor might be chosen. The `Manual(position)` override is the escape hatch.
- **`TraverseAt` for position out of range (≥ word length) silently yields nothing** → This is correct behavior; callers constructing patterns from valid board positions should not hit this case.

## Open Questions

- Should `query.Search` accept `*index.Index` directly, or an interface (e.g., `Traverser`) to allow mocking in tests? Given the plan's emphasis on testability, an interface may be cleaner but adds indirection. Recommend concrete type for Phase 3 and revisit if needed.
