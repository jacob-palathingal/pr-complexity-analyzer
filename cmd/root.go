package cmd

import (
	"fmt"
	"os"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/runner"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "pr-complexity",
	Short:         "Analyze cyclomatic complexity changes in a pull request",
	Long:          "pr-complexity computes per-function cyclomatic complexity deltas for every supported function touched by a pull request diff.",
	Version:       versionString(),
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if exitErr, ok := err.(interface{ ExitCode() int }); ok {
			if err.Error() != "" {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(exitErr.ExitCode())
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(runner.ExitToolError)
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
