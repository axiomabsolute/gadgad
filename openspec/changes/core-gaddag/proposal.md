## Why

`gadgad` is a new Go library for word search in text-based language games, and
it needs a foundation before any user-facing API can be built. The GADDAG index
is that foundation — it is the data structure that makes efficient
partial-pattern word search possible, and nothing in the query engine or
convenience API can exist without it.

## What Changes

- New Go module `github.com/axiomabsolute/gadgad` initialized with `go.mod`
- New package `index/`: GADDAG graph construction from a `[]string` word list,
  internal node/edge representation, DAWG minimization, and single-anchor
  traversal primitives
- New package `dict/`: `Source` interface plus `FileSource` (line-delimited
  text file) and `SliceSource` (`[]string`) implementations, wired into
  `index.Build()`
- Benchmark suite establishing construction-time and memory baselines against a
  standard wordlist (TWL06, ~180k words)
- Small curated word list checked into `testdata/` for deterministic unit tests

## Capabilities

### New Capabilities

- `gaddag-index`: Build a minimized GADDAG from a collection of words and
  expose single-anchor traversal primitives; the foundational index layer that
  all higher-level query and game logic will build upon
- `dict-source`: Load a word collection from a pluggable source (line-delimited
  file or `[]string`), with normalization, into a form suitable for
  `index.Build()`

### Modified Capabilities

## Impact

- Introduces the Go module and establishes the package structure described in
  `PLAN.md`
- No existing code is affected (greenfield)
- Go 1.23+ required (range-over-func iterator in later phases; module targets
  this from the start to avoid a future breaking change)
- `testdata/small_*.txt` wordlists are checked in; full wordlists (TWL06,
  SOWPODS) are gitignored and must be supplied by the user at runtime or test
  time
