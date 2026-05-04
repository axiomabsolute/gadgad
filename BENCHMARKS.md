# Benchmark Baselines

Hardware: Apple M3 Max, darwin/arm64, Go 1.23.

## Current (post-optimization)

<!-- bench:current:start -->
Recorded at commit `38e9ce7` on 2026-05-04. Run with `-benchmem -benchtime=3s`.

### Index layer (`index/`)

#### Small dictionary (`testdata/small_words.txt`, 60 words)

```
BenchmarkBuild_SmallWords-14      	   12309	    283332 ns/op	  281813 B/op	    4900 allocs/op
BenchmarkTraverse_E-14            	  184414	     18985 ns/op	    5352 B/op	     167 allocs/op
```

#### Full dictionary (`testdata/twl06.txt`, 178691 words)

```
BenchmarkBuild_TWL06-14           	       2	2000905729 ns/op	1340928152 B/op	26885781 allocs/op
BenchmarkTraverse_TWL06_S-14      	      40	  86856894 ns/op	16660779 B/op	  389749 allocs/op
BenchmarkTraverseAt_TWL06_S-14    	     676	   5170809 ns/op	  983280 B/op	   21337 allocs/op
```

### Query layer (`query/`)

#### Full dictionary (`testdata/twl06.txt`, 178691 words)

```
BenchmarkSearch_Crossword_TWL06-14           	    1232	   3027649 ns/op	  378880 B/op	   14120 allocs/op
BenchmarkSearch_Scrabble_TWL06-14            	    1208	   2984447 ns/op	  352024 B/op	   14107 allocs/op
BenchmarkSearch_Anagram_TWL06_Common-14      	      42	  87613333 ns/op	16666992 B/op	  389764 allocs/op
BenchmarkSearch_Anagram_TWL06_Moderate-14    	    1834	   2018465 ns/op	  707544 B/op	   12132 allocs/op
```

_Note: `Anagram_TWL06_Common` (rack `AEINTRS`) exercises the worst-case path through_
_`Traverse` on a high-frequency anchor. `Anagram_TWL06_Moderate` (rack `ABORVXZ`) reflects_
_typical performance when the rack contains at least one rare letter._
<!-- bench:current:end -->

---

## Previous baseline (pre-optimization)

Recorded against the initial Layer 1 implementation (`c8cf565`).

### Small dictionary

```
BenchmarkBuild_SmallWords-14       4189    285246 ns/op    288028 B/op    5470 allocs/op
BenchmarkTraverse_E-14            46046     24271 ns/op     36040 B/op     441 allocs/op
```

TWL06 benchmarks were not recorded for this baseline.

---

## Delta (small dictionary, pre → post)

| Benchmark             | ns/op  | B/op    | allocs/op |
|-----------------------|--------|---------|-----------|
| Build_SmallWords      | -0.9%  | -2.2%   | -10.4%    |
| Traverse_E            | -25.8% | -85.2%  | -62.1%    |

Traverse improvements are driven by the `traversalOrder` package-level var
eliminating a 27-element `[]rune` allocation on every DFS node visit.
Build improvements are modest at small scale; the buffer-reuse gains are
proportionally larger for full-dictionary builds.

---

TWL06 benchmarks (`BenchmarkBuild_TWL06`, `BenchmarkTraverse_TWL06_S`) are skipped
when `testdata/twl06.txt` is absent. Supply the file to establish full-dictionary baselines.
