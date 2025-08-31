/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type ProjectConfig struct {
	Name        string `json:"name"`
	RootDir     string `json:"rootDir"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type Config struct {
	Provider            string            `json:"provider"`
	Models              map[string]string `json:"models"`
	Log                 LogConfig         `json:"log"`
	Safety              SafetyConfig      `json:"safety"`
	Limits              LimitsConfig      `json:"limits"`
	ConfidenceThreshold float64           `json:"confidenceThreshold"`
	OutDir              string            `json:"outDir"`
	Secrets             map[string]string `json:"secrets,omitempty"`
	CurrentProject      *ProjectConfig    `json:"currentProject,omitempty"`
}

type LogConfig struct {
	Level               string `json:"level"`
	Redact              bool   `json:"redact"`
	StoreChainOfThought bool   `json:"storeChainOfThought"`
}

type SafetyConfig struct {
	Mode         string `json:"mode"`
	PIIRedaction bool   `json:"piiRedaction"`
}

type LimitsConfig struct {
	MaxSteps       int     `json:"maxSteps"`
	RunTimeoutSec  int     `json:"runTimeoutSec"`
	TokenBudget    int     `json:"tokenBudget"`
	CostCeilingUsd float64 `json:"costCeilingUsd"`
}

func defaultConfig() *Config {
	return &Config{
		Provider: "openai",
		Models: map[string]string{
			"openai":    "gpt-5-mini",
			"anthropic": "claude-3.5-sonnet",
			"google":    "gemini-2.5-flash",
		},
		Log: LogConfig{
			Level:               "info",
			Redact:              true,
			StoreChainOfThought: false,
		},
		Safety: SafetyConfig{
			Mode:         "standard",
			PIIRedaction: true,
		},
		Limits: LimitsConfig{
			MaxSteps:       25,
			RunTimeoutSec:  300,
			TokenBudget:    100000,
			CostCeilingUsd: 1.0,
		},
		ConfidenceThreshold: 0.90,
		OutDir:              "~/mermaid-agent-documenter/output",
	}
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "mermaid-agent-documenter")
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project or the global environment",
	Long: `Initialize a new project with its own directory structure and update the global configuration.

If no project name is provided, initializes the global environment.
If a project name is provided, creates the project in the current directory and sets it as the current project.

Examples:
  mad init                    # Initialize global environment
  mad init my-project         # Initialize new project called "my-project"
  mad init ecommerce-app      # Initialize project for e-commerce application`,
	Run: func(cmd *cobra.Command, args []string) {
		// First, ensure global config directory exists
		globalConfigDir := getConfigDir()
		if err := os.MkdirAll(globalConfigDir, 0755); err != nil {
			fmt.Printf("Error creating global config dir: %v\n", err)
			os.Exit(1)
		}

		// Load or create global config
		globalConfigPath := filepath.Join(globalConfigDir, "config.json")
		var config *Config

		if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
			config = defaultConfig()
		} else {
			data, err := os.ReadFile(globalConfigPath)
			if err != nil {
				fmt.Printf("Error reading global config: %v\n", err)
				os.Exit(1)
			}
			config = &Config{}
			if err := json.Unmarshal(data, config); err != nil {
				fmt.Printf("Error parsing global config: %v\n", err)
				os.Exit(1)
			}
		}

		if len(args) > 0 {
			// Initialize a project
			projectName := args[0]
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}

			projectDir := filepath.Join(cwd, projectName)

			// Create project directory structure
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				fmt.Printf("Error creating project dir: %v\n", err)
				os.Exit(1)
			}

			// Create subdirectories
			transcriptsDir := filepath.Join(projectDir, "transcripts")
			if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
				fmt.Printf("Error creating transcripts dir: %v\n", err)
				os.Exit(1)
			}

			outDir := filepath.Join(projectDir, "out")
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Printf("Error creating output dir: %v\n", err)
				os.Exit(1)
			}

			logsDir := filepath.Join(projectDir, "logs")
			if err := os.MkdirAll(logsDir, 0755); err != nil {
				fmt.Printf("Error creating logs dir: %v\n", err)
				os.Exit(1)
			}

			// Update global config with current project
			config.CurrentProject = &ProjectConfig{
				Name:      projectName,
				RootDir:   projectDir,
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			fmt.Printf("Project '%s' initialized at %s\n", projectName, projectDir)
			fmt.Printf("Project structure:\n")
			fmt.Printf("  📁 %s/\n", projectName)
			fmt.Printf("    📁 transcripts/     (place your transcript files here)\n")
			fmt.Printf("    📁 out/            (generated diagrams will be saved here)\n")
			fmt.Printf("    📁 logs/           (execution logs)\n")
			fmt.Printf("\nProject set as current in global config.\n")

		} else {
			// Initialize global environment only
			fmt.Printf("Global environment initialized at %s\n", globalConfigDir)
		}

		// Save global config
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(globalConfigPath, data, 0644); err != nil {
			fmt.Printf("Error writing global config: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
