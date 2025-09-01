package providers

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"path/filepath"
)

// Local config struct to avoid import cycle
type testConfig struct {
	Secrets map[string]string `json:"secrets,omitempty"`
}

func loadTestConfig() (*testConfig, error) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, "mermaid-agent-documenter", "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config testConfig
	err = json.Unmarshal(data, &config)
	return &config, err
}

func TestListGeminiModels(t *testing.T) {
	config, err := loadTestConfig()
	if err != nil {
		t.Skipf("Error loading config: %v", err)
	}
	// fmt.Println("config.Secrets[google]:", string(config.Secrets["google"]))
	provider := &GeminiProvider{}
	models, err := provider.ListModels(context.Background(), config.Secrets["google"])
	if err != nil {
		t.Fatalf("Error listing models with API key ending with (...%s): %v", config.Secrets["google"][:3], err)
	}
	t.Logf("Listed %d models", len(models))
	for _, model := range models {
		t.Logf("Model: %s", model.ID)
	}
}

func TestListOpenAIModels(t *testing.T) {
	config, err := loadTestConfig()
	if err != nil {
		t.Skipf("Error loading config: %v", err)
	}
	// t.Log("config.Secrets[openai]:", config.Secrets["openai"])
	provider := &OpenAIProvider{}
	models, err := provider.ListModels(context.Background(), config.Secrets["openai"])
	if err != nil {
		t.Fatalf("Error listing models with API key ending with (...%s): %v", config.Secrets["openai"][:3], err)
	}
	t.Logf("Listed %d models", len(models))
	for _, model := range models {
		t.Logf("Model: %s", model.ID)
	}
}

func TestListAnthropicModels(t *testing.T) {
	config, err := loadTestConfig()
	if err != nil {
		t.Skipf("Error loading config: %v", err)
	}
	// t.Log("config.Secrets[anthropic]:", config.Secrets["anthropic"])
	provider := &AnthropicProvider{}
	models, err := provider.ListModels(context.Background(), config.Secrets["anthropic"])
	if err != nil {
		t.Skipf("Error listing models with API key ending with (...%s): %v", config.Secrets["anthropic"][:min(len(config.Secrets["anthropic"]), 3)], err)
	}
	t.Logf("Listed %d models", len(models))
	for _, model := range models {
		t.Logf("Model: %s", model.ID)
	}
}

/*
curl https://api.openai.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": "gpt-5-nano",
    "input": "Tell me a three sentence bedtime story about a unicorn."
  }'


*/
