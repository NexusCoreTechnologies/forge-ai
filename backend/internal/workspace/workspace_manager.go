package workspace

import (
    "fmt"
    "os"
    "path/filepath"

    "forgeai/backend/internal/config"
    "forgeai/backend/internal/filesystem"
    "forgeai/backend/internal/logger"
    "forgeai/backend/internal/models"
)

// WorkspaceManager handles workspace discovery and validation.
type WorkspaceManager struct {
    config    *config.Config
    logger    *logger.JSONLogger
    workspace string
}

func NewWorkspaceManager(cfg *config.Config, lg *logger.JSONLogger) *WorkspaceManager {
    return &WorkspaceManager{config: cfg, logger: lg}
}

func (w *WorkspaceManager) OpenWorkspace(path string) error {
    if !filesystem.Exists(path) {
        return fmt.Errorf("workspace not found: %s", path)
    }
    w.workspace = path
    return nil
}

func (w *WorkspaceManager) ValidateWorkspace() error {
    required := []string{
        w.workspace,
        filepath.Join(w.workspace, "generated"),
        filepath.Join(w.workspace, "reports"),
        filepath.Join(w.workspace, "logs"),
        filepath.Join(w.workspace, "context"),
        filepath.Join(w.workspace, "memory"),
        filepath.Join(w.workspace, "cache"),
        filepath.Join(w.workspace, "sessions"),
    }
    for _, path := range required {
        if !filesystem.Exists(path) {
            if err := filesystem.Create(path); err != nil {
                return err
            }
        }
    }
    if !filesystem.Exists(w.config.PromptFolder) {
        return fmt.Errorf("prompt folder not found: %s", w.config.PromptFolder)
    }
    return nil
}

func (w *WorkspaceManager) DetectProjects() ([]models.Project, error) {
    entries, err := os.ReadDir(w.workspace)
    if err != nil {
        return nil, err
    }
    var projects []models.Project
    for _, entry := range entries {
        if entry.IsDir() {
            projects = append(projects, models.Project{Name: entry.Name(), Path: filepath.Join(w.workspace, entry.Name())})
        }
    }
    return projects, nil
}

func (w *WorkspaceManager) DetectPrompts() ([]models.Prompt, error) {
    files, err := filesystem.List(w.config.PromptFolder)
    if err != nil {
        return nil, err
    }
    var prompts []models.Prompt
    for _, path := range files {
        if filepath.Ext(path) == ".md" {
            prompts = append(prompts, models.Prompt{Name: filepath.Base(path), Path: path})
        }
    }
    return prompts, nil
}
