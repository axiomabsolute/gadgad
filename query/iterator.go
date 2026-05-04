package query

import "iter"

// Iterator wraps an iter.Seq[string] and adds a Collect convenience method
// for callers that prefer an eager []string result.
//
// For range-based iteration, use the underlying sequence directly:
//
//	for word := range query.Search(idx, p) { ... }
//
// For eager materialization:
//
//	words := query.NewIterator(query.Search(idx, p)).Collect()
type Iterator struct {
	seq iter.Seq[string]
}

// NewIterator wraps seq in an Iterator.
func NewIterator(seq iter.Seq[string]) Iterator {
	return Iterator{seq: seq}
}

// All returns the underlying iter.Seq[string] for use with range.
func (it Iterator) All() iter.Seq[string] { return it.seq }

// Collect materializes all results into a slice.
func (it Iterator) Collect() []string {
	var out []string
	for s := range it.seq {
		out = append(out, s)
	}
	return out
}
