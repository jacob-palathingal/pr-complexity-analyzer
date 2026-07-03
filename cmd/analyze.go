package cmd

import (
	"fmt"
	"time"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/runner"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze complexity changes between two git refs",
	Long: `Compares cyclomatic complexity of every supported function in files changed between --base and --head.

Exit codes:
  0  success, no functions breached the threshold
  1  one or more functions met or exceeded --threshold
  2  tool error

Examples:
  pr-complexity analyze --base main --head feature/my-branch
  pr-complexity analyze --base HEAD~1 --head HEAD --threshold 5
  pr-complexity analyze --base main --head HEAD --min-delta 3
  pr-complexity analyze --base main --head HEAD --format json
  pr-complexity analyze --base main --head HEAD --include 'services/payments/**' --exclude 'services/payments/mocks/**'
`,
	RunE: runAnalyze,
}

var (
	flagBase         string
	flagHead         string
	flagThreshold    int
	flagMinDelta     int
	flagFormat       string
	flagLang         string
	flagUnchanged    bool
	flagRepoDir      string
	flagTimeout      time.Duration
	flagInclude      []string
	flagExclude      []string
	flagIncludeTests bool
)

func init() {
	analyzeCmd.Flags().StringVar(&flagBase, "base", "", "Base git ref (branch, tag, or commit SHA) (required)")
	analyzeCmd.Flags().StringVar(&flagHead, "head", "HEAD", "Head git ref (branch, tag, or commit SHA)")
	analyzeCmd.Flags().IntVar(&flagThreshold, "threshold", 0, "Exit 1 if any function delta meets or exceeds this value (0 = disabled)")
	analyzeCmd.Flags().IntVar(&flagMinDelta, "min-delta", 0, "Only report functions with delta greater than or equal to this value")
	analyzeCmd.Flags().StringVar(&flagFormat, "format", "text", "Output format: text, json, or markdown")
	analyzeCmd.Flags().StringVar(&flagLang, "lang", "", "Restrict analysis to a specific language (all, python, py, go, golang)")
	analyzeCmd.Flags().BoolVar(&flagUnchanged, "include-unchanged", false, "Include functions with no complexity change")
	analyzeCmd.Flags().StringVar(&flagRepoDir, "repo", "", "Path to the git repository (default: current directory)")
	analyzeCmd.Flags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "Per-git-command timeout")
	analyzeCmd.Flags().StringArrayVar(&flagInclude, "include", nil, "Only analyze paths matching this glob; can be repeated")
	analyzeCmd.Flags().StringArrayVar(&flagExclude, "exclude", nil, "Exclude paths matching this glob; can be repeated")
	analyzeCmd.Flags().BoolVar(&flagIncludeTests, "include-tests", false, "Include test files in analysis")

	_ = analyzeCmd.MarkFlagRequired("base")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg := runner.Config{
		RepoDir:          flagRepoDir,
		BaseRef:          flagBase,
		HeadRef:          flagHead,
		MinDelta:         flagMinDelta,
		IncludeUnchanged: flagUnchanged,
		Format:           flagFormat,
		LangFilter:       flagLang,
		Threshold:        flagThreshold,
		Timeout:          flagTimeout,
		Include:          flagInclude,
		Exclude:          flagExclude,
		IncludeTests:     flagIncludeTests,
	}

	result, err := runner.Run(cmd.OutOrStdout(), cfg)
	if err != nil {
		return runner.NewExitError(runner.ExitToolError, fmt.Sprintf("error: %v", err))
	}

	if result.ExitCode != runner.ExitOK {
		return runner.NewExitError(result.ExitCode, "")
	}

	return nil
}
