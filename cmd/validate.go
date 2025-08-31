/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a manifest or Mermaid file",
	Long: `Validate a generated manifest or Mermaid file for syntax correctness.

If a current project is set in the global config, the path will be resolved relative to the project's out/ directory.

Examples:
  mad validate docs/diagrams/auth/sequence-login.md    # Global validation
  mad validate auth/sequence-login.md                 # Project-specific validation`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Validating: %s\n", args[0])

		// Load global config to check current project
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if config.CurrentProject != nil {
			fmt.Printf("Project: %s\n", config.CurrentProject.Name)
		}
		fmt.Println("Validation feature - checks Mermaid syntax and manifests (TODO: implement)")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
