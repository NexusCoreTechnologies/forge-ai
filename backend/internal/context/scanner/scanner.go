package scanner

import (
    "io/fs"
    "path/filepath"
)

// ScanWorkspace walks the workspace and returns candidate file paths.
func ScanWorkspace(root string, ignored []string) ([]string, error) {
    var files []string
    walk := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        // skip directories we don't want
        if d.IsDir() {
            for _, ig := range ignored {
                if d.Name() == ig {
                    return filepath.SkipDir
                }
            }
            return nil
        }
        files = append(files, path)
        return nil
    }
    err := filepath.WalkDir(root, walk)
    return files, err
}
