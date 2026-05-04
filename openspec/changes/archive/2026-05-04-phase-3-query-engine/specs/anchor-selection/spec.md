## ADDED Requirements

### Requirement: Auto anchor selects the rarest fixed letter
When `Anchor` is `Auto` and the pattern contains at least one `Fixed` slot, the system SHALL select the `Fixed` slot whose letter has the lowest frequency among letters in the dictionary as the anchor. This minimizes the number of candidates yielded by `TraverseAt`.

#### Scenario: Multiple fixed slots — rarest chosen
- **WHEN** a pattern has `Fixed('E')` at position 1 and `Fixed('Q')` at position 3
- **THEN** 'Q' at position 3 is selected as the anchor (Q is rarer than E)

#### Scenario: Single fixed slot
- **WHEN** a pattern has exactly one `Fixed` slot
- **THEN** that letter and position are selected as the anchor unconditionally

### Requirement: Auto anchor for all-Free patterns uses rarest Available letter
When `Anchor` is `Auto` and all slots are `Free` or `Any` (no `Fixed` slots), the system SHALL select the rarest letter in the `Available` set as the anchor.

#### Scenario: Anagram anchor selection
- **WHEN** all slots are `Free` and `Available` contains `{A, E, R, T, S}`
- **THEN** the rarest of those letters (e.g. 'R' or 'T' depending on frequency table) is selected as the anchor

### Requirement: Manual anchor overrides auto-selection
When `Anchor` is `Manual(position)`, the system SHALL use the slot at `position` as the anchor without applying the frequency heuristic.

#### Scenario: Manual anchor used as-is
- **WHEN** `Anchor` is `Manual(4)` and position 4 contains `Fixed('E')`
- **THEN** 'E' at position 4 is used as the anchor regardless of whether rarer fixed letters exist elsewhere

#### Scenario: Manual anchor on a Free slot
- **WHEN** `Anchor` is `Manual(2)` and position 2 is `Free`
- **THEN** the system MUST return an error or panic — a Free slot cannot be the anchor

### Requirement: Frequency table is static and deterministic
The letter frequency used for anchor selection SHALL be derived from a static table (not computed at query time from the index). Selection between two letters of equal frequency SHALL be deterministic (e.g., alphabetical tiebreak).

#### Scenario: Deterministic selection with equal-frequency letters
- **WHEN** two `Fixed` slots have letters of equal frequency in the table
- **THEN** the same slot is always selected across multiple calls with the same pattern
