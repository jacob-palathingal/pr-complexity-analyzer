package cmd

import "github.com/spf13/cobra"

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze complexity changes between two git refs",
	Long: `Compares cyclomatic complexity of every function in files changed
between --base and --head git refs.

Example:
  pr-complexity analyze --base main --head feature/my-branch
  pr-complexity analyze --base HEAD~1 --head HEAD --threshold 3
  pr-complexity analyze --base abc123 --head def456 --json
`,
	RunE: runAnalyze,
}

var (
	flagBase      string
	flagHead      string
	flagThreshold int
	flagJSON      bool
	flagLang      string
	flagUnchanged bool
	flagRepoDir   string
)

func init() {
	analyzeCmd.Flags().StringVar(&flagBase, "base", "", "Base git ref (branch, tag, or commit SHA) (required)")
	analyzeCmd.Flags().StringVar(&flagHead, "head", "HEAD", "Head git ref (branch, tag, or commit SHA)")
	analyzeCmd.Flags().IntVar(&flagThreshold, "threshold", 0, "Only report functions with complexity increase >= this value")
	analyzeCmd.Flags().BoolVar(&flagJSON, "json", false, "Output results as JSON instead of a table")
	analyzeCmd.Flags().StringVar(&flagLang, "lang", "", "Restrict analysis to a specific language (e.g. python)")
	analyzeCmd.Flags().BoolVar(&flagUnchanged, "include-unchanged", false, "Include functions with no complexity change")
	analyzeCmd.Flags().StringVar(&flagRepoDir, "repo", "", "Path to the git repository (default: current directory)")
	_ = analyzeCmd.MarkFlagRequired("base")
}

// runAnalyze is wired up in commit 07. Stub so binary compiles now.
func runAnalyze(cmd *cobra.Command, args []string) error {
	cmd.Println("analyze: not yet implemented")
	return nil
}
