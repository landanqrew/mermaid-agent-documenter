/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/landanqrew/mermaid-agent-documenter/internal/providers"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage configuration settings for the Mermaid Agent Documenter.

This command provides subcommands to manage:
- API keys for different model providers (secrets)
- Current project settings (project)
- Default provider and model selection (provider, model)
- View current configuration`,
}

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage API keys and secrets",
	Long: `Manage API keys and secrets for different model providers.

Supported providers: openai, anthropic, google`,
}

// secretsSetCmd represents the secrets set command
var secretsSetCmd = &cobra.Command{
	Use:   "set <provider> <api-key>",
	Short: "Set API key for a model provider",
	Long: `Set the API key for a specific model provider.

Supported providers:
- openai: OpenAI API key
- anthropic: Anthropic API key
- google: Google AI API key

Example:
  mad config secrets set openai "sk-your-openai-key-here"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		provider := strings.ToLower(args[0])
		apiKey := args[1]

		// Validate provider
		validProviders := map[string]bool{
			"openai":    true,
			"anthropic": true,
			"google":    true,
		}

		if !validProviders[provider] {
			fmt.Printf("Error: Invalid provider '%s'. Supported providers: openai, anthropic, google\n", provider)
			os.Exit(1)
		}

		// Load current config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Initialize secrets if not exists
		if config.Secrets == nil {
			config.Secrets = make(map[string]string)
		}

		// Set the API key
		config.Secrets[provider] = apiKey

		// Save config
		configDir := getConfigDir()
		configPath := filepath.Join(configDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ API key for '%s' has been set successfully\n", provider)
	},
}

// secretsListCmd represents the secrets list command
var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured API keys (without showing actual keys)",
	Long: `List all configured API keys without showing the actual key values.

This shows which providers have API keys configured.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🔑 Configured API Keys:")
		fmt.Println()

		providers := []string{"openai", "anthropic", "google"}
		hasAnyKeys := false

		for _, provider := range providers {
			if config.Secrets != nil && config.Secrets[provider] != "" {
				// Show first 4 and last 4 characters for verification
				key := config.Secrets[provider]
				maskedKey := ""
				if len(key) > 8 {
					maskedKey = key[:4] + "..." + key[len(key)-4:]
				} else {
					maskedKey = "***hidden***"
				}
				fmt.Printf("✅ %s: %s\n", provider, maskedKey)
				hasAnyKeys = true
			} else {
				fmt.Printf("❌ %s: Not configured\n", provider)
			}
		}

		if !hasAnyKeys {
			fmt.Println()
			fmt.Println("No API keys are currently configured.")
			fmt.Println("Use 'mad config secrets set <provider> <api-key>' to configure API keys.")
		}
	},
}

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage project settings",
	Long: `Manage current project settings.

This allows you to set which project directory is currently active.`,
}

// projectSetCmd represents the project set command
var projectSetCmd = &cobra.Command{
	Use:   "set <project-directory>",
	Short: "Set the current project directory",
	Long: `Set the current project directory for the Mermaid Agent Documenter.

The project directory should contain transcripts/, out/, and logs/ subdirectories.
You can specify either an absolute path or a relative path from the current directory.

Examples:
  mad config project set /path/to/my-project
  mad config project set ./my-auth-app
  mad config project set ../projects/ecommerce`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]

		// Convert to absolute path if relative
		if !filepath.IsAbs(projectPath) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			projectPath = filepath.Join(cwd, projectPath)
		}

		// Verify the directory exists
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			fmt.Printf("Error: Project directory '%s' does not exist\n", projectPath)
			fmt.Println("Make sure to create the project first with 'mad init <project-name>'")
			os.Exit(1)
		}

		// Verify it's a directory
		if info, err := os.Stat(projectPath); err != nil || !info.IsDir() {
			fmt.Printf("Error: '%s' is not a directory\n", projectPath)
			os.Exit(1)
		}

		// Check for required subdirectories
		requiredDirs := []string{"transcripts", "out", "logs"}
		for _, dir := range requiredDirs {
			dirPath := filepath.Join(projectPath, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				fmt.Printf("Warning: Required directory '%s' not found in project\n", dirPath)
				fmt.Println("The project may not be properly initialized.")
			}
		}

		// Load current config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Extract project name from path
		projectName := filepath.Base(projectPath)

		// Update current project
		config.CurrentProject = &ProjectConfig{
			Name:      projectName,
			RootDir:   projectPath,
			CreatedAt: fmt.Sprintf("Updated %s", "now"), // Could use proper timestamp
		}

		// Save config
		configDir := getConfigDir()
		configPath := filepath.Join(configDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Current project set to: %s\n", projectPath)
		fmt.Printf("📁 Project name: %s\n", projectName)
	},
}

// projectListCmd represents the project list command
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current project",
	Long: `List current project settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		currentProject := ""
		if config.CurrentProject != nil {
			currentProject = config.CurrentProject.Name
		}

		if currentProject == "" {
			fmt.Println("No current project defined")
			fmt.Println("You can set your current project configurations with 'mad config project set <project-directory>'")
			return
		}
		
		fmt.Printf("Current Project: %s\n", currentProject)
		fmt.Printf("Project Directory: %s\n", config.CurrentProject.RootDir)
	},
}

