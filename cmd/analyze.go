package cmd

import (
	"fmt"
	"os"

	"github.com/jacob-palathingal/pr-complexity-analyzer/internal/runner"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze complexity changes between two git refs",
	Long: `Compares cyclomatic complexity of every function in files changed
between --base and --head git refs.

Exit codes:
  0 — success, no functions breached the threshold
  1 — one or more functions met or exceeded --threshold
  2 — tool error (bad ref, missing git, analyzer failure)

Examples:
  pr-complexity analyze --base main --head feature/my-branch
  pr-complexity analyze --base HEAD~1 --head HEAD --threshold 5
  pr-complexity analyze --base main --head HEAD --format json
  pr-complexity analyze --base main --head HEAD --format markdown
`,
	RunE: runAnalyze,
}

var (
	flagBase      string
	flagHead      string
	flagThreshold int
	flagFormat    string
	flagLang      string
	flagUnchanged bool
	flagRepoDir   string
)

func init() {
	analyzeCmd.Flags().StringVar(&flagBase, "base", "", "Base git ref (branch, tag, or commit SHA) (required)")
	analyzeCmd.Flags().StringVar(&flagHead, "head", "HEAD", "Head git ref (branch, tag, or commit SHA)")
	analyzeCmd.Flags().IntVar(&flagThreshold, "threshold", 0, "Exit 1 if any function delta meets or exceeds this value (0 = disabled)")
	analyzeCmd.Flags().StringVar(&flagFormat, "format", "text", "Output format: text, json, or markdown")
	analyzeCmd.Flags().StringVar(&flagLang, "lang", "", "Restrict analysis to a specific language (e.g. python, go)")
	analyzeCmd.Flags().BoolVar(&flagUnchanged, "include-unchanged", false, "Include functions with no complexity change")
	analyzeCmd.Flags().StringVar(&flagRepoDir, "repo", "", "Path to the git repository (default: current directory)")

	_ = analyzeCmd.MarkFlagRequired("base")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg := runner.Config{
		RepoDir:          flagRepoDir,
		BaseRef:          flagBase,
		HeadRef:          flagHead,
		IncludeUnchanged: flagUnchanged,
		Format:           flagFormat,
		LangFilter:       flagLang,
		Threshold:        flagThreshold,
	}

	result, err := runner.Run(os.Stdout, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(runner.ExitError)
	}
	if result.ExitCode != runner.ExitOK {
		os.Exit(result.ExitCode)
	}
	return nil
}
