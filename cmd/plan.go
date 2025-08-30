/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan [transcript]",
	Short: "Plan the agent's actions without executing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		fmt.Printf("Planning for transcript: %s (yes: %v)\n", args[0], yes)
		// TODO: Implement planning logic similar to run but without execution
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
}
