package filters

// DefaultIgnored returns default folders to ignore when scanning.
func DefaultIgnored() []string {
    return []string{".git", "node_modules", "dist", "build", "vendor", ".cache", "bin", "tmp"}
}

// IsIgnored checks if a path fragment should be ignored.
func IsIgnored(path string, ignored []string) bool {
    for _, p := range ignored {
        if p != "" && (path == p || len(path) >= len(p) && path[:len(p)] == p) {
            return true
        }
    }
    return false
}
