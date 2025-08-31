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

func (p *GeminiProvider) ListModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	// Google Gemini doesn't have a direct models API endpoint in the same way
	// We'll return the known Gemini models for now
	// In a future version, we could use the REST API directly

	// For now, return our known models since Gemini's SDK doesn't expose ListModels
	knownModels := []ModelInfo{
		{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro"},
		{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash"},
		{ID: "gemini-1.5-pro-002", Name: "Gemini 1.5 Pro 002"},
		{ID: "gemini-1.5-flash-002", Name: "Gemini 1.5 Flash 002"},
		{ID: "gemini-pro", Name: "Gemini Pro"},
		{ID: "gemini-pro-vision", Name: "Gemini Pro Vision"},
	}

	return knownModels, nil
}
