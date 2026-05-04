## Why

The GADDAG index (Layer 1) can find all words containing a given letter, but callers have no way to express structured queries — fixed positions, available tile sets, length constraints, or exclusions. A query engine (Layer 2) is needed to translate game-meaningful patterns into index traversals and return results lazily.

## What Changes

- **New**: `index.Index.TraverseAt(anchor rune, position int) iter.Seq[string]` — position-aware traversal that prunes wrong-position branches during DFS rather than post-hoc, and returns decoded words so Layer 2 never touches the rotation encoding.
- **New**: `query/` package with `Pattern`, `Slot`, `LengthConstraint`, `LetterSet`, `AnchorStrategy` types.
- **New**: `query.Search(idx, pattern) iter.Seq[string]` — translates a `Pattern` into a `TraverseAt` call, applies slot/length/rack/exclusion filters inline as words are yielded.
- **New**: `query.Iterator` type wrapping `iter.Seq[string]` with a `.Collect() []string` convenience method.

## Capabilities

### New Capabilities

- `position-aware-traversal`: `index.TraverseAt` — traverse the GADDAG from a given anchor letter constrained to a specific word position, returning decoded words.
- `pattern-search`: `query.Search` — structured word search with fixed slots, free slots, available tile set, exclusion set, and length constraints; lazy iterator output.
- `anchor-selection`: Auto-selection of the anchor letter and position from a pattern that minimizes search fan-out (prefer rarest fixed letter).

### Modified Capabilities

_(none — no existing spec-level requirements are changing)_

## Impact

- `index/index.go`: add `TraverseAt` method; existing `Traverse` and `TraverseRaw` unchanged.
- New `query/` package: `pattern.go`, `engine.go`, `anchor.go`, `iterator.go`.
- No changes to `dict/` or the public `Build` API.
- Go 1.23+ required (range-over-func) — already established by Layer 1.
