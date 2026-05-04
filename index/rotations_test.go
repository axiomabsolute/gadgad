package index

import (
	"strings"
	"testing"
)

func TestRotations_SingleLetter(t *testing.T) {
	rots := rotations("A")
	if len(rots) != 1 {
		t.Fatalf("expected 1 rotation, got %d", len(rots))
	}
	if rots[0] != "A+" {
		t.Errorf("expected \"A+\", got %q", rots[0])
	}
}

func TestRotations_TwoLetters(t *testing.T) {
	rots := rotations("AB")
	// position 0: A+ B  → "A+B"
	// position 1: BA+   → "BA+"
	want := []string{"A+B", "BA+"}
	if len(rots) != len(want) {
		t.Fatalf("expected %d rotations, got %d: %v", len(want), len(rots), rots)
	}
	for i, w := range want {
		if rots[i] != w {
			t.Errorf("rotation[%d]: expected %q, got %q", i, w, rots[i])
		}
	}
}

func TestRotations_CARED(t *testing.T) {
	rots := rotations("CARED")
	want := []string{
		"C+ARED",
		"AC+RED",
		"RAC+ED",
		"ERAC+D",
		"DERAC+",
	}
	if len(rots) != len(want) {
		t.Fatalf("expected %d rotations, got %d: %v", len(want), len(rots), rots)
	}
	for i, w := range want {
		if rots[i] != w {
			t.Errorf("rotation[%d]: expected %q, got %q", i, w, rots[i])
		}
	}
}

func TestRotations_AnchorIsFirstCharacter(t *testing.T) {
	// Per Gordon (1994): rotation i = w_i, w_{i-1}, ..., w_1, '+', w_{i+1}, ..., w_n
	// The anchor letter is always the FIRST character of the rotation string.
	for _, word := range []string{"A", "GO", "CAT", "CRANE", "STARED"} {
		runes := []rune(word)
		rots := rotations(word)
		if len(rots) != len(runes) {
			t.Errorf("%s: expected %d rotations, got %d", word, len(runes), len(rots))
			continue
		}
		for i, rot := range rots {
			rotRunes := []rune(rot)
			if rotRunes[0] != runes[i] {
				t.Errorf("%s rotation[%d] %q: expected first char %c (anchor), got %c",
					word, i, rot, runes[i], rotRunes[0])
			}
		}
	}
}

func TestRotations_ExactlyOneSeparator(t *testing.T) {
	for _, word := range []string{"A", "CAT", "CARED", "TRACES"} {
		for _, rot := range rotations(word) {
			count := strings.Count(rot, "+")
			if count != 1 {
				t.Errorf("%s rotation %q: expected exactly 1 '+', got %d", word, rot, count)
			}
		}
	}
}
