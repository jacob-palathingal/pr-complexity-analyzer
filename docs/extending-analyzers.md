# Adding a new language analyzer

`pr-complexity` uses a pluggable analyzer interface. Adding JavaScript, Ruby, Go, or any other language takes four steps and touches two files.

## Step 1 — Implement the interface

Create a new package under `internal/analyzers/<language>/`.

```go
// internal/analyzers/javascript/escomplex.go
package javascript

import (
    "path/filepath"
    "strings"

    "github.com/yourusername/pr-complexity-analyzer/internal/interfaces"
)

type Analyzer struct{}

func New() *Analyzer { return &Analyzer{} }

func (a *Analyzer) Name() string { return "javascript/escomplex" }

func (a *Analyzer) Supports(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx"
}

func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
    // Shell out to your complexity tool, parse the output,
    // build and return []interfaces.FunctionDelta.
    //
    // See internal/analyzers/python/radon.go for a complete example.
    panic("not yet implemented")
}
```

The full `interfaces.Analyzer` contract is in `internal/interfaces/analyzer.go`.

## Step 2 — Register it in the runner

Open `internal/runner/runner.go` and add your analyzer to the `registry` slice:

```go
import (
    "github.com/yourusername/pr-complexity-analyzer/internal/analyzers/javascript"
    "github.com/yourusername/pr-complexity-analyzer/internal/analyzers/python"
    // ...
)

var registry = []interfaces.Analyzer{
    python.New(),
    javascript.New(), // ← add this line
}
```

That's the only core file you need to touch.

## Step 3 — Write tests

Add `internal/analyzers/<language>/analyzer_test.go`. Follow the pattern in `internal/analyzers/python/radon_test.go`:

- `skipIfNo<Tool>` — skip integration tests when the external tool is absent.
- `TestSupports` — verify the file extension matching.
- `TestAnalyze_IncreaseDetected` — confirm deltas are positive when complexity grows.
- `TestAnalyze_NewFile` — empty `OldContent` should produce deltas with `OldComplexity == 0`.
- `TestBuildDeltas_*` — unit-test the delta-building logic in isolation.

## Step 4 — Update the Dockerfile

If your analyzer shells out to an external tool, install it in the `Runtime image` stage of the `Dockerfile`:

```dockerfile
RUN pip install --no-cache-dir radon==6.*
# Add your tool:
RUN npm install -g escomplex-cli
```

And note it in `scripts/install_radon.sh` (or add a separate install script).

## Interface contract summary

| Method | Signature | Notes |
|---|---|---|
| `Name()` | `string` | Human-readable, e.g. `"javascript/escomplex"`. Used by `--lang` filter. |
| `Supports(path)` | `bool` | Return true for file extensions your tool handles. |
| `Analyze(path, oldContent, newContent)` | `([]FunctionDelta, error)` | `oldContent` is empty for new files. Never receive deleted files. |

`FunctionDelta` fields:

| Field | Type | Meaning |
|---|---|---|
| `FilePath` | string | Pass through the `path` argument. |
| `FunctionName` | string | Qualified name, e.g. `"ClassName.method"`. |
| `OldComplexity` | int | 0 if the function didn't exist before. |
| `NewComplexity` | int | Cyclomatic complexity at head ref. |
| `Delta` | int | `NewComplexity - OldComplexity`. Compute this yourself. |
