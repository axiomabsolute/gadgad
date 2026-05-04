# Benchmark Baselines

Recorded against `testdata/small_words.txt` (57 words).
Hardware: Apple M3 Max, darwin/arm64, Go 1.23.

```
BenchmarkBuild_SmallWords-14      4189    285246 ns/op    288028 B/op    5470 allocs/op
BenchmarkTraverse_E-14           46046     24271 ns/op     36040 B/op     441 allocs/op
```

TWL06 benchmarks (`BenchmarkBuild_TWL06`, `BenchmarkTraverse_TWL06_S`) are skipped
when `testdata/twl06.txt` is absent. Supply the file to establish full-dictionary baselines.