// providerCmd represents the provider command
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage default provider settings",
	Long: `Manage the default LLM provider selection.

This allows you to set which provider (openai, anthropic, google) is used by default.`,
}

// providerSetCmd represents the provider set command
var providerSetCmd = &cobra.Command{
	Use:   "set <provider>",
	Short: "Set the default LLM provider",
	Long: `Set the default LLM provider for the Mermaid Agent Documenter.

Supported providers:
- openai: OpenAI models
- anthropic: Anthropic Claude models
- google: Google Gemini models

Example:
  mad config provider set openai`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := strings.ToLower(args[0])

		// Validate provider
		validProviders := map[string]bool{
			"openai":    true,
			"anthropic": true,
			"google":    true,
		}

		if !validProviders[provider] {
			fmt.Printf("Error: Invalid provider '%s'. Supported providers: openai, anthropic, google\n", provider)
			os.Exit(1)
		}

		// Load current config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if API key is configured for this provider
		if config.Secrets == nil || config.Secrets[provider] == "" {
			fmt.Printf("⚠️  Warning: No API key configured for '%s'\n", provider)
			fmt.Printf("   Configure it using: mad config secrets set %s \"your-api-key\"\n", provider)
			fmt.Println()
		}

		// Set the provider
		config.Provider = provider

		// Save config
		configDir := getConfigDir()
		configPath := filepath.Join(configDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Default provider set to: %s\n", provider)
	},
}

// providerListCmd represents the provider list command
var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available providers and current selection",
	Long:  `List all available LLM providers and show which one is currently selected as default.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🤖 Available LLM Providers:")
		fmt.Println()

		providers := []struct {
			name string
			desc string
		}{
			{"openai", "OpenAI GPT models"},
			{"anthropic", "Anthropic Claude models"},
			{"google", "Google Gemini models"},
		}

		for _, p := range providers {
			if config.Provider == p.name {
				fmt.Printf("✅ %s: %s (current)\n", p.name, p.desc)
			} else {
				fmt.Printf("○ %s: %s\n", p.name, p.desc)
			}
		}

		fmt.Println()
		fmt.Printf("Current default: %s\n", config.Provider)
	},
}

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage model settings for the current provider",
	Long: `Manage the model selection for the currently configured provider.

This allows you to set which specific model to use within your selected provider.`,
}

// modelSetCmd represents the model set command
var modelSetCmd = &cobra.Command{
	Use:   "set <model>",
	Short: "Set the model for the current provider",
	Long: `Set the specific model to use for the currently configured provider.

You can use any model name that the provider supports. The system will attempt to use
the model you specify, even if it's not in our known models list.

Examples:
  mad config model set gpt-4o           # Known OpenAI model
  mad config model set claude-3-haiku   # Known Anthropic model
  mad config model set custom-model-xyz # Custom/unknown model (will attempt to use)

