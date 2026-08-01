package collector

import (
    "io/ioutil"
)

// CollectFiles reads provided file paths and returns a map[path]content.
func CollectFiles(paths []string) (map[string][]byte, error) {
    out := map[string][]byte{}
    for _, p := range paths {
        b, err := ioutil.ReadFile(p)
        if err != nil {
            // skip unreadable files
            continue
        }
        out[p] = b
    }
    return out, nil
}
