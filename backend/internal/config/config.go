package config

import (
    "errors"
    "os"
    "gopkg.in/yaml.v3"
)

// Config defines the application configuration values.
type Config struct {
    Workspace    string `yaml:"workspace"`
    PromptFolder string `yaml:"promptFolder"`
    OutputFolder string `yaml:"outputFolder"`
    LogFolder    string `yaml:"logFolder"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, cfg.Validate()
}

func (c *Config) Validate() error {
    if c.Workspace == "" {
        return errors.New("workspace is required")
    }
    if c.PromptFolder == "" {
        return errors.New("promptFolder is required")
    }
    if c.OutputFolder == "" {
        return errors.New("outputFolder is required")
    }
    if c.LogFolder == "" {
        return errors.New("logFolder is required")
    }
    return nil
}