Note: If you use a custom model that's not in our known list, the system will still
try to use it. You'll get an error only if the provider's API rejects the model name.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		model := args[0]

		// Load current config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Initialize models map if not exists
		if config.Models == nil {
			config.Models = make(map[string]string)
		}

		// Check if this is a known model
		isKnown := isKnownModel(config.Provider, model)

		// Set the model for the current provider
		config.Models[config.Provider] = model

		// Save config
		configDir := getConfigDir()
		configPath := filepath.Join(configDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		modelType := "known"
		if !isKnown {
			modelType = "custom"
		}

		fmt.Printf("✅ Model for '%s' set to: %s (%s)\n", config.Provider, model, modelType)

		if !isKnown {
			fmt.Println()
			fmt.Println("ℹ️  Note: This appears to be a custom model not in our known list.")
			fmt.Println("   The system will attempt to use it, but it may not be available.")
			fmt.Println("   Check the provider's documentation for the correct model name.")
		}
	},
}

// getKnownModels returns a map of known models for each provider
func getKnownModels() map[string][]string {
	return map[string][]string{
		"openai": {
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-4-turbo",
			"gpt-4-turbo-preview",
			"gpt-4",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
		},
		"anthropic": {
			"claude-3-5-sonnet-20241022",
			"claude-3-5-sonnet-20240620",
			"claude-3-5-haiku-20241022",
			"claude-3-haiku-20240307",
			"claude-3-sonnet-20240229",
			"claude-3-opus-20240229",
			"claude-2.1",
			"claude-2.0",
		},
		"google": {
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.5-pro-002",
			"gemini-1.5-flash-002",
			"gemini-pro",
			"gemini-pro-vision",
		},
	}
}

// isKnownModel checks if a model is in our known list
func isKnownModel(provider, model string) bool {
	knownModels := getKnownModels()
	if models, exists := knownModels[provider]; exists {
		for _, knownModel := range models {
			if knownModel == model {
				return true
			}
		}
	}
	return false
}

// modelListCmd represents the model list command
var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models for the current provider",
	Long: `List all known models for the currently configured provider and show which one is selected.

Note: Model availability can change frequently. If you don't see a model you want to use,
you can still set it with 'mad config model set <model>' and the system will attempt to use it.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		currentModel := ""
		if config.Models != nil {
			currentModel = config.Models[config.Provider]
		}

		fmt.Printf("🧠 Models for %s:\n", strings.Title(config.Provider))
		fmt.Println()

		knownModels := getKnownModels()
		models := knownModels[config.Provider]

		if models == nil {
			fmt.Printf("No known models defined for provider: %s\n", config.Provider)
			fmt.Println("You can still set custom models with 'mad config model set <model>'")
			return
		}

		fmt.Println("📋 Known Models:")
		for _, model := range models {
			if currentModel == model {
				fmt.Printf("✅ %s (current, known)\n", model)
			} else {
				fmt.Printf("○ %s (known)\n", model)
			}
		}

		fmt.Println()
		fmt.Println("💡 Custom Models:")

		// Show custom models that have been set but aren't in our known list
		customModels := []string{}
		if config.Models != nil {
			for provider, model := range config.Models {
				if provider == config.Provider && model != "" && !isKnownModel(provider, model) {
					customModels = append(customModels, model)
				}
			}
		}

		if len(customModels) == 0 {
			fmt.Println("○ No custom models configured")
		} else {
			for _, model := range customModels {
				if currentModel == model {
					fmt.Printf("✅ %s (current, custom)\n", model)
				} else {
					fmt.Printf("○ %s (custom)\n", model)
				}
			}
		}

		fmt.Println()
		if currentModel != "" {
			modelType := "known"
			if !isKnownModel(config.Provider, currentModel) {
				modelType = "custom"
			}
			fmt.Printf("Current model: %s (%s)\n", currentModel, modelType)
		} else {
			fmt.Printf("No model set for %s.\n", config.Provider)
			fmt.Printf("Use 'mad config model set <model>' to set one.\n")
			fmt.Printf("You can use any model name - the system will attempt to use it.\n")
		}

		fmt.Println()
		fmt.Println("ℹ️  Note: Model availability changes frequently.")
		fmt.Println("   If a model you want isn't listed, you can still use it.")
	},
}

// modelRefreshCmd represents the model refresh command
var modelRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Query provider APIs for current model availability",
	Long: `Query the current provider's API to get the most up-to-date list of available models.

This command will:
• Connect to the provider's API using your configured API key (if available)
• Fetch the latest list of available models
• Fall back to known models if API is unavailable
• Display models with their current status
• Help you discover new models that aren't in our known list

