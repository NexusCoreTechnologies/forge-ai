package interfaces

import "forgeai/backend/internal/providers/models"

// Provider defines the methods the core engine will call.
type Provider interface {
    Generate(prompt string) (*models.Response, error)
    GenerateStream(prompt string) (<-chan string, error)
    ListModels() ([]string, error)
    HealthCheck() error
    ValidateConfiguration() error
    ProviderInfo() models.Info
}
