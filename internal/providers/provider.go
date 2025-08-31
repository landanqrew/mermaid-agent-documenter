package providers

import (
	"context"
)

type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Created int64  `json:"created,omitempty"`
}

type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string, model string, apiKey string) (string, error)
	ListModels(ctx context.Context, apiKey string) ([]ModelInfo, error)
}

func GetProvider(providerName string) LLMProvider {
	switch providerName {
	case "openai":
		return &OpenAIProvider{}
	case "anthropic":
		return &AnthropicProvider{}
	case "google":
		return &GeminiProvider{}
	default:
		return &OpenAIProvider{} // default
	}
}
