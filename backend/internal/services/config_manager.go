package services

import (
    "forgeai/backend/internal/config"
    "forgeai/backend/internal/logger"
)

// ConfigManager wraps configuration operations.
type ConfigManager struct {
    config *config.Config
    logger *logger.JSONLogger
}

func NewConfigManager(cfg *config.Config, logger *logger.JSONLogger) *ConfigManager {
    return &ConfigManager{config: cfg, logger: logger}
}

func (c *ConfigManager) Config() *config.Config {
    return c.config
}
