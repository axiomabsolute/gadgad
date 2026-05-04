## 1. Code Organization

- [x] 1.1 Move `isASCIIUpper` from `index/node.go` to `index/index.go`
- [x] 1.2 Simplify `FileSource.Words` in `dict/file.go`: replace manual `TrimSpace` + blank/comment checks with `normalize(scanner.Text())` first, then check the result
- [x] 1.3 Fix error message in `Build` (`index/index.go:51`): "contains non-ASCII or non-uppercase runes" → "contains characters outside A–Z"

## 2. Build Hot Path Optimizations

- [x] 2.1 Replace `words[:0]` dedup pattern in `Build` with `make([]string, 0, len(words))` (`index/index.go:43`)
- [x] 2.2 Replace `[]rune` conversion in `commonPrefixLen` with direct byte comparison (`index/index.go:207`)
- [x] 2.3 Lift `path`, `edges`, and `runes` allocations out of `insertSuffix`; allocate once in `Build` and pass as buffers, re-slicing to `[:0]` on each call (`index/index.go:182`)

## 3. Query Hot Path Optimizations

- [x] 3.1 Add package-level `traversalOrder []rune` (Separator + A–Z) to `index/index.go` and replace the `append([]rune{Separator}, alphabetRunes()...)` expression in `TraverseRaw` with it (`index/index.go:122`)

## 4. Stats Caching

- [x] 4.1 Add `nodeCount` and `edgeCount` fields to `Index` struct (`index/index.go`)
- [x] 4.2 Populate `nodeCount` and `edgeCount` via `countNodesEdges` at the end of `Build` (`index/index.go`)
- [x] 4.3 Rewrite `Stats()` to return a struct copy from stored fields; remove the `countNodesEdges` call from `Stats()` (`index/index.go`)

## 5. Export RotationToWord

- [x] 5.1 Rename `rotationToWord` → `RotationToWord` in `index/index.go` and update all internal call sites
- [x] 5.2 Remove the duplicate `rotationToWord` from `index/index_test.go` and update `TestTraverse_ConsistencyWithRaw` to use `index.RotationToWord`
- [x] 5.3 Add focused unit tests for `RotationToWord` covering: single-letter word, anchor at first/middle/last position, and consistency with `Traverse` (per spec `rotation-word-reconstruction`)

## 6. Missing Test

- [x] 6.1 Add `TestTraverse_SingleLetterWord` to `index/index_test.go`: build index from `["A"]`, assert `Traverse('A')` returns exactly `["A"]`

## 7. Verification

- [x] 7.1 Run `go test ./...` and confirm all tests pass
- [x] 7.2 Run `go test -bench=. ./index/` and confirm benchmark results are not regressed
