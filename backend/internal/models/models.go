package models

import "time"

// Project represents a workspace project.
type Project struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

// Prompt represents a prompt artifact.
type Prompt struct {
    Name    string `json:"name"`
    Path    string `json:"path"`
    Content string `json:"content"`
}

// ExecutionReport is the persisted result of an execution.
type ExecutionReport struct {
    ExecutionID    string    `json:"executionId"`
    Project        string    `json:"project"`
    Prompt         string    `json:"prompt"`
    StartTime      time.Time `json:"startTime"`
    EndTime        time.Time `json:"endTime"`
    DurationSeconds float64  `json:"durationSeconds"`
    Status         string    `json:"status"`
    Errors         []string  `json:"errors,omitempty"`
    Warnings       []string  `json:"warnings,omitempty"`
    Logs           []string  `json:"logs,omitempty"`
}
