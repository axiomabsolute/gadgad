package dict

// SliceSource is a Source backed by a string slice.
type SliceSource struct {
	words []string
}

// Slice returns a Source that yields words from the provided slice.
// Words are normalized to uppercase with whitespace trimmed.
func Slice(words []string) Source {
	return &SliceSource{words: words}
}

func (s *SliceSource) Words() ([]string, error) {
	out := make([]string, 0, len(s.words))
	for _, w := range s.words {
		if n := normalize(w); n != "" {
			out = append(out, n)
		}
	}
	return out, nil
}
