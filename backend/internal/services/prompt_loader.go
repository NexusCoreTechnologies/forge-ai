package services

import (
    "path/filepath"

    "forgeai/backend/internal/config"
    "forgeai/backend/internal/filesystem"
    "forgeai/backend/internal/logger"
    "forgeai/backend/internal/models"
)

// PromptLoader loads prompt files from disk.
type PromptLoader struct {
    config *config.Config
    logger *logger.JSONLogger
}

func NewPromptLoader(cfg *config.Config, logger *logger.JSONLogger) *PromptLoader {
    return &PromptLoader{config: cfg, logger: logger}
}

func (p *PromptLoader) ListPrompts() ([]models.Prompt, error) {
    folder := p.config.PromptFolder
    paths, err := filesystem.List(folder)
    if err != nil {
        return nil, err
    }
    var prompts []models.Prompt
    for _, path := range paths {
        if filepath.Ext(path) == ".md" {
            prompts = append(prompts, models.Prompt{Name: filepath.Base(path), Path: path})
        }
    }
    return prompts, nil
}

func (p *PromptLoader) LoadPrompt(name string) (*models.Prompt, error) {
    path := filepath.Join(p.config.PromptFolder, name)
    data, err := filesystem.Read(path)
    if err != nil {
        return nil, err
    }
    return &models.Prompt{Name: name, Path: path, Content: string(data)}, nil
}
