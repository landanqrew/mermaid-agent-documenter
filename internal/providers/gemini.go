package providers

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiProvider struct{}

func (p *GeminiProvider) GenerateContent(ctx context.Context, prompt string, model string, apiKey string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		nil, // no config needed for basic text generation
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return result.Text(), nil
}
