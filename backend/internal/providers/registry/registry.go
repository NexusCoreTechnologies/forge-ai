package registry

import "forgeai/backend/internal/providers/models"

// Constructor returns a Provider instance from a config.
type Constructor func(cfg *models.Config) (any, error)

var registry = map[string]Constructor{}

// Register registers a provider constructor under a name.
func Register(name string, c Constructor) {
    registry[name] = c
}

// Get returns a constructor for a provider name, or nil.
func Get(name string) Constructor {
    if c, ok := registry[name]; ok {
        return c
    }
    return nil
}
