package mock

import (
    "forgeai/backend/internal/providers/interfaces"
    pmodels "forgeai/backend/internal/providers/models"
    "forgeai/backend/internal/providers/registry"
)

type MockProvider struct{}

func NewMock(cfg *pmodels.Config) (interfaces.Provider, error) { return &MockProvider{}, nil }

func (m *MockProvider) Generate(prompt string) (*pmodels.Response, error) {
    return &pmodels.Response{Text: "mock response for: " + prompt}, nil
}
func (m *MockProvider) GenerateStream(prompt string) (<-chan string, error) {
    ch := make(chan string, 1)
    ch <- "mock response for: " + prompt
    close(ch)
    return ch, nil
}
func (m *MockProvider) ListModels() ([]string, error) { return []string{"mock"}, nil }
func (m *MockProvider) HealthCheck() error                       { return nil }
func (m *MockProvider) ValidateConfiguration() error             { return nil }
func (m *MockProvider) ProviderInfo() pmodels.Info                { return pmodels.Info{Name: "mock", Version: "v0", Model: "mock"} }

func init() {
    registry.Register("mock", func(cfg *pmodels.Config) (any, error) { return NewMock(cfg) })
}
