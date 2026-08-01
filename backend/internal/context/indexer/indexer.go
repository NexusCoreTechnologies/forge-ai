package indexer

import "strings"

// BuildIndex creates a simple inverted index mapping terms to file paths.
func BuildIndex(docs map[string]string) map[string][]string {
    idx := map[string][]string{}
    for path, content := range docs {
        words := strings.Fields(content)
        seen := map[string]bool{}
        for _, w := range words {
            w = strings.ToLower(w)
            if seen[w] {
                continue
            }
            seen[w] = true
            idx[w] = append(idx[w], path)
        }
    }
    return idx
}
