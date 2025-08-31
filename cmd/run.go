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
	"time"

	"github.com/landanqrew/mermaid-agent-documenter/internal/agent"
	"github.com/spf13/cobra"
)

func loadConfig() (*Config, error) {
	// Always load from global config
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func getAPIKey(provider string, config *Config) string {
	// First check config for stored API keys
	if config.Secrets != nil {
		if key, exists := config.Secrets[provider]; exists && key != "" {
			return key
		}
	}

	// Fall back to environment variables
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "google":
		return os.Getenv("GOOGLE_API_KEY")
	default:
		return ""
	}
}

func readTranscript(path string, config *Config) (string, error) {
	var fullPath string

	if config.CurrentProject != nil {
		// Use current project's transcripts directory
		transcriptsDir := filepath.Join(config.CurrentProject.RootDir, "transcripts")

		// If path doesn't contain directory separators, assume it's in transcripts dir
		if !strings.Contains(path, "/") && !strings.Contains(path, "\\") {
			fullPath = filepath.Join(transcriptsDir, path)
		} else {
			fullPath = path
		}
	} else {
		fullPath = path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(fullPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		fullPath = strings.Replace(fullPath, "~", home, 1)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if config.CurrentProject != nil && strings.Contains(err.Error(), "no such file") {
			return "", fmt.Errorf("transcript file not found. Make sure '%s' exists in the '%s/transcripts/' directory", path, config.CurrentProject.Name)
		}
		return "", err
	}

	return string(data), nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [transcript]",
	Short: "Run the agent on a transcript",
	Long: `Run the Mermaid Documenter Agent on a transcript file to generate diagrams and documentation.

If a current project is set in the global config, the transcript will be read from the project's
transcripts/ directory and output will be saved to the project's out/ directory.

Examples:
  mad run transcript.txt                    # Use current project or global config
  mad run auth.txt --dry-run              # Dry run with current project`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		yes, _ := cmd.Flags().GetBool("yes")

		// Load global config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Get API key from config or environment
		apiKey := getAPIKey(config.Provider, config)
		if apiKey == "" {
			fmt.Printf("Error: API key for provider '%s' not found\n", config.Provider)
			fmt.Printf("Configure it using: mad config secrets set %s \"your-api-key\"\n", config.Provider)
			fmt.Printf("Or set environment variable: %s_API_KEY\n", strings.ToUpper(config.Provider))
			os.Exit(1)
		}

		// Read transcript (project-aware)
		transcript, err := readTranscript(args[0], config)
		if err != nil {
			fmt.Printf("Error reading transcript: %v\n", err)
			os.Exit(1)
		}

		// Create agent config
		agentConfig := &agent.AgentConfig{
			Provider:            config.Provider,
			Model:               config.Models[config.Provider],
			APIKey:              apiKey,
			MaxSteps:            config.Limits.MaxSteps,
			TimeoutSec:          config.Limits.RunTimeoutSec,
			TokenBudget:         config.Limits.TokenBudget,
			CostCeilingUsd:      config.Limits.CostCeilingUsd,
			ConfidenceThreshold: config.ConfidenceThreshold,
			OutputDir:           config.OutDir,
			RedactPII:           config.Safety.PIIRedaction,
			StoreChainOfThought: config.Log.StoreChainOfThought,
		}

		// Create and run agent
		mermaidAgent := agent.NewMermaidDocumenterAgent(agentConfig)
		mermaidAgent.SetTranscript(transcript)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Limits.RunTimeoutSec)*time.Second)
		defer cancel()

		if config.CurrentProject != nil {
			fmt.Printf("Running Mermaid Documenter Agent on project: %s\n", config.CurrentProject.Name)
			fmt.Printf("Transcript: transcripts/%s\n", args[0])
		} else {
			fmt.Printf("Running Mermaid Documenter Agent on transcript: %s\n", args[0])
		}
		fmt.Printf("Provider: %s, Model: %s\n", config.Provider, agentConfig.Model)
		fmt.Printf("Output directory: %s\n", config.OutDir)
		fmt.Printf("Dry run: %v\n", dryRun)

		if !dryRun {
			if !yes {
				fmt.Print("Proceed with agent execution? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled.")
					return
				}
			}

			err = mermaidAgent.Run(ctx)
			if err != nil {
				fmt.Printf("Agent execution failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Agent execution completed successfully!")
		} else {
			fmt.Println("Dry run mode - agent execution skipped.")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("dry-run", false, "Print planned actions without executing")
	runCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
}
