# Benchmark Baselines

Hardware: Apple M3 Max, darwin/arm64, Go 1.23.

## Current (post-optimization)

<!-- bench:current:start -->
Recorded at commit `2b00af3` on 2026-05-04. Run with `-benchmem -benchtime=1s`.

### Small dictionary (`testdata/small_words.txt`, 60 words)

```
BenchmarkBuild_SmallWords-14    	    4148	    280390 ns/op	  281814 B/op	    4900 allocs/op
BenchmarkTraverse_E-14          	   68082	     17699 ns/op	    5352 B/op	     167 allocs/op
```

### Full dictionary (`testdata/twl06.txt`, 178691 words)

```
BenchmarkBuild_TWL06-14         	       1	1871523584 ns/op	1340932656 B/op	26885780 allocs/op
BenchmarkTraverse_TWL06_S-14    	      14	  80346250 ns/op	16648657 B/op	  389748 allocs/op
```
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
