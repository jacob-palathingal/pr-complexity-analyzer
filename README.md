# pr-complexity

![Release](https://github.com/jacob-palathingal/pr-complexity-analyzer/actions/workflows/release.yml/badge.svg)
![PR Complexity Check](https://github.com/jacob-palathingal/pr-complexity-analyzer/actions/workflows/pr-complexity.yml/badge.svg)

> Per-function cyclomatic complexity deltas, scoped to exactly what your PR touches.

`pr-complexity` fills the gap between line-count tools and whole-codebase scanners. It compares the base and head versions of changed files, runs language-specific analyzers only where needed, and ranks functions by how much more complex they became — fast enough for local review and strict enough for CI.

## What it does

- Analyzes only files changed in a pull request
- Reports function-level complexity before, after, and delta
- Supports Python through Radon and Go through the standard-library AST
- Emits text, JSON, or GitHub-flavored Markdown
- Fails CI when complexity increases meet a configured threshold
- Runs locally, in Docker, or as a GitHub Actions PR comment workflow

## Quick start

**With Docker:**

```bash
docker run --rm -v $(pwd):/repo ghcr.io/jacob-palathingal/pr-complexity-analyzer:latest \
  analyze --base main --head HEAD
```

Docker images are published by tagged releases such as `v0.1.0`.

**Without Docker:**

```bash
# Required only for Python analysis. Go analysis uses the Go standard library.
bash scripts/install_radon.sh

# Build the binary
go build -o pr-complexity .

# Run against your PR or local branch
./pr-complexity analyze --base main --head feature/my-branch
```

## Example output

```text
  PR Complexity Report — 6 function(s) changed

  File                        Function                   Before  After  Delta
  --------------------------  -------------------------  ------  -----  -----
  payments/processor.py       PaymentProcessor.charge         3     11   +8 ▲
  auth/session.py             SessionManager.validate         2      9   +7 ▲
  service/router.go           Server.Route                    2      7   +5 ▲
  payments/processor.py       PaymentProcessor.refund         4      8   +4 ▲
  auth/session.py             create_token                    1      4   +3 ▲
  utils/retry.py              with_retry                      0      3   +3 ▲
```

See [docs/example-output.txt](docs/example-output.txt) for text, Markdown, and JSON examples.

## Flags

| Flag | Default | Description |
|---|---:|---|
| `--base` | *(required)* | Base git ref, branch, tag, or commit SHA |
| `--head` | `HEAD` | Head git ref, branch, tag, or commit SHA |
| `--threshold` | `0` | Exit 1 if any function delta is greater than or equal to this value. `0` disables CI failure. |
| `--min-delta` | `0` | Only include functions with delta greater than or equal to this value in the report. |
| `--format` | `text` | Output format: `text`, `json`, or `markdown` |
| `--lang` | *(all)* | Restrict analysis to one language: `python`, `py`, `go`, or `golang` |
| `--include-unchanged` | `false` | Include functions with no complexity change |
| `--repo` | cwd | Path to the git repository |

`--threshold` controls CI enforcement. `--min-delta` controls report filtering.

## GitHub Actions PR comments

This repo includes `.github/workflows/pr-complexity.yml`, which builds the CLI, analyzes each pull request, posts a sticky Markdown report comment, and fails the check when a function meets the configured threshold.

To reuse the workflow in another repo, copy the workflow and adjust the command:

```yaml
./pr-complexity analyze \
  --base "origin/${{ github.base_ref }}" \
  --head "HEAD" \
  --format markdown \
  --threshold 5
```

## Architecture

```text
main.go
└── cmd/              Cobra commands (root, analyze)
└── internal/
    ├── diff/         Git ref resolver + file snapshot fetcher
    ├── interfaces/   Analyzer interface + shared result types
    ├── analyzers/
    │   ├── python/   Radon-backed Python complexity analyzer
    │   └── goast/    Go AST complexity analyzer
    ├── report/       Ranked delta report: text, JSON, Markdown
    └── runner/       Orchestrates diff → analyzer dispatch → report → threshold
```

**Key design decision:** analyzers implement a shared `interfaces.Analyzer` contract. Adding a new language requires implementing `Supports`, `Analyze`, and `Name`, then registering it in the runner. Diff parsing and report formatting do not change.

## Adding a new language

See [docs/extending-analyzers.md](docs/extending-analyzers.md) for a step-by-step guide.

## Running tests

```bash
# Unit tests that do not require Radon
go test ./internal/diff/... ./internal/report/... ./internal/runner/... ./internal/analyzers/goast/...

# Full test suite, including Python/Radon integration tests
pip install radon==6.*
go test ./...
```

## Tech stack

- **Go 1.22 + Cobra** — CLI binary, command parsing, fast startup
- **Go AST** — native Go function complexity without external tools
- **Python + Radon** — battle-tested Python cyclomatic complexity
- **Docker** — zero-setup distribution with git, Radon, and the compiled CLI
- **GitHub Actions** — release builds, GHCR publishing, PR comments, threshold checks

## License

MIT
