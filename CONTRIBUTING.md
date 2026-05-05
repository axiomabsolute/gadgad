# Contributing

## Prerequisites

- Go 1.23 or later
- [pre-commit](https://pre-commit.com/#installation) (`brew install pre-commit`)
- [golangci-lint](https://golangci-lint.run/welcome/install/) v2+ (`brew install golangci-lint`)

After cloning, run once to install the git hooks:

```bash
make setup
```

## Common Commands

```bash
make setup   # install pre-commit hooks (run once after clone)
make build   # go build ./...
make test    # go test ./...
make lint    # golangci-lint run ./...
```

Raw commands if preferred:

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for a specific package
go test ./index/
go test ./dict/
go test ./query/

# Run benchmarks (without updating BENCHMARKS.md)
go test ./index/ ./query/ -bench=. -benchmem -run='^$'

# Check for vet issues
go vet ./...
```

## Benchmarks

Run `scripts/bench.sh` to execute the full benchmark suite and update the
Current section of `BENCHMARKS.md` with fresh results:

```bash
./scripts/bench.sh
```

The script records the git commit hash and date alongside the numbers so results
are always traceable. The Previous baseline and Delta sections in `BENCHMARKS.md`
are not touched — update those manually when a significant before/after comparison
is worth preserving.

To run a faster pass during development without updating the file:

```bash
go test ./index/ ./query/ -bench=. -benchmem -run='^$'
```

**Full dictionary benchmarks** are skipped automatically when `testdata/twl06.txt`
is absent. Supply the file to include them. The default benchmark time is 5 seconds
per benchmark; override with the `BENCHTIME` environment variable:

```bash
BENCHTIME=1s ./scripts/bench.sh   # quick check
BENCHTIME=10s ./scripts/bench.sh  # higher-precision run
```

The query layer has two anagram benchmarks that cover different performance
characteristics:

| Benchmark | Rack | Notes |
|---|---|---|
| `Anagram_TWL06_Common` | `AEINTRS` | Worst case: common letters, anchor `S` appears in most words |
| `Anagram_TWL06_Moderate` | `ABORVXZ` | Typical case: rare anchor letter, fast path |

## Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/).
The convention is chosen to align with [git-cliff](https://git-cliff.org/) for
automated changelog generation, even though git-cliff is not currently in use.
Adopting the format now avoids retroactive cleanup if it is added later.

Format:

```
<type>(<optional scope>): <description>
```

Types used in this project:

| Type | When to use |
|------|-------------|
| `feat` | new capability |
| `fix` | bug correction |
| `test` | adding or updating tests only |
| `docs` | documentation only (README, comments, PLAN.md) |
| `refactor` | restructuring without behavior change |
| `perf` | measurable performance improvement |
| `chore` | maintenance — deps, .gitignore, CI config, tooling |

Scope is optional but encouraged when the change is confined to a package or
layer (e.g. `index`, `dict`, `query`).

Examples:

```
feat(index): add TraverseRaw with raw rotation output
fix(index): correct Daciuk replaceOrRegister slice bounds
test(dict): add FileSource missing-file error case
docs: add GADDAG algorithm explainer to README
perf(index): replace map edges with sorted slice for hot path
chore: add .gitignore and go.mod scaffold
```

Breaking changes should include `!` after the type and a `BREAKING CHANGE:`
footer:

```
feat(index)!: change Build signature to accept context

BREAKING CHANGE: Build now takes a context.Context as its first argument.
```
