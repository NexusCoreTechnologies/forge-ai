package cache

import (
    "os"
    "path/filepath"
)

// Save writes data to a simple file-based cache under workspace/.contextcache
func Save(workspace, key string, data []byte) error {
    dir := filepath.Join(workspace, ".contextcache")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    path := filepath.Join(dir, key)
    return os.WriteFile(path, data, 0o644)
}

// Load reads data from cache; returns ok=false if missing.
func Load(workspace, key string) ([]byte, bool, error) {
    path := filepath.Join(workspace, ".contextcache", key)
    b, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, false, nil
        }
        return nil, false, err
    }
    return b, true, nil
}
