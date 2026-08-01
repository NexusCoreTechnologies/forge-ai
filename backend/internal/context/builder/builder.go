package builder

import (
    "time"

    models "forgeai/backend/internal/context/models"
)

// BuildExecutionContext composes an ExecutionContext from pieces.
func BuildExecutionContext(project, workspace, branch, version, sprint, arch string, docs map[string]string, relevant []string, deps, tasks, risks []string) *models.ExecutionContext {
    return &models.ExecutionContext{
        Project:       project,
        Workspace:     workspace,
        Branch:        branch,
        Version:       version,
        CurrentSprint: sprint,
        Architecture:  arch,
        Documentation: docs,
        RelevantFiles: relevant,
        Dependencies:  deps,
        OpenTasks:     tasks,
        KnownRisks:    risks,
        GeneratedAt:   time.Now(),
    }
}
