# Adding a new language analyzer

`pr-complexity` uses a pluggable analyzer interface. Adding JavaScript, Ruby, Java, or another language should not require changes to git diff parsing or report formatting.

## Step 1 — Implement the interface

Create a new package under `internal/analyzers/<language>/`.

```go
// internal/analyzers/javascript/escomplex.go
package javascript

import (
    "path/filepath"
    "strings"

    "github.com/jacob-palathingal/pr-complexity-analyzer/internal/interfaces"
)

type Analyzer struct{}

func New() *Analyzer { return &Analyzer{} }

func (a *Analyzer) Name() string { return "javascript/escomplex" }

func (a *Analyzer) Supports(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx"
}

func (a *Analyzer) Analyze(path, oldContent, newContent string) ([]interfaces.FunctionDelta, error) {
    // Shell out to a complexity tool or parse source directly.
    // Return one FunctionDelta per function seen in either version.
    panic("not yet implemented")
}
```

The full `interfaces.Analyzer` contract is in `internal/interfaces/analyzer.go`.

## Step 2 — Register it in the runner

Open `internal/runner/runner.go` and add your analyzer to the `registry` slice:

```go
import (
    "github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/goast"
    "github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/javascript"
    "github.com/jacob-palathingal/pr-complexity-analyzer/internal/analyzers/python"
    // ...
)

var registry = []interfaces.Analyzer{
    python.New(),
    goast.New(),
    javascript.New(), // ← add this line
}
```

That's the only core runtime file you should need to touch.

## Step 3 — Write tests

Add `internal/analyzers/<language>/analyzer_test.go`. Follow the pattern in the Python and Go analyzer tests:

- `TestSupports` — verify extension matching.
- `TestAnalyze_IncreaseDetected` — confirm deltas are positive when complexity grows.
- `TestAnalyze_NewFile` — empty `OldContent` should produce deltas with `OldComplexity == 0`.
- Syntax/error tests for malformed source or missing external tools.
- Unit tests for delta-building logic or parser-specific edge cases.

If the analyzer shells out to an external tool, make the integration tests skip cleanly when the tool is not installed.

## Step 4 — Update Docker and setup docs

If your analyzer requires an external runtime or CLI, install it in the runtime stage of the `Dockerfile`:

```dockerfile
RUN pip install --no-cache-dir radon==6.*
# Add your tool, for example:
RUN npm install -g escomplex-cli
```

Then update the README setup section and, if helpful, add a dedicated script under `scripts/`.

## Interface contract summary

| Method | Signature | Notes |
|---|---|---|
| `Name()` | `string` | Human-readable, e.g. `"javascript/escomplex"`. Used by `--lang` filtering. |
| `Supports(path)` | `bool` | Return true for file extensions your analyzer handles. |
| `Analyze(path, oldContent, newContent)` | `([]FunctionDelta, error)` | `oldContent` is empty for new files. Deleted files are filtered upstream. |

`FunctionDelta` fields:

| Field | Type | Meaning |
|---|---|---|
| `FilePath` | string | Pass through the `path` argument. |
| `FunctionName` | string | Qualified name, e.g. `ClassName.method` or `Server.Route`. |
| `OldComplexity` | int | 0 if the function did not exist before. |
| `NewComplexity` | int | Cyclomatic complexity at head ref. |
| `Delta` | int | `NewComplexity - OldComplexity`. |
