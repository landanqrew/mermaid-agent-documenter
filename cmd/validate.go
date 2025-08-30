/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a manifest or Mermaid file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Validating: %s\n", args[0])
		// TODO: Implement validation logic for Mermaid syntax and manifests
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
