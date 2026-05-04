# gadgad

A Go library for efficient word search in text-based language games. Given a
dictionary of words, `gadgad` builds an optimized index that answers queries
like:

> *"What 7-letter words match the pattern `??A??ED` using letters from the rack
> `[P, R, I, Z, S, O, U]`?"*

Useful for Scrabble engines, crossword solvers, Wordle analysis, and any
application that needs to search a word list by partial positional information.

```go
idx, _ := gadgad.Load(dict.File("twl06.txt"))
for word := range gadgad.Search(idx, "??A??ED", gadgad.WithAvailable("PRIZSO")) {
    fmt.Println(word)
}
```

**Status:** under active development. See [PLAN.md](PLAN.md) for the roadmap.

---

## Features

- Pattern matching with fixed and wildcard positions (`??A??ED`)
- Available-letter constraints (Scrabble rack: only use tiles you hold)
- Exclusion constraints (Wordle gray tiles: letter cannot appear)
- Anagram search (find all words using exactly a given set of letters)
- Lazy iterator (`range`-over-func, Go 1.23+) and eager `Collect() []string`
- Multiple dictionary sources: text file or `[]string`
- Layered API — use the convenience wrapper or reach into the index directly

---

## Installation

```
go get github.com/axiomabsolute/gadgad
```

Requires Go 1.23 or later.

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/axiomabsolute/gadgad"
    "github.com/axiomabsolute/gadgad/dict"
)

func main() {
    // Load a dictionary from a line-delimited text file
    idx, err := gadgad.Load(dict.File("words.txt"))
    if err != nil {
        panic(err)
    }

    // Crossword: find all 7-letter words matching a positional pattern
    fmt.Println("Crossword candidates:")
    for word := range gadgad.Search(idx, "??A??ED") {
        fmt.Println(" ", word)
    }

    // Scrabble: same pattern, but only use letters from the rack
    fmt.Println("Scrabble candidates (rack: PRIZSO):")
    for word := range gadgad.Search(idx, "??A??ED", gadgad.WithAvailable("PRIZSO")) {
        fmt.Println(" ", word)
    }

    // Wordle: 5-letter words ending in E, not containing T or R
    fmt.Println("Wordle candidates:")
    for word := range gadgad.Search(idx, "????E", gadgad.WithExcluded("TR")) {
        fmt.Println(" ", word)
    }

    // Anagram: all words using exactly these letters
    fmt.Println("Anagrams of LISTEN:")
    for word := range gadgad.Search(idx, "??????", gadgad.WithAvailable("LISTEN"), gadgad.AnagramOnly()) {
        fmt.Println(" ", word)
    }
}
```

---

## Dictionary Sources

```go
// From a line-delimited text file (one word per line, # for comments)
dict.File("twl06.txt")

// From a string slice (useful for tests and embedded wordlists)
dict.Slice([]string{"APPLE", "APPLY", "AMPLE"})
```

Words are normalized to uppercase automatically.

---

## Query Patterns

A pattern is a string where each character is either a fixed letter or `?`
(wildcard). The length of the pattern constrains the result length exactly.

| Pattern | Meaning |
|---------|---------|
| `??A??ED` | 7-letter word with A at position 3, E at 6, D at 7 |
| `?????` | any 5-letter word |
| `RE????` | 6-letter word starting with RE |
| `??????` | any 6-letter word |

### Query Options

| Option | Description |
|--------|-------------|
| `WithAvailable(letters)` | wildcard positions may only use these letters (each consumed once) |
| `WithExcluded(letters)` | wildcard positions must not contain these letters |
| `WithAnchor(position)` | manually specify the anchor position for traversal |
| `AnagramOnly()` | all available letters must be used exactly once |

---

## Layered API

The library exposes three layers. Most users only need the top.

```
gadgad.go          ← convenience: Load, Search, options
query/             ← Pattern type, query engine, iterator
index/             ← GADDAG graph, Build, Traverse (traversal primitives)
dict/              ← Source interface, FileSource, SliceSource
```

To work directly with the index:

```go
import "github.com/axiomabsolute/gadgad/index"

idx, _ := index.Build(words)
iter := idx.Traverse('A', backwardConstraint, forwardConstraint)
for word := range iter {
    fmt.Println(word)
}
```

---

## How It Works: The GADDAG

`gadgad` is built around a **GADDAG** — a data structure designed by Steven
Gordon (1994) specifically for Scrabble move generation, but applicable to any
word search problem involving positional constraints.

### The problem with a standard trie

A trie efficiently answers "what words start with X?" but struggles with
"what words have X at position 4?" — you'd have to scan the entire dictionary.
A reverse trie helps with suffix queries, but neither handles the general case
of a letter at an arbitrary position.

### Rotations and the `+` separator

GADDAG solves this by storing every word *multiple times*, once for each
possible anchor position, using a reserved `+` character as a separator.

For the word `CARED`:

```
Anchor at C (pos 0):  C + A R E D
Anchor at A (pos 1):  A C + R E D     ← letters before anchor are reversed
Anchor at R (pos 2):  R A C + E D
Anchor at E (pos 3):  E R A C + D
Anchor at D (pos 4):  D E R A C +
```

All five rotations are inserted into a single graph. Common substrings across
all words in the dictionary are shared (the DAWG minimization step), keeping
memory usage manageable.

### Querying

To find all words with `A` at position 1 (as in the `??A??ED` example with `A`
as the anchor):

```
         backward          +          forward
    ┌────────────────┐     │     ┌──────────────────┐
    │  2 free slots  │─────┼────▶│  free, free, E, D │
    │  (from rack)   │     │     │  (rack + fixed)   │
    └────────────────┘     │     └──────────────────┘
           ▲               │
           │           separator
      start: 'A'
```

Traversal prunes branches that violate constraints — if a rack letter is
exhausted, that branch is abandoned. The result is a set of candidate words
without ever scanning the full dictionary.

### DAWG minimization

After generating all rotations, the resulting trie is minimized into a
**Directed Acyclic Word Graph** using Daciuk's incremental construction
algorithm. This collapses common suffixes across the dictionary into shared
nodes, significantly reducing memory usage compared to an unminimized trie.

---

## References

- Gordon, S. (1994). *A faster Scrabble move generation algorithm.* Software —
  Practice and Experience, 24(2), 219–232.
- Daciuk, J., Mihov, S., Watson, B. W., & Watson, R. E. (2000). *Incremental
  construction of minimal acyclic finite-state automata.* Computational
  Linguistics, 26(1), 3–16.
- Appel, A. W., & Jacobson, G. J. (1988). *The world's fastest Scrabble
  program.* Communications of the ACM, 31(5), 572–578.

---

## License

MIT
