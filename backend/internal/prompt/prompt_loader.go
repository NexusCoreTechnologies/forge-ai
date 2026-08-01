package prompt

import (
    "fmt"
    "path/filepath"

    "forgeai/backend/internal/config"
    "forgeai/backend/internal/filesystem"
    "forgeai/backend/internal/models"
)

// PromptLoader discovers and validates prompt files.
type PromptLoader struct {
    config *config.Config
}

func NewPromptLoader(cfg *config.Config) *PromptLoader {
    return &PromptLoader{config: cfg}
}

func (p *PromptLoader) DiscoverPrompts() ([]models.Prompt, error) {
    files, err := filesystem.List(p.config.PromptFolder)
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

func (p *PromptLoader) ReadPrompt(name string) (*models.Prompt, error) {
    path := filepath.Join(p.config.PromptFolder, name)
    if !filesystem.Exists(path) {
        return nil, fmt.Errorf("prompt not found: %s", name)
    }
    data, err := filesystem.Read(path)
    if err != nil {
        return nil, err
    }
    prompt := &models.Prompt{Name: name, Path: path, Content: string(data)}
    return prompt, nil
}

func (p *PromptLoader) ValidatePrompt(prompt *models.Prompt) error {
    if prompt == nil {
        return fmt.Errorf("prompt is nil")
    }
    if prompt.Content == "" {
        return fmt.Errorf("prompt content is empty")
    }
    if filepath.Ext(prompt.Path) != ".md" {
        return fmt.Errorf("unsupported prompt format: %s", prompt.Path)
    }
    return nil
}
