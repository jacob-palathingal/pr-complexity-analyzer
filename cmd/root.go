package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pr-complexity",
	Short: "Analyze cyclomatic complexity changes in a pull request",
	Long: `pr-complexity computes per-function cyclomatic complexity deltas
for every function touched by a PR diff.

It checks out both the base and head snapshots, runs language-specific
analyzers only on changed files, and outputs a ranked report of functions
by complexity increase.

Supported languages: Python (via Radon)
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
