package providers

import (
	"context"
	"strings"
	"testing"
)

func TestGeminiProvider_GenerateContent(t *testing.T) {
	provider := &GeminiProvider{}

	t.Run("missing API key", func(t *testing.T) {
		ctx := context.Background()
		_, err := provider.GenerateContent(ctx, "test prompt", "gemini-1.5-flash", "")

		if err == nil {
			t.Error("Expected error for missing API key, got nil")
		}

		if !strings.Contains(err.Error(), "API key") {
			t.Errorf("Expected error message to contain 'API key', got: %v", err)
		}
	})

	t.Run("invalid model", func(t *testing.T) {
		ctx := context.Background()
		_, err := provider.GenerateContent(ctx, "test prompt", "invalid-model", "fake-key")

		if err == nil {
			t.Error("Expected error for invalid model, got nil")
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		ctx := context.Background()
		_, err := provider.GenerateContent(ctx, "", "gemini-1.5-flash", "fake-key")

		// This might succeed or fail depending on Gemini's behavior with empty prompts
		// For now, we'll just check that it doesn't panic
		if err != nil {
			t.Logf("Empty prompt error (expected): %v", err)
		}
	})
}

func TestGeminiProvider_ListModels(t *testing.T) {
	provider := &GeminiProvider{}

	t.Run("list models without API key", func(t *testing.T) {
		ctx := context.Background()
		models, err := provider.ListModels(ctx, "")

		// This should work since we return static models
		if err != nil {
			t.Errorf("Expected no error for static model list, got: %v", err)
		}

		if len(models) == 0 {
			t.Error("Expected some models to be returned")
		}

		// Check that we have expected models
		expectedModels := []string{
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.5-pro-002",
			"gemini-1.5-flash-002",
			"gemini-pro",
			"gemini-pro-vision",
		}

		if len(models) != len(expectedModels) {
			t.Errorf("Expected %d models, got %d", len(expectedModels), len(models))
		}

		for i, expected := range expectedModels {
			if i >= len(models) {
				t.Errorf("Missing expected model: %s", expected)
				continue
			}

			if models[i].ID != expected {
				t.Errorf("Expected model %s at index %d, got %s", expected, i, models[i].ID)
			}
		}
	})

	t.Run("model info structure", func(t *testing.T) {
		ctx := context.Background()
		models, err := provider.ListModels(ctx, "")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		for _, model := range models {
			if model.ID == "" {
				t.Error("Model ID should not be empty")
			}

			if model.Name == "" {
				t.Error("Model Name should not be empty")
			}

			if model.ID != model.Name && !strings.Contains(model.Name, " ") {
				t.Errorf("Model name should be descriptive, got: %s", model.Name)
			}
		}
	})
}

// Test with real API key (if available)
func TestGeminiProvider_RealAPI(t *testing.T) {
	// Skip this test if no real API key is available
	apiKey := "test-key" // This should be set via environment variable in real testing
	if apiKey == "test-key" {
		t.Skip("Skipping real API test - no API key provided")
	}

	provider := &GeminiProvider{}
	ctx := context.Background()

	t.Run("real API call with valid model", func(t *testing.T) {
		model := "gemini-1.5-flash"
		prompt := "Say hello in exactly 2 words."

		response, err := provider.GenerateContent(ctx, prompt, model, apiKey)

		if err != nil {
			t.Logf("API call failed (might be expected with test key): %v", err)
			return
		}

		if response == "" {
			t.Error("Expected non-empty response from API")
		}

		t.Logf("API Response: %s", response)
	})

	t.Run("real API call with custom model", func(t *testing.T) {
		model := "gemini-2.5-flash" // The model that's causing issues
		prompt := "Say hello in exactly 2 words."

		response, err := provider.GenerateContent(ctx, prompt, model, apiKey)

		if err != nil {
			t.Logf("Custom model API call failed: %v", err)
			// This might fail if the model doesn't exist
			return
		}

		if response == "" {
			t.Error("Expected non-empty response from API")
		}

		t.Logf("Custom model API Response: %s", response)
	})
}

func TestGeminiProvider_ErrorHandling(t *testing.T) {
	provider := &GeminiProvider{}

	t.Run("network timeout simulation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to simulate timeout

		_, err := provider.GenerateContent(ctx, "test", "gemini-1.5-flash", "fake-key")

		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}
	})

	t.Run("very long prompt", func(t *testing.T) {
		ctx := context.Background()
		longPrompt := strings.Repeat("This is a long prompt. ", 1000)

		_, err := provider.GenerateContent(ctx, longPrompt, "gemini-1.5-flash", "fake-key")

		// This might succeed or fail depending on Gemini's limits
		if err != nil {
			t.Logf("Long prompt error: %v", err)
		}
	})
}
