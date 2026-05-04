## ADDED Requirements

### Requirement: Source interface
The system SHALL provide a `dict.Source` interface with a single method
`Words() ([]string, error)` that returns a collection of words suitable for
passing to `index.Build`.

#### Scenario: Source returns words on success
- **WHEN** a `Source` implementation can successfully read its underlying data
- **THEN** `Words()` returns a non-nil slice and a nil error

#### Scenario: Source propagates errors
- **WHEN** a `Source` implementation encounters an error reading its underlying data
- **THEN** `Words()` returns nil and a non-nil error

### Requirement: FileSource loads words from a line-delimited text file
`dict.File(path string) Source` SHALL return a `Source` that reads words from a
UTF-8 text file with one word per line.

#### Scenario: Valid file with words is loaded
- **WHEN** `Words()` is called on a `FileSource` pointing to a readable file
- **THEN** it returns one entry per non-blank, non-comment line

#### Scenario: Blank lines are skipped
- **WHEN** the file contains one or more blank lines
- **THEN** blank lines do not appear as entries in the returned slice

#### Scenario: Comment lines are skipped
- **WHEN** a line begins with `#`
- **THEN** that line does not appear as an entry in the returned slice

#### Scenario: Missing file returns an error
- **WHEN** `Words()` is called on a `FileSource` pointing to a non-existent path
- **THEN** it returns a non-nil error

### Requirement: SliceSource wraps a string slice
`dict.Slice(words []string) Source` SHALL return a `Source` that yields the
provided slice as its word collection.

#### Scenario: SliceSource returns the provided words
- **WHEN** `Words()` is called on a `SliceSource`
- **THEN** it returns the same words that were passed to `dict.Slice`

#### Scenario: SliceSource with empty slice returns empty result
- **WHEN** `Words()` is called on a `SliceSource` wrapping an empty slice
- **THEN** it returns an empty slice and a nil error

### Requirement: Word normalization
All `Source` implementations SHALL normalize words to uppercase and trim leading
and trailing whitespace before returning them.

#### Scenario: Lowercase input is uppercased
- **WHEN** the source contains a word in lowercase (e.g., `"apple"`)
- **THEN** `Words()` returns it as `"APPLE"`

#### Scenario: Mixed-case input is uppercased
- **WHEN** the source contains a word in mixed case (e.g., `"Apple"`)
- **THEN** `Words()` returns it as `"APPLE"`

#### Scenario: Surrounding whitespace is trimmed
- **WHEN** a word has leading or trailing whitespace (e.g., `"  WORD  "`)
- **THEN** `Words()` returns it without whitespace (`"WORD"`)
