/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Config struct {
	Provider            string            `json:"provider"`
	Models              map[string]string `json:"models"`
	Log                 LogConfig         `json:"log"`
	Safety              SafetyConfig      `json:"safety"`
	Limits              LimitsConfig      `json:"limits"`
	ConfidenceThreshold float64           `json:"confidenceThreshold"`
	OutDir              string            `json:"outDir"`
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
			MaxSteps:       12,
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
	Use:   "init",
	Short: "Initialize the working directory and config",
	Run: func(cmd *cobra.Command, args []string) {
		configDir := getConfigDir()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Error creating config dir: %v\n", err)
			os.Exit(1)
		}
		outputDir := filepath.Join(configDir, "output")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output dir: %v\n", err)
			os.Exit(1)
		}
		logsDir := filepath.Join(configDir, "logs")
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			fmt.Printf("Error creating logs dir: %v\n", err)
			os.Exit(1)
		}
		config := defaultConfig()
		configPath := filepath.Join(configDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling config: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			fmt.Printf("Error writing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Initialized at %s\n", configDir)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
