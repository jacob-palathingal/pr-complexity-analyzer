# pr-complexity

> Per-function cyclomatic complexity deltas, scoped to exactly what your PR touches.

`pr-complexity` fills the gap between line-count tools (Danger.js, GitHub's native diff view) and whole-codebase scanners (SonarQube, CodeClimate). It checks out both sides of a diff, runs complexity analysis only on changed files, and ranks functions by how much more complex they became — in seconds, with no SaaS account.

## Quick start

**With Docker (no local setup needed):**

```bash
docker run --rm -v $(pwd):/repo ghcr.io/yourusername/pr-complexity:latest \
  analyze --base main --head HEAD
```

**Without Docker (requires Go + Python + Radon):**

```bash
# Install Radon once
bash scripts/install_radon.sh

# Build the binary
go build -o pr-complexity .

# Run against your PR
./pr-complexity analyze --base main --head feature/my-branch
```

## Example output

```
  PR Complexity Report — 6 function(s) changed

  File                        Function                   Before  After  Delta
  --------------------------  -------------------------  ------  -----  -----
  payments/processor.py       PaymentProcessor.charge         3     11   +8 ▲
  auth/session.py             SessionManager.validate         2      9   +7 ▲
  payments/processor.py       PaymentProcessor.refund         4      8   +4 ▲
  auth/session.py             create_token                    1      4   +3 ▲
  utils/retry.py              with_retry                      0      3   +3 ▲
  models/order.py             Order.apply_discount            5      7   +2 ▲
```

See [docs/example-output.txt](docs/example-output.txt) for JSON output examples.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--base` | *(required)* | Base git ref (branch, tag, or commit SHA) |
| `--head` | `HEAD` | Head git ref |
| `--threshold` | `0` | Only report functions with delta ≥ this value |
| `--json` | `false` | Machine-readable JSON output |
| `--lang` | *(all)* | Restrict to a language, e.g. `python` |
| `--include-unchanged` | `false` | Include functions with no complexity change |
| `--repo` | cwd | Path to the git repository |

## Architecture

```
main.go
└── cmd/              Cobra commands (root, analyze)
└── internal/
    ├── diff/         Git ref resolver + file snapshot fetcher
    ├── interfaces/   Analyzer interface + shared types (pluggable)
    ├── analyzers/
    │   └── python/   Radon-backed cyclomatic complexity (first implementation)
    ├── report/       Ranked delta report — text table and JSON formatters
    └── runner/       Orchestrates diff → analyzer dispatch → report
```

**Key design decision:** the `interfaces.Analyzer` interface decouples language analysis from everything else. Python/Radon is the first implementation; adding JavaScript, Ruby, or Go analysis means implementing two methods and adding one line to the registry — no other code changes.

## Adding a new language

See [docs/extending-analyzers.md](docs/extending-analyzers.md) for a step-by-step guide.

## Running tests

```bash
# Unit tests (no external tools needed)
go test ./internal/diff/... ./internal/report/...

# All tests including Python/Radon integration
pip install radon
go test ./...
```

## Tech stack

- **Go 1.22 + Cobra** — CLI binary, fast startup, single static binary distribution
- **Python + Radon** — battle-tested cyclomatic complexity for Python
- **Docker** — zero-setup distribution; bundles Go binary + Python + Radon in one image

## License

MIT
