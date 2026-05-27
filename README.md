# pr-complexity

> Per-function cyclomatic complexity deltas, scoped to exactly what your PR touches.

`pr-complexity` fills a gap between line-count tools (Danger.js, GitHub native) and
whole-codebase scanners (SonarQube, CodeClimate). It checks out both sides of a diff,
runs complexity analysis only on changed files, and ranks functions by how much more
complex they became — in seconds, with no SaaS account.

## Status

🚧 Under construction.

## Planned usage

```bash
pr-complexity analyze --base main --head feature/my-branch
pr-complexity analyze --base HEAD~1 --head HEAD --threshold 3 --json
```

## Tech stack

- **Go + Cobra** — CLI binary, fast startup, single static binary
- **Python + Radon** — cyclomatic complexity for Python files
- **Docker** — zero-setup distribution

## Architecture

```
main.go
└── cmd/         Cobra commands (analyze, root)
└── internal/
    ├── diff/        Git ref resolver + unified diff parser
    ├── interfaces/  Analyzer interface (pluggable per-language)
    ├── analyzers/
    │   └── python/  Radon-backed implementation
    ├── report/      Ranked delta report (text table + JSON)
    └── runner/      Orchestrates diff → dispatch → report
```
