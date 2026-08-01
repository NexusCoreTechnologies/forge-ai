package services

import (
    "path/filepath"

    "forgeai/backend/internal/config"
    "forgeai/backend/internal/filesystem"
    "forgeai/backend/internal/logger"
    "forgeai/backend/internal/models"
)

// WorkspaceManager manages project workspace folders.
type WorkspaceManager struct {
    config *config.Config
    logger *logger.JSONLogger
}

func NewWorkspaceManager(cfg *config.Config, logger *logger.JSONLogger) *WorkspaceManager {
    return &WorkspaceManager{config: cfg, logger: logger}
}

func (w *WorkspaceManager) ListProjects() ([]models.Project, error) {
    root := w.config.Workspace
    files, err := filesystem.List(root)
    if err != nil {
        return nil, err
    }
    var projects []models.Project
    for _, path := range files {
        projects = append(projects, models.Project{Name: filepath.Base(path), Path: path})
    }
    return projects, nil
}
