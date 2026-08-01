package factory

import (
    "errors"
    "forgeai/backend/internal/providers/models"
    "forgeai/backend/internal/providers/registry"
    "forgeai/backend/internal/providers/interfaces"
)

// NewProvider instantiates a provider by name using the registered constructor.
func NewProvider(name string, cfg *models.Config) (interfaces.Provider, error) {
    c := registry.Get(name)
    if c == nil {
        return nil, errors.New("provider not registered: " + name)
    }
    inst, err := c(cfg)
    if err != nil {
        return nil, err
    }
    // The constructor may return a concrete type that implements interfaces.Provider
    p, ok := inst.(interfaces.Provider)
    if !ok {
        return nil, errors.New("constructor did not return a Provider implementation")
    }
    return p, nil
}
