package openai

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "os"
    "time"

    "forgeai/backend/internal/providers/interfaces"
    pmodels "forgeai/backend/internal/providers/models"
    "forgeai/backend/internal/providers/registry"
)

type OpenAIProvider struct {
    apiKey string
    model  string
    client *http.Client
}

func NewOpenAI(cfg *pmodels.Config) (interfaces.Provider, error) {
    keyEnv := "OPENAI_API_KEY"
    model := "gpt-3.5-turbo"
    if cfg != nil {
        if cfg.APIKeyEnv != "" {
            keyEnv = cfg.APIKeyEnv
        }
        if cfg.DefaultModel != "" {
            model = cfg.DefaultModel
        }
    }
    key := os.Getenv(keyEnv)
    if key == "" {
        return nil, errors.New("openai: missing API key in environment " + keyEnv)
    }
    p := &OpenAIProvider{apiKey: key, model: model, client: &http.Client{Timeout: 30 * time.Second}}
    return p, nil
}

func (p *OpenAIProvider) Generate(prompt string) (*pmodels.Response, error) {
    body := map[string]interface{}{
        "model": p.model,
        "messages": []map[string]string{{"role": "user", "content": prompt}},
        "max_tokens": 512,
    }
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(b))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    resp, err := p.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        return nil, errors.New("openai: non-OK response")
    }
    var rr struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
        return nil, err
    }
    text := ""
    if len(rr.Choices) > 0 {
        text = rr.Choices[0].Message.Content
    }
    return &pmodels.Response{Text: text}, nil
}

func (p *OpenAIProvider) GenerateStream(prompt string) (<-chan string, error) {
    // Streaming placeholder: return a closed channel with full text as single message.
    ch := make(chan string, 1)
    resp, err := p.Generate(prompt)
    if err != nil {
        close(ch)
        return ch, err
    }
    ch <- resp.Text
    close(ch)
    return ch, nil
}

func (p *OpenAIProvider) ListModels() ([]string, error) { return []string{p.model}, nil }
func (p *OpenAIProvider) HealthCheck() error                       { return nil }
func (p *OpenAIProvider) ValidateConfiguration() error             { return nil }
func (p *OpenAIProvider) ProviderInfo() pmodels.Info {
    return pmodels.Info{Name: "openai", Version: "v1", Model: p.model}
}

func init() {
    registry.Register("openai", func(cfg *pmodels.Config) (any, error) { return NewOpenAI(cfg) })
}
