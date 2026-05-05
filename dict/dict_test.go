package dict_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
)

// --- SliceSource ---

func TestSliceSource_Empty(t *testing.T) {
	words, err := dict.Slice(nil).Words()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 0 {
		t.Fatalf("expected 0 words, got %d", len(words))
	}
}

func TestSliceSource_Lowercase(t *testing.T) {
	words, err := dict.Slice([]string{"apple"}).Words()
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 1 || words[0] != "APPLE" {
		t.Fatalf("expected [APPLE], got %v", words)
	}
}

func TestSliceSource_MixedCase(t *testing.T) {
	words, err := dict.Slice([]string{"Apple", "bAnAnA"}).Words()
	if err != nil {
		t.Fatal(err)
	}
	if words[0] != "APPLE" || words[1] != "BANANA" {
		t.Fatalf("expected [APPLE BANANA], got %v", words)
	}
}

func TestSliceSource_WhitespaceTrimmed(t *testing.T) {
	words, err := dict.Slice([]string{"  WORD  ", "\tTAB\t"}).Words()
	if err != nil {
		t.Fatal(err)
	}
	if words[0] != "WORD" || words[1] != "TAB" {
		t.Fatalf("expected [WORD TAB], got %v", words)
	}
}

func TestSliceSource_BlankLinesSkipped(t *testing.T) {
	words, err := dict.Slice([]string{"HELLO", "", "  ", "WORLD"}).Words()
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d: %v", len(words), words)
	}
}

// --- FileSource ---

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "words*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close() //nolint:errcheck
	return f.Name()
}

func TestFileSource_ValidFile(t *testing.T) {
	path := writeTempFile(t, "APPLE\nBANANA\nCHERRY\n")
	words, err := dict.File(path).Words()
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 3 {
		t.Fatalf("expected 3 words, got %d: %v", len(words), words)
	}
}

func TestFileSource_BlankLinesSkipped(t *testing.T) {
	path := writeTempFile(t, "APPLE\n\nBANANA\n\n")
	words, err := dict.File(path).Words()
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}

func TestFileSource_CommentLinesSkipped(t *testing.T) {
	path := writeTempFile(t, "# a comment\nAPPLE\n# another\nBANANA\n")
	words, err := dict.File(path).Words()
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d: %v", len(words), words)
	}
}

func TestFileSource_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.txt")
	_, err := dict.File(path).Words()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestFileSource_Normalization(t *testing.T) {
	path := writeTempFile(t, "apple\n  Banana  \n")
	words, err := dict.File(path).Words()
	if err != nil {
		t.Fatal(err)
	}
	if words[0] != "APPLE" || words[1] != "BANANA" {
		t.Fatalf("expected [APPLE BANANA], got %v", words)
	}
}
