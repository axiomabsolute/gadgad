## ADDED Requirements

### Requirement: Pattern type encodes structured word constraints
The `query` package SHALL provide a `Pattern` type with the following fields:
- `Slots []Slot` — one entry per letter position; each slot is `Fixed(rune)`, `Free`, or `Any`
- `Length LengthConstraint` — `Exact(n)`, `Min(n)`, `Max(n)`, or `Unconstrained`
- `Available LetterSet` — multiset of runes available to fill `Free` slots; nil means no rack constraint
- `Excluded LetterSet` — set of runes that MUST NOT appear in `Free` slots
- `Anchor AnchorStrategy` — `Auto` or `Manual(position)`

#### Scenario: Pattern construction
- **WHEN** a `Pattern` is created with slots `[Fixed('A'), Free, Fixed('E')]` and `Length: Exact(3)`
- **THEN** the pattern is valid and represents the three-letter word shape `A_E`

### Requirement: Fixed slots must match exactly
A candidate word SHALL be rejected if any `Fixed(r)` slot does not match the corresponding character in the word.

#### Scenario: Fixed slot matches
- **WHEN** the pattern has `Fixed('E')` at position 2 and the candidate word has 'E' at position 2
- **THEN** the word passes the fixed-slot check for that position

#### Scenario: Fixed slot mismatch
- **WHEN** the pattern has `Fixed('E')` at position 2 and the candidate word has 'A' at position 2
- **THEN** the word is excluded from results

### Requirement: Free slots consume from Available rack
When `Available` is non-nil, every `Free` slot SHALL be filled by a letter present in the remaining `Available` multiset, consuming it. A candidate word is rejected if the letters needed for all `Free` slots cannot be satisfied by the `Available` multiset.

#### Scenario: Rack satisfied
- **WHEN** the pattern has two `Free` slots, `Available` contains `{R, E}`, and the candidate word uses 'R' and 'E' for those slots
- **THEN** the word is included in results

#### Scenario: Rack exhausted
- **WHEN** the pattern has two `Free` slots both requiring 'R', `Available` contains only one 'R'
- **THEN** the word is excluded from results

#### Scenario: Available is nil — no rack constraint
- **WHEN** `Available` is nil and the pattern has `Free` slots
- **THEN** any letter is accepted for those slots regardless of what it is

### Requirement: Any slots accept any letter and ignore Available
A `Any` slot SHALL match any letter in the candidate word and SHALL NOT consume from the `Available` multiset.

#### Scenario: Any slot matches any letter
- **WHEN** a slot is `Any` and the candidate word has 'Q' at that position
- **THEN** 'Q' is accepted and the `Available` multiset is unchanged

### Requirement: Excluded set forbids letters in Free slots
A candidate word SHALL be rejected if any `Free` slot is filled by a letter present in the `Excluded` set.

#### Scenario: Excluded letter in Free slot
- **WHEN** `Excluded` contains 'S' and a `Free` slot is filled by 'S'
- **THEN** the word is excluded from results

#### Scenario: Excluded letter in Fixed slot is not affected
- **WHEN** `Excluded` contains 'S' and 'S' appears only in a `Fixed('S')` slot
- **THEN** the word is NOT excluded on that basis (exclusions apply only to Free slots)

### Requirement: Length constraint filters by word length
The `LengthConstraint` SHALL be applied to candidate words before slot-level checks.

#### Scenario: Exact length match
- **WHEN** `Length` is `Exact(5)` and the candidate word has 5 letters
- **THEN** the word passes the length check

#### Scenario: Exact length mismatch
- **WHEN** `Length` is `Exact(5)` and the candidate word has 4 letters
- **THEN** the word is excluded

#### Scenario: Unconstrained length
- **WHEN** `Length` is `Unconstrained`
- **THEN** words of any length pass the length check

### Requirement: Search returns a lazy iterator
`query.Search(idx *index.Index, p Pattern) iter.Seq[string]` SHALL return a lazy `iter.Seq[string]`. Filtering (length, fixed slots, rack, exclusions) SHALL be applied inline as each candidate word is yielded from `TraverseAt`.

#### Scenario: Results are lazy
- **WHEN** the caller breaks after the first result
- **THEN** remaining candidates are not evaluated

### Requirement: Search results contain no duplicates
A given word SHALL appear at most once in the results of `query.Search`, regardless of how many positions the anchor letter appears in the word.

#### Scenario: Word with repeated anchor letter
- **WHEN** searching with anchor auto-selected from a pattern that matches "BANANA"
- **THEN** "BANANA" appears at most once in the results

### Requirement: Iterator.Collect materializes results eagerly
`Iterator.Collect() []string` SHALL consume the full `iter.Seq[string]` and return all results as a slice.

#### Scenario: Collect returns all results
- **WHEN** `Collect()` is called on a search that matches three words
- **THEN** the returned slice contains exactly those three words

### Requirement: Crossword query mode
A pattern with `Fixed` and `Any` slots and `Available: nil` SHALL correctly return all dictionary words matching the fixed positions.

#### Scenario: Crossword pattern
- **WHEN** the pattern is `[Any, Any, Fixed('A'), Any, Any, Fixed('E'), Fixed('D')]` with no rack
- **THEN** only words matching `_ _ A _ _ E D` are returned

### Requirement: Scrabble query mode
A pattern with `Fixed` and `Free` slots and a non-nil `Available` rack SHALL return only words where free slots are satisfiable from the rack.

#### Scenario: Scrabble pattern with rack
- **WHEN** the pattern has `Fixed('A')` at position 2, `Free` elsewhere, `Available: {P,R,I,Z,S,O,U}`
- **THEN** only words containing 'A' at position 2 whose remaining letters can be drawn from the rack are returned

### Requirement: Wordle query mode
A pattern with some `Fixed` slots, some `Free` slots, non-nil `Available` (known good letters), and non-nil `Excluded` (gray letters) SHALL return only words satisfying all constraints.

#### Scenario: Wordle hard mode pattern
- **WHEN** the pattern has `Fixed('E')` at position 4, `Free` elsewhere, `Available: {known letters}`, `Excluded: {gray letters}`
- **THEN** only five-letter words ending in 'E' that use the known letters and avoid gray letters are returned

### Requirement: Anagram query mode
A pattern with all `Free` slots, `Length: Exact(n)`, and a non-nil `Available` set SHALL return all words of length n whose letters are a sub-multiset of `Available`.

#### Scenario: Anagram pattern
- **WHEN** the pattern has 5 `Free` slots, `Length: Exact(5)`, `Available: {A,E,R,T,S}`
- **THEN** all 5-letter words formable from those letters are returned
