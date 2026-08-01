package backlog

import (
    "bufio"
    "os"
    "strings"
)

// ReadBacklog reads MASTER_BACKLOG.md and returns lines.
func ReadBacklog(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var lines []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

// ParseSprints extracts sprint headings like 'Sprint' from lines.
func ParseSprints(lines []string) []string {
    var sprints []string
    for _, l := range lines {
        if strings.HasPrefix(strings.ToLower(strings.TrimSpace(l)), "sprint") {
            sprints = append(sprints, strings.TrimSpace(l))
        }
    }
    return sprints
}
