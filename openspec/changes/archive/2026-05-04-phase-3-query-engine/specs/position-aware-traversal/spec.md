## ADDED Requirements

### Requirement: TraverseAt returns words with anchor at given position
`index.Index` SHALL expose `TraverseAt(anchor rune, position int) iter.Seq[string]` that yields all words in the index where `anchor` appears at zero-indexed position `position`.

#### Scenario: Known anchor and position
- **WHEN** `TraverseAt('A', 1)` is called on an index containing "CARE"
- **THEN** "CARE" is yielded (C=0, A=1, R=2, E=3)

#### Scenario: Anchor appears at multiple positions in the dictionary
- **WHEN** `TraverseAt('A', 0)` is called on an index containing "ACE" and "CARE"
- **THEN** "ACE" is yielded (A at position 0) and "CARE" is NOT yielded (A is at position 1, not 0)

#### Scenario: Unknown anchor yields nothing
- **WHEN** `TraverseAt('Z', 0)` is called on an index with no words containing 'Z'
- **THEN** the iterator yields no values

#### Scenario: Position out of range yields nothing
- **WHEN** `TraverseAt('A', 99)` is called
- **THEN** the iterator yields no values without error

### Requirement: TraverseAt returns decoded words, not rotation strings
The strings yielded by `TraverseAt` SHALL be fully formed words with no `+` separator character.

#### Scenario: Returned words contain no separator
- **WHEN** any value is yielded by `TraverseAt`
- **THEN** that value does not contain the `+` character

### Requirement: TraverseAt yields no duplicates
For a given `(anchor, position)` pair, each matching word SHALL appear at most once in the iteration.

#### Scenario: Word with anchor letter appearing once at the given position
- **WHEN** `TraverseAt('A', 1)` is called on an index containing "CARE"
- **THEN** "CARE" appears exactly once across the full iteration

### Requirement: TraverseAt is lazy
`TraverseAt` SHALL return an `iter.Seq[string]` that only performs traversal work as values are consumed by the caller.

#### Scenario: Early termination stops traversal
- **WHEN** the caller breaks out of the range loop after receiving the first result
- **THEN** no further GADDAG nodes beyond that point are visited
