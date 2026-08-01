package core

import "forgeai/backend/internal/models"

type Logger interface {
    Info(message string, fields map[string]any)
    Error(message string, fields map[string]any)
    Debug(message string, fields map[string]any)
}

type FileSystem interface {
    Exists(path string) bool
    Read(path string) ([]byte, error)
    Write(path string, data []byte) error
    List(path string) ([]string, error)
    Create(path string) error
    Delete(path string) error
}

type WorkspaceManager interface {
    OpenWorkspace(path string) error
    ValidateWorkspace() error
    DetectProjects() ([]models.Project, error)
    DetectPrompts() ([]models.Prompt, error)
}

type PromptLoader interface {
    DiscoverPrompts() ([]models.Prompt, error)
    ReadPrompt(name string) (*models.Prompt, error)
    ValidatePrompt(prompt *models.Prompt) error
}

type ExecutionEngine interface {
    Execute(project string, promptName string) (*models.ExecutionReport, error)
}

type ReportManager interface {
    SaveExecutionReport(report *models.ExecutionReport) error
}
