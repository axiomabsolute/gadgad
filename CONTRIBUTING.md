# Contributing

## Prerequisites

- Go 1.23 or later

## Common Commands

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

# Run benchmarks (small wordlist)
go test ./index/ -bench=. -benchmem -run='^$'

# Run benchmarks against TWL06 (supply testdata/twl06.txt first)
go test ./index/ -bench=. -benchmem -run='^$'

# Check for vet issues
go vet ./...
```

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
