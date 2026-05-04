package query

import "testing"

func TestSelectAnchor_RarestFixed(t *testing.T) {
	// E(freq=12) at pos 0, Q(freq=1) at pos 3 — Q should win.
	p := Pattern{
		Slots:  []Slot{Fixed('E'), Free, Free, Fixed('Q')},
		Anchor: Auto,
	}
	got, err := selectAnchor(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.letter != 'Q' || got.position != 3 {
		t.Errorf("expected Q at position 3, got %c at %d", got.letter, got.position)
	}
}

func TestSelectAnchor_SingleFixed(t *testing.T) {
	p := Pattern{
		Slots:  []Slot{Free, Fixed('A'), Free},
		Anchor: Auto,
	}
	got, err := selectAnchor(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.letter != 'A' || got.position != 1 {
		t.Errorf("expected A at position 1, got %c at %d", got.letter, got.position)
	}
}

func TestSelectAnchor_AllFreePicksRarestFromAvailable(t *testing.T) {
	// Available: A(9), E(12), R(6) — R is rarest.
	avail := NewLetterSet("AER")
	p := Pattern{
		Slots:     []Slot{Free, Free, Free},
		Available: &avail,
		Anchor:    Auto,
	}
	got, err := selectAnchor(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.letter != 'R' {
		t.Errorf("expected R (rarest), got %c", got.letter)
	}
	if got.position != -1 {
		t.Errorf("expected position -1 for all-Free pattern, got %d", got.position)
	}
}

func TestSelectAnchor_ManualPassthrough(t *testing.T) {
	p := Pattern{
		Slots:  []Slot{Free, Fixed('E'), Free},
		Anchor: Manual(1),
	}
	got, err := selectAnchor(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.letter != 'E' || got.position != 1 {
		t.Errorf("expected E at position 1, got %c at %d", got.letter, got.position)
	}
}

func TestSelectAnchor_ManualOnFreeSlotPanics(t *testing.T) {
	p := Pattern{
		Slots:  []Slot{Free, Free, Free},
		Anchor: Manual(0),
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for Manual anchor on Free slot, got none")
		}
	}()
	_, _ = selectAnchor(p) //nolint:errcheck
}

func TestSelectAnchor_DeterministicTiebreak(t *testing.T) {
	// J(freq=1) at pos 0, Z(freq=1) at pos 2 — alphabetical tiebreak picks J.
	p := Pattern{
		Slots:  []Slot{Fixed('J'), Free, Fixed('Z')},
		Anchor: Auto,
	}
	got1, err := selectAnchor(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got2, _ := selectAnchor(p)
	if got1 != got2 {
		t.Errorf("non-deterministic: first=%v second=%v", got1, got2)
	}
	if got1.letter != 'J' {
		t.Errorf("expected J (alphabetically first among equal-freq), got %c", got1.letter)
	}
}

func TestSelectAnchor_NoFixedNilAvailableError(t *testing.T) {
	p := Pattern{
		Slots:  []Slot{Free, Free},
		Anchor: Auto,
	}
	_, err := selectAnchor(p)
	if err == nil {
		t.Error("expected error for all-Free pattern with nil Available, got nil")
	}
}

func TestSelectAnchor_NoFixedEmptyAvailableError(t *testing.T) {
	empty := NewLetterSet("")
	p := Pattern{
		Slots:     []Slot{Free, Free},
		Available: &empty,
		Anchor:    Auto,
	}
	_, err := selectAnchor(p)
	if err == nil {
		t.Error("expected error for all-Free pattern with empty Available, got nil")
	}
}
