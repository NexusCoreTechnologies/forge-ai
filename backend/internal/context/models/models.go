package models

import "time"

// ExecutionContext is the assembled context for a prompt execution.
type ExecutionContext struct {
    Project           string            `json:"project"`
    Workspace         string            `json:"workspace"`
    Branch            string            `json:"branch"`
    Version           string            `json:"version"`
    CurrentSprint     string            `json:"currentSprint"`
    Architecture      string            `json:"architecture"`
    Documentation     map[string]string `json:"documentation"`
    RelevantFiles     []string          `json:"relevantFiles"`
    Dependencies      []string          `json:"dependencies"`
    OpenTasks         []string          `json:"openTasks"`
    KnownRisks        []string          `json:"knownRisks"`
    GeneratedAt       time.Time         `json:"generatedAt"`
}
