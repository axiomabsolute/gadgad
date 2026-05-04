## ADDED Requirements

### Requirement: Build index from word collection
The system SHALL construct a GADDAG index from a `dict.Source` by generating
all letter-position rotations for each word and minimizing the resulting graph
into a DAWG. The index SHALL be immutable after construction.

#### Scenario: Build from non-empty word list
- **WHEN** `index.Build` is called with a `Source` returning one or more words
- **THEN** it returns a non-nil `*Index` and a nil error

#### Scenario: Build from empty word list
- **WHEN** `index.Build` is called with a `Source` returning zero words
- **THEN** it returns a non-nil `*Index` (empty) and a nil error

#### Scenario: Duplicate words are deduplicated
- **WHEN** the word list contains the same word more than once
- **THEN** the index behaves identically to one built with a deduplicated list

#### Scenario: Source error is propagated
- **WHEN** the `Source` returns an error from `Words()`
- **THEN** `index.Build` returns a nil index and a non-nil error

### Requirement: Every word is reachable via every anchor
For every word W in the dictionary and every letter position P in W, the index
SHALL return W when traversed starting from the letter at position P.

#### Scenario: Word reachable from its first letter
- **WHEN** `Traverse` is called with the first letter of a word W
- **THEN** W appears in the traversal results

#### Scenario: Word reachable from a middle letter
- **WHEN** `Traverse` is called with a letter that appears in the middle of word W
- **THEN** W appears in the traversal results

#### Scenario: Word reachable from its last letter
- **WHEN** `Traverse` is called with the last letter of a word W
- **THEN** W appears in the traversal results

### Requirement: Word-level traversal (high-level)
The index SHALL expose a `Traverse(anchor rune) iter.Seq[string]` method that
returns all complete words in the dictionary containing the given letter at any
position. The separator character `+` SHALL NOT appear in any yielded string.

#### Scenario: Traversal from a common letter returns multiple words
- **WHEN** `Traverse` is called with a letter that appears in many words
- **THEN** all words containing that letter are present in the results, with no duplicates

#### Scenario: Traversal from a letter not in the dictionary returns empty
- **WHEN** `Traverse` is called with a rune that appears in no dictionary word
- **THEN** the resulting sequence is empty

#### Scenario: Traversal results contain no duplicates
- **WHEN** `Traverse` is called for any anchor letter
- **THEN** each word in the dictionary appears at most once in the results

#### Scenario: Traversal results never contain the separator character
- **WHEN** `Traverse` is called for any anchor letter
- **THEN** no yielded string contains the `+` character

### Requirement: Raw rotation traversal (low-level)
The index SHALL expose a `TraverseRaw(anchor rune) iter.Seq[string]` method
that yields raw GADDAG rotation strings, with the `+` separator present,
for all rotations in the index whose anchor letter matches. This is the
primitive form intended for Layer 2 consumers that require structural knowledge
of the backward/forward boundary around the anchor.

#### Scenario: Raw traversal exposes the separator character
- **WHEN** `TraverseRaw` is called with a letter present in the dictionary
- **THEN** every yielded string contains exactly one `+` character

#### Scenario: Raw traversal anchor letter is the first character of each rotation
- **WHEN** `TraverseRaw` is called with anchor letter A
- **THEN** every yielded string begins with A (per Gordon 1994: rotation i = w_i, w_{i-1}, ..., w_1, '+', w_{i+1}, ..., w_n)

#### Scenario: Raw traversal is consistent with word-level traversal
- **WHEN** `TraverseRaw` is called for anchor A and rotation strings are
  reconstructed into complete words by reversing the segment before `+` and
  concatenating with the segment after `+`
- **THEN** the resulting word set equals the result of `Traverse` for the same anchor

#### Scenario: Raw traversal from a letter not in the dictionary returns empty
- **WHEN** `TraverseRaw` is called with a rune that appears in no dictionary word
- **THEN** the resulting sequence is empty

### Requirement: Index statistics
The index SHALL expose a `Stats() IndexStats` method returning node count, edge
count, and word count.

#### Scenario: Stats reflect dictionary size
- **WHEN** `Stats()` is called on an index built from N distinct words
- **THEN** `Stats().WordCount` equals N
