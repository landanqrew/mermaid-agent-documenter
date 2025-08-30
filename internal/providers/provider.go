package providers

import (
	"context"
)

type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string, model string, apiKey string) (string, error)
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
