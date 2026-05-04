package index_test

import (
	"os"
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
	"github.com/axiomabsolute/gadgad/index"
)

func BenchmarkBuild_SmallWords(b *testing.B) {
	src := dict.File("../testdata/small_words.txt")
	b.ResetTimer()
	for range b.N {
		if _, err := index.Build(src); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuild_TWL06 runs only when testdata/twl06.txt exists.
func BenchmarkBuild_TWL06(b *testing.B) {
	path := "../testdata/twl06.txt"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		b.Skip("testdata/twl06.txt not present")
	}
	src := dict.File(path)
	b.ResetTimer()
	for range b.N {
		if _, err := index.Build(src); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTraverse_E(b *testing.B) {
	idx, err := index.Build(dict.File("../testdata/small_words.txt"))
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		for range idx.Traverse('E') {
		}
	}
}

func BenchmarkTraverse_TWL06_S(b *testing.B) {
	path := "../testdata/twl06.txt"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		b.Skip("testdata/twl06.txt not present")
	}
	idx, err := index.Build(dict.File(path))
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		for range idx.Traverse('S') {
		}
	}
}
