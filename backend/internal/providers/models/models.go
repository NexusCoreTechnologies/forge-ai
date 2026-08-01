package models

// Response is a normalized provider response.
type Response struct {
    Text   string
    Tokens int
}

// Info provides provider metadata.
type Info struct {
    Name    string
    Version string
    Model   string
}

// Config captures provider-specific configuration.
type Config struct {
    Name         string `yaml:"name"`
    APIKeyEnv    string `yaml:"apiKeyEnv"`
    DefaultModel string `yaml:"defaultModel"`
}
