package executionplan

import (
    "encoding/json"
    "os"
)

type Plan struct {
    SelectedSprint string   `json:"selectedSprint"`
    Tasks          []string `json:"tasks"`
}

func WritePlan(path string, p *Plan) error {
    b, _ := json.MarshalIndent(p, "", "  ")
    return os.WriteFile(path, b, 0o644)
}
