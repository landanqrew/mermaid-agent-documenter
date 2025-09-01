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
	knownModels := []ModelInfo{
		{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro"},
		{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash"},
		{ID: "gemini-1.5-pro-002", Name: "Gemini 1.5 Pro 002"},
		{ID: "gemini-1.5-flash-002", Name: "Gemini 1.5 Flash 002"},
		{ID: "gemini-pro", Name: "Gemini Pro"},
		{ID: "gemini-pro-vision", Name: "Gemini Pro Vision"},
	}

	if apiKey == "" {
		return knownModels, fmt.Errorf("API key is required")
	}

	// ctx = context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		
	}


	// Retrieve the list of models.
	models, err := client.Models.List(ctx, &genai.ListModelsConfig{})
	if err != nil {
		return knownModels, fmt.Errorf("Error listing models: %w", err)
	}

	fmt.Println("List of models that support generateContent:")
	for _, m := range models.Items {
		for _, action := range m.SupportedActions {
			if action == "generateContent" {
				fmt.Println(m.Name)
				break
			}
		}
	}

	modelInfo := []ModelInfo{}
	fmt.Println("\nList of models that support embedContent:")
	for _, m := range models.Items {
		for _, action := range m.SupportedActions {
			if action == "embedContent" {
				modelInfo = append(modelInfo, ModelInfo{
					ID:   m.Name,
					Name: m.Name,
				})
				break
			}
		}
	}

	if len(modelInfo) > 0 {
		return modelInfo, nil
	}


	return knownModels, fmt.Errorf("No models found")
}
