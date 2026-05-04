## ADDED Requirements

### Requirement: RotationToWord reconstructs the original word from a GADDAG rotation string
The `index` package SHALL export a `RotationToWord(rot string) string` function that reconstructs the original dictionary word from any GADDAG rotation string. The rotation format is `<reversed-prefix-including-anchor>+<suffix>`. The function SHALL reverse the portion before `+` and concatenate it with the portion after `+` to produce the original word.

#### Scenario: Single-letter word
- **WHEN** `RotationToWord` is called with `"A+"`
- **THEN** it returns `"A"`

#### Scenario: Anchor at first position
- **WHEN** `RotationToWord` is called with `"C+ARED"`
- **THEN** it returns `"CARED"`

#### Scenario: Anchor at middle position
- **WHEN** `RotationToWord` is called with `"RAC+ED"`
- **THEN** it returns `"CARED"`

#### Scenario: Anchor at last position
- **WHEN** `RotationToWord` is called with `"DERAC+"`
- **THEN** it returns `"CARED"`

#### Scenario: Consistency with Traverse
- **WHEN** `RotationToWord` is applied to every string yielded by `TraverseRaw(anchor)`
- **THEN** the resulting set of words SHALL equal the set yielded by `Traverse(anchor)` for the same anchor
