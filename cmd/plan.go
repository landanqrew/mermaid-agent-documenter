/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan [transcript]",
	Short: "Plan the agent's actions without executing",
	Long: `Plan the Mermaid Documenter Agent's actions on a transcript without actually executing them.
This shows what diagrams and documentation would be generated.

If a current project is set in the global config, the transcript will be read from the project's transcripts/ directory.

Examples:
  mad plan transcript.txt                    # Use current project or global config
  mad plan auth.txt                         # Plan with current project`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Planning for transcript: %s\n", args[0])

		// Load global config to check current project
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if config.CurrentProject != nil {
			fmt.Printf("Project: %s\n", config.CurrentProject.Name)
		}
		fmt.Println("Planning feature - shows what would be generated (TODO: implement)")
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
}