Note: Works best with a valid API key, but will show known models as fallback.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Get API key for current provider
		apiKey := getAPIKey(config.Provider, config)

		fmt.Printf("🔄 Refreshing models for %s...\n", strings.Title(config.Provider))

		var models []providers.ModelInfo
		var fetchSource string

		if apiKey != "" {
			// Try to fetch from API
			fmt.Println("📡 Fetching from provider API...")
			provider := providers.GetProvider(config.Provider)
			ctx := context.Background()
			apiModels, err := provider.ListModels(ctx, apiKey)
			if err != nil {
				fmt.Printf("⚠️  API call failed: %v\n", err)
				fmt.Println("Falling back to known models...")
			} else {
				models = apiModels
				fetchSource = "API"
			}
		}

		// If API call failed or no API key, use known models
		if len(models) == 0 {
			if apiKey == "" {
				fmt.Println("📋 Using known models (no API key configured)...")
			} else {
				fmt.Println("📋 Using known models as fallback...")
			}

			knownModels := getKnownModels()
			if providerModels, exists := knownModels[config.Provider]; exists {
				for _, modelName := range providerModels {
					models = append(models, providers.ModelInfo{
						ID:   modelName,
						Name: modelName,
					})
				}
			}
			fetchSource = "known list"
		}

		if len(models) == 0 {
			fmt.Printf("❌ No models available for provider '%s'\n", config.Provider)
			return
		}

		fmt.Printf("✅ Found %d models from %s:\n", len(models), fetchSource)
		fmt.Println()

		knownModels := getKnownModels()
		knownModelMap := make(map[string]bool)
		if providerModels, exists := knownModels[config.Provider]; exists {
			for _, model := range providerModels {
				knownModelMap[model] = true
			}
		}

		currentModel := ""
		if config.Models != nil {
			currentModel = config.Models[config.Provider]
		}

		// Group models by type
		var knownAvailable []providers.ModelInfo
		var newModels []providers.ModelInfo
		var customModels []providers.ModelInfo

		for _, model := range models {
			if knownModelMap[model.ID] {
				knownAvailable = append(knownAvailable, model)
			} else {
				// Check if this is a custom model we've configured
				isCustom := false
				if config.Models != nil {
					for _, configuredModel := range config.Models {
						if configuredModel == model.ID {
							isCustom = true
							break
						}
					}
				}
				if isCustom {
					customModels = append(customModels, model)
				} else {
					newModels = append(newModels, model)
				}
			}
		}

		// Display known models
		if len(knownAvailable) > 0 {
			fmt.Println("📋 Known Models (available via API):")
			for _, model := range knownAvailable {
				if currentModel == model.ID {
					fmt.Printf("✅ %s (current)\n", model.ID)
				} else {
					fmt.Printf("○ %s\n", model.ID)
				}
			}
			fmt.Println()
		}

		// Display custom models
		if len(customModels) > 0 {
			fmt.Println("💡 Your Custom Models:")
			for _, model := range customModels {
				if currentModel == model.ID {
					fmt.Printf("✅ %s (current, custom)\n", model.ID)
				} else {
					fmt.Printf("○ %s (custom)\n", model.ID)
				}
			}
			fmt.Println()
		}

		// Display new models
		if len(newModels) > 0 {
			fmt.Println("🆕 New/Discovered Models (not in our known list):")
			for _, model := range newModels {
				fmt.Printf("○ %s", model.ID)
				if model.Name != "" && model.Name != model.ID {
					fmt.Printf(" (%s)", model.Name)
				}
				fmt.Println()
			}
			fmt.Println()
			fmt.Println("💡 Tip: You can use these new models with:")
			fmt.Printf("   mad config model set <model-name>\n")
		}

		fmt.Println()
		fmt.Printf("📊 Summary: %d total models, %d known, %d custom, %d new\n",
			len(models), len(knownAvailable), len(customModels), len(newModels))

		if currentModel != "" {
			modelType := "known"
			if !knownModelMap[currentModel] {
				modelType = "custom"
			}
			fmt.Printf("Current model: %s (%s)\n", currentModel, modelType)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add secrets subcommand
	configCmd.AddCommand(secretsCmd)
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsListCmd)

	// Add project subcommand
	configCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add provider subcommand
	configCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerSetCmd)
	providerCmd.AddCommand(providerListCmd)

	// Add model subcommand
	configCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelSetCmd)
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelRefreshCmd)
}
