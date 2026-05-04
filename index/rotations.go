package index

// rotations returns all GADDAG rotation strings for word.
//
// For a word of length N, N rotations are produced — one per letter position.
// Each rotation has the form: <reversed prefix up to and including anchor> + <suffix after anchor>
//
// For "CARED" (indices 0–4):
//
//	position 0: "C+"       + "ARED"  → "C+ARED"
//	position 1: "AC"       + "+RED"  → "AC+RED"
//	position 2: "RAC"      + "+ED"   → "RAC+ED"
//	position 3: "ERAC"     + "+D"    → "ERAC+D"
//	position 4: "DERAC"    + "+"     → "DERAC+"
//
// The anchor letter is always immediately before the '+' separator.
func rotations(word string) []string {
	runes := []rune(word)
	n := len(runes)
	result := make([]string, n)
	for i := 0; i < n; i++ {
		// Reversed prefix: runes[i], runes[i-1], ..., runes[0]
		prefix := make([]rune, i+1)
		for j := 0; j <= i; j++ {
			prefix[j] = runes[i-j]
		}
		// Suffix: runes[i+1..n-1]
		suffix := runes[i+1:]
		// Compose: prefix + Separator + suffix
		rot := make([]rune, 0, len(prefix)+1+len(suffix))
		rot = append(rot, prefix...)
		rot = append(rot, Separator)
		rot = append(rot, suffix...)
		result[i] = string(rot)
	}
	return result
}
