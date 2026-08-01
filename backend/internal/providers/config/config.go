package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

// ProvidersConfig represents providers.yaml structure.
type ProvidersConfig struct {
    Default string `yaml:"default_provider"`
}

// LoadProvidersConfig loads a providers.yaml file from path.
func LoadProvidersConfig(path string) (*ProvidersConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg ProvidersConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
