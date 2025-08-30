/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/landanqrew/mermaid-agent-documenter/internal/agent"
	"github.com/spf13/cobra"
)

func loadConfig() (*Config, error) {
	configDir := getConfigDir()
	configPath := os.ExpandEnv(fmt.Sprintf("%s/config.json", configDir))
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

func getAPIKey(provider string) string {
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

func readTranscript(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", home, 1)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [transcript]",
	Short: "Run the agent on a transcript",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		yes, _ := cmd.Flags().GetBool("yes")

		// Load config
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Get API key from environment
		apiKey := getAPIKey(config.Provider)
		if apiKey == "" {
			fmt.Printf("Error: API key for provider '%s' not found in environment variables\n", config.Provider)
			fmt.Printf("Please set one of: OPENAI_API_KEY, ANTHROPIC_API_KEY, or GOOGLE_API_KEY\n")
			os.Exit(1)
		}

		// Read transcript
		transcript, err := readTranscript(args[0])
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

		fmt.Printf("Running Mermaid Documenter Agent on transcript: %s\n", args[0])
		fmt.Printf("Provider: %s, Model: %s\n", config.Provider, agentConfig.Model)
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
