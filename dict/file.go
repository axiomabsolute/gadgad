package dict

import (
	"bufio"
	"os"
	"strings"
)

// FileSource is a Source backed by a line-delimited text file.
type FileSource struct {
	path string
}

// File returns a Source that reads words from a UTF-8 line-delimited text file.
// Blank lines and lines beginning with '#' are skipped. Words are normalized
// to uppercase with whitespace trimmed.
func File(path string) Source {
	return &FileSource{path: path}
}

func (f *FileSource) Words() ([]string, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var out []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, normalize(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
