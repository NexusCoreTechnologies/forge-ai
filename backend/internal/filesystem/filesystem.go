package filesystem

import (
    "io/fs"
    "os"
    "path/filepath"
)

func Exists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

func Read(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func Write(path string, data []byte) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    return os.WriteFile(path, data, 0o644)
}

func List(path string) ([]string, error) {
    var entries []string
    err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() {
            entries = append(entries, p)
        }
        return nil
    })
    return entries, err
}

func Create(path string) error {
    return os.MkdirAll(path, 0o755)
}

func Delete(path string) error {
    return os.RemoveAll(path)
}
