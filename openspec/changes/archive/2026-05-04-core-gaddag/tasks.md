## 1. Module and Package Scaffold

- [x] 1.1 Initialize Go module (`go mod init github.com/axiomabsolute/gadgad`) targeting Go 1.23
- [x] 1.2 Create empty package stubs: `index/`, `dict/`
- [x] 1.3 Add small curated word list to `testdata/small_words.txt` (50–100 words, hand-chosen for good anchor coverage)

## 2. Dictionary Loading (`dict/`)

- [x] 2.1 Define `Source` interface (`Words() ([]string, error)`) in `dict/dict.go`
- [x] 2.2 Implement `SliceSource` in `dict/slice.go` with normalization (uppercase + trim)
- [x] 2.3 Implement `FileSource` in `dict/file.go`: read UTF-8 line-delimited file, skip blanks and `#` comments, normalize
- [x] 2.4 Write unit tests for `SliceSource`: empty slice, lowercase, mixed-case, whitespace trimming
- [x] 2.5 Write unit tests for `FileSource`: valid file, blank lines, comment lines, missing file error

## 3. GADDAG Node and Edge Representation (`index/`)

- [x] 3.1 Define `Node` struct in `index/node.go`: `children map[rune]*Node`, `terminal bool`
- [x] 3.2 Define `SEPARATOR` constant (`+`, U+002B) in `index/node.go`
- [x] 3.3 Write a helper `rotations(word string) []string` that generates all GADDAG rotations for a word (unit-testable in isolation)
- [x] 3.4 Write unit tests for `rotations`: single-letter word, two-letter word, known multi-letter word, verify separator placement and reversal

## 4. GADDAG Construction and Minimization (`index/`)

- [x] 4.1 Implement unminimized trie insertion of a single rotation string into a node tree (`insertPath`)
- [x] 4.2 Implement `buildRotationList(words []string) []string`: generate and lexicographically sort all rotations for all words
- [x] 4.3 Implement Daciuk incremental DAWG minimization in `index/minimize.go`: `minimize(root *Node)` that collapses equivalent subgraphs
- [x] 4.4 Implement `index.Build(src dict.Source) (*Index, error)` in `index/index.go`, wiring rotation generation → sorted insertion → minimization
- [x] 4.5 Implement `Index.Stats() IndexStats` returning node count, edge count, and word count

## 5. Single-Anchor Traversal (`index/`)

- [x] 5.1 Implement `Index.TraverseRaw(anchor rune) iter.Seq[string]`: yields raw rotation strings with `+` present; anchor letter is always immediately before `+`
- [x] 5.2 Implement `Index.Traverse(anchor rune) iter.Seq[string]`: reconstructs complete words from raw rotations, `+` never appears in output; built on top of `TraverseRaw`
- [x] 5.3 Implement `Collect(seq iter.Seq[string]) []string` helper for eager materialization
- [x] 5.4 Write unit tests for `TraverseRaw`: separator present in every result, anchor letter immediately before `+`, empty result for unknown anchor
- [x] 5.5 Write unit tests for `Traverse` against `testdata/small_words.txt`: every word reachable from every letter position, no duplicates, no `+` in any result
- [x] 5.6 Write unit test asserting consistency: reconstructing words from `TraverseRaw` output equals `Traverse` output for the same anchor
- [x] 5.7 Write property test: for any word in the index and any position P, `Traverse(word[P])` includes that word

## 6. Integration and Benchmarks

- [x] 6.1 Write integration test: load `testdata/small_words.txt` via `dict.FileSource`, build index, assert `Stats().WordCount` matches file word count
- [x] 6.2 Write integration test: for a handful of known words, verify `Traverse` returns expected results using both `FileSource` and `SliceSource`
- [x] 6.3 Add `BenchmarkBuild` in `index/` benchmarking construction from the small wordlist (expand to TWL06 if available at test time)
- [x] 6.4 Add `BenchmarkTraverse` benchmarking a traversal from a common letter (`E`, `S`)
- [x] 6.5 Run benchmarks, record baselines in a `BENCHMARKS.md` or as comments in the bench file
