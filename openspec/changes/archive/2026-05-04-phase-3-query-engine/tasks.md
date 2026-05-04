## 1. Layer 1 Addition: TraverseAt

- [x] 1.1 Add `TraverseAt(anchor rune, position int) iter.Seq[string]` to `index/index.go` — DFS that only follows branches at depth == position, yields decoded words, no duplicates
- [x] 1.2 Write unit tests for `TraverseAt` in `index/index_test.go`: known position, anchor at multiple positions, unknown anchor, out-of-range position, no separator in results, no duplicates, early termination
- [x] 1.3 Add `BenchmarkTraverseAt_TWL06_S` benchmark to `index/bench_test.go` (anchor='S', position=0) to establish a baseline

## 2. Query Types

- [x] 2.1 Create `query/pattern.go` with `Slot` (Fixed/Free/Any), `LengthConstraint` (Exact/Min/Max/Unconstrained), `LetterSet` (multiset backed by `[26]int`), `AnchorStrategy` (Auto/Manual), and `Pattern` types
- [x] 2.2 Add a `Pattern.Validate() error` method that rejects Manual anchors pointing at Free/Any slots

## 3. Anchor Selection

- [x] 3.1 Create `query/anchor.go` with a static letter-frequency table (derived from English/TWL06 ordering) and `selectAnchor(p Pattern) (letter rune, position int, err error)`
- [x] 3.2 Handle the all-Free case: pick rarest letter from `Available`; return error if Available is also nil/empty
- [x] 3.3 Write unit tests in `query/anchor_test.go`: rarest fixed letter chosen, single fixed slot, all-Free picks rarest from Available, Manual passthrough, Manual on Free slot errors, deterministic tiebreak

## 4. Search Engine

- [x] 4.1 Create `query/engine.go` with `Search(idx *index.Index, p Pattern) iter.Seq[string]` — calls `selectAnchor`, calls `idx.TraverseAt`, applies length/fixed-slot/rack/exclusion filters inline
- [x] 4.2 Implement rack consumption check: extract letters needed for `Free` slots, verify they are a sub-multiset of `Available` (using `LetterSet` arithmetic)
- [x] 4.3 Implement `Any` slot bypass: `Any` slots pass without consuming from Available
- [x] 4.4 Implement exclusion check: reject words where any `Free` slot letter is in `Excluded`
- [x] 4.5 Implement deduplication within `Search` for patterns where anchor letter appears multiple times in a word (use a `seen` map)

## 5. Iterator

- [x] 5.1 Create `query/iterator.go` with `Iterator` type wrapping `iter.Seq[string]` and `Collect() []string`
- [x] 5.2 Update `Search` to return `iter.Seq[string]` directly (not `Iterator`); Iterator is constructed by callers who need Collect

## 6. Tests: All Four Query Modes

- [x] 6.1 Write crossword mode test in `query/engine_test.go`: pattern `[Any,Any,Fixed('A'),Any,Any,Fixed('E'),Fixed('D')]`, Available=nil, verify known words returned and no non-matching words included
- [x] 6.2 Write Scrabble mode test: pattern with Fixed and Free slots, Available rack, verify rack constraint respected (including multiset depletion)
- [x] 6.3 Write Wordle hard-mode test: five-letter pattern with Fixed('E') at position 4, Available = known letters, Excluded = gray letters
- [x] 6.4 Write anagram mode test: all Free, Exact(5), Available={A,E,R,T,S}, verify all valid anagrams returned and no extras
- [x] 6.5 Write no-results test: pattern that matches zero words returns empty iteration without error

## 7. Benchmarks

- [x] 7.1 Add `BenchmarkSearch_Crossword_TWL06` to `query/bench_test.go`: a 7-slot crossword pattern on TWL06, measure query latency
- [x] 7.2 Add `BenchmarkSearch_Scrabble_TWL06`: 7-slot pattern with 2 fixed and 5 free, Available={common rack}, measure query latency
- [x] 7.3 Add `BenchmarkSearch_Anagram_TWL06`: all-Free length-7, Available={A,E,R,T,I,N,S}, measure query latency
- [x] 7.4 Run all benchmarks, record results, confirm no query exceeds 5ms on TWL06 as a soft ceiling; document findings
  <!-- Results (Apple M3 Max): Crossword=3.1ms ✓  Scrabble=3.0ms ✓  Anagram=77.5ms ✗
       Anagram exceeds ceiling: Traverse('S') on TWL06 returns ~100k+ candidates.
       Fix: for Exact-length anagram, iterate TraverseAt(letter,0..N-1) instead of Traverse. Deferred. -->
